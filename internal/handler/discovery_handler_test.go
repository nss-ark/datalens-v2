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
