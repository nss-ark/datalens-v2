package m365

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/stretchr/testify/assert"
)

func TestOutlookConnector_Capabilities(t *testing.T) {
	c := NewOutlookConnector(&config.Config{})
	caps := c.Capabilities()
	assert.True(t, caps.CanDiscover)
	assert.True(t, caps.CanSample)
	assert.False(t, caps.CanDelete)
}

func TestOutlookConnector_DiscoverSchema(t *testing.T) {
	// Mock Graph API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1.0/me/mailFolders" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"value": [
					{"id": "id-inbox", "displayName": "Inbox", "totalItemCount": 10},
					{"id": "id-sent", "displayName": "Sent Items", "totalItemCount": 5}
				]
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create Connector within test with mocked client
	c := NewOutlookConnector(&config.Config{})

	// Inject client directly (white-box test)
	// We need to override the Base URL which is hardcoded in specific methods or use transport interception?
	// The connector hardcodes "https://graph.microsoft.com".
	// We should probably make the BaseURL configurable or use a transport that intercepts.

	// Better approach: Connector methods use full URL.
	// We can't change the URL easily without changing code to use a baseURL variable.
	// OR we can use a custom Transport that redirects graph.microsoft.com to our test server.

	// Let's modify the client transport.
	client := server.Client()
	client.Transport = &rewriteTransport{
		Transport: client.Transport,
		URL:       server.URL,
	}
	c.client = client

	// Test DiscoverSchema
	ctx := context.Background()
	inv, entities, err := c.DiscoverSchema(ctx, discovery.DiscoveryInput{})

	assert.NoError(t, err)
	assert.NotNil(t, inv)
	assert.Equal(t, 2, inv.TotalEntities)
	assert.Len(t, entities, 2)
	assert.Equal(t, "Inbox", entities[0].Name)
	assert.Equal(t, int64(10), *entities[0].RowCount)
}

func TestOutlookConnector_GetFields(t *testing.T) {
	c := NewOutlookConnector(&config.Config{})
	fields, err := c.GetFields(context.Background(), "Inbox")
	assert.NoError(t, err)
	assert.NotEmpty(t, fields)
	assert.Equal(t, "Subject", fields[0].Name)
}

// rewriteTransport rewrites requests to graph.microsoft.com to the test server
type rewriteTransport struct {
	Transport http.RoundTripper
	URL       string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite URL to test server
	// We just replace Scheme and Host
	// The path in the code is /v1.0/... which matches our mock
	req.URL.Scheme = "http"
	req.URL.Host = t.URL[7:] // remove http://
	return t.Transport.RoundTrip(req)
}
