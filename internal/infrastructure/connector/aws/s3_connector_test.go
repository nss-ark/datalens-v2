package aws

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MockS3Client implements S3ClientInterface for testing.
type MockS3Client struct {
	GetObjectFunc     func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	ListObjectsV2Func func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.GetObjectFunc != nil {
		return m.GetObjectFunc(ctx, params, optFns...)
	}
	return &s3.GetObjectOutput{}, nil
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if m.ListObjectsV2Func != nil {
		return m.ListObjectsV2Func(ctx, params, optFns...)
	}
	return &s3.ListObjectsV2Output{}, nil
}

func TestS3_SampleData_CSV(t *testing.T) {
	csvContent := `id,name,email
1,Alice,alice@example.com
2,Bob,bob@example.com`

	mockClient := &MockS3Client{
		GetObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(strings.NewReader(csvContent)),
			}, nil
		},
	}

	c := NewS3Connector()
	c.client = mockClient
	c.bucket = "test-bucket"

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

func TestS3_SampleData_JSON(t *testing.T) {
	jsonContent := `[
		{"id": 1, "user": {"email": "alice@example.com"}},
		{"id": 2, "user": {"email": "bob@example.com"}}
	]`

	mockClient := &MockS3Client{
		GetObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(strings.NewReader(jsonContent)),
			}, nil
		},
	}

	c := NewS3Connector()
	c.client = mockClient
	c.bucket = "test-bucket"

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
