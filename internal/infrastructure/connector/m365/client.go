package m365

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	GraphAPI        = "https://graph.microsoft.com/v1.0"
	DefaultPageSize = 100
)

// ListUsers returns a list of users in the tenant.
func (c *GraphClient) ListUsers(ctx context.Context) ([]User, error) {
	var users []User
	nextLink := "/users?$top=100&$select=id,displayName,mail,jobTitle"

	for nextLink != "" {
		var response struct {
			Value    []User `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}

		if err := c.get(ctx, nextLink, &response); err != nil {
			return nil, err
		}

		users = append(users, response.Value...)
		nextLink = response.NextLink

		// Cap for now
		if len(users) >= 1000 {
			break
		}
	}
	return users, nil
}

// ListSites returns a list of SharePoint sites in the tenant.
func (c *GraphClient) ListSites(ctx context.Context) ([]Site, error) {
	var sites []Site
	// Search=* is required to list all sites via Search API, distinct from /sites endpoint behavior
	nextLink := "/sites?search=*"

	for nextLink != "" {
		var response struct {
			Value    []Site `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}

		if err := c.get(ctx, nextLink, &response); err != nil {
			return nil, err
		}

		sites = append(sites, response.Value...)
		nextLink = response.NextLink

		if len(sites) >= 1000 {
			break
		}
	}
	return sites, nil
}

// GraphClient wraps the HTTP client for Microsoft Graph API.
type GraphClient struct {
	client  *http.Client
	BaseURL string
}

// NewGraphClient initializes a new GraphClient with an OAuth2 token source.
// It handles token refresh automatically using the refreshToken.
func NewGraphClient(ctx context.Context, refreshToken string, configJSON string) (*GraphClient, error) {
	// Parse config to get client ID/Secret/Tenant
	var cfg map[string]string
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	clientID := cfg["client_id"]
	clientSecret := cfg["client_secret"]
	tenantID := cfg["tenant_id"]

	// Allow overriding base URL for testing
	baseURL := cfg["graph_endpoint"]
	if baseURL == "" {
		baseURL = GraphAPI
	}

	if clientID == "" || clientSecret == "" || tenantID == "" {
		return nil, fmt.Errorf("missing client_id, client_secret, or tenant_id in config")
	}

	// Create TokenSource
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID),
		},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-1 * time.Hour), // Force refresh
	}

	return &GraphClient{
		client:  conf.Client(ctx, token),
		BaseURL: baseURL,
	}, nil
}

// NewClient creates a GraphClient from an existing HTTP client.
func NewClient(client *http.Client) *GraphClient {
	return &GraphClient{
		client:  client,
		BaseURL: GraphAPI,
	}
}

// GetRootDrive returns the user's default drive (OneDrive).
func (c *GraphClient) GetRootDrive(ctx context.Context) (*Drive, error) {
	var drive Drive
	if err := c.get(ctx, "/me/drive", &drive); err != nil {
		return nil, err
	}
	return &drive, nil
}

// GetSites returns a list of SharePoint sites matching the search query.
func (c *GraphClient) GetSites(ctx context.Context, search string) ([]Site, error) {
	query := url.Values{}
	query.Set("search", search)

	var response struct {
		Value []Site `json:"value"`
	}

	if err := c.get(ctx, "/sites?"+query.Encode(), &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

// GetDrives returns all document libraries (drives) for a site.
func (c *GraphClient) GetDrives(ctx context.Context, siteID string) ([]Drive, error) {
	var response struct {
		Value []Drive `json:"value"`
	}
	if err := c.get(ctx, fmt.Sprintf("/sites/%s/drives", siteID), &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

// GetDriveChildren lists items in a drive folder.
func (c *GraphClient) GetDriveChildren(ctx context.Context, driveID, itemID string) ([]DriveItem, error) {
	var items []DriveItem
	nextLink := fmt.Sprintf("/drives/%s/items/%s/children", driveID, itemID)

	for nextLink != "" {
		var response struct {
			Value    []DriveItem `json:"value"`
			NextLink string      `json:"@odata.nextLink"`
		}

		// Handle full URL in NextLink vs relative path
		path := nextLink
		if strings.HasPrefix(nextLink, c.BaseURL) {
			path = strings.TrimPrefix(nextLink, c.BaseURL)
		} else if strings.HasPrefix(nextLink, GraphAPI) {
			// fallback for standard API
			path = strings.TrimPrefix(nextLink, GraphAPI)
		}

		if err := c.get(ctx, path, &response); err != nil {
			return nil, err
		}

		items = append(items, response.Value...)
		nextLink = response.NextLink

		// Safety limit just in case
		if len(items) > 5000 {
			break
		}
	}

	return items, nil
}

// GetFileContent retrieves the content of a file.
func (c *GraphClient) GetFileContent(ctx context.Context, driveID, itemID string) ([]byte, error) {
	// Construct full URL dynamically based on BaseURL
	path := fmt.Sprintf("%s/drives/%s/items/%s/content", c.BaseURL, driveID, itemID)
	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (c *GraphClient) get(ctx context.Context, path string, target interface{}) error {
	if !strings.HasPrefix(path, "http") {
		// Ensure path starts with / if not empty, though M365 usually has one.
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		path = c.BaseURL + path
	}

	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api error: %s - %s", resp.Status, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
