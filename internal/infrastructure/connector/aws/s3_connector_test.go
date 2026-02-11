package aws

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/discovery"
)

// =============================================================================
// Mocks
// =============================================================================

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

// =============================================================================
// Tests
// =============================================================================

func TestS3_ParseCSV(t *testing.T) {
	// Setup
	connector := NewS3Connector()
	mockClient := new(MockS3Client)
	connector.client = mockClient
	connector.bucket = "test-bucket"

	csvContent := `id,name,email
1,Alice,alice@example.com
2,Bob,bob@example.com`
	body := io.NopCloser(bytes.NewReader([]byte(csvContent)))

	// Mock GetObject
	mockClient.On("GetObject", mock.Anything, mock.MatchedBy(func(input *s3.GetObjectInput) bool {
		return *input.Key == "data.csv"
	})).Return(&s3.GetObjectOutput{Body: body}, nil)

	// Execute
	samples, err := connector.SampleData(context.Background(), "data.csv", "email", 10)

	// Verify
	require.NoError(t, err)
	assert.Len(t, samples, 2)
	assert.Equal(t, "alice@example.com", samples[0])
	assert.Equal(t, "bob@example.com", samples[1])
}

func TestS3_ParseJSON(t *testing.T) {
	// Setup
	connector := NewS3Connector()
	mockClient := new(MockS3Client)
	connector.client = mockClient
	connector.bucket = "test-bucket"

	jsonContent := `[
		{"id": 1, "user": {"email": "alice@example.com"}},
		{"id": 2, "user": {"email": "bob@example.com"}}
	]`
	body := io.NopCloser(bytes.NewReader([]byte(jsonContent)))

	// Mock GetObject
	mockClient.On("GetObject", mock.Anything, mock.MatchedBy(func(input *s3.GetObjectInput) bool {
		return *input.Key == "data.json"
	})).Return(&s3.GetObjectOutput{Body: body}, nil)

	// Execute
	samples, err := connector.SampleData(context.Background(), "data.json", "user.email", 10)

	// Verify
	require.NoError(t, err)
	assert.Len(t, samples, 2)
	assert.Equal(t, "alice@example.com", samples[0])
	assert.Equal(t, "bob@example.com", samples[1])
}

func TestS3_ParseJSONL(t *testing.T) {
	// Setup
	connector := NewS3Connector()
	mockClient := new(MockS3Client)
	connector.client = mockClient
	connector.bucket = "test-bucket"

	jsonlContent := `{"id": 1, "email": "alice@example.com"}
{"id": 2, "email": "bob@example.com"}
`
	body := io.NopCloser(bytes.NewReader([]byte(jsonlContent)))

	// Mock GetObject
	mockClient.On("GetObject", mock.Anything, mock.MatchedBy(func(input *s3.GetObjectInput) bool {
		return *input.Key == "data.jsonl"
	})).Return(&s3.GetObjectOutput{Body: body}, nil)

	// Execute
	samples, err := connector.SampleData(context.Background(), "data.jsonl", "email", 10)

	// Verify
	require.NoError(t, err)
	assert.Len(t, samples, 2)
	assert.Equal(t, "alice@example.com", samples[0])
	assert.Equal(t, "bob@example.com", samples[1])
}

func TestS3_IncrementalScan(t *testing.T) {
	// Setup
	connector := NewS3Connector()
	mockClient := new(MockS3Client)
	connector.client = mockClient
	connector.bucket = "test-bucket"

	now := time.Now()
	oldTime := now.Add(-24 * time.Hour)
	newTime := now

	// Mock ListObjects
	mockClient.On("ListObjectsV2", mock.Anything, mock.Anything).Return(&s3.ListObjectsV2Output{
		Contents: []types.Object{
			{Key: aws.String("old.txt"), LastModified: &oldTime},
			{Key: aws.String("new.txt"), LastModified: &newTime},
		},
		IsTruncated: aws.Bool(false),
	}, nil)

	// Execute
	input := discovery.DiscoveryInput{
		ChangedSince: now.Add(-1 * time.Hour), // 1 hour ago
	}
	_, entities, err := connector.DiscoverSchema(context.Background(), input)

	// Verify
	require.NoError(t, err)
	assert.Len(t, entities, 1)
	assert.Equal(t, "new.txt", entities[0].Name)
}

func TestS3_Connect_ManualClient(t *testing.T) {
	// Verify that we can inject a client (which we are doing in other tests, but good to double check Connect doesn't overwrite it)
	connector := NewS3Connector()
	mockClient := new(MockS3Client)
	connector.client = mockClient

	ds := &discovery.DataSource{
		Database: "my-bucket",
	}

	err := connector.Connect(context.Background(), ds)
	require.NoError(t, err)
	assert.Equal(t, mockClient, connector.client)
	assert.Equal(t, "my-bucket", connector.bucket)
}
