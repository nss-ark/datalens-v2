package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/handler"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/types"
)

// MockDSRRepo for Admin tests
type MockDSRRepo struct {
	mock.Mock
}

func (m *MockDSRRepo) Create(ctx context.Context, dsr *compliance.DSR) error {
	args := m.Called(ctx, dsr)
	return args.Error(0)
}

func (m *MockDSRRepo) GetByID(ctx context.Context, id types.ID) (*compliance.DSR, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*compliance.DSR), args.Error(1)
}

func (m *MockDSRRepo) GetByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination, statusFilter *compliance.DSRStatus) (*types.PaginatedResult[compliance.DSR], error) {
	args := m.Called(ctx, tenantID, pagination, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PaginatedResult[compliance.DSR]), args.Error(1)
}

func (m *MockDSRRepo) GetAll(ctx context.Context, pagination types.Pagination, statusFilter *compliance.DSRStatus, typeFilter *compliance.DSRRequestType) (*types.PaginatedResult[compliance.DSR], error) {
	args := m.Called(ctx, pagination, statusFilter, typeFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PaginatedResult[compliance.DSR]), args.Error(1)
}

func (m *MockDSRRepo) GetOverdue(ctx context.Context, tenantID types.ID) ([]compliance.DSR, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]compliance.DSR), args.Error(1)
}

func (m *MockDSRRepo) Update(ctx context.Context, dsr *compliance.DSR) error {
	return m.Called(ctx, dsr).Error(0)
}

func (m *MockDSRRepo) CreateTask(ctx context.Context, task *compliance.DSRTask) error {
	return m.Called(ctx, task).Error(0)
}

func (m *MockDSRRepo) GetTasksByDSR(ctx context.Context, dsrID types.ID) ([]compliance.DSRTask, error) {
	args := m.Called(ctx, dsrID)
	return args.Get(0).([]compliance.DSRTask), args.Error(1)
}

func (m *MockDSRRepo) UpdateTask(ctx context.Context, task *compliance.DSRTask) error {
	return m.Called(ctx, task).Error(0)
}

// Helper to create AdminHandler with mocks
func setupAdminHandler(t *testing.T) (*handler.AdminHandler, *MockDSRRepo) {
	dsrRepo := new(MockDSRRepo)
	// We need to mock other repos too if AdminService uses them, but for DSRs specifically
	// we only need DSRRepo. AdminService constructor requires others, so we might need nil or mocks.
	// Since we are unit testing the handler -> service -> repo flow for DSRs, we can pass nil for others
	// IF the service methods we call don't use them.
	// AdminService.GetAllDSRs uses dsrRepo.GetAll.
	// AdminService.GetDSR uses dsrRepo.GetByID.
	// So passing nil for others should be safe for THESE tests.

	// Logger is needed
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// TenantService is tricky, it's a struct pointer. We might need a real one or mock if we touch it.
	// But for DSR methods we don't touch tenantSvc.

	adminSvc := service.NewAdminService(nil, nil, nil, dsrRepo, nil, nil, nil, nil, nil, logger)
	h := handler.NewAdminHandler(adminSvc, nil)
	return h, dsrRepo
}

func TestAdminHandler_ListDSRs(t *testing.T) {
	h, repo := setupAdminHandler(t)

	t.Run("success", func(t *testing.T) {
		dsrs := []compliance.DSR{
			{ID: types.NewID(), TenantID: types.NewID(), SubjectEmail: "test@example.com"},
		}
		result := &types.PaginatedResult[compliance.DSR]{
			Items: dsrs,
			Total: 1,
		}

		repo.On("GetAll", mock.Anything, mock.MatchedBy(func(p types.Pagination) bool {
			return p.Page == 1 && p.PageSize == 20
		}), (*compliance.DSRStatus)(nil), (*compliance.DSRRequestType)(nil)).Return(result, nil)

		req := httptest.NewRequest("GET", "/admin/dsr?page=1&page_size=20", nil)
		w := httptest.NewRecorder()

		h.ListDSRs(w, req)

		var resp struct {
			Data types.PaginatedResult[compliance.DSR] `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp.Data.Total)
		assert.Equal(t, dsrs[0].ID, resp.Data.Items[0].ID)
	})
}

func TestAdminHandler_GetDSR(t *testing.T) {
	h, repo := setupAdminHandler(t)

	t.Run("success", func(t *testing.T) {
		id := types.NewID()
		dsr := &compliance.DSR{ID: id, TenantID: types.NewID(), SubjectEmail: "test@example.com"}

		repo.On("GetByID", mock.Anything, id).Return(dsr, nil)

		r := chi.NewRouter()
		r.Get("/admin/dsr/{id}", h.GetDSR)

		req := httptest.NewRequest("GET", "/admin/dsr/"+id.String(), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Data compliance.DSR `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, id, resp.Data.ID)
	})

	t.Run("not found", func(t *testing.T) {
		id := types.NewID()
		repo.On("GetByID", mock.Anything, id).Return(nil, types.NewNotFoundError("DSR", id))

		r := chi.NewRouter()
		r.Get("/admin/dsr/{id}", h.GetDSR)

		req := httptest.NewRequest("GET", "/admin/dsr/"+id.String(), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
