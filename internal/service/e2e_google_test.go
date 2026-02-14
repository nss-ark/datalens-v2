//go:build integration

package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/infrastructure/connector"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/crypto"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGoogle_E2E_Mock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// 1. Setup Postgres
	pool := setupPostgres(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 2. Setup Mock Google API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock Google API Request: %s %s", r.Method, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		// OAuth2 Token
		case "/token":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "mock-google-access-token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))

		// Drive: List Files
		case "/drive/v3/files":
			// Return a single file
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"files": [
					{
						"id": "g-file-1",
						"name": "project_specs.pdf",
						"mimeType": "application/pdf",
						"size": "1024"
					}
				]
			}`))

		// Drive: Download File
		case "/drive/v3/files/g-file-1":
			if r.URL.Query().Get("alt") == "media" {
				w.WriteHeader(http.StatusOK)
				// PII Content
				w.Write([]byte(`confidential@google-project.com`))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "g-file-1",
					"name": "project_specs.pdf",
					"mimeType": "application/pdf"
				}`))
			}

		// Gmail: List Labels (Discovery Schema)
		case "/gmail/v1/users/me/labels":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"labels": [
					{"id": "INBOX", "name": "INBOX", "type": "system"}
				]
			}`))

		// Gmail: List Messages
		case "/gmail/v1/users/me/messages":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"messages": [
					{"id": "msg-1", "threadId": "thread-1"}
				],
				"resultSizeEstimate": 1
			}`))

		// Gmail: Get Message
		case "/gmail/v1/users/me/messages/msg-1":
			// Return payload with PII in body
			body := "Meeting with client@gmail.com regarding the merge."
			w.WriteHeader(http.StatusOK)
			// Minimal Gmail Message Resource with payload
			// Payload structure: { mimeType, body: { data }, parts: [] }
			w.Write([]byte(fmt.Sprintf(`{
				"id": "msg-1",
				"snippet": "Meeting...",
				"payload": {
					"mimeType": "text/plain",
					"body": {
						"data": "%s"
					}
				}
			}`, base64UrlEncode(body))))

		default:
			t.Logf("Unhandled URL: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// 3. Setup Repositories
	dsRepo := repository.NewDataSourceRepo(pool)
	scanRepo := repository.NewScanRunRepo(pool)
	invRepo := repository.NewDataInventoryRepo(pool)
	entityRepo := repository.NewDataEntityRepo(pool)
	fieldRepo := repository.NewDataFieldRepo(pool)
	piiRepo := repository.NewPIIClassificationRepo(pool)
	tenantRepo := repository.NewTenantRepo(pool)

	// Mock Queue
	scanQueue := &E2EMockScanQueue{}
	eventBus := newMockEventBus()

	// 4. Setup Services
	encryptionKey := "12345678901234567890123456789012"
	cfg := &config.Config{
		App: config.AppConfig{
			SecretKey: encryptionKey,
		},
		Google: config.GoogleConfig{
			ClientID:     "mock-google-client-id",
			ClientSecret: "mock-google-client-secret",
		},
	}

	// Setup Detector
	patternStrategy := detection.NewPatternStrategy()
	detector := detection.NewComposableDetector(patternStrategy)

	// Registry with detector
	registry := connector.NewConnectorRegistry(cfg, detector, nil)

	// Discovery Service
	discoverySvc := NewDiscoveryService(
		dsRepo, invRepo, entityRepo, fieldRepo, piiRepo, scanRepo,
		registry, detector, eventBus, logger,
	)

	// Scan Service
	scanSvc := NewScanService(scanRepo, dsRepo, scanQueue, discoverySvc, logger)
	scanSvc.StartWorker(context.Background())

	// 5. Create Data (Tenant + DS)
	tenantID := types.NewID()
	ctx := context.Background()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	// In test, we need to instruct the Google Connector to use our client/transport
	// The `google` package looks for `oauth2.HTTPClient` in the context when creating the service
	httpClient := &http.Client{
		Transport: &GoogleMockTransport{MockServerURL: mockServer.URL},
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	// Create Tenant
	tenant := &identity.Tenant{
		BaseEntity: types.BaseEntity{ID: tenantID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name:       "Google Test Org", Domain: "google.test", Plan: identity.PlanStarter, Status: identity.TenantActive,
		Settings: identity.TenantSettings{RetentionDays: 30},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	// Encrypt Creds
	creds := map[string]string{
		"refresh_token": "mock-refresh-token",
		"access_token":  "mock-access-token",
	}
	credsBytes, _ := json.Marshal(creds)
	encryptedCreds, _ := crypto.Encrypt(string(credsBytes), encryptionKey)

	// Create DataSource
	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
			TenantID:   tenant.ID,
		},
		Name:        "Google Workspace",
		Type:        types.DataSourceGoogleDrive, // Using Google Drive type for both Drive/Gmail mostly
		Credentials: encryptedCreds,
		Config:      "{}",
		Status:      discovery.ConnectionStatusConnected,
	}
	require.NoError(t, dsRepo.Create(ctx, ds))

	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenant.ID)

	// 6. Enqueue Scan
	run, err := scanSvc.EnqueueScan(ctx, ds.ID, tenant.ID, discovery.ScanTypeFull)
	require.NoError(t, err)

	// 7. Wait for Completion
	require.Eventually(t, func() bool {
		updatedRun, err := scanRepo.GetByID(ctx, run.ID)
		if err != nil {
			return false
		}
		if updatedRun.Status == discovery.ScanStatusFailed {
			t.Logf("Scan Failed: %s", *updatedRun.ErrorMessage)
			return false
		}
		return updatedRun.Status == discovery.ScanStatusCompleted
	}, 10*time.Second, 500*time.Millisecond, "Scan timeout")

	// 8. Verify
	// Check PII - expecting email in Drive file and Gmail body
	piiList, err := piiRepo.GetByDataSource(ctx, ds.ID, types.Pagination{Page: 1, PageSize: 100})
	require.NoError(t, err)

	foundDrivePII := false
	foundGmailPII := false

	// Drive PII: confidential@google-project.com (Email)
	// Gmail PII: client@gmail.com (Email)

	for _, p := range piiList.Items {
		if p.Type == types.PIITypeEmail {
			t.Logf("Found PII: %v", p)
			foundDrivePII = true
			foundGmailPII = true
		}
	}
	assert.True(t, foundDrivePII, "Should return PII from Drive")
	assert.True(t, foundGmailPII, "Should return PII from Gmail")
}

// GoogleMockTransport redirects all traffic to MockServer
type GoogleMockTransport struct {
	MockServerURL string
}

func (t *GoogleMockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Parse Mock URL to get Scheme and Host
	mockURL, err := url.Parse(t.MockServerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid mock server url: %w", err)
	}

	// Override Scheme and Host
	req.URL.Scheme = mockURL.Scheme
	req.URL.Host = mockURL.Host

	return http.DefaultTransport.RoundTrip(req)
}

func base64UrlEncode(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}
