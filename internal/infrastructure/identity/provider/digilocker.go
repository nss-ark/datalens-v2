package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/complyark/datalens/internal/domain/identity"
)

const (
	DigiLockerBaseURL  = "https://api.digilocker.gov.in"
	DigiLockerAuthURL  = "https://digilocker.gov.in/public/oauth2/1/authorize"
	DigiLockerTokenURL = "https://api.digilocker.gov.in/public/oauth2/1/token"
	DigiLockerUserURL  = "https://api.digilocker.gov.in/public/oauth2/1/user"
	DigiLockerFilesURL = "https://api.digilocker.gov.in/public/oauth2/1/files"
)

type DigiLockerProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

// SetHTTPClient sets the HTTP client for testing purposes.
func (p *DigiLockerProvider) SetHTTPClient(client *http.Client) {
	p.httpClient = client
}

func NewDigiLockerProvider(clientID, clientSecret, redirectURI string) *DigiLockerProvider {
	return &DigiLockerProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *DigiLockerProvider) Name() string {
	return "DigiLocker"
}

func (p *DigiLockerProvider) GetAuthorizationURL(state string) string {
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", p.clientID)
	params.Add("redirect_uri", p.redirectURI)
	params.Add("state", state)
	// Additional scopes can be added here if needed

	return fmt.Sprintf("%s?%s", DigiLockerAuthURL, params.Encode())
}

func (p *DigiLockerProvider) ExchangeToken(ctx context.Context, code string) (*identity.TokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", p.clientID)
	data.Set("client_secret", p.clientSecret)
	data.Set("redirect_uri", p.redirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", DigiLockerTokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("digilocker error (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp identity.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

func (p *DigiLockerProvider) GetUserProfile(ctx context.Context, token string) (*identity.UserProfile, error) {
	// DigiLocker user endpoint requires HMAC signing of parameters if any,
	// but for user profile it's usually just the bearer token.
	// However, some implementations require HMAC of the access token itself or body.
	// Based on standard DigiLocker docs, it's a Bearer token.

	req, err := http.NewRequestWithContext(ctx, "GET", DigiLockerUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// NOTE: If DigiLocker requires HMAC for the User API specifically, implementing it here.
	// Usually HMAC is for file access or specific partner APIs.
	// Assuming standard OAuth2 UserInfo behavior for now, but adding HMAC helper below just in case.

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user profile, status: %d", resp.StatusCode)
	}

	// DigiLocker returns user info in a specific format
	var result struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Mobile       string `json:"mobile"`
		DOB          string `json:"dob"`
		Gender       string `json:"gender"`
		DigiLockerID string `json:"digilockerid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode user profile: %w", err)
	}

	return &identity.UserProfile{
		ProviderID:  result.DigiLockerID,
		Name:        result.Name,
		Email:       result.Email,
		Phone:       result.Mobile,
		DateOfBirth: result.DOB,
		Gender:      result.Gender,
	}, nil
}

func (p *DigiLockerProvider) FetchDocuments(ctx context.Context, token string) ([]identity.IdentityDocument, error) {
	// This endpoint lists issued documents
	req, err := http.NewRequestWithContext(ctx, "GET", DigiLockerFilesURL+"/issued", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch documents, status: %d", resp.StatusCode)
	}

	var fileList struct {
		Items []struct {
			Uri    string `json:"uri"`
			Name   string `json:"name"`
			Type   string `json:"type"`
			Date   string `json:"date"`
			Issuer string `json:"issuer"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&fileList); err != nil {
		return nil, fmt.Errorf("failed to decode document list: %w", err)
	}

	var documents []identity.IdentityDocument
	for _, item := range fileList.Items {
		// Map DigiLocker types to our internal types
		docType := mapDigiLockerTypeToInternal(item.Type)
		if docType == "" {
			continue // Skip unknown document types
		}

		documents = append(documents, identity.IdentityDocument{
			Type:        docType,
			ReferenceID: item.Uri, // URI serves as unique ref
			Issuer:      item.Issuer,
			VerifiedAt:  time.Now(), // Verified at moment of fetch
			Metadata: map[string]any{
				"original_name": item.Name,
				"issue_date":    item.Date,
			},
		})
	}

	return documents, nil
}

// mapDigiLockerTypeToInternal maps DigiLocker doctypes to our domain types
func mapDigiLockerTypeToInternal(dlType string) identity.DocumentType {
	switch dlType {
	case "ADHAR":
		return identity.DocumentTypeAadhaar
	case "PANCR":
		return identity.DocumentTypePAN
	case "DRVLC":
		return identity.DocumentTypeDrivingLicense
	// Add more mappings as needed
	default:
		return ""
	}
}

// computeHMAC256 calculates the HMAC-SHA256 signature
func computeHMAC256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyGuardian logic would go here, possibly using a similar flow or a specific API
func (p *DigiLockerProvider) VerifyGuardian(ctx context.Context, guardianID string) (bool, error) {
	// Implementation depends on specific API for guardian verification
	return false, fmt.Errorf("not implemented")
}
