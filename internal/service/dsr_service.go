package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/queue"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// DSRService handles Data Subject Request business logic.
type DSRService struct {
	dsrRepo        compliance.DSRRepository
	dataSourceRepo discovery.DataSourceRepository
	dsrQueue       queue.DSRQueue
	eventBus       eventbus.EventBus
	logger         *slog.Logger
}

// NewDSRService creates a new DSRService.
func NewDSRService(
	dsrRepo compliance.DSRRepository,
	dataSourceRepo discovery.DataSourceRepository,
	dsrQueue queue.DSRQueue,
	eventBus eventbus.EventBus,
	logger *slog.Logger,
) *DSRService {
	return &DSRService{
		dsrRepo:        dsrRepo,
		dataSourceRepo: dataSourceRepo,
		dsrQueue:       dsrQueue,
		eventBus:       eventBus,
		logger:         logger,
	}
}

// CreateDSR creates a new DSR.
func (s *DSRService) CreateDSR(ctx context.Context, req CreateDSRRequest) (*compliance.DSR, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("tenant id is required")
	}

	// Calculate SLA (default 30 days for GDPR/DPDPA)
	slaDeadline := time.Now().AddDate(0, 0, 30)

	dsr := &compliance.DSR{
		ID:                 types.NewID(),
		TenantID:           tenantID,
		RequestType:        req.RequestType,
		Status:             compliance.DSRStatusPending,
		SubjectName:        req.SubjectName,
		SubjectEmail:       req.SubjectEmail,
		SubjectIdentifiers: req.SubjectIdentifiers,
		Priority:           req.Priority,
		Notes:              req.Notes,
		SLADeadline:        slaDeadline,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	if err := s.dsrRepo.Create(ctx, dsr); err != nil {
		return nil, fmt.Errorf("create dsr: %w", err)
	}

	// Emit event
	event := eventbus.NewEvent(eventbus.EventDSRCreated, "dsr_service", tenantID, map[string]any{
		"dsr_id":       dsr.ID,
		"request_type": dsr.RequestType,
	})
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("failed to publish dsr.created event", "error", err)
		// Don't fail the request, just log
	}

	return dsr, nil
}

// ApproveDSR transitions DSR to APPROVED and decomposes into tasks.
func (s *DSRService) ApproveDSR(ctx context.Context, id types.ID) (*compliance.DSR, error) {
	dsr, err := s.dsrRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := dsr.ValidateTransition(compliance.DSRStatusApproved); err != nil {
		return nil, types.NewValidationError("invalid transition", map[string]any{"status": err.Error()})
	}

	// Create tasks for all data sources (Naive implementation: verify against all)
	// In reality, we might filter by data sources that actually have PII for this subject
	// For now, we create a task for every data source to check.
	dataSources, err := s.dataSourceRepo.GetByTenant(ctx, dsr.TenantID)
	if err != nil {
		return nil, fmt.Errorf("get data sources for decomposition: %w", err)
	}

	for _, ds := range dataSources {
		task := &compliance.DSRTask{
			ID:           types.NewID(),
			DSRID:        dsr.ID,
			DataSourceID: ds.ID,
			TenantID:     dsr.TenantID,
			TaskType:     dsr.RequestType,
			Status:       compliance.TaskStatusPending,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		if err := s.dsrRepo.CreateTask(ctx, task); err != nil {
			return nil, fmt.Errorf("create dsr task: %w", err)
		}
	}

	dsr.Status = compliance.DSRStatusApproved
	// Move to IN_PROGRESS immediately if we have tasks, or strict APPROVED if manual intervention needed
	// The prompt schema says APPROVED -> IN_PROGRESS. Let's keep it APPROVED for now,
	// and assume an async worker or another call picks it up to IN_PROGRESS.
	// Or we can just set it to IN_PROGRESS here if we started tasks.
	// Let's stick to APPROVED as the explicit action result.

	if err := s.dsrRepo.Update(ctx, dsr); err != nil {
		return nil, fmt.Errorf("update dsr status: %w", err)
	}

	// Emit event
	s.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventDSRExecuting, "dsr_service", dsr.TenantID, map[string]any{
		"dsr_id": dsr.ID,
	}))

	// Queue for execution
	if err := s.dsrQueue.Enqueue(ctx, dsr.ID.String()); err != nil {
		s.logger.Error("failed to enqueue dsr for execution",
			slog.String("tenant_id", dsr.TenantID.String()),
			slog.String("dsr_id", dsr.ID.String()),
			slog.String("error", err.Error()),
		) // Don't fail approval, execution can be triggered manually
	}

	return dsr, nil
}

// RejectDSR reject a DSR.
func (s *DSRService) RejectDSR(ctx context.Context, id types.ID, reason string) (*compliance.DSR, error) {
	dsr, err := s.dsrRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := dsr.ValidateTransition(compliance.DSRStatusRejected); err != nil {
		return nil, types.NewValidationError("invalid transition", map[string]any{"status": err.Error()})
	}

	dsr.Status = compliance.DSRStatusRejected
	dsr.Reason = reason
	completedAt := time.Now().UTC()
	dsr.CompletedAt = &completedAt

	if err := s.dsrRepo.Update(ctx, dsr); err != nil {
		return nil, fmt.Errorf("update dsr status: %w", err)
	}

	// Emit event
	s.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventDSRRejected, "dsr_service", dsr.TenantID, map[string]any{
		"dsr_id": dsr.ID,
		"reason": reason,
	}))

	return dsr, nil
}

// GetDSR retrieves a DSR with its tasks.
func (s *DSRService) GetDSR(ctx context.Context, id types.ID) (*DSRWithTasks, error) {
	dsr, err := s.dsrRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tasks, err := s.dsrRepo.GetTasksByDSR(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get dsr tasks: %w", err)
	}

	return &DSRWithTasks{
		DSR:   dsr,
		Tasks: tasks,
	}, nil
}

// GetDSRs lists DSRs.
func (s *DSRService) GetDSRs(ctx context.Context, pagination types.Pagination, status *compliance.DSRStatus) (*types.PaginatedResult[compliance.DSR], error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("tenant id is required")
	}
	return s.dsrRepo.GetByTenant(ctx, tenantID, pagination, status)
}

// GetOverdue returns DSRs that have passed their SLA deadline.
func (s *DSRService) GetOverdue(ctx context.Context) ([]compliance.DSR, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("tenant id is required")
	}
	return s.dsrRepo.GetOverdue(ctx, tenantID)
}

// DTOs

type CreateDSRRequest struct {
	RequestType        compliance.DSRRequestType `json:"request_type"`
	SubjectName        string                    `json:"subject_name"`
	SubjectEmail       string                    `json:"subject_email"`
	SubjectIdentifiers map[string]string         `json:"subject_identifiers"`
	Priority           string                    `json:"priority"`
	Notes              string                    `json:"notes"`
}

type DSRWithTasks struct {
	*compliance.DSR
	Tasks []compliance.DSRTask `json:"tasks"`
}
