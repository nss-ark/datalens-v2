package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/gmail/v1"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/crypto"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// GoogleAuthService handles Google Workspace authentication flows.
type GoogleAuthService struct {
	oauthConfig *oauth2.Config
	dsRepo      discovery.DataSourceRepository
	eventBus    eventbus.EventBus
	cfg         *config.Config
	logger      *slog.Logger
}

// NewGoogleAuthService creates a new GoogleAuthService.
func NewGoogleAuthService(cfg *config.Config, dsRepo discovery.DataSourceRepository, eb eventbus.EventBus, logger *slog.Logger) *GoogleAuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			drive.DriveReadonlyScope,
			gmail.GmailReadonlyScope,
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleAuthService{
		oauthConfig: oauthConfig,
		dsRepo:      dsRepo,
		eventBus:    eb,
		cfg:         cfg,
		logger:      logger.With("service", "google_auth"),
	}
}

// GetAuthURL returns the URL to start the OAuth2 flow.
func (s *GoogleAuthService) GetAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeAndConnect exchanges the auth code for tokens and creates/updates a DataSource.
func (s *GoogleAuthService) ExchangeAndConnect(ctx context.Context, code string, tenantID types.ID) (*discovery.DataSource, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange token: %w", err)
	}

	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token received (check if offline_access is working)")
	}

	// 1. Fetch User Profile
	client := s.oauthConfig.Client(ctx, token)
	// oauth2Service not needed, using direct API call
	// Actually, simple GET to userinfo endpoint is easier
	userInfo, err := s.fetchUserInfo(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("fetch user info: %w", err)
	}

	// 2. Encrypt Credentials (Refresh Token)
	creds := map[string]string{
		"refresh_token": token.RefreshToken,
	}
	credsJSON, _ := json.Marshal(creds)

	key := s.cfg.App.SecretKey
	if len(key) < 32 {
		key = fmt.Sprintf("%-32s", key)
	}
	key = key[:32]

	encryptedCreds, err := crypto.Encrypt(string(credsJSON), key)
	if err != nil {
		return nil, fmt.Errorf("encrypt credentials: %w", err)
	}

	dsName := fmt.Sprintf("Google Workspace - %s", userInfo.Email)

	// Create Config JSON
	configMap := map[string]string{
		"email": userInfo.Email,
		"name":  userInfo.Name,
		"hd":    userInfo.HD, // Hosted Domain
	}
	configJSON, _ := json.Marshal(configMap)

	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID()},
			TenantID:   tenantID,
		},
		Name:        dsName,
		Type:        types.DataSourceGoogleWorkspace,
		Description: fmt.Sprintf("Connected via account %s", userInfo.Email),
		Host:        "googleapis.com",
		Port:        443,
		Database:    "drive,gmail",
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

	s.logger.InfoContext(ctx, "google data source created", "id", ds.ID, "user", userInfo.Email)
	return ds, nil
}

type googleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	HD    string `json:"hd"` // Hosted Domain
}

func (s *GoogleAuthService) fetchUserInfo(ctx context.Context, client *http.Client) (*googleUserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// makeOauth2Service - not used, we use raw http helper above
func makeOauth2Service(ctx context.Context, client *http.Client) (interface{}, error) {
	return nil, nil
}
