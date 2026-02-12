package m365

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector/shared"
	"github.com/complyark/datalens/pkg/crypto" // Assuming crypto package exists and has Decrypt
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

// OutlookConnector implements discovery.Connector for Outlook/M365 Mail.
type OutlookConnector struct {
	client *http.Client
	cfg    *config.Config
}

// NewOutlookConnector creates a new OutlookConnector.
func NewOutlookConnector(cfg *config.Config) *OutlookConnector {
	return &OutlookConnector{cfg: cfg}
}

// Compile-time check
var _ discovery.Connector = (*OutlookConnector)(nil)

// Capabilities returns the supported operations.
func (c *OutlookConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:             true,
		CanSample:               true,
		CanDelete:               false,
		CanUpdate:               false,
		SupportsStreaming:       false, // Not implementing streaming yet
		SupportsSchemaDiscovery: true,
		SupportsDataSampling:    true,
		SupportsParallelScan:    false,
		MaxConcurrency:          1,
	}
}

// Connect authenticates with Microsoft Graph API using the stored refresh token.
func (c *OutlookConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	// 1. Decrypt Credentials
	// Credentials stored as encrypted JSON: {"refresh_token": "..."}
	credsJSON, err := crypto.Decrypt(ds.Credentials, c.cfg.App.SecretKey[:32]) // Ensure 32 bytes
	if err != nil {
		return fmt.Errorf("decrypt credentials: %w", err)
	}

	creds, err := shared.ParseCredentials(credsJSON)
	if err != nil {
		return fmt.Errorf("parse credentials: %w", err)
	}

	refreshToken, ok := creds["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return fmt.Errorf("refresh_token not found in credentials")
	}

	// 2. Setup OAuth2 Config
	oauthConfig := &oauth2.Config{
		ClientID:     c.cfg.Microsoft.ClientID,
		ClientSecret: c.cfg.Microsoft.ClientSecret,
		Endpoint:     microsoft.AzureADEndpoint(c.cfg.Microsoft.TenantID),
		Scopes:       []string{"User.Read", "Mail.Read", "Files.Read.All", "Sites.Read.All"},
	}

	// 3. Create Token Source
	// We only have refresh token. We construct a token with it.
	// AccessToken is likely expired, so TokenSource will refresh it immediately.
	initialToken := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-1 * time.Hour), // Force refresh
	}

	tokenSource := oauthConfig.TokenSource(ctx, initialToken)

	// 4. Create Client
	c.client = oauth2.NewClient(ctx, tokenSource)

	// verify connection
	resp, err := c.client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		return fmt.Errorf("verify connection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection verification failed: status %d", resp.StatusCode)
	}

	return nil
}

// DiscoverSchema lists Mail Folders as DataEntities.
func (c *OutlookConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.client == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	// Fetch Mail Folders
	// https://graph.microsoft.com/v1.0/me/mailFolders
	endpoint := "https://graph.microsoft.com/v1.0/me/mailFolders?$top=50"

	var entities []discovery.DataEntity

	for endpoint != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("create request: %w", err)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("list folders: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, nil, fmt.Errorf("list folders failed: status %d", resp.StatusCode)
		}

		var result struct {
			Value []struct {
				ID             string `json:"id"`
				DisplayName    string `json:"displayName"`
				TotalItemCount int    `json:"totalItemCount"`
			} `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, nil, fmt.Errorf("decode folders: %w", err)
		}

		for _, folder := range result.Value {
			if folder.TotalItemCount > 0 { // Only list non-empty folders? Or all? Let's list all relevant.
				entities = append(entities, discovery.DataEntity{
					Name:     folder.DisplayName,        // Using DisplayName as the Entity Name
					Schema:   "Mail",                    // Grouping
					Type:     discovery.EntityTypeTable, // Best fit
					RowCount: func() *int64 { i := int64(folder.TotalItemCount); return &i }(),
				})
			}
		}

		endpoint = result.NextLink
	}

	inventory := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
	}

	return inventory, entities, nil
}

// GetFields returns standard email fields.
func (c *OutlookConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	// For Outlook, fields are standard across all folders.
	// entityID is the Folder Name (e.g. "Inbox").

	return []discovery.DataField{
		{Name: "Subject", DataType: "string", Nullable: false},
		{Name: "Body", DataType: "string", Nullable: false}, // Body content
		{Name: "Sender", DataType: "string", Nullable: false},
		{Name: "ToRecipients", DataType: "string", Nullable: true},
		{Name: "Attachments", DataType: "array", Nullable: true}, // Special handling
	}, nil
}

// SampleData fetches messages from the folder.
func (c *OutlookConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	// We need the Folder ID. But entity is the Display Name.
	// We should ideally cache the ID map or look it up.
	// For now, I'll fetch folders again to find the ID matching the name.
	// Optimization: Store ID in Entity.Name? No, user friendly names are better.
	// Maybe I can assume entity IS the ID if I change DiscoverSchema?
	// The prompt uses DisplayName.
	// I'll quickly look up ID by name.

	// Helper to find folder ID (could be optimized)
	folderID, err := c.getFolderIDByName(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("resolve folder id: %w", err)
	}

	// Fetch Messages
	// $select=subject,body,sender,toRecipients,hasAttachments
	endpoint := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/mailFolders/%s/messages?$top=%d&$select=subject,body,sender,toRecipients,hasAttachments,id", folderID, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch messages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch messages failed: status %d", resp.StatusCode)
	}

	var result struct {
		Value []struct {
			ID      string `json:"id"`
			Subject string `json:"subject"`
			Body    struct {
				Content string `json:"content"`
			} `json:"body"`
			Sender struct {
				EmailAddress struct {
					Address string `json:"address"`
					Name    string `json:"name"`
				} `json:"emailAddress"`
			} `json:"sender"`
			ToRecipients []struct {
				EmailAddress struct {
					Address string `json:"address"`
				} `json:"emailAddress"`
			} `json:"toRecipients"`
			HasAttachments bool `json:"hasAttachments"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode messages: %w", err)
	}

	var samples []string

	for _, msg := range result.Value {
		switch field {
		case "Subject":
			samples = append(samples, msg.Subject)
		case "Body":
			// Strip HTML? For PII detection, raw HTML is usually fine or might need stripping.
			// Let's return raw content for now.
			samples = append(samples, msg.Body.Content)
		case "Sender":
			samples = append(samples, fmt.Sprintf("%s <%s>", msg.Sender.EmailAddress.Name, msg.Sender.EmailAddress.Address))
		case "ToRecipients":
			var toes []string
			for _, t := range msg.ToRecipients {
				toes = append(toes, t.EmailAddress.Address)
			}
			samples = append(samples, strings.Join(toes, ", "))
		case "Attachments":
			if msg.HasAttachments {
				// Fetch attachments for this message
				// Only fetch checks/content if needed.
				// The prompt says "download if size < 10MB".
				// I should fetch attachment content here.
				attContent, err := c.fetchAttachmentsContent(ctx, msg.ID)
				if err == nil && attContent != "" {
					samples = append(samples, attContent)
				}
			}
		}
	}

	return samples, nil
}

func (c *OutlookConnector) getFolderIDByName(ctx context.Context, name string) (string, error) {
	// Simple pagination loop
	endpoint := "https://graph.microsoft.com/v1.0/me/mailFolders?$select=id,displayName&$top=50"
	for endpoint != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			return "", err
		}
		resp, err := c.client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		var result struct {
			Value []struct {
				ID          string `json:"id"`
				DisplayName string `json:"displayName"`
			} `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}

		for _, f := range result.Value {
			if f.DisplayName == name {
				return f.ID, nil
			}
		}
		endpoint = result.NextLink
	}
	return "", fmt.Errorf("folder not found: %s", name)
}

func (c *OutlookConnector) fetchAttachmentsContent(ctx context.Context, messageID string) (string, error) {
	// https://graph.microsoft.com/v1.0/me/messages/{id}/attachments
	endpoint := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/messages/%s/attachments?$select=id,name,size,contentType", messageID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Size        int    `json:"size"`
			ContentType string `json:"contentType"`
			// Note: FileAttachment has 'contentBytes' but we need to fetch specific type or $value?
			// The list endpoint returns metadata. We might need to fetch individual or assume type.
			// Actually standard list usually returns types.
			ODataType string `json:"@odata.type"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	var contents []string
	for _, att := range result.Value {
		// Limit 10MB
		if att.Size > 10*1024*1024 {
			continue
		}

		// Only process #microsoft.graph.fileAttachment
		if att.ODataType == "#microsoft.graph.fileAttachment" {
			// Fetch content.
			// We can get contentBytes if we requested it?
			// Let's request it: $select=...,contentBytes
			// Re-fetch or specific fetch.
			// Let's try to fetch specific attachment with $value to get raw bytes.
			// GET /me/messages/{id}/attachments/{id}/$value

			// Scan text-based only
			if isTextType(att.ContentType, att.Name) {
				content, err := c.downloadAttachmentContent(ctx, messageID, att.ID)
				if err == nil {
					contents = append(contents, fmt.Sprintf("[Attachment: %s]\n%s", att.Name, content))
				}
			}
		} else if att.ODataType == "#microsoft.graph.itemAttachment" {
			// Item (email inside email). Skip for now.
		}
	}

	return strings.Join(contents, "\n\n"), nil
}

func (c *OutlookConnector) downloadAttachmentContent(ctx context.Context, messageID, attachmentID string) (string, error) {
	// GET /me/messages/{id}/attachments/{id}/$value
	endpoint := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/messages/%s/attachments/%s/$value", messageID, attachmentID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func isTextType(contentType, name string) bool {
	contentType = strings.ToLower(contentType)
	name = strings.ToLower(name)

	if strings.Contains(contentType, "text/") || strings.Contains(contentType, "json") || strings.Contains(contentType, "xml") {
		return true
	}

	exts := []string{".txt", ".csv", ".json", ".xml", ".html", ".htm", ".md", ".log"}
	for _, ext := range exts {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

// Export is a stub.
func (c *OutlookConnector) Export(ctx context.Context, entity string, filter map[string]string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("export not supported for outlook")
}

// Delete is a stub.
func (c *OutlookConnector) Delete(ctx context.Context, entity string, filter map[string]string) (int64, error) {
	return 0, fmt.Errorf("delete not supported for outlook")
}

func (c *OutlookConnector) Close() error {
	// Client doesn't need closing
	return nil
}
