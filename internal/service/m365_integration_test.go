//go:build integration

package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/pkg/crypto"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestM365AuthService_Integration(t *testing.T) {
	pool := setupPostgres(t)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	eventBus := newMockEventBus()

	// Mock Microsoft Graph API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			// Access Token Endpoint
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"access_token": "fake-access-token",
				"refresh_token": "fake-refresh-token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
		case "/me":
			// User Profile Endpoint
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"id": "m365-user-123",
				"displayName": "Test User",
				"userPrincipalName": "testuser@example.com"
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	// Config with mock server endpoint
	cfg := &config.Config{
		App: config.AppConfig{
			SecretKey: "12345678901234567890123456789012", // 32 bytes
		},
		Microsoft: config.MicrosoftConfig{
			ClientID:     "mock-client-id",
			ClientSecret: "mock-client-secret",
			TenantID:     "mock-tenant-id",
			RedirectURL:  "http://localhost:8080/callback",
		},
	}

	// Override Endpoint in service creation or via config if possible.
	// The service uses microsoft.AzureADEndpoint which is hardcoded.
	// We need to inject the HTTP client into the context to intercept the request.

	dsRepo := repository.NewDataSourceRepo(pool)
	authService := NewM365AuthService(cfg, dsRepo, eventBus, logger)

	// Context with Mock HTTP Client
	ctx := context.Background()
	// Create a client that routes requests to our mock server
	// But OAuth2 library uses hardcoded endpoints if we use microsoft.AzureADEndpoint.
	// We can't easily change the endpoint URL in the service without modifying the service code
	// OR we can use a custom Transport that redirects specific domains to localhost.
	// Easier approach: The service uses `oauthConfig.Exchange(ctx, ...)` which uses the client from context.
	// We need to make sure the Exchange call goes to our mock server.
	//
	// Wait, the service initializes oauthConfig with `microsoft.AzureADEndpoint(tenantID)`.
	// That URL is `https://login.microsoftonline.com/...`.
	// Our mock server is `127.0.0.1:xxxxx`.
	// We can use a custom Transport in the client to intercept `login.microsoftonline.com` and `graph.microsoft.com`.

	mockTransport := &MockTransport{
		RoundTripper: http.DefaultTransport,
		BaseURL:      mockServer.URL,
	}
	httpClient := &http.Client{Transport: mockTransport}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	tenantID := types.NewID()

	t.Run("Exchange Flow and DataSource Creation", func(t *testing.T) {
		ds, err := authService.ExchangeAndConnect(ctx, "fake-auth-code", tenantID)
		require.NoError(t, err)
		require.NotNil(t, ds)

		assert.Equal(t, "Microsoft 365 - testuser@example.com", ds.Name)
		assert.Equal(t, types.DataSourceMicrosoft365, ds.Type)
		assert.Equal(t, tenantID, ds.TenantID)

		// Verify Encryption
		// Decrypt credentials to check
		decryptedJSON, err := crypto.Decrypt(ds.Credentials, cfg.App.SecretKey)
		require.NoError(t, err, "failed to decrypt credentials")

		var creds map[string]string
		err = json.Unmarshal([]byte(decryptedJSON), &creds)
		require.NoError(t, err)
		assert.Equal(t, "fake-refresh-token", creds["refresh_token"])

		// Verify Config
		var conf map[string]string
		err = json.Unmarshal([]byte(ds.Config), &conf)
		require.NoError(t, err)
		assert.Equal(t, "testuser@example.com", conf["email"])
		assert.Equal(t, "m365-user-123", conf["user_id"])
	})
}

// MockTransport redirects requests to the mock server
type MockTransport struct {
	RoundTripper http.RoundTripper
	BaseURL      string
}

func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite URL to mock server
	// We need to keep the path but change scheme/host
	// The mock server handles /token (from login.microsoftonline.com) and /me (from graph.microsoft.com)
	//
	// `oauth2` package might append path to the Endpoint URL.
	// The service sets Endpoint to `https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token`
	// So requests will go there.
	// We need to map that path to `/token` on our mock server.

	newURL := *req.URL
	newURL.Scheme = "http"
	newURL.Host = strings.TrimPrefix(t.BaseURL, "http://")

	if strings.Contains(req.URL.Path, "/oauth2/v2.0/token") {
		newURL.Path = "/token"
	} else if strings.Contains(req.URL.Path, "/v1.0/me") {
		newURL.Path = "/me"
	}

	req.URL = &newURL
	return t.RoundTripper.RoundTrip(req)
}
