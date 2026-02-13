package handler_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/handler"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementation must act "real enough" for the service logic
type mockDSRRepo struct {
	dsrs map[types.ID]*compliance.DSR
}

func (r *mockDSRRepo) Create(ctx context.Context, dsr *compliance.DSR) error {
	if r.dsrs == nil {
		r.dsrs = make(map[types.ID]*compliance.DSR)
	}
	r.dsrs[dsr.ID] = dsr
	return nil
}

func (r *mockDSRRepo) GetAll(ctx context.Context, pagination types.Pagination, statusFilter *compliance.DSRStatus, typeFilter *compliance.DSRRequestType) (*types.PaginatedResult[compliance.DSR], error) {
	var items []compliance.DSR
	for _, d := range r.dsrs {
		items = append(items, *d)
	}
	return &types.PaginatedResult[compliance.DSR]{Items: items, Total: len(items)}, nil
}

// ... other mocks ...
func (r *mockDSRRepo) GetByID(ctx context.Context, id types.ID) (*compliance.DSR, error) {
	return nil, nil
}
func (r *mockDSRRepo) GetByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination, statusFilter *compliance.DSRStatus) (*types.PaginatedResult[compliance.DSR], error) {
	return nil, nil
}
func (r *mockDSRRepo) GetOverdue(ctx context.Context, tenantID types.ID) ([]compliance.DSR, error) {
	return nil, nil
}
func (r *mockDSRRepo) Update(ctx context.Context, dsr *compliance.DSR) error          { return nil }
func (r *mockDSRRepo) CreateTask(ctx context.Context, task *compliance.DSRTask) error { return nil }
func (r *mockDSRRepo) GetTasksByDSR(ctx context.Context, dsrID types.ID) ([]compliance.DSRTask, error) {
	return nil, nil
}
func (r *mockDSRRepo) UpdateTask(ctx context.Context, task *compliance.DSRTask) error { return nil }

type mockTenantRepo struct{ identity.TenantRepository }

func (m *mockTenantRepo) GetStats(ctx context.Context) (*identity.TenantStats, error) {
	return &identity.TenantStats{}, nil
}

// We need Search for ListTenants if used, but ListDSRs doesn't use it.

type mockUserRepo struct{ identity.UserRepository }

func (m *mockUserRepo) CountGlobal(ctx context.Context) (int64, error) { return 0, nil }

type mockRoleRepo struct{ identity.RoleRepository }

func TestBatch19_AdminHandler_ListDSRs(t *testing.T) {
	dsrRepo := &mockDSRRepo{}

	dsrRepo.Create(context.Background(), &compliance.DSR{ID: types.NewID(), Status: compliance.DSRStatusPending})
	dsrRepo.Create(context.Background(), &compliance.DSR{ID: types.NewID(), Status: compliance.DSRStatusApproved})

	adminSvc := service.NewAdminService(
		&mockTenantRepo{},
		&mockUserRepo{},
		&mockRoleRepo{},
		dsrRepo,
		nil,
		slog.Default(),
	)

	h := handler.NewAdminHandler(adminSvc)

	req := httptest.NewRequest(http.MethodGet, "/admin/dsr?page=1&limit=10", nil)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/admin/dsr", h.ListDSRs)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp httputil.Response
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)

	// Optional: verify count
	dataBytes, _ := json.Marshal(resp.Data)
	var result types.PaginatedResult[compliance.DSR]
	_ = json.Unmarshal(dataBytes, &result)
	assert.Equal(t, 2, result.Total)
}
