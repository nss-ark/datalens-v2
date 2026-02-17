package service

import (
	"context"
	"testing"
	"time"

	"log/slog"

	"github.com/complyark/datalens/internal/domain/audit"
	"github.com/complyark/datalens/internal/domain/breach"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockBreachRepository struct {
	mock.Mock
}

func (m *MockBreachRepository) Create(ctx context.Context, b *breach.BreachIncident) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func (m *MockBreachRepository) GetByID(ctx context.Context, id types.ID) (*breach.BreachIncident, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*breach.BreachIncident), args.Error(1)
}

func (m *MockBreachRepository) Update(ctx context.Context, b *breach.BreachIncident) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func (m *MockBreachRepository) List(ctx context.Context, tenantID types.ID, filter breach.Filter, pagination types.Pagination) (*types.PaginatedResult[breach.BreachIncident], error) {
	args := m.Called(ctx, tenantID, filter, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PaginatedResult[breach.BreachIncident]), args.Error(1)
}

func (m *MockBreachRepository) LogNotification(ctx context.Context, notification *breach.BreachNotification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockBreachRepository) GetNotificationsForPrincipal(ctx context.Context, tenantID types.ID, principalID types.ID, pagination types.Pagination) (*types.PaginatedResult[breach.BreachNotification], error) {
	args := m.Called(ctx, tenantID, principalID, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PaginatedResult[breach.BreachNotification]), args.Error(1)
}

type MockBreachEventBus struct {
	mock.Mock
}

func (m *MockBreachEventBus) Publish(ctx context.Context, event eventbus.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockBreachEventBus) Subscribe(ctx context.Context, pattern string, handler eventbus.EventHandler) (eventbus.Subscription, error) {
	args := m.Called(ctx, pattern, handler)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(eventbus.Subscription), args.Error(1)
}

func (m *MockBreachEventBus) Close() error {
	return nil
}

type MockBreachAuditRepository struct {
	mock.Mock
}

func (m *MockBreachAuditRepository) Create(ctx context.Context, log *audit.AuditLog) error {
	return nil
}

func (m *MockBreachAuditRepository) GetByTenant(ctx context.Context, tenantID types.ID, limit int) ([]audit.AuditLog, error) {
	return nil, nil
}

func (m *MockBreachAuditRepository) ListByTenant(ctx context.Context, tenantID types.ID, filters audit.AuditFilters, pagination types.Pagination) (*types.PaginatedResult[audit.AuditLog], error) {
	return &types.PaginatedResult[audit.AuditLog]{}, nil
}

// Tests

func TestBreachService_CreateIncident(t *testing.T) {
	mockRepo := new(MockBreachRepository)
	mockEventBus := new(MockBreachEventBus)
	mockAuditRepo := new(MockBreachAuditRepository)

	logger := slog.Default()
	auditService := NewAuditService(mockAuditRepo, logger)
	service := NewBreachService(mockRepo, nil, nil, auditService, mockEventBus, logger)

	ctx := context.Background()
	tenantID := types.NewID()
	userID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)
	ctx = context.WithValue(ctx, types.ContextKeyUserID, userID)

	req := CreateIncidentRequest{
		Title:      "Test Breach",
		DetectedAt: time.Now(),
		Severity:   breach.SeverityHigh,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*breach.BreachIncident")).Return(nil)
	mockEventBus.On("Publish", ctx, mock.AnythingOfType("eventbus.Event")).Return(nil)

	incident, err := service.CreateIncident(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, incident)
	assert.Equal(t, req.Title, incident.Title)
	assert.True(t, incident.IsReportableToCertIn) // High severity

	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}
