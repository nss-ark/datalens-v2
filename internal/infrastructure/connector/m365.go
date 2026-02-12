package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector/m365"
	"github.com/complyark/datalens/internal/infrastructure/connector/shared"
	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/crypto"
	"github.com/complyark/datalens/pkg/types"
)

const (
	GraphAPIEndpoint = "https://graph.microsoft.com/v1.0"
	MaxFileSize      = 10 * 1024 * 1024 // 10MB
)

// M365Connector implements discovery.Connector for Microsoft 365 (OneDrive).
type M365Connector struct {
	client      *http.Client
	fileScanner *shared.FileScanner
	logger      *slog.Logger
	cfg         *config.Config
}

// NewM365Connector creates a new M365Connector.
func NewM365Connector(detector *detection.ComposableDetector) *M365Connector {
	// We load config temporarily to get secrets if needed, but Connect loads it too.
	// Actually we need config for ClientID/Secret.
	// We'll load it in Connect or here?
	// It's better to load once.
	cfg, _ := config.Load() // Ignore error, Connect will catch if missing or we handle it.

	return &M365Connector{
		fileScanner: shared.NewFileScanner(detector, slog.Default()),
		logger:      slog.Default().With("connector", "m365"),
		cfg:         cfg,
	}
}

// Capabilities returns the supported operations.
func (c *M365Connector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:          true,
		SupportsStreaming:    true,
		SupportsParallelScan: false, // Simple traversal for now
	}
}

// Connect establishes a connection to Microsoft Graph using the stored credentials.
func (c *M365Connector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	if c.cfg == nil {
		var err error
		c.cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
	}

	// 1. Decrypt Credentials
	if ds.Credentials == "" {
		return fmt.Errorf("credentials required")
	}

	// Key padding logic (same as in m365_auth_service.go, should be refactored but copying for speed)
	key := c.cfg.App.SecretKey
	if len(key) < 32 {
		key = fmt.Sprintf("%-32s", key)
	}
	key = key[:32]

	credsJSON, err := crypto.Decrypt(ds.Credentials, key)
	if err != nil {
		return fmt.Errorf("decrypt credentials: %w", err)
	}

	var creds map[string]string
	if err := json.Unmarshal([]byte(credsJSON), &creds); err != nil {
		return fmt.Errorf("unmarshal credentials: %w", err)
	}

	refreshToken, ok := creds["refresh_token"]
	if !ok || refreshToken == "" {
		return fmt.Errorf("refresh token not found in credentials")
	}

	// 2. Setup OAuth2 Config
	oauthConfig := &oauth2.Config{
		ClientID:     c.cfg.Microsoft.ClientID,
		ClientSecret: c.cfg.Microsoft.ClientSecret,
		Endpoint:     microsoft.AzureADEndpoint(c.cfg.Microsoft.TenantID),
		RedirectURL:  c.cfg.Microsoft.RedirectURL,
		Scopes:       []string{"offline_access", "User.Read", "Files.Read.All", "Sites.Read.All"},
	}

	// 3. Create Token Source & Client
	// We start with just the refresh token. The TokenSource will fetch a new AccessToken immediately if needed.
	token := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-1 * time.Hour), // Force refresh
	}

	tokenSource := oauthConfig.TokenSource(ctx, token)
	c.client = oauth2.NewClient(ctx, tokenSource)

	return nil
}

// DiscoverSchema traverses OneDrive to find files and folders.
func (c *M365Connector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.client == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	var entities []discovery.DataEntity
	// Traverse root
	if err := c.traverse(ctx, "root", &entities, input); err != nil {
		return nil, nil, err
	}

	inv := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
		LastScannedAt: time.Now(),
	}

	return inv, entities, nil
}

// traverse recursively lists children of a drive item.
func (c *M365Connector) traverse(ctx context.Context, itemID string, entities *[]discovery.DataEntity, input discovery.DiscoveryInput) error {
	url := fmt.Sprintf("%s/me/drive/items/%s/children", GraphAPIEndpoint, itemID)
	if itemID == "root" {
		url = fmt.Sprintf("%s/me/drive/root/children", GraphAPIEndpoint)
	}

	for url != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			// Rate limit
			// Simple retry logic could go here, for now just fail or log
			return fmt.Errorf("rate limited")
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("graph api error: %d", resp.StatusCode)
		}

		var result struct {
			Value []struct {
				ID     string    `json:"id"`
				Name   string    `json:"name"`
				Folder *struct{} `json:"folder"`
				File   *struct {
					MimeType string `json:"mimeType"`
				} `json:"file"`
				LastModifiedDateTime time.Time `json:"lastModifiedDateTime"`
			} `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		for _, item := range result.Value {
			entityType := discovery.EntityTypeFile
			if item.Folder != nil {
				entityType = discovery.EntityTypeFolder
			}

			// Check ChangedSince
			if !input.ChangedSince.IsZero() && item.LastModifiedDateTime.Before(input.ChangedSince) {
				// If folder, we might still need to traverse if children changed?
				// Graph API delta query is better for this, but recursive simple traversal assumes checking everything.
				// If strictly enforcing changedSince on file:
				if entityType == discovery.EntityTypeFile {
					continue
				}
			}

			entity := discovery.DataEntity{
				Name:   item.Name,
				Type:   entityType,
				Schema: itemID, // Parent ID as schema/path?
			}

			// For this implementation, we just list flattened
			*entities = append(*entities, entity)

			if item.Folder != nil {
				// Recurse
				// Prevent too deep?
				if err := c.traverse(ctx, item.ID, entities, input); err != nil {
					return err
				}
			}
		}

		url = result.NextLink
	}
	return nil
}

// GetFields returns "content" for files.
func (c *M365Connector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	// We assume entityID is the Name. But duplicate names exist in folders.
	// Real implementation should use Item ID.
	// But DataEntity.Name is what DiscoveryService passes.
	// Ideally we store ID in DataEntity.
	// DataEntity has ID (its DB ID), but not external ID field.
	// We used Name.
	// If we assume Name is unique or we don't support GetFields correctly without ID.
	// We'll return a generic "content" field.
	return []discovery.DataField{
		{Name: "content", DataType: "string"},
	}, nil
}

// SampleData returns empty because we use Scan for M365.
func (c *M365Connector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	return []string{}, nil
}

// Close is a no-op.
func (c *M365Connector) Close() error {
	return nil
}

// Scan performs the deep scan using FileScanner.
func (c *M365Connector) Scan(ctx context.Context, ds *discovery.DataSource, onFinding func(discovery.PIIClassification)) error {
	if c.client == nil {
		return fmt.Errorf("not connected")
	}

	// 1. Traverse and Scan
	// We need to traverse again or use an inventory?
	// The `Scan` method usually runs on a schedule.
	// We'll traverse and scan on the fly.
	return c.traverseAndScan(ctx, "root", ds.ID, onFinding)
}

func (c *M365Connector) traverseAndScan(ctx context.Context, itemID string, dsID types.ID, onFinding func(discovery.PIIClassification)) error {
	url := fmt.Sprintf("%s/me/drive/items/%s/children", GraphAPIEndpoint, itemID)
	if itemID == "root" {
		url = fmt.Sprintf("%s/me/drive/root/children", GraphAPIEndpoint)
	}

	for url != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}
		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return fmt.Errorf("graph api error during scan: %d", resp.StatusCode)
		}

		var result struct {
			Value []struct {
				ID     string                     `json:"id"`
				Name   string                     `json:"name"`
				Folder *struct{}                  `json:"folder"`
				File   *struct{ MimeType string } `json:"file"`
				Size   int64                      `json:"size"`
			} `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()

		for _, item := range result.Value {
			if item.Folder != nil {
				if err := c.traverseAndScan(ctx, item.ID, dsID, onFinding); err != nil {
					return err
				}
			} else if item.File != nil {
				// Process File
				if err := c.scanFile(ctx, item.ID, item.Name, item.Size, dsID, onFinding); err != nil {
					// Log error but continue
					c.logger.Warn("failed to scan file", "file", item.Name, "error", err)
				}
			}
		}
		url = result.NextLink
	}
	return nil
}

func (c *M365Connector) scanFile(ctx context.Context, itemID, name string, size int64, dsID types.ID, onFinding func(discovery.PIIClassification)) error {
	// Download URL
	// GET /me/drive/items/{item-id}/content
	url := fmt.Sprintf("%s/me/drive/items/%s/content", GraphAPIEndpoint, itemID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %d", resp.StatusCode)
	}

	// Stream to FileScanner
	limit := int64(MaxFileSize)
	if size > limit {
		// Just read first 10MB
	}
	// FileScanner takes reader and limit
	findings, err := c.fileScanner.ScanStream(ctx, resp.Body, name, limit)
	if err != nil {
		return err
	}

	for _, f := range findings {
		f.DataSourceID = dsID
		// Fill other fields if needed, FileScanner fills most.
		// EntityName is name.
		onFinding(f)
	}
	return nil
}

// Delete is a stub for M365.
func (c *M365Connector) Delete(ctx context.Context, entity string, filter map[string]string) (int64, error) {
	c.logger.Warn("delete requested but not supported for m365", "entity", entity)
	return 0, fmt.Errorf("delete not supported for m365")
}

// Export is a stub for M365.
func (c *M365Connector) Export(ctx context.Context, entity string, filter map[string]string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("export not supported for m365")
}

// Compile-time check
// ListUsers lists all users in the tenant.
func (c *M365Connector) ListUsers(ctx context.Context) ([]m365.User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	gc := m365.NewClient(c.client)
	return gc.ListUsers(ctx)
}

// ListSites lists all SharePoint sites.
func (c *M365Connector) ListSites(ctx context.Context) ([]m365.Site, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}
	gc := m365.NewClient(c.client)
	return gc.ListSites(ctx)
}

var _ ScannableConnector = (*M365Connector)(nil)
