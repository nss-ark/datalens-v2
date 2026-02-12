package m365

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector/shared"
	"github.com/complyark/datalens/pkg/crypto"
)

// Microsoft365Connector implements discovery.Connector for OneDrive and SharePoint.
type Microsoft365Connector struct {
	client    *GraphClient
	secretKey string
	logger    *slog.Logger
}

// NewMicrosoft365Connector creates a new M365 connector.
func NewMicrosoft365Connector(secretKey string) *Microsoft365Connector {
	return &Microsoft365Connector{
		secretKey: secretKey,
		logger:    slog.Default().With("connector", "microsoft365"),
	}
}

// Connect authenticates with the Microsoft Graph API.
func (c *Microsoft365Connector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	if c.client != nil {
		return nil
	}

	// 1. Decrypt Credentials
	// Ensure secret key is correct length (32 bytes for AES-256)
	// Just locally truncating or padding if needed, assuming shared.Encrypt logic aligns
	key := c.secretKey
	if len(key) < 32 {
		key = fmt.Sprintf("%-32s", key)
	}
	key = key[:32]

	decryptedJSON, err := crypto.Decrypt(ds.Credentials, key)
	if err != nil {
		return fmt.Errorf("decrypt credentials: %w", err)
	}

	creds, err := shared.ParseCredentials(decryptedJSON)
	if err != nil {
		return fmt.Errorf("parse credentials: %w", err)
	}

	refreshToken, ok := creds["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return fmt.Errorf("missing refresh token")
	}

	// 2. Initialize Graph Client
	client, err := NewGraphClient(ctx, refreshToken, ds.Config)
	if err != nil {
		return fmt.Errorf("init graph client: %w", err)
	}

	c.client = client
	return nil
}

// DiscoverSchema lists all accessible files in OneDrive and SharePoint.
func (c *Microsoft365Connector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.client == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	var entities []discovery.DataEntity

	// A. Scan OneDrive (User's Drive)
	c.logger.InfoContext(ctx, "scanning onedrive")
	rootDrive, err := c.client.GetRootDrive(ctx)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get root drive", "error", err)
	} else {
		files, err := c.scanDrive(ctx, rootDrive.ID, input.ChangedSince)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to scan root drive", "error", err)
		}
		entities = append(entities, files...)
	}

	// B. Scan SharePoint Sites
	c.logger.InfoContext(ctx, "scanning sharepoint sites")
	sites, err := c.client.GetSites(ctx, "*")
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to list sites", "error", err)
	}

	for _, site := range sites {
		drives, err := c.client.GetDrives(ctx, site.ID)
		if err != nil {
			c.logger.WarnContext(ctx, "failed to get drives for site", "site_id", site.ID, "error", err)
			continue
		}

		for _, drive := range drives {
			files, err := c.scanDrive(ctx, drive.ID, input.ChangedSince)
			if err != nil {
				c.logger.WarnContext(ctx, "failed to scan drive", "drive_id", drive.ID, "error", err)
				continue
			}
			entities = append(entities, files...)

			if len(entities) > 5000 {
				c.logger.WarnContext(ctx, "hit scan limit of 5000 entities", "total_so_far", len(entities))
				break
			}
		}
		if len(entities) > 5000 {
			break
		}
	}

	inv := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
		LastScannedAt: time.Now(),
	}

	return inv, entities, nil
}

// scanDrive recurses through a drive to find files.
func (c *Microsoft365Connector) scanDrive(ctx context.Context, driveID string, changedSince time.Time) ([]discovery.DataEntity, error) {
	var entities []discovery.DataEntity

	// Recursive helper
	var walk func(folderID string) error
	walk = func(folderID string) error {
		items, err := c.client.GetDriveChildren(ctx, driveID, folderID)
		if err != nil {
			return err
		}

		for _, item := range items {
			// Check modification time
			if !changedSince.IsZero() && item.LastModified.Before(changedSince) {
				continue
			}

			if item.Folder != nil {
				// Recurse
				if err := walk(item.ID); err != nil {
					return err
				}
			} else if item.File != nil {
				// It's a file
				name := fmt.Sprintf("%s|%s|%s", driveID, item.ID, item.Name)

				// Map Size to RowCount
				size := item.Size

				entity := discovery.DataEntity{
					Name:     name,
					Type:     discovery.EntityTypeFile,
					Schema:   item.WebURL, // Storing WebURL in Schema field
					RowCount: &size,
				}
				entities = append(entities, entity)
			}
		}
		return nil
	}

	// Start at root
	if err := walk("root"); err != nil {
		return nil, err
	}

	return entities, nil
}

// GetFields returns generic file metadata fields.
func (c *Microsoft365Connector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	return []discovery.DataField{
		{Name: "content", DataType: "blob"},
		{Name: "name", DataType: "string"},
		{Name: "size", DataType: "integer"},
		{Name: "mime_type", DataType: "string"},
		{Name: "created_by", DataType: "string"},
		{Name: "last_modified", DataType: "datetime"},
	}, nil
}

// SampleData retrieves file content for PII inspection.
func (c *Microsoft365Connector) SampleData(ctx context.Context, entityID, field string, limit int) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	parts := strings.Split(entityID, "|")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid entity ID format (expected driveId|itemId|name)")
	}
	driveID := parts[0]
	itemID := parts[1]

	// Download content
	content, err := c.client.GetFileContent(ctx, driveID, itemID)
	if err != nil {
		return nil, err
	}

	// Limit content length
	if len(content) > 10000 {
		content = content[:10000]
	}

	// Simple text check
	if strings.Contains(http.DetectContentType(content), "text") || strings.Contains(strings.ToLower(parts[2]), ".json") || strings.Contains(strings.ToLower(parts[2]), ".csv") {
		return []string{string(content)}, nil
	}

	return []string{string(content)}, nil // Return raw bytes as string for now, detector handles it?
}

func (c *Microsoft365Connector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:         true,
		CanSample:           true,
		SupportsIncremental: true,
	}
}

func (c *Microsoft365Connector) Close() error {
	return nil
}
