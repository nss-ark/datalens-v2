package m365

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphClient_ListUsers(t *testing.T) {
	// Mock Graph API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users", r.URL.Path)
		assert.Equal(t, "id,displayName,mail,jobTitle", r.URL.Query().Get("$select"))

		response := `{
			"value": [
				{"id": "1", "displayName": "User One", "mail": "user1@example.com", "userPrincipalName": "user1@example.com"},
				{"id": "2", "displayName": "User Two", "mail": "user2@example.com", "userPrincipalName": "user2@example.com"}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := &GraphClient{
		client:  server.Client(),
		BaseURL: server.URL,
	}

	users, err := client.ListUsers(context.Background())
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "User One", users[0].DisplayName)
	assert.Equal(t, "user1@example.com", users[0].Mail)
}

func TestGraphClient_ListSites(t *testing.T) {
	// Mock Graph API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/sites", r.URL.Path)
		assert.Equal(t, "*", r.URL.Query().Get("search"))

		response := `{
			"value": [
				{"id": "site1", "name": "Site One", "webUrl": "https://sharepoint.com/sites/site1"},
				{"id": "site2", "name": "Site Two", "webUrl": "https://sharepoint.com/sites/site2"}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := &GraphClient{
		client:  server.Client(),
		BaseURL: server.URL,
	}

	sites, err := client.ListSites(context.Background())
	assert.NoError(t, err)
	assert.Len(t, sites, 2)
	assert.Equal(t, "Site One", sites[0].Name)
	assert.Equal(t, "https://sharepoint.com/sites/site1", sites[0].WebURL)
}
