package connector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector/shared"
	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/crypto"
	"github.com/complyark/datalens/pkg/types"
)

// GoogleConnector implements discovery.Connector for Google Workspace.
type GoogleConnector struct {
	client      *http.Client
	driveSvc    *drive.Service
	gmailSvc    *gmail.Service
	fileScanner *shared.FileScanner
	logger      *slog.Logger
	cfg         *config.Config
}

// NewGoogleConnector creates a new GoogleConnector.
func NewGoogleConnector(cfg *config.Config, detector *detection.ComposableDetector) *GoogleConnector {
	if cfg == nil {
		cfg, _ = config.Load()
	}
	return &GoogleConnector{
		fileScanner: shared.NewFileScanner(detector, slog.Default()),
		logger:      slog.Default().With("connector", "google"),
		cfg:         cfg,
	}
}

// Capabilities returns the supported operations.
func (c *GoogleConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:          true,
		SupportsStreaming:    true,
		SupportsParallelScan: false,
	}
}

// Connect establishes connection using stored credentials.
func (c *GoogleConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	if c.cfg == nil {
		var err error
		c.cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
	}

	if ds.Credentials == "" {
		return fmt.Errorf("credentials required")
	}

	// Decrypt
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
		return fmt.Errorf("refresh token not found")
	}

	// Token Source
	oauthConfig := &oauth2.Config{
		ClientID:     c.cfg.Google.ClientID,
		ClientSecret: c.cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{drive.DriveReadonlyScope, gmail.GmailReadonlyScope},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-1 * time.Hour), // Force refresh
	}

	tokenSource := oauthConfig.TokenSource(ctx, token)
	c.client = oauth2.NewClient(ctx, tokenSource)

	// Create Services
	c.driveSvc, err = drive.NewService(ctx, option.WithHTTPClient(c.client))
	if err != nil {
		return fmt.Errorf("create drive service: %w", err)
	}

	c.gmailSvc, err = gmail.NewService(ctx, option.WithHTTPClient(c.client))
	if err != nil {
		return fmt.Errorf("create gmail service: %w", err)
	}

	return nil
}

// DiscoverSchema lists Drive folders and Gmail labels.
func (c *GoogleConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.client == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	var entities []discovery.DataEntity

	// 1. Discover Drive
	// We'll just list root files/folders for schema discovery.
	// Recursive traversal is expensive for large drives, maybe just root?
	// Spec says "List files".
	if err := c.discoverDrive(ctx, &entities); err != nil {
		c.logger.ErrorContext(ctx, "discover drive failed", "error", err)
		// Continue to Gmail?
	}

	// 2. Discover Gmail
	if err := c.discoverGmail(ctx, &entities); err != nil {
		c.logger.ErrorContext(ctx, "discover gmail failed", "error", err)
	}

	inv := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
		LastScannedAt: time.Now(),
	}

	return inv, entities, nil
}

func (c *GoogleConnector) discoverDrive(ctx context.Context, entities *[]discovery.DataEntity) error {
	// List files in root
	q := "'root' in parents and trashed = false"
	err := c.driveSvc.Files.List().Q(q).Fields("nextPageToken, files(id, name, mimeType)").Pages(ctx, func(files *drive.FileList) error {
		for _, file := range files.Files {
			entityType := discovery.EntityTypeFile
			if file.MimeType == "application/vnd.google-apps.folder" {
				entityType = discovery.EntityTypeFolder
			}
			*entities = append(*entities, discovery.DataEntity{
				Name:   file.Name,
				Type:   entityType,
				Schema: "drive",
				// Path field removed as it doesn't exist in DataEntity
			})
		}
		return nil
	})
	return err
}

func (c *GoogleConnector) discoverGmail(ctx context.Context, entities *[]discovery.DataEntity) error {
	// List Labels
	r, err := c.gmailSvc.Users.Labels.List("me").Do()
	if err != nil {
		return err
	}

	for _, label := range r.Labels {
		*entities = append(*entities, discovery.DataEntity{
			Name:   label.Name,
			Type:   discovery.EntityTypeTable, // Abuse 'Table' for 'Label/Folder'
			Schema: "gmail",
			// Path field removed
		})
	}
	return nil
}

// GetFields returns generic fields.
func (c *GoogleConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	return []discovery.DataField{
		{Name: "content", DataType: "string"},
	}, nil
}

// SampleData returns empty.
func (c *GoogleConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	return []string{}, nil
}

// Close is a no-op.
func (c *GoogleConnector) Close() error {
	return nil
}

// Scan performs deep scan of Drive and Gmail.
func (c *GoogleConnector) Scan(ctx context.Context, ds *discovery.DataSource, onFinding func(discovery.PIIClassification)) error {
	if c.client == nil {
		return fmt.Errorf("not connected")
	}

	// 1. Scan Drive
	if err := c.scanDrive(ctx, ds.ID, onFinding); err != nil {
		c.logger.ErrorContext(ctx, "scan drive failed", "error", err)
	}

	// 2. Scan Gmail
	if err := c.scanGmail(ctx, ds.ID, onFinding); err != nil {
		c.logger.ErrorContext(ctx, "scan gmail failed", "error", err)
	}

	return nil
}

func (c *GoogleConnector) scanDrive(ctx context.Context, dsID types.ID, onFinding func(discovery.PIIClassification)) error {
	// Full traversal using Q
	// We want all non-trashed files.
	// We'll page through everything.
	q := "trashed = false and mimeType != 'application/vnd.google-apps.folder'"

	return c.driveSvc.Files.List().Q(q).
		Fields("nextPageToken, files(id, name, mimeType, size)").
		Pages(ctx, func(files *drive.FileList) error {
			for _, file := range files.Files {
				// Initialize FileScanner checks
				// Convert GDocs? Only if export links are available, but `files.get` media only works for binary files.
				// For GDocs, we need to export.
				// Let's stick to binary files for now (PDF, Docx, etc) + simple text.

				if isGoogleDoc(file.MimeType) {
					// Skip GDocs for now or implement export
					continue
				}

				if err := c.scanDriveFile(ctx, file, dsID, onFinding); err != nil {
					c.logger.WarnContext(ctx, "scan drive file failed", "file", file.Name, "error", err)
				}
			}
			return nil
		})
}

func isGoogleDoc(mimeType string) bool {
	return strings.HasPrefix(mimeType, "application/vnd.google-apps.")
}

func (c *GoogleConnector) scanDriveFile(ctx context.Context, file *drive.File, dsID types.ID, onFinding func(discovery.PIIClassification)) error {
	resp, err := c.driveSvc.Files.Get(file.Id).Download()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	findings, err := c.fileScanner.ScanStream(ctx, resp.Body, file.Name, file.Size)
	if err != nil {
		return err
	}

	for _, f := range findings {
		f.DataSourceID = dsID
		// Location field removed
		onFinding(f)
	}
	return nil
}

func (c *GoogleConnector) scanGmail(ctx context.Context, dsID types.ID, onFinding func(discovery.PIIClassification)) error {
	// List messages (all)
	// 'me' is the user
	req := c.gmailSvc.Users.Messages.List("me").MaxResults(50) // Batch of 50

	return req.Pages(ctx, func(msgs *gmail.ListMessagesResponse) error {
		for _, msg := range msgs.Messages {
			if err := c.scanEmail(ctx, msg.Id, dsID, onFinding); err != nil {
				c.logger.WarnContext(ctx, "scan email failed", "msg_id", msg.Id, "error", err)
			}
		}
		return nil
	})
}

func (c *GoogleConnector) scanEmail(ctx context.Context, msgID string, dsID types.ID, onFinding func(discovery.PIIClassification)) error {
	// Get full message with payload
	msg, err := c.gmailSvc.Users.Messages.Get("me", msgID).Format("full").Do()
	if err != nil {
		return err
	}

	// 1. Scan Body (Snippet or full text)
	// Extract body from payload parts
	bodyText := extractBody(msg.Payload)
	if bodyText != "" {
		// Use FileScanner on text content
		reader := strings.NewReader(bodyText)
		findings, err := c.fileScanner.ScanStream(ctx, reader, "email_body.txt", int64(len(bodyText)))
		if err == nil {
			for _, f := range findings {
				f.DataSourceID = dsID
				// Location field removed
				onFinding(f)
			}
		}
	}

	// 2. Scan Attachments
	for _, part := range msg.Payload.Parts {
		if part.Filename != "" && part.Body.AttachmentId != "" {
			// Download attachment
			att, err := c.gmailSvc.Users.Messages.Attachments.Get("me", msgID, part.Body.AttachmentId).Do()
			if err != nil {
				continue
			}

			data, err := base64.URLEncoding.DecodeString(att.Data)
			if err != nil {
				continue
			}

			reader := strings.NewReader(string(data))
			findings, err := c.fileScanner.ScanStream(ctx, reader, part.Filename, int64(len(data)))
			if err == nil {
				for _, f := range findings {
					f.DataSourceID = dsID
					// Location field removed
					onFinding(f)
				}
			}
		}
	}
	return nil
}

func extractBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}
	if payload.Body != nil && payload.Body.Data != "" {
		data, _ := base64.URLEncoding.DecodeString(payload.Body.Data)
		return string(data)
	}

	var sb strings.Builder
	for _, part := range payload.Parts {
		// Prefer text/plain or text/html
		if part.MimeType == "text/plain" || part.MimeType == "text/html" {
			if part.Body != nil && part.Body.Data != "" {
				data, _ := base64.URLEncoding.DecodeString(part.Body.Data)
				sb.Write(data)
				sb.WriteString("\n")
			}
		}
		// Recurse (e.g. multipart/alternative)
		if len(part.Parts) > 0 {
			sb.WriteString(extractBody(part))
		}
	}
	return sb.String()
}

var _ ScannableConnector = (*GoogleConnector)(nil)
