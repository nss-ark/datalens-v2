package provider_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/complyark/datalens/internal/infrastructure/identity/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTransport allows mocking HTTP responses
type MockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestDigiLockerProvider_GetUserProfile_Mock(t *testing.T) {
	// Mock DigiLocker User Profile JSON
	mockUserResponse := `{
		"name": "Jane Doe",
		"email": "jane.doe@example.com",
		"mobile": "9876543210",
		"dob": "01-01-1990",
		"gender": "F",
		"digilockerid": "in.gov.digilocker.12345678"
	}`

	mockTransport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request URL and params
			assert.Equal(t, "https://api.digilocker.gov.in/public/oauth2/1/user", req.URL.String())
			assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(mockUserResponse)),
				Header:     make(http.Header),
			}, nil
		},
	}

	mockClient := &http.Client{
		Transport: mockTransport,
	}

	p := provider.NewDigiLockerProvider("client-id", "client-secret", "redirect-uri")
	p.SetHTTPClient(mockClient)

	ctx := context.Background()
	userProfile, err := p.GetUserProfile(ctx, "test-token")
	require.NoError(t, err)
	require.NotNil(t, userProfile)

	// Verify the mocked fields
	assert.Equal(t, "Jane Doe", userProfile.Name)
	assert.Equal(t, "jane.doe@example.com", userProfile.Email)
	assert.Equal(t, "9876543210", userProfile.Phone)
	assert.Equal(t, "01-01-1990", userProfile.DateOfBirth)
	assert.Equal(t, "F", userProfile.Gender)
	assert.Equal(t, "in.gov.digilocker.12345678", userProfile.ProviderID)
}
