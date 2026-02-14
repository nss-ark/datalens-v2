package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/service/ai"
	"github.com/complyark/datalens/internal/service/detection"
)

// FileUploadConfig holds configuration for file-upload-based data sources.
type FileUploadConfig struct {
	FilePath     string `json:"file_path"`
	MimeType     string `json:"mime_type"`
	OriginalName string `json:"original_name"`
}

// fileUploadConnector implements discovery.Connector and discovery.ScannableConnector
// for FILE_UPLOAD data sources.
type fileUploadConnector struct {
	parser   ai.ParsingService
	detector *detection.ComposableDetector
	ds       *discovery.DataSource
	config   *FileUploadConfig
}

// Compile-time checks
var _ discovery.Connector = (*fileUploadConnector)(nil)
var _ discovery.ScannableConnector = (*fileUploadConnector)(nil)

// NewFileUploadConnectorFactory returns a ConnectorFactory for file upload connectors.
func NewFileUploadConnectorFactory(parser ai.ParsingService, detector *detection.ComposableDetector) ConnectorFactory {
	return func() discovery.Connector {
		return &fileUploadConnector{
			parser:   parser,
			detector: detector,
		}
	}
}

// Capabilities returns what this connector supports.
func (c *fileUploadConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:             true,
		CanSample:               true,
		CanDelete:               false,
		CanUpdate:               false,
		CanExport:               false,
		SupportsStreaming:       false,
		SupportsIncremental:     false,
		SupportsSchemaDiscovery: true,
		SupportsDataSampling:    true,
		SupportsParallelScan:    false,
		MaxConcurrency:          1,
	}
}

// Connect validates configuration and prepares the connector.
func (c *fileUploadConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	var cfg FileUploadConfig
	if err := json.Unmarshal([]byte(ds.Config), &cfg); err != nil {
		return fmt.Errorf("invalid file upload config: %w", err)
	}
	if cfg.FilePath == "" {
		return fmt.Errorf("file_path is required in config")
	}
	c.ds = ds
	c.config = &cfg
	return nil
}

// DiscoverSchema returns a minimal schema representing the file content.
func (c *fileUploadConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.config == nil {
		return nil, nil, fmt.Errorf("connector not connected")
	}

	name := c.config.OriginalName
	if name == "" {
		name = "uploaded_file"
	}

	inventory := &discovery.DataInventory{
		TotalEntities: 1,
		TotalFields:   1,
		SchemaVersion: "1.0",
	}

	entities := []discovery.DataEntity{
		{
			Name: name,
			Type: discovery.EntityTypeFile,
		},
	}

	return inventory, entities, nil
}

// GetFields returns the fields for a given entity.
func (c *fileUploadConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	return []discovery.DataField{
		{Name: "content", DataType: "text"},
	}, nil
}

// SampleData returns sample data from the file.
func (c *fileUploadConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	if c.config == nil {
		return nil, fmt.Errorf("connector not connected")
	}

	text, err := c.parser.Parse(ctx, c.config.FilePath, c.config.MimeType)
	if err != nil {
		return nil, fmt.Errorf("parse file for samples: %w", err)
	}

	if text == "" {
		return []string{}, nil
	}

	// Return a truncated sample
	runes := []rune(text)
	if len(runes) > 1000 {
		runes = runes[:1000]
	}
	return []string{string(runes)}, nil
}

// Delete is not supported for file uploads.
func (c *fileUploadConnector) Delete(ctx context.Context, entity string, filter map[string]string) (int64, error) {
	return 0, fmt.Errorf("delete not supported for file uploads")
}

// Export is not supported for file uploads.
func (c *fileUploadConnector) Export(ctx context.Context, entity string, filter map[string]string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("export not supported for file uploads")
}

// Close releases resources.
func (c *fileUploadConnector) Close() error {
	return nil
}

// Scan extracts text from the uploaded file and runs PII detection on it.
func (c *fileUploadConnector) Scan(ctx context.Context, ds *discovery.DataSource, onFinding func(discovery.PIIClassification)) error {
	// 1. Parse config
	var cfg FileUploadConfig
	if err := json.Unmarshal([]byte(ds.Config), &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// 2. Extract text
	text, err := c.parser.Parse(ctx, cfg.FilePath, cfg.MimeType)
	if err != nil {
		return fmt.Errorf("parse file: %w", err)
	}

	if text == "" {
		return nil // Nothing to scan
	}

	entityName := cfg.OriginalName
	if entityName == "" {
		entityName = "uploaded_file"
	}

	// 3. Scan content in chunks
	chunkSize := 4000
	runes := []rune(text)

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunk := string(runes[i:end])

		input := detection.Input{
			TableName:  entityName,
			ColumnName: "content",
			DataType:   "text",
			Samples:    []string{chunk},
		}

		report, err := c.detector.Detect(ctx, input)
		if err != nil {
			slog.Warn("detection error on chunk", "error", err, "chunk_start", i)
			continue
		}

		if report == nil || !report.IsPII {
			continue
		}

		// Report top match as finding
		if report.TopMatch != nil {
			onFinding(discovery.PIIClassification{
				EntityName:      entityName,
				FieldName:       "content",
				Category:        report.TopMatch.Category,
				Type:            report.TopMatch.Type,
				Sensitivity:     report.TopMatch.Sensitivity,
				Confidence:      report.TopMatch.FinalConfidence,
				DetectionMethod: report.TopMatch.Methods[0],
				Reasoning:       report.TopMatch.Reasoning,
			})
		}
	}

	return nil
}
