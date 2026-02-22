package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/types"
)

// PurposeAssignmentService provides business logic for purpose scope assignments.
type PurposeAssignmentService struct {
	repo        governance.PurposeAssignmentRepository
	purposeRepo governance.PurposeRepository
	auditSvc    *AuditService
	logger      *slog.Logger
}

// NewPurposeAssignmentService creates a new PurposeAssignmentService.
func NewPurposeAssignmentService(
	repo governance.PurposeAssignmentRepository,
	purposeRepo governance.PurposeRepository,
	auditSvc *AuditService,
	logger *slog.Logger,
) *PurposeAssignmentService {
	return &PurposeAssignmentService{
		repo:        repo,
		purposeRepo: purposeRepo,
		auditSvc:    auditSvc,
		logger:      logger.With("service", "purpose_assignment"),
	}
}

// AssignPurposeInput holds input for assigning a purpose at a scope level.
type AssignPurposeInput struct {
	PurposeID types.ID             `json:"purpose_id"`
	ScopeType governance.ScopeType `json:"scope_type"`
	ScopeID   string               `json:"scope_id"`
	ScopeName string               `json:"scope_name"`
}

// Assign creates a purpose assignment at a specific scope level.
func (s *PurposeAssignmentService) Assign(ctx context.Context, input AssignPurposeInput) (*governance.PurposeAssignment, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	userID, _ := types.UserIDFromContext(ctx)

	// Validate scope type
	if !isValidScopeType(input.ScopeType) {
		return nil, types.NewValidationError("invalid scope_type", map[string]any{
			"scope_type": input.ScopeType,
			"valid":      governance.ValidScopeTypes,
		})
	}

	if input.ScopeID == "" {
		return nil, types.NewValidationError("scope_id is required", nil)
	}

	assignment := &governance.PurposeAssignment{
		TenantID:   tenantID,
		PurposeID:  input.PurposeID,
		ScopeType:  input.ScopeType,
		ScopeID:    input.ScopeID,
		ScopeName:  input.ScopeName,
		Inherited:  false,
		AssignedBy: &userID,
	}

	if err := s.repo.Create(ctx, assignment); err != nil {
		return nil, fmt.Errorf("create purpose assignment: %w", err)
	}

	// Audit log
	s.auditSvc.Log(ctx, userID, "PURPOSE_ASSIGN", "PURPOSE_ASSIGNMENT", assignment.ID, nil,
		map[string]any{"purpose_id": input.PurposeID.String(), "scope_type": string(input.ScopeType), "scope_id": input.ScopeID}, tenantID)

	s.logger.Info("purpose assigned",
		slog.String("tenant_id", tenantID.String()),
		slog.String("purpose_id", input.PurposeID.String()),
		slog.String("scope_type", string(input.ScopeType)),
		slog.String("scope_id", input.ScopeID),
	)

	return assignment, nil
}

// Remove deletes a purpose assignment.
func (s *PurposeAssignmentService) Remove(ctx context.Context, id types.ID) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}
	userID, _ := types.UserIDFromContext(ctx)

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit log
	s.auditSvc.Log(ctx, userID, "PURPOSE_UNASSIGN", "PURPOSE_ASSIGNMENT", id, nil, nil, tenantID)

	s.logger.Info("purpose unassigned",
		slog.String("tenant_id", tenantID.String()),
		slog.String("assignment_id", id.String()),
	)

	return nil
}

// GetByScope retrieves direct assignments at a specific scope level.
func (s *PurposeAssignmentService) GetByScope(ctx context.Context, scopeType governance.ScopeType, scopeID string) ([]governance.PurposeAssignment, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.GetByScope(ctx, tenantID, scopeType, scopeID)
}

// GetEffective resolves effective purpose assignments at a scope level,
// including inherited assignments from parent scopes.
//
// Scope ID convention:
//   - SERVER:   server_name                              e.g. "prod-db-01"
//   - DATABASE: db_name                                  e.g. "customers_db"
//   - SCHEMA:   db_name.schema_name                      e.g. "customers_db.public"
//   - TABLE:    db_name.schema_name.table_name            e.g. "customers_db.public.users"
//   - COLUMN:   db_name.schema_name.table_name.col_name   e.g. "customers_db.public.users.email"
func (s *PurposeAssignmentService) GetEffective(ctx context.Context, scopeType governance.ScopeType, scopeID string) ([]governance.PurposeAssignment, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	// Build list of scopes to query (from current level up to SERVER)
	scopeLevels := buildScopeHierarchy(scopeType, scopeID)

	// Collect all assignments at each level
	purposeMap := make(map[types.ID]governance.PurposeAssignment) // purposeID → best assignment
	purposeLevel := make(map[types.ID]int)                        // purposeID → scope level

	for _, sl := range scopeLevels {
		assignments, err := s.repo.GetByScope(ctx, tenantID, sl.scopeType, sl.scopeID)
		if err != nil {
			return nil, fmt.Errorf("get scope %s/%s: %w", sl.scopeType, sl.scopeID, err)
		}

		level := governance.ScopeHierarchy[sl.scopeType]

		for _, a := range assignments {
			existingLevel, exists := purposeLevel[a.PurposeID]
			if !exists || level > existingLevel {
				// Lower-level (more specific) assignments take precedence
				a.Inherited = sl.scopeType != scopeType || sl.scopeID != scopeID
				purposeMap[a.PurposeID] = a
				purposeLevel[a.PurposeID] = level
			}
		}
	}

	// Flatten map to slice
	var result []governance.PurposeAssignment
	for _, a := range purposeMap {
		result = append(result, a)
	}

	return result, nil
}

// GetByPurpose retrieves all assignments for a specific purpose.
func (s *PurposeAssignmentService) GetByPurpose(ctx context.Context, purposeID types.ID) ([]governance.PurposeAssignment, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.GetByPurpose(ctx, tenantID, purposeID)
}

// GetAll retrieves all purpose assignments for the tenant.
func (s *PurposeAssignmentService) GetAll(ctx context.Context) ([]governance.PurposeAssignment, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.GetByTenant(ctx, tenantID)
}

// scopeLevel represents a scope type and ID pair.
type scopeLevel struct {
	scopeType governance.ScopeType
	scopeID   string
}

// buildScopeHierarchy builds the list of parent scopes for a given scope.
// Convention: scope_id segments are dot-separated.
func buildScopeHierarchy(scopeType governance.ScopeType, scopeID string) []scopeLevel {
	levels := []scopeLevel{{scopeType: scopeType, scopeID: scopeID}}
	parts := strings.Split(scopeID, ".")

	switch scopeType {
	case governance.ScopeTypeColumn:
		// COLUMN: db.schema.table.column → TABLE=db.schema.table, SCHEMA=db.schema, DATABASE=db, SERVER=*
		if len(parts) >= 4 {
			levels = append(levels, scopeLevel{governance.ScopeTypeTable, strings.Join(parts[:3], ".")})
		}
		if len(parts) >= 3 {
			levels = append(levels, scopeLevel{governance.ScopeTypeSchema, strings.Join(parts[:2], ".")})
		}
		if len(parts) >= 2 {
			levels = append(levels, scopeLevel{governance.ScopeTypeDatabase, parts[0]})
		}
		if len(parts) >= 1 {
			levels = append(levels, scopeLevel{governance.ScopeTypeServer, "*"})
		}

	case governance.ScopeTypeTable:
		// TABLE: db.schema.table → SCHEMA=db.schema, DATABASE=db, SERVER=*
		if len(parts) >= 3 {
			levels = append(levels, scopeLevel{governance.ScopeTypeSchema, strings.Join(parts[:2], ".")})
		}
		if len(parts) >= 2 {
			levels = append(levels, scopeLevel{governance.ScopeTypeDatabase, parts[0]})
		}
		if len(parts) >= 1 {
			levels = append(levels, scopeLevel{governance.ScopeTypeServer, "*"})
		}

	case governance.ScopeTypeSchema:
		// SCHEMA: db.schema → DATABASE=db, SERVER=*
		if len(parts) >= 2 {
			levels = append(levels, scopeLevel{governance.ScopeTypeDatabase, parts[0]})
		}
		if len(parts) >= 1 {
			levels = append(levels, scopeLevel{governance.ScopeTypeServer, "*"})
		}

	case governance.ScopeTypeDatabase:
		// DATABASE: db → SERVER=*
		levels = append(levels, scopeLevel{governance.ScopeTypeServer, "*"})

	case governance.ScopeTypeServer:
		// SERVER: no parents
	}

	return levels
}

// isValidScopeType checks if the given scope type is valid.
func isValidScopeType(st governance.ScopeType) bool {
	for _, v := range governance.ValidScopeTypes {
		if v == st {
			return true
		}
	}
	return false
}
