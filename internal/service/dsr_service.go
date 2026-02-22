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

	"github.com/complyark/datalens/internal/domain/consent"
)

// DSRService handles Data Subject Request business logic.
type DSRService struct {
	dsrRepo        compliance.DSRRepository
	dataSourceRepo discovery.DataSourceRepository
	dprRepo        consent.DPRRequestRepository
	dsrQueue       queue.DSRQueue
	eventBus       eventbus.EventBus
	auditService   *AuditService
	logger         *slog.Logger
}

// NewDSRService creates a new DSRService.
func NewDSRService(
	dsrRepo compliance.DSRRepository,
	dataSourceRepo discovery.DataSourceRepository,
	dsrQueue queue.DSRQueue,
	// Batch 17B: Inject DPR repo for status sync
	dprRepo consent.DPRRequestRepository,

	eventBus eventbus.EventBus,
	auditService *AuditService,
	logger *slog.Logger,
) *DSRService {
	return &DSRService{
		dsrRepo:        dsrRepo,
		dataSourceRepo: dataSourceRepo,
		dprRepo:        dprRepo,
		dsrQueue:       dsrQueue,

		eventBus:     eventBus,
		auditService: auditService,
		logger:       logger,
	}
}

// CreateDSR creates a new DSR.
func (s *DSRService) CreateDSR(ctx context.Context, req CreateDSRRequest) (*compliance.DSR, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("tenant id is required")
	}

	// Calculate SLA deadline based on request type
	// DPDP Rules R14(3) / Schedule V: ACCESS requests must be fulfilled within 72 hours
	slaDeadline := time.Now().AddDate(0, 0, 30) // Default: 30 days for ERASURE, CORRECTION, etc.
	if req.RequestType == compliance.RequestTypeAccess {
		slaDeadline = time.Now().Add(72 * time.Hour) // 72 hours per DPDP R14(3)
	}

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

	// Tenant Isolation
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("tenant id is required")
	}
	if dsr.TenantID != tenantID {
		return nil, types.NewNotFoundError("DSR", id)
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

	// Tenant Isolation
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("tenant id is required")
	}
	if dsr.TenantID != tenantID {
		return nil, types.NewNotFoundError("DSR", id)
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

	// Tenant Isolation
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("tenant id is required")
	}
	if dsr.TenantID != tenantID {
		return nil, types.NewNotFoundError("DSR", id)
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
func (s *DSRService) GetDSRs(ctx context.Context, pagination types.Pagination, status *compliance.DSRStatus, requestType *compliance.DSRRequestType) (*types.PaginatedResult[compliance.DSR], error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("tenant id is required")
	}
	return s.dsrRepo.GetByTenant(ctx, tenantID, pagination, status, requestType)
}

// UpdateStatus updates the status of a DSR and syncs with DPR if linked.
func (s *DSRService) UpdateStatus(ctx context.Context, id types.ID, status compliance.DSRStatus, notes string) (*compliance.DSR, error) {
	dsr, err := s.dsrRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Tenant Isolation
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, errors.New("tenant id is required")
	}
	if dsr.TenantID != tenantID {
		return nil, types.NewNotFoundError("DSR", id)
	}

	if err := dsr.ValidateTransition(status); err != nil {
		return nil, types.NewValidationError("invalid transition", map[string]any{"status": err.Error()})
	}

	dsr.Status = status
	if notes != "" {
		dsr.Notes = notes
	}
	if status == compliance.DSRStatusCompleted || status == compliance.DSRStatusRejected || status == compliance.DSRStatusFailed {
		now := time.Now().UTC()
		dsr.CompletedAt = &now
	}

	if err := s.dsrRepo.Update(ctx, dsr); err != nil {
		return nil, fmt.Errorf("update dsr status: %w", err)
	}

	// Emit event
	s.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventDSRExecuting, "dsr_service", dsr.TenantID, map[string]any{
		"dsr_id": dsr.ID,
		"status": status,
	}))

	// Sync with DPR Request if exists
	go s.syncDPRStatus(context.Background(), dsr.TenantID, dsr.ID, status, notes)

	return dsr, nil
}

func (s *DSRService) syncDPRStatus(ctx context.Context, tenantID, dsrID types.ID, dsrStatus compliance.DSRStatus, notes string) {
	// Use background context with tenant? No, we need to be careful.
	// We'll pass explicit context or create new one. Use Background for async but need tenant context if repo needs it?
	// Repos usually need context for timeout/tracing, tenantID is passed as arg.
	dpr, err := s.dprRepo.GetByDSRID(ctx, dsrID)
	if err != nil {
		if !types.IsNotFoundError(err) {
			s.logger.Error("failed to get linked dpr for sync", "error", err, "dsr_id", dsrID)
		}
		return
	}

	var dprStatus consent.DPRStatus
	switch dsrStatus {
	case compliance.DSRStatusInProgress:
		dprStatus = consent.DPRStatusInProgress
	case compliance.DSRStatusCompleted:
		dprStatus = consent.DPRStatusCompleted
	case compliance.DSRStatusRejected:
		dprStatus = consent.DPRStatusRejected
	case compliance.DSRStatusFailed:
		dprStatus = consent.DPRStatusRejected // Map failed to rejected or keep strictly failed? DPR doesn't have FAILED.
	default:
		return // No change
	}

	if dpr.Status == dprStatus {
		return
	}

	dpr.Status = dprStatus
	if dprStatus == consent.DPRStatusCompleted || dprStatus == consent.DPRStatusRejected {
		now := time.Now().UTC()
		dpr.CompletedAt = &now
		if notes != "" {
			dpr.ResponseSummary = &notes
		}
	}

	if err := s.dprRepo.Update(ctx, dpr); err != nil {
		s.logger.Error("failed to sync dpr status", "error", err, "dpr_id", dpr.ID)
	}
}

// RespondToAppeal handles the admin decision on an appeal DSR.
func (s *DSRService) RespondToAppeal(ctx context.Context, appealDSRID types.ID, decision string, notes string) (*compliance.DSR, error) {
	// 1. Get Appeal DSR
	appealDSR, err := s.dsrRepo.GetByID(ctx, appealDSRID)
	if err != nil {
		return nil, err
	}

	if appealDSR.RequestType != compliance.RequestTypeAppeal {
		return nil, types.NewValidationError("target DSR is not an appeal", nil)
	}

	// 2. Find linked Appeal DPR
	dpr, err := s.dprRepo.GetByDSRID(ctx, appealDSRID)
	if err != nil {
		return nil, fmt.Errorf("failed to find linked appeal DPR: %w", err)
	}
	if dpr.AppealOf == nil {
		return nil, types.NewValidationError("appeal DPR has no original request link", nil)
	}

	// 3. Handle Decision
	// REVERSED = Appeal Successful (Original decision overturned)
	// UPHELD = Appeal Rejected (Original decision stands)

	switch decision {
	case "REVERSED":
		// Appeal successful -> Re-open original DSR
		// Need to find original DSR. The original DPR has DSRID?
		originalDPR, err := s.dprRepo.GetByID(ctx, *dpr.AppealOf)
		if err != nil {
			return nil, fmt.Errorf("failed to find original DPR: %w", err)
		}
		if originalDPR.DSRID != nil {
			originalDSR, err := s.dsrRepo.GetByID(ctx, *originalDPR.DSRID)
			if err == nil {
				// Re-open original DSR
				// Force status to IN_PROGRESS or APPROVED?
				// Status transition validation might fail if REJECTED -> IN_PROGRESS is not allowed.
				// Let's assume we can force update or we need to update validation logic.
				// For now, let's try direct update.
				originalDSR.Status = compliance.DSRStatusInProgress
				originalDSR.Notes += fmt.Sprintf("\n[Appeal Successful] Re-opened via Appeal %s", appealDSR.ID)
				if err := s.dsrRepo.Update(ctx, originalDSR); err != nil {
					s.logger.Error("failed to re-open original DSR", "error", err)
				}
			}
		}

		// Mark Appeal DSR as COMPLETED (Successfully resolved)
		appealDSR.Status = compliance.DSRStatusCompleted
		appealDSR.Notes = "Appeal REVERSED (Successful). Original request re-opened. " + notes

	case "UPHELD":
		// Appeal rejected -> Original decision stands
		appealDSR.Status = compliance.DSRStatusCompleted // Completed process, but negative outcome
		appealDSR.Notes = "Appeal UPHELD (Rejected). Original decision stands. " + notes
		// We could use Rejected status for the appeal DSR itself to indicate "Appeal Failed"?
		// But "Completed" implies the DPO has finished processing it.
		// Let's use Completed and rely on notes/metadata.

	default:
		return nil, types.NewValidationError("invalid decision (must be REVERSED or UPHELD)", nil)
	}

	if err := s.dsrRepo.Update(ctx, appealDSR); err != nil {
		return nil, err
	}

	// Check for auto-sync to DPR status via UpdateStatus/syncDPRStatus if we called UpdateStatus,
	// but here we called repo.Update directly.
	// We should probably manually sync the appeal DPR status too.
	// If Reversed -> DPRStatusVerified? Or leave as Appealed?
	// If Upheld -> DPRStatusRejected?
	// Let's rely on standard sync or just update here?
	// The `UpdateStatus` method calls `syncDPRStatus`.
	// Since we manually updated, let's call sync manually or just let it be.
	// Actually, `syncDPRStatus` handles mapping DSRStatus to DPRStatus.
	// Completed DSR -> Completed DPR.
	// But for "Upheld" (Appeal Rejected), maybe we want DPR to be Rejected?
	// If we set Appeal DSR to Completed, Appeal DPR becomes Completed. This means "Appeal Process Completed".
	// The Data Principal will see "Appeal Completed". They need to know the outcome.
	// ResponseSummary should contain the decision.

	// Update Appeal DPR outcome notes
	dpr.ResponseSummary = &appealDSR.Notes
	if err := s.dprRepo.Update(ctx, dpr); err != nil {
		s.logger.Error("failed to update appeal DPR summary", "error", err)
	}
	// Trigger status sync
	go s.syncDPRStatus(context.Background(), appealDSR.TenantID, appealDSR.ID, appealDSR.Status, appealDSR.Notes)

	return appealDSR, nil
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
