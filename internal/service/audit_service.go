package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/audit"
	"github.com/complyark/datalens/pkg/types"
)

// AuditService handles the creation and management of audit logs.
type AuditService struct {
	repo   audit.Repository
	logger *slog.Logger
}

// NewAuditService creates a new AuditService.
func NewAuditService(repo audit.Repository, logger *slog.Logger) *AuditService {
	return &AuditService{
		repo:   repo,
		logger: logger.With("service", "audit"),
	}
}

// Log records an audit entry asynchronously.
// It uses a detached context to ensure the log is written even if the parent request context is cancelled.
// However, for simplicity in this MVP, we will use a goroutine and a background context.
//
// Parameters:
//   - ctx: Context (used to extract tenant_id if not explicitly passed, though for now we pass explicit IDs)
//   - actorID: ID of the user performing the action
//   - action: unique string identifier for the action (e.g., "LOGIN", "DSR_APPROVE")
//   - resourceType: type of resource being acted upon (e.g., "USER", "DSR")
//   - resourceID: ID of the resource
//   - changes: map of changes (optional)
//   - tenantID: tenant ID
func (s *AuditService) Log(
	ctx context.Context,
	userID types.ID,
	action string,
	resourceType string,
	resourceID types.ID,
	oldValues map[string]any,
	newValues map[string]any,
	tenantID types.ID,
) {
	// Extract IP and UserAgent from context if available
	ip, _ := ctx.Value(types.ContextKeyIP).(string)
	ua, _ := ctx.Value(types.ContextKeyUserAgent).(string)

	// Launch async logging
	go func() {
		// Use a new context with timeout for the DB operation
		logCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logEntry := &audit.AuditLog{
			ID:           types.NewID(),
			TenantID:     tenantID,
			UserID:       userID,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			OldValues:    oldValues,
			NewValues:    newValues,
			IPAddress:    ip,
			UserAgent:    ua,
			CreatedAt:    time.Now().UTC(),
		}

		if err := s.repo.Create(logCtx, logEntry); err != nil {
			s.logger.Error("failed to create audit log",
				slog.String("error", err.Error()),
				slog.String("action", action),
				slog.String("tenant_id", tenantID.String()),
			)
		}
	}()
}
