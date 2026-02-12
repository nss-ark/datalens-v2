package azure

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector/shared"
)

// BlobClientInterface defines the subset of Blob operations we need.
type BlobClientInterface interface {
	NewListBlobsFlatPager(containerName string, options *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse]
	DownloadStream(ctx context.Context, containerName string, blobName string, options *azblob.DownloadStreamOptions) (io.ReadCloser, error)
}

// azBlobClientAdapter adapts the official Generic Client to our interface
type azBlobClientAdapter struct {
	client *azblob.Client
}

func (a *azBlobClientAdapter) NewListBlobsFlatPager(containerName string, options *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse] {
	return a.client.NewListBlobsFlatPager(containerName, options)
}

func (a *azBlobClientAdapter) DownloadStream(ctx context.Context, containerName string, blobName string, options *azblob.DownloadStreamOptions) (io.ReadCloser, error) {
	resp, err := a.client.DownloadStream(ctx, containerName, blobName, options)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// BlobConnector implements the Connector interface for Azure Blob Storage.
type BlobConnector struct {
	client    BlobClientInterface
	url       string
	logger    *slog.Logger
	container string
}

// NewBlobConnector creates a new Azure Blob connector.
func NewBlobConnector() *BlobConnector {
	return &BlobConnector{
		logger: slog.Default().With("connector", "azure_blob"),
	}
}

// Connect establishes a connection to Azure Blob Storage.
func (c *BlobConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	c.container = ds.Database

	if c.client != nil {
		return nil
	}

	// ds.Credentials should contain connection string or components as JSON
	creds, err := shared.ParseCredentials(ds.Credentials)
	if err != nil {
		return fmt.Errorf("parse credentials: %w", err)
	}

	// Connection string takes precedence
	if connStr, ok := creds["connection_string"].(string); ok && connStr != "" {
		client, err := azblob.NewClientFromConnectionString(connStr, nil)
		if err != nil {
			return fmt.Errorf("create client from connection string: %w", err)
		}
		c.client = &azBlobClientAdapter{client: client}
		return nil
	}

	// Try to build from account/key
	accountName, ok1 := creds["account_name"].(string)
	accountKey, ok2 := creds["account_key"].(string)
	if ok1 && ok2 {
		cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			return fmt.Errorf("invalid credentials: %w", err)
		}
		url := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
		client, err := azblob.NewClientWithSharedKeyCredential(url, cred, nil)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}
		c.client = &azBlobClientAdapter{client: client}
		c.url = url
		return nil
	}

	return fmt.Errorf("missing connection string or account/key")
}

// DiscoverSchema lists blobs in the container.
func (c *BlobConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.client == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	if c.container == "" {
		return nil, nil, fmt.Errorf("container name not specified in data source")
	}

	pager := c.client.NewListBlobsFlatPager(c.container, &azblob.ListBlobsFlatOptions{})

	var entities []discovery.DataEntity
	count := 0

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list blobs: %w", err)
		}

		for _, blob := range resp.Segment.BlobItems {
			// Check ChangedSince
			if !input.ChangedSince.IsZero() && blob.Properties.LastModified != nil && blob.Properties.LastModified.Before(input.ChangedSince) {
				continue
			}

			name := *blob.Name
			entity := discovery.DataEntity{
				Name: name,
				Type: discovery.EntityTypeFile, // Blobs are files
			}
			entities = append(entities, entity)
			count++
			if count >= 1000 {
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

func (c *BlobConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	return []discovery.DataField{}, nil
}

// SampleData reads the blob and extracts values.
func (c *BlobConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	if c.container == "" {
		// Use default container if not set or expect error?
		// S3 reuses database name. Here we do too.
		return nil, fmt.Errorf("container name not specified")
	}

	// Download blob
	body, err := c.client.DownloadStream(ctx, c.container, entity, nil)
	if err != nil {
		return nil, fmt.Errorf("download blob: %w", err)
	}
	defer body.Close()

	if strings.HasSuffix(strings.ToLower(entity), ".csv") {
		return parseCSV(body, field, limit)
	} else if strings.HasSuffix(strings.ToLower(entity), ".json") {
		return parseJSON(body, field, limit)
	}

	return nil, fmt.Errorf("unsupported file type: %s", entity)
}

// Helper functions (implemented similarly to S3)

func parseCSV(r io.Reader, field string, limit int) ([]string, error) {
	reader := csv.NewReader(r)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	colIdx := -1
	for i, h := range headers {
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

func parseJSON(r io.Reader, field string, limit int) ([]string, error) {
	decoder := json.NewDecoder(r)

	// Read opening bracket '['
	t, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("expected JSON array")
	}

	var samples []string
	for decoder.More() && len(samples) < limit {
		var obj map[string]interface{}
		if err := decoder.Decode(&obj); err != nil {
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

func (c *BlobConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:         true,
		CanSample:           true,
		SupportsIncremental: true,
	}
}

// Delete is a stub for Azure Blob.
func (c *BlobConnector) Delete(ctx context.Context, entity string, filter map[string]string) (int64, error) {
	return 0, fmt.Errorf("delete not supported for azure blob")
}

// Export is a stub for Azure Blob.
func (c *BlobConnector) Export(ctx context.Context, entity string, filter map[string]string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("export not supported for azure blob")
}

func (c *BlobConnector) Close() error {
	return nil
}
