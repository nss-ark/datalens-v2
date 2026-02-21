package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/pkg/types"
)

// RetentionService provides business logic for retention policy management.
type RetentionService struct {
	repo   compliance.RetentionPolicyRepository
	logger *slog.Logger
}

// NewRetentionService creates a new RetentionService.
func NewRetentionService(repo compliance.RetentionPolicyRepository, logger *slog.Logger) *RetentionService {
	return &RetentionService{
		repo:   repo,
		logger: logger.With("service", "retention"),
	}
}

// CreateRetentionPolicyRequest holds input for creating a retention policy.
type CreateRetentionPolicyRequest struct {
	PurposeID        types.ID `json:"purpose_id"`
	MaxRetentionDays int      `json:"max_retention_days"`
	DataCategories   []string `json:"data_categories"`
	AutoErase        bool     `json:"auto_erase"`
	Description      string   `json:"description"`
}

// UpdateRetentionPolicyRequest holds input for updating a retention policy.
type UpdateRetentionPolicyRequest struct {
	PurposeID        *types.ID `json:"purpose_id,omitempty"`
	MaxRetentionDays *int      `json:"max_retention_days,omitempty"`
	DataCategories   []string  `json:"data_categories,omitempty"`
	Status           *string   `json:"status,omitempty"`
	AutoErase        *bool     `json:"auto_erase,omitempty"`
	Description      *string   `json:"description,omitempty"`
}

// Create creates a new retention policy for the tenant.
func (s *RetentionService) Create(ctx context.Context, req CreateRetentionPolicyRequest) (*compliance.RetentionPolicy, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	if req.MaxRetentionDays <= 0 {
		return nil, types.NewValidationError("max_retention_days must be positive", nil)
	}

	policy := &compliance.RetentionPolicy{
		TenantID:         tenantID,
		PurposeID:        req.PurposeID,
		MaxRetentionDays: req.MaxRetentionDays,
		DataCategories:   req.DataCategories,
		Status:           compliance.RetentionPolicyActive,
		AutoErase:        req.AutoErase,
		Description:      req.Description,
	}

	if err := s.repo.Create(ctx, policy); err != nil {
		return nil, fmt.Errorf("create retention policy: %w", err)
	}

	s.logger.Info("retention policy created",
		slog.String("tenant_id", tenantID.String()),
		slog.String("policy_id", policy.ID.String()),
	)

	return policy, nil
}

// GetByID retrieves a retention policy by ID.
func (s *RetentionService) GetByID(ctx context.Context, id types.ID) (*compliance.RetentionPolicy, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByTenant retrieves all retention policies for a tenant.
func (s *RetentionService) GetByTenant(ctx context.Context) ([]compliance.RetentionPolicy, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.GetByTenant(ctx, tenantID)
}

// Update updates an existing retention policy.
func (s *RetentionService) Update(ctx context.Context, id types.ID, req UpdateRetentionPolicyRequest) (*compliance.RetentionPolicy, error) {
	policy, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.PurposeID != nil {
		policy.PurposeID = *req.PurposeID
	}
	if req.MaxRetentionDays != nil {
		if *req.MaxRetentionDays <= 0 {
			return nil, types.NewValidationError("max_retention_days must be positive", nil)
		}
		policy.MaxRetentionDays = *req.MaxRetentionDays
	}
	if req.DataCategories != nil {
		policy.DataCategories = req.DataCategories
	}
	if req.Status != nil {
		policy.Status = compliance.RetentionPolicyStatus(*req.Status)
	}
	if req.AutoErase != nil {
		policy.AutoErase = *req.AutoErase
	}
	if req.Description != nil {
		policy.Description = *req.Description
	}

	policy.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, policy); err != nil {
		return nil, fmt.Errorf("update retention policy: %w", err)
	}

	s.logger.Info("retention policy updated",
		slog.String("policy_id", id.String()),
	)

	return policy, nil
}

// Delete removes a retention policy by ID.
func (s *RetentionService) Delete(ctx context.Context, id types.ID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.logger.Info("retention policy deleted",
		slog.String("policy_id", id.String()),
	)

	return nil
}

// GetLogs retrieves paginated retention logs, optionally filtered by policy ID.
func (s *RetentionService) GetLogs(ctx context.Context, policyID *types.ID, pagination types.Pagination) (*types.PaginatedResult[compliance.RetentionLog], error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.GetLogs(ctx, tenantID, policyID, pagination)
}
