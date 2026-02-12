package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/crypto"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

// M365AuthService handles Microsoft 365 authentication flows.
type M365AuthService struct {
	oauthConfig *oauth2.Config
	dsRepo      discovery.DataSourceRepository
	eventBus    eventbus.EventBus
	cfg         *config.Config // Need full config for secret key
	logger      *slog.Logger
}

// NewM365AuthService creates a new M365AuthService.
func NewM365AuthService(cfg *config.Config, dsRepo discovery.DataSourceRepository, eb eventbus.EventBus, logger *slog.Logger) *M365AuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.Microsoft.ClientID,
		ClientSecret: cfg.Microsoft.ClientSecret,
		RedirectURL:  cfg.Microsoft.RedirectURL,
		Scopes:       []string{"offline_access", "User.Read", "Files.Read.All", "Sites.Read.All"},
		Endpoint:     microsoft.AzureADEndpoint(cfg.Microsoft.TenantID),
	}

	return &M365AuthService{
		oauthConfig: oauthConfig,
		dsRepo:      dsRepo,
		eventBus:    eb,
		cfg:         cfg,
		logger:      logger.With("service", "m365_auth"),
	}
}

// GetAuthURL returns the URL to start the OAuth2 flow.
func (s *M365AuthService) GetAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeAndConnect exchanges the auth code for tokens and creates/updates a DataSource.
// It encrypts the refresh token before storage.
func (s *M365AuthService) ExchangeAndConnect(ctx context.Context, code string, tenantID types.ID) (*discovery.DataSource, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange token: %w", err)
	}

	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token received (check if offline_access scope is requested)")
	}

	// 1. Fetch User Profile (to name the data source)
	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		return nil, fmt.Errorf("fetch user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("fetch user profile failed: status %d", resp.StatusCode)
	}

	var profile struct {
		DisplayName       string `json:"displayName"`
		UserPrincipalName string `json:"userPrincipalName"`
		ID                string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("decode user profile: %w", err)
	}

	// 2. Encrypt Credentials (Refresh Token)
	// We store the whole token struct or just the refresh token?
	// Just refresh token is enough to get new access tokens.
	// But let's store the whole token as JSON for completeness (expiry, etc).
	// ACTUALLY, usually we just need Refresh Token for background jobs.
	// Let's store a JSON with refresh_token.

	creds := map[string]string{
		"refresh_token": token.RefreshToken,
	}
	credsJSON, _ := json.Marshal(creds)

	// Encrypt using App Secret Key (must be 32 bytes, ensure padding or hashing in real app if not)
	// For this task, assuming APP_SECRET_KEY is valid 32 chars or we need to handle it.
	// We'll use a derived key or just slice if it's long enough.
	// The config loaded "change-me-in-prod" which is < 32 bytes...
	// Wait, AES-256 needs 32 bytes.
	// If APP_SECRET_KEY is short, we should pad it.
	key := s.cfg.App.SecretKey
	if len(key) < 32 {
		key = fmt.Sprintf("%-32s", key) // Pad with spaces
	}
	key = key[:32] // Truncate to 32

	encryptedCreds, err := crypto.Encrypt(string(credsJSON), key)
	if err != nil {
		return nil, fmt.Errorf("encrypt credentials: %w", err)
	}

	// 3. Create or Update DataSource
	// We check if a datasource with this specific external ID (profile.ID) exists?
	// The DataSource entity doesn't have an ExternalID field.
	// We can check by Name or just always create new?
	// Duplicate names might be allowed or we append ID.
	// Let's Name it: "OneDrive - <UserPrincipalName>"

	dsName := fmt.Sprintf("Microsoft 365 - %s", profile.UserPrincipalName)

	// Create Config JSON
	configMap := map[string]string{
		"user_id":   profile.ID,
		"email":     profile.UserPrincipalName,
		"tenant_id": s.cfg.Microsoft.TenantID, // The M365 tenant
	}
	configJSON, _ := json.Marshal(configMap)

	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID()},
			TenantID:   tenantID,
		},
		Name:        dsName,
		Type:        types.DataSourceMicrosoft365,
		Description: fmt.Sprintf("Connected via account %s", profile.UserPrincipalName),
		Host:        "graph.microsoft.com",
		Port:        443,
		Database:    "onedrive", // logical name
		Credentials: encryptedCreds,
		Config:      string(configJSON),
		Status:      discovery.ConnectionStatusConnected,
		LastSyncAt:  types.Ptr(time.Now()),
	}

	if err := s.dsRepo.Create(ctx, ds); err != nil {
		return nil, fmt.Errorf("create data source: %w", err)
	}

	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(
		eventbus.EventDataSourceCreated, "discovery", tenantID,
		map[string]any{"id": ds.ID, "name": ds.Name, "type": string(ds.Type)},
	))

	s.logger.InfoContext(ctx, "m365 data source created", "id", ds.ID, "user", profile.UserPrincipalName)
	return ds, nil
}
