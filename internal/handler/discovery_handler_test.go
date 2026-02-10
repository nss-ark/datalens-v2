package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
	// Assuming this is where middleware is
)

// Mocks

type MockScanOrchestrator struct {
	mock.Mock
}

func (m *MockScanOrchestrator) EnqueueScan(ctx context.Context, dataSourceID types.ID, tenantID types.ID, scanType discovery.ScanType) (*discovery.ScanRun, error) {
	args := m.Called(ctx, dataSourceID, tenantID, scanType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*discovery.ScanRun), args.Error(1)
}

func (m *MockScanOrchestrator) GetScan(ctx context.Context, id types.ID) (*discovery.ScanRun, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*discovery.ScanRun), args.Error(1)
}

func (m *MockScanOrchestrator) GetHistory(ctx context.Context, dataSourceID types.ID) ([]discovery.ScanRun, error) {
	args := m.Called(ctx, dataSourceID)
	return args.Get(0).([]discovery.ScanRun), args.Error(1)
}

// Minimal Repository Mocks needed for Handler
type MockInventoryRepo struct{ mock.Mock }

func (m *MockInventoryRepo) Create(ctx context.Context, d *discovery.DataInventory) error { return nil }
func (m *MockInventoryRepo) Update(ctx context.Context, d *discovery.DataInventory) error { return nil }
func (m *MockInventoryRepo) GetByDataSource(ctx context.Context, id types.ID) (*discovery.DataInventory, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*discovery.DataInventory), args.Error(1)
}

// Tests

func TestDiscoveryHandler_ScanDataSource(t *testing.T) {
	// Setup
	scanSvc := new(MockScanOrchestrator)
	handler := NewDiscoveryHandler(nil, scanSvc, nil, nil, nil) // other deps not needed for this endpoint

	r := chi.NewRouter()
	r.Post("/data-sources/{sourceID}/scan", handler.ScanDataSource)

	dsID := types.NewID()
	tenantID := types.NewID()
	runID := types.NewID()

	// Mock Enqueue
	expectedRun := &discovery.ScanRun{
		BaseEntity: types.BaseEntity{ID: runID},
		Status:     discovery.ScanStatusPending,
	}
	scanSvc.On("EnqueueScan", mock.Anything, dsID, tenantID, discovery.ScanTypeFull).Return(expectedRun, nil)

	// Request
	req := httptest.NewRequest("POST", "/data-sources/"+dsID.String()+"/scan", nil)
	// Add Tenant Context (Middleware simulation)
	ctx := context.WithValue(req.Context(), types.ContextKeyTenantID, tenantID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Verify
	require.Equal(t, http.StatusAccepted, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, runID.String(), resp["job_id"])
	assert.Equal(t, "Scan queued", resp["message"])

	scanSvc.AssertExpectations(t)
}

func TestDiscoveryHandler_GetScanHistory(t *testing.T) {
	scanSvc := new(MockScanOrchestrator)
	handler := NewDiscoveryHandler(nil, scanSvc, nil, nil, nil)

	r := chi.NewRouter()
	r.Get("/data-sources/{sourceID}/scan/history", handler.GetScanHistory)

	dsID := types.NewID()

	history := []discovery.ScanRun{
		{BaseEntity: types.BaseEntity{ID: types.NewID()}},
	}
	scanSvc.On("GetHistory", mock.Anything, dsID).Return(history, nil)

	req := httptest.NewRequest("GET", "/data-sources/"+dsID.String()+"/scan/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// JSON Array check...
	scanSvc.AssertExpectations(t)
}

// =============================================================================
// Mock DiscoveryOrchestrator for GetClassifications
// =============================================================================

type MockDiscoveryOrchestrator struct {
	mock.Mock
}

func (m *MockDiscoveryOrchestrator) ScanDataSource(ctx context.Context, dataSourceID types.ID) error {
	args := m.Called(ctx, dataSourceID)
	return args.Error(0)
}

func (m *MockDiscoveryOrchestrator) TestConnection(ctx context.Context, dataSourceID types.ID) error {
	args := m.Called(ctx, dataSourceID)
	return args.Error(0)
}

func (m *MockDiscoveryOrchestrator) GetClassifications(ctx context.Context, tenantID types.ID, filter discovery.ClassificationFilter) (*types.PaginatedResult[discovery.PIIClassification], error) {
	args := m.Called(ctx, tenantID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PaginatedResult[discovery.PIIClassification]), args.Error(1)
}

func TestDiscoveryHandler_GetClassifications(t *testing.T) {
	// Setup
	discSvc := new(MockDiscoveryOrchestrator)
	handler := NewDiscoveryHandler(discSvc, nil, nil, nil, nil)

	r := chi.NewRouter()
	r.Get("/classifications", handler.GetClassifications)

	tenantID := types.NewID()

	expectedResult := &types.PaginatedResult[discovery.PIIClassification]{
		Items: []discovery.PIIClassification{
			{Type: "EMAIL", Confidence: 0.95},
			{Type: "PHONE", Confidence: 0.87},
		},
		Total:      2,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}

	// Mock: accept any filter since we're just testing the endpoint wiring
	discSvc.On("GetClassifications", mock.Anything, tenantID, mock.Anything).Return(expectedResult, nil)

	// Request with query params
	req := httptest.NewRequest("GET", "/classifications?status=PENDING&page=1&page_size=20", nil)
	ctx := context.WithValue(req.Context(), types.ContextKeyTenantID, tenantID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Verify
	require.Equal(t, http.StatusOK, w.Code)

	var resp types.PaginatedResult[discovery.PIIClassification]
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Len(t, resp.Items, 2)
	assert.Equal(t, 2, resp.Total)
	assert.Equal(t, types.PIIType("EMAIL"), resp.Items[0].Type)

	discSvc.AssertExpectations(t)
}
