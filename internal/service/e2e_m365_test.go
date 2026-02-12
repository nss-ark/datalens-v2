//go:build integration

package service

import (
	"context"
	"encoding/json"
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

// E2EMockTransport intercepts requests to login.microsoftonline.com and redirects to mock server
type E2EMockTransport struct {
	MockServerURL string
}

func (t *E2EMockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Intercept Token Request
	if req.URL.Host == "login.microsoftonline.com" {
		// Redirect to mock server
		// We can just use the path /token if we set it up that way
		u, _ := url.Parse(t.MockServerURL + "/token")
		req.URL = u
		req.Host = u.Host
		req.URL.Scheme = "http"
	}
	return http.DefaultTransport.RoundTrip(req)
}

// E2EMockScanQueue implements queue.ScanQueue for testing
type E2EMockScanQueue struct {
	Handler func(ctx context.Context, jobID string) error
}

func (q *E2EMockScanQueue) Enqueue(ctx context.Context, jobID string) error {
	// Execute immediately in a goroutine to simulate async
	go func() {
		if q.Handler != nil {
			q.Handler(ctx, jobID)
		}
	}()
	return nil
}

func (q *E2EMockScanQueue) Subscribe(ctx context.Context, handler func(ctx context.Context, jobID string) error) error {
	q.Handler = handler
	return nil
}

func TestM365_E2E_Mock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// 1. Setup Postgres (uses helper from setup_integration_test.go)
	pool := setupPostgres(t)
	// setupSchema is implicitly handled by TestMain in setup_integration_test.go calling applyMigrations

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 2. Setup Mock Graph API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock Graph API Request: %s %s", r.Method, r.URL.Path)

		switch r.URL.Path {
		case "/me/drive":
			// Return root drive
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "drive-1",
				"name": "OneDrive",
				"driveType": "personal",
				"webUrl": "https://onedrive.live.com/"
			}`))
		case "/drives/drive-1/items/root/children":
			// Return root children
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"value": [
					{
						"id": "file-1",
						"name": "sensitive_doc.docx",
						"size": 1024,
						"webUrl": "https://onedrive.live.com/file-1",
						"file": {"mimeType": "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
						"lastModifiedDateTime": "2023-10-01T12:00:00Z"
					}
				]
			}`))
		case "/drives/drive-1/items/file-1/content":
			// Return file content with PII
			// NOTE: PatternStrategy uses anchored regex (^...$), so we provide ONLY the PII to ensure match.
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`test@example.com`))
		case "/sites":
			// Return empty sites to keep test simple
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"value": []}`))
		case "/token":
			// Mock Token Response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "mock-new-access-token",
				"token_type": "Bearer",
				"expires_in": 3600,
				"scope": "Files.Read.All"
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// 3. Setup Repositories (Real Postgres)
	dsRepo := repository.NewDataSourceRepo(pool)
	scanRepo := repository.NewScanRunRepo(pool)
	invRepo := repository.NewDataInventoryRepo(pool)
	entityRepo := repository.NewDataEntityRepo(pool)
	fieldRepo := repository.NewDataFieldRepo(pool)
	piiRepo := repository.NewPIIClassificationRepo(pool)
	tenantRepo := repository.NewTenantRepo(pool)

	// Mock Queue
	scanQueue := &E2EMockScanQueue{}
	eventBus := newMockEventBus() // Helper from mocks_test.go (same package)

	encryptionKey := "12345678901234567890123456789012" // 32 bytes

	// 4. Setup Registry and Services
	cfg := &config.Config{
		App: config.AppConfig{
			SecretKey: encryptionKey,
		},
		Microsoft: config.MicrosoftConfig{
			ClientID:     "mock-client-id",
			ClientSecret: "mock-client-secret",
			TenantID:     "mock-tenant-id",
		},
	}
	// Need to register with config that supports M365 (via SecretKey in AppConfig)
	registry := connector.NewConnectorRegistry(cfg)

	// Setup Detection Service
	// Use PatternStrategy to detect the email
	patternStrategy := detection.NewPatternStrategy()
	detector := detection.NewComposableDetector(patternStrategy)

	// Setup Discovery Service
	// Signature: (dsRepo, invRepo, entityRepo, fieldRepo, piiRepo, scanRunRepo, registry, detector, eb, logger)
	discoverySvc := NewDiscoveryService(
		dsRepo, invRepo, entityRepo, fieldRepo, piiRepo, scanRepo,
		registry, detector, eventBus, logger,
	)

	// Setup Scan Service
	scanSvc := NewScanService(scanRepo, dsRepo, scanQueue, discoverySvc, logger)
	// Start worker to process queue
	scanSvc.StartWorker(context.Background())

	// 5. Create Data Setup
	tenantID := types.NewID()
	ctx := context.Background()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	// Inject custom HTTP client for OAuth2
	httpClient := &http.Client{
		Transport: &E2EMockTransport{MockServerURL: mockServer.URL},
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	// Create Tenant
	tenant := &identity.Tenant{
		BaseEntity: types.BaseEntity{
			ID:        tenantID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:     "Test Org",
		Domain:   "test.org",
		Industry: "Tech",
		Country:  "US",
		Plan:     identity.PlanStarter,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{
			RetentionDays: 30,
		},
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Encrypt credentials
	creds := map[string]interface{}{
		"refresh_token": "mock-refresh-token",
		"access_token":  "mock-access-token",
		"expires_in":    3600,
	}
	credsJSON, _ := json.Marshal(creds)
	encryptedCreds, err := crypto.Encrypt(string(credsJSON), encryptionKey)
	require.NoError(t, err)

	// Config with Graph Endpoint override
	mockConfig := map[string]string{
		"client_id":      "mock-client-id",
		"client_secret":  "mock-client-secret",
		"tenant_id":      "mock-tenant-id",
		"graph_endpoint": mockServer.URL, // KEY: Pointing to mock server
	}
	mockConfigJSON, _ := json.Marshal(mockConfig)

	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{
				ID:        types.NewID(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			TenantID: tenant.ID,
		},
		Name:        "Mock OneDrive",
		Type:        types.DataSourceMicrosoft365, // types package
		Credentials: encryptedCreds,
		Config:      string(mockConfigJSON),
		Status:      discovery.ConnectionStatusConnected,
		// Sensitivity removed as it doesn't exist in DataSource struct
	}

	// Create DS in DB
	err = dsRepo.Create(ctx, ds)
	require.NoError(t, err)

	// Update context with correct tenant ID
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenant.ID)

	// 6. Enqueue Scan
	run, err := scanSvc.EnqueueScan(ctx, ds.ID, tenant.ID, discovery.ScanTypeFull)
	require.NoError(t, err)
	assert.NotNil(t, run)
	assert.Equal(t, discovery.ScanStatusPending, run.Status)

	// 7. Wait for Completion
	// Since we use in-memory queue and StartWorker, it runs in background.
	// We poll DB for status.
	require.Eventually(t, func() bool {
		updatedRun, err := scanRepo.GetByID(ctx, run.ID)
		if err != nil {
			return false
		}
		if updatedRun.Status == discovery.ScanStatusFailed {
			t.Logf("Scan failed: %s", *updatedRun.ErrorMessage)
			return false // Fail immediately or let timeout handle it
		}
		return updatedRun.Status == discovery.ScanStatusCompleted
	}, 10*time.Second, 500*time.Millisecond, "Scan did not complete in time")

	// 8. Verify Findings
	// Check Data Inventory
	inv, err := invRepo.GetByDataSource(ctx, ds.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, inv.TotalEntities, 1)

	// Check Data Entity
	entities, err := entityRepo.GetByInventory(ctx, inv.ID)
	require.NoError(t, err)
	foundFile := false
	for _, e := range entities {
		if e.Name == "drive-1|file-1|sensitive_doc.docx" { // Name format from M365 connector
			foundFile = true
			break
		}
	}
	assert.True(t, foundFile, "Should have found sensitive_doc.docx")

	// Check PII Classifications
	// We expect 1 classification for Email
	pgResult, err := piiRepo.GetByDataSource(ctx, ds.ID, types.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	foundEmail := false
	if pgResult != nil {
		for _, pii := range pgResult.Items {
			if pii.Type == types.PIITypeEmail {
				foundEmail = true
				break
			}
		}
	}
	assert.True(t, foundEmail, "Should have detected Email PII")
}
