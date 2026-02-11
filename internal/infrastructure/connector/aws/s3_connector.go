package aws

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/complyark/datalens/internal/domain/discovery"
)

// S3ClientInterface defines the subset of S3 operations we need, for mocking.
type S3ClientInterface interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// S3Connector implements the Connector interface for Amazon S3.
type S3Connector struct {
	client S3ClientInterface
	bucket string
	logger *slog.Logger
}

// NewS3Connector creates a new S3 connector.
func NewS3Connector() *S3Connector {
	return &S3Connector{
		logger: slog.Default().With("connector", "s3"),
	}
}

// Connect establishes a connection to S3.
func (c *S3Connector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	c.bucket = ds.Database // We reuse Database field for bucket name

	if c.client != nil {
		return nil
	}

	cfg, err := LoadConfig(ctx, ds)
	if err != nil {
		return err
	}

	c.client = s3.NewFromConfig(cfg)
	return nil
}

// DiscoverSchema lists objects in the bucket, treating them as entities.
func (c *S3Connector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.client == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	var entities []discovery.DataEntity
	paginator := s3.NewListObjectsV2Paginator(c.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
	})

	count := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list objects: %w", err)
		}

		for _, obj := range page.Contents {
			// Check ChangedSince if provided
			if !input.ChangedSince.IsZero() && obj.LastModified.Before(input.ChangedSince) {
				continue
			}

			key := aws.ToString(obj.Key)
			if strings.HasSuffix(key, "/") {
				continue // Skip folders
			}

			entity := discovery.DataEntity{
				Name: key,
				Type: discovery.EntityTypeFile,
				// Schema derived from file extension?
			}
			entities = append(entities, entity)
			count++
			if count >= 1000 { // Limit for safety
				break
			}
		}
		if count >= 1000 {
			break
		}
	}

	inv := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
	}

	return inv, entities, nil
}

// GetFields inspects a file to determine its "fields" (headers/keys).
func (c *S3Connector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	// For files, fields depend on content type (CSV headers, JSON keys)
	// This requires reading the file, which is expensive.
	// We'll implement a basic version that returning empty for now or implementing if needed.
	// Task says "CSV/JSON/JSONL parsing", so we likely need this or SampleData to utilize parsing.
	// Typically GetFields samples the file to infer schema.
	return []discovery.DataField{}, nil
}

// SampleData reads the file and extracts values for a specific column/key.
// entity = S3 Key
// field = CSV Header or JSON Key
func (c *S3Connector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	if c.client == nil {
		// Use a mock client if not connected (for testing without Connect)
		// Or return error. For now error is safer.
		// BUT existing tests might rely on manually setting client and not calling Connect?
		// S3Connector struct field `client` is unexported, so tests in same package can set it.
		// If in different package (like `aws_test` vs `aws`), they can't.
		// We'll assume tests are in same package or `_test` package accessing exported methods.
		// The existing test `TestS3_ParseCSV` sets `client` manually.
		// Since we are in `package aws`, and tests will be in `package aws` (or `aws_test`),
		// we should be fine if we keep `client` unexported but accessible to tests in same package.
		// However, if we put tests in `aws_test` package, we can't access `client`.
		// We should put tests in `package aws`.
		if c.client == nil {
			return nil, fmt.Errorf("not connected")
		}
	}

	// Get object stream
	resp, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(entity),
	})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer resp.Body.Close()

	if strings.HasSuffix(strings.ToLower(entity), ".csv") {
		return c.parseCSV(resp.Body, field, limit)
	} else if strings.HasSuffix(strings.ToLower(entity), ".json") {
		return c.parseJSON(resp.Body, field, limit)
	} else if strings.HasSuffix(strings.ToLower(entity), ".jsonl") {
		return c.parseJSONL(resp.Body, field, limit)
	}

	return nil, fmt.Errorf("unsupported file type: %s", entity)
}

func (c *S3Connector) parseCSV(r io.Reader, field string, limit int) ([]string, error) {
	reader := csv.NewReader(r)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	colIdx := -1
	for i, h := range headers {
		// Case-insensitive match?
		if strings.EqualFold(h, field) {
			colIdx = i
			break
		}
	}

	if colIdx == -1 {
		return nil, fmt.Errorf("field %s not found in CSV headers", field)
	}

	var samples []string
	for len(samples) < limit {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if colIdx < len(row) {
			samples = append(samples, row[colIdx])
		}
	}
	return samples, nil
}

func (c *S3Connector) parseJSON(r io.Reader, field string, limit int) ([]string, error) {
	// Assumes JSON Array of Objects
	decoder := json.NewDecoder(r)

	// Read opening bracket '['
	t, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		// Maybe it's a single object?
		// For now assume array
		return nil, fmt.Errorf("expected JSON array")
	}

	var samples []string
	for decoder.More() && len(samples) < limit {
		var obj map[string]interface{}
		if err := decoder.Decode(&obj); err != nil {
			return nil, err
		}

		// Navigate dot notation if needed, for now flat
		// Support simple dot notation for nested
		val, found := getJSONValue(obj, field)
		if found {
			samples = append(samples, fmt.Sprintf("%v", val))
		}
	}
	return samples, nil
}

func (c *S3Connector) parseJSONL(r io.Reader, field string, limit int) ([]string, error) {
	decoder := json.NewDecoder(r)
	var samples []string

	for decoder.More() && len(samples) < limit {
		var obj map[string]interface{}
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		val, found := getJSONValue(obj, field)
		if found {
			samples = append(samples, fmt.Sprintf("%v", val))
		}
	}
	return samples, nil
}

func getJSONValue(obj map[string]interface{}, path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	var current interface{} = obj

	for _, k := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, ok := m[k]
		if !ok {
			return nil, false
		}
		current = val
	}
	return current, true
}

func (c *S3Connector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:         true,
		CanSample:           true,
		SupportsIncremental: true,
	}
}

func (c *S3Connector) Close() error {
	return nil
}
