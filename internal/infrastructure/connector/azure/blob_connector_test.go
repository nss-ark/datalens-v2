package azure

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/discovery"
)

// =============================================================================
// Mocks
// =============================================================================

type MockBlobClient struct {
	mock.Mock
}

func (m *MockBlobClient) NewListBlobsFlatPager(containerName string, options *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse] {
	args := m.Called(containerName, options)
	return args.Get(0).(*runtime.Pager[azblob.ListBlobsFlatResponse])
}

func (m *MockBlobClient) DownloadStream(ctx context.Context, containerName string, blobName string, options *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error) {
	args := m.Called(ctx, containerName, blobName, options)
	return args.Get(0).(azblob.DownloadStreamResponse), args.Error(1)
}

// =============================================================================
// Tests
// =============================================================================

func TestBlob_Connect_ManualClient(t *testing.T) {
	connector := NewBlobConnector()
	mockClient := new(MockBlobClient)
	connector.client = mockClient

	ds := &discovery.DataSource{
		Database: "my-container",
	}

	err := connector.Connect(context.Background(), ds)
	require.NoError(t, err)
	assert.Equal(t, mockClient, connector.client)
	assert.Equal(t, "my-container", connector.container)
}

func TestBlob_SampleData_Mock(t *testing.T) {
	// Since parsing logic is currently stubbed/empty in implementation,
	// we just test that SampleData calls DownloadStream and returns expected (empty) result from stub.
	// If we implemented parsing, we'd test parsing here.

	connector := NewBlobConnector()
	mockClient := new(MockBlobClient)
	connector.client = mockClient
	connector.container = "test-container"

	blobName := "data.csv"
	bodyContent := "name,email\nalice,alice@example.com"
	body := io.NopCloser(strings.NewReader(bodyContent))

	response := azblob.DownloadStreamResponse{}
	response.Body = body

	mockClient.On("DownloadStream", mock.Anything, "test-container", blobName, mock.Anything).Return(response, nil)

	data, err := connector.SampleData(context.Background(), blobName, "email", 10)
	require.NoError(t, err)
	// Currently parsing stub returns empty slice
	assert.Empty(t, data)

	mockClient.AssertExpectations(t)
}
