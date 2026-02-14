package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector"
	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/types"
)

// MockParsingService
type LocalMockParsingService struct {
	mock.Mock
}

func (m *LocalMockParsingService) Parse(ctx context.Context, filePath string, mimeType string) (string, error) {
	args := m.Called(ctx, filePath, mimeType)
	return args.String(0), args.Error(1)
}

func (m *LocalMockParsingService) IsOCRAvailable() bool {
	return true
}

// FileUploadMockScanQueue
type FileUploadMockScanQueue struct {
	Handler func(context.Context, string) error
}

func (m *FileUploadMockScanQueue) Enqueue(ctx context.Context, jobID string) error {
	// Execute immediately for testing (simplification)
	if m.Handler != nil {
		go m.Handler(ctx, jobID)
	}
	return nil
}

func (m *FileUploadMockScanQueue) Subscribe(ctx context.Context, handler func(context.Context, string) error) error {
	m.Handler = handler
	return nil
}

func (m *FileUploadMockScanQueue) Close() error { return nil }

// FileUploadMockStrategy
type FileUploadMockStrategy struct {
	mock.Mock
}

func (m *FileUploadMockStrategy) Name() string                  { return "FileUploadMockStrategy" }
func (m *FileUploadMockStrategy) Method() types.DetectionMethod { return types.DetectionMethodAI }
func (m *FileUploadMockStrategy) Weight() float64               { return 1.0 }
func (m *FileUploadMockStrategy) Detect(ctx context.Context, input detection.Input) ([]detection.Result, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]detection.Result), args.Error(1)
}

func TestScanning_FileUpload_Flow(t *testing.T) {
	// 1. Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()
	tenantID := types.NewID()

	// Mocks
	dsRepo := newMockDataSourceRepo()
	invRepo := newMockDataInventoryRepo()
	entityRepo := newMockDataEntityRepo()
	fieldRepo := newMockDataFieldRepo()
	piiRepo := newMockPIIClassificationRepo()
	scanRunRepo := newMockScanRunRepo()
	eb := newMockEventBus()
	parsingSvc := new(LocalMockParsingService)

	// Mock Detector
	mockStrategy := new(FileUploadMockStrategy)

	// Expect detection on the extracted text
	// Relaxed matcher for debugging
	mockStrategy.On("Detect", mock.Anything, mock.Anything).Return([]detection.Result{
		{
			Category:    types.PIICategoryContact,
			Type:        types.PIITypeEmail,
			Sensitivity: types.SensitivityHigh,
			Confidence:  0.99,
			Method:      types.DetectionMethodAI,
			Reasoning:   "AI detected email pattern",
		},
	}, nil)

	detector := detection.NewComposableDetector(mockStrategy)

	// Registry
	registry := connector.NewConnectorRegistry(&config.Config{}, detector, parsingSvc)

	// Services
	discoverySvc := NewDiscoveryService(dsRepo, invRepo, entityRepo, fieldRepo, piiRepo, scanRunRepo, registry, detector, eb, logger)

	// Mock Queue (execute immediately)
	mockQueue := &FileUploadMockScanQueue{
		Handler: nil,
	}

	scanSvc := NewScanService(scanRunRepo, dsRepo, mockQueue, discoverySvc, logger)

	// Wrap handler to wait for completion
	done := make(chan struct{})
	mockQueue.Handler = func(ctx context.Context, jobID string) error {
		defer close(done)
		return scanSvc.ProcessScanJob(ctx, jobID)
	}

	// 2. Create Data Source (FILE_UPLOAD)
	fileConfig := connector.FileUploadConfig{
		FilePath:     "/tmp/uploads/123-resume.pdf",
		OriginalName: "resume.pdf",
		MimeType:     "application/pdf",
	}
	configBytes, _ := json.Marshal(fileConfig)

	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID()},
			TenantID:   tenantID,
		},
		Name:   "Uploaded Resume",
		Type:   types.DataSourceFileUpload,
		Config: string(configBytes),
	}
	require.NoError(t, dsRepo.Create(ctx, ds))

	// 3. Mock Parser Behavior
	parsingSvc.On("Parse", mock.Anything, "/tmp/uploads/123-resume.pdf", "application/pdf").
		Return("John Doe email@example.com", nil)

	// 4. Enqueue Scan
	run, err := scanSvc.EnqueueScan(ctx, ds.ID, tenantID, discovery.ScanTypeFull)
	require.NoError(t, err)

	// Wait for async execution
	select {
	case <-done:
		// Completed
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for scan job")
	}

	// 6. Verify
	updatedRun, _ := scanRunRepo.GetByID(ctx, run.ID)
	if assert.NotNil(t, updatedRun) {
		assert.Equal(t, discovery.ScanStatusCompleted, updatedRun.Status)
	}

	pii, err := piiRepo.GetByDataSource(ctx, ds.ID, types.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	require.Len(t, pii.Items, 1)
	if len(pii.Items) > 0 {
		assert.Equal(t, "resume.pdf", pii.Items[0].EntityName)
		assert.Equal(t, types.PIITypeEmail, pii.Items[0].Type)
	}

	parsingSvc.AssertExpectations(t)
	mockStrategy.AssertExpectations(t)
}
