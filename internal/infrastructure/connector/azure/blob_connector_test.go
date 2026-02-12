package azure

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// MockBlobClient implements BlobClientInterface for testing.
type MockBlobClient struct {
	DownloadStreamFunc        func(ctx context.Context, containerName string, blobName string, options *azblob.DownloadStreamOptions) (io.ReadCloser, error)
	NewListBlobsFlatPagerFunc func(containerName string, options *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse]
}

func (m *MockBlobClient) DownloadStream(ctx context.Context, containerName string, blobName string, options *azblob.DownloadStreamOptions) (io.ReadCloser, error) {
	if m.DownloadStreamFunc != nil {
		return m.DownloadStreamFunc(ctx, containerName, blobName, options)
	}
	return nil, nil // Should verify error handling if nil
}

func (m *MockBlobClient) NewListBlobsFlatPager(containerName string, options *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse] {
	if m.NewListBlobsFlatPagerFunc != nil {
		return m.NewListBlobsFlatPagerFunc(containerName, options)
	}
	return nil
}

func TestBlob_SampleData_CSV(t *testing.T) {
	csvContent := `id,name,email
1,Alice,alice@example.com
2,Bob,bob@example.com`

	mockClient := &MockBlobClient{
		DownloadStreamFunc: func(ctx context.Context, container string, blob string, options *azblob.DownloadStreamOptions) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(csvContent)), nil
		},
	}

	c := NewBlobConnector()
	c.client = mockClient
	c.container = "test-container"

	samples, err := c.SampleData(context.Background(), "data.csv", "email", 10)
	if err != nil {
		t.Fatalf("SampleData failed: %v", err)
	}

	if len(samples) != 2 {
		t.Errorf("Expected 2 samples, got %d", len(samples))
	}
	if samples[0] != "alice@example.com" {
		t.Errorf("Expected first sample 'alice@example.com', got '%s'", samples[0])
	}
}

func TestBlob_SampleData_JSON(t *testing.T) {
	jsonContent := `[
		{"id": 1, "user": {"email": "alice@example.com"}},
		{"id": 2, "user": {"email": "bob@example.com"}}
	]`

	mockClient := &MockBlobClient{
		DownloadStreamFunc: func(ctx context.Context, container string, blob string, options *azblob.DownloadStreamOptions) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(jsonContent)), nil
		},
	}

	c := NewBlobConnector()
	c.client = mockClient
	c.container = "test-container"

	samples, err := c.SampleData(context.Background(), "data.json", "user.email", 10)
	if err != nil {
		t.Fatalf("SampleData failed: %v", err)
	}

	if len(samples) != 2 {
		t.Errorf("Expected 2 samples, got %d", len(samples))
	}
	if samples[0] != "alice@example.com" {
		t.Errorf("Expected first sample 'alice@example.com', got '%s'", samples[0])
	}
}
