package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/types"
)

// PostgresPolicyRepository implements governance.PolicyRepository.
type PostgresPolicyRepository struct {
	db *pgxpool.Pool
}

// NewPostgresPolicyRepository creates a new PostgresPolicyRepository.
func NewPostgresPolicyRepository(db *pgxpool.Pool) *PostgresPolicyRepository {
	return &PostgresPolicyRepository{db: db}
}

// Create persists a new policy.
func (r *PostgresPolicyRepository) Create(ctx context.Context, p *governance.Policy) error {
	query := `
		INSERT INTO policies (
			id, tenant_id, name, description, type, rules, severity, actions, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.TenantID, p.Name, p.Description, p.Type, p.Rules, p.Severity, p.Actions, p.IsActive, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create policy: %w", err)
	}
	return nil
}

// GetByID retrieves a policy by ID.
func (r *PostgresPolicyRepository) GetByID(ctx context.Context, id types.ID) (*governance.Policy, error) {
	query := `
		SELECT id, tenant_id, name, description, type, rules, severity, actions, is_active, created_at, updated_at
		FROM policies
		WHERE id = $1
	`
	var p governance.Policy
	row := r.db.QueryRow(ctx, query, id)
	if err := row.Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Description, &p.Type, &p.Rules, &p.Severity, &p.Actions, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("policy not found", nil)
		}
		return nil, fmt.Errorf("get policy by id: %w", err)
	}
	return &p, nil
}

// GetActive retrieves all active policies for a tenant.
func (r *PostgresPolicyRepository) GetActive(ctx context.Context, tenantID types.ID) ([]governance.Policy, error) {
	query := `
		SELECT id, tenant_id, name, description, type, rules, severity, actions, is_active, created_at, updated_at
		FROM policies
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get active policies: %w", err)
	}
	defer rows.Close()

	var policies []governance.Policy
	for rows.Next() {
		var p governance.Policy
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.Name, &p.Description, &p.Type, &p.Rules, &p.Severity, &p.Actions, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan policy: %w", err)
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// GetByType retrieves policies by type for a tenant.
func (r *PostgresPolicyRepository) GetByType(ctx context.Context, tenantID types.ID, policyType governance.PolicyType) ([]governance.Policy, error) {
	query := `
		SELECT id, tenant_id, name, description, type, rules, severity, actions, is_active, created_at, updated_at
		FROM policies
		WHERE tenant_id = $1 AND type = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, policyType)
	if err != nil {
		return nil, fmt.Errorf("get policies by type: %w", err)
	}
	defer rows.Close()

	var policies []governance.Policy
	for rows.Next() {
		var p governance.Policy
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.Name, &p.Description, &p.Type, &p.Rules, &p.Severity, &p.Actions, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan policy: %w", err)
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// Update updates an existing policy.
func (r *PostgresPolicyRepository) Update(ctx context.Context, p *governance.Policy) error {
	query := `
		UPDATE policies
		SET name = $1, description = $2, type = $3, rules = $4, severity = $5, actions = $6, is_active = $7, updated_at = $8
		WHERE id = $9
	`
	cmdTag, err := r.db.Exec(ctx, query,
		p.Name, p.Description, p.Type, p.Rules, p.Severity, p.Actions, p.IsActive, p.UpdatedAt, p.ID,
	)
	if err != nil {
		return fmt.Errorf("update policy: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return types.NewNotFoundError("policy not found", nil)
	}
	return nil
}

// Delete removes a policy.
func (r *PostgresPolicyRepository) Delete(ctx context.Context, id types.ID) error {
	query := `DELETE FROM governance_policies WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return types.NewNotFoundError("policy not found", nil)
	}
	return nil
}

// PostgresViolationRepository implements governance.ViolationRepository.
type PostgresViolationRepository struct {
	db *pgxpool.Pool
}

// NewPostgresViolationRepository creates a new PostgresViolationRepository.
func NewPostgresViolationRepository(db *pgxpool.Pool) *PostgresViolationRepository {
	return &PostgresViolationRepository{db: db}
}

// Create persists a new violation.
func (r *PostgresViolationRepository) Create(ctx context.Context, v *governance.Violation) error {
	query := `
		INSERT INTO violations (
			id, tenant_id, policy_id, data_source_id, entity_name, field_name, status, severity, detected_at, resolved_at, resolved_by, resolution
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query,
		v.ID, v.TenantID, v.PolicyID, v.DataSourceID, v.EntityName, v.FieldName, v.Status, v.Severity, v.DetectedAt, v.ResolvedAt, v.ResolvedBy, v.Resolution,
	)
	if err != nil {
		return fmt.Errorf("create violation: %w", err)
	}
	return nil
}

// GetByID retrieves a violation by ID.
func (r *PostgresViolationRepository) GetByID(ctx context.Context, id types.ID) (*governance.Violation, error) {
	query := `
		SELECT id, tenant_id, policy_id, data_source_id, entity_name, field_name, status, severity, detected_at, resolved_at, resolved_by, resolution
		FROM violations
		WHERE id = $1
	`
	var v governance.Violation
	row := r.db.QueryRow(ctx, query, id)
	if err := row.Scan(
		&v.ID, &v.TenantID, &v.PolicyID, &v.DataSourceID, &v.EntityName, &v.FieldName, &v.Status, &v.Severity, &v.DetectedAt, &v.ResolvedAt, &v.ResolvedBy, &v.Resolution,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("violation not found", nil)
		}
		return nil, fmt.Errorf("get violation by id: %w", err)
	}
	return &v, nil
}

// GetByTenant retrieves violations by tenant, optionally filtering by status.
func (r *PostgresViolationRepository) GetByTenant(ctx context.Context, tenantID types.ID, status *governance.ViolationStatus) ([]governance.Violation, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `
			SELECT id, tenant_id, policy_id, data_source_id, entity_name, field_name, status, severity, detected_at, resolved_at, resolved_by, resolution
			FROM violations
			WHERE tenant_id = $1 AND status = $2
			ORDER BY detected_at DESC
		`
		args = []interface{}{tenantID, *status}
	} else {
		query = `
			SELECT id, tenant_id, policy_id, data_source_id, entity_name, field_name, status, severity, detected_at, resolved_at, resolved_by, resolution
			FROM violations
			WHERE tenant_id = $1
			ORDER BY detected_at DESC
		`
		args = []interface{}{tenantID}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get violations by tenant: %w", err)
	}
	defer rows.Close()

	var violations []governance.Violation
	for rows.Next() {
		var v governance.Violation
		if err := rows.Scan(
			&v.ID, &v.TenantID, &v.PolicyID, &v.DataSourceID, &v.EntityName, &v.FieldName, &v.Status, &v.Severity, &v.DetectedAt, &v.ResolvedAt, &v.ResolvedBy, &v.Resolution,
		); err != nil {
			return nil, fmt.Errorf("scan violation: %w", err)
		}
		violations = append(violations, v)
	}
	return violations, nil
}

// GetByPolicy retrieves violations for a specific policy.
func (r *PostgresViolationRepository) GetByPolicy(ctx context.Context, policyID types.ID) ([]governance.Violation, error) {
	query := `
		SELECT id, tenant_id, policy_id, data_source_id, entity_name, field_name, status, severity, detected_at, resolved_at, resolved_by, resolution
		FROM violations
		WHERE policy_id = $1
		ORDER BY detected_at DESC
	`
	rows, err := r.db.Query(ctx, query, policyID)
	if err != nil {
		return nil, fmt.Errorf("get violations by policy: %w", err)
	}
	defer rows.Close()

	var violations []governance.Violation
	for rows.Next() {
		var v governance.Violation
		if err := rows.Scan(
			&v.ID, &v.TenantID, &v.PolicyID, &v.DataSourceID, &v.EntityName, &v.FieldName, &v.Status, &v.Severity, &v.DetectedAt, &v.ResolvedAt, &v.ResolvedBy, &v.Resolution,
		); err != nil {
			return nil, fmt.Errorf("scan violation: %w", err)
		}
		violations = append(violations, v)
	}
	return violations, nil
}

// GetByDataSource retrieves violations for a specific data source.
func (r *PostgresViolationRepository) GetByDataSource(ctx context.Context, dataSourceID types.ID) ([]governance.Violation, error) {
	query := `
		SELECT id, tenant_id, policy_id, data_source_id, entity_name, field_name, status, severity, detected_at, resolved_at, resolved_by, resolution
		FROM violations
		WHERE data_source_id = $1
		ORDER BY detected_at DESC
	`
	rows, err := r.db.Query(ctx, query, dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("get violations by data source: %w", err)
	}
	defer rows.Close()

	var violations []governance.Violation
	for rows.Next() {
		var v governance.Violation
		if err := rows.Scan(
			&v.ID, &v.TenantID, &v.PolicyID, &v.DataSourceID, &v.EntityName, &v.FieldName, &v.Status, &v.Severity, &v.DetectedAt, &v.ResolvedAt, &v.ResolvedBy, &v.Resolution,
		); err != nil {
			return nil, fmt.Errorf("scan violation: %w", err)
		}
		violations = append(violations, v)
	}
	return violations, nil
}

// UpdateStatus updates the status of a violation (e.g., resolving it).
func (r *PostgresViolationRepository) UpdateStatus(ctx context.Context, id types.ID, status governance.ViolationStatus, resolvedBy *types.ID, resolution *string) error {
	query := `
		UPDATE violations
		SET status = $1, resolved_by = $2, resolution = $3, resolved_at = $4
		WHERE id = $5
	`
	var resolvedAt *time.Time
	if status == governance.ViolationStatusResolved {
		now := time.Now()
		resolvedAt = &now
	}

	cmdTag, err := r.db.Exec(ctx, query, status, resolvedBy, resolution, resolvedAt, id)
	if err != nil {
		return fmt.Errorf("update violation status: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return types.NewNotFoundError("violation not found", nil)
	}
	return nil
}

// PostgresDataMappingRepository implements governance.DataMappingRepository.
type PostgresDataMappingRepository struct {
	db *pgxpool.Pool
}

// NewPostgresDataMappingRepository creates a new PostgresDataMappingRepository.
func NewPostgresDataMappingRepository(db *pgxpool.Pool) *PostgresDataMappingRepository {
	return &PostgresDataMappingRepository{db: db}
}

// Create persists a new data mapping.
func (r *PostgresDataMappingRepository) Create(ctx context.Context, dm *governance.DataMapping) error {
	query := `
		INSERT INTO data_mappings (
			id, tenant_id, classification_id, purpose_ids, retention_days, third_party_ids, notes, mapped_by, mapped_at, cross_border, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query,
		dm.ID, dm.TenantID, dm.ClassificationID, dm.PurposeIDs, dm.RetentionDays, dm.ThirdPartyIDs, dm.Notes, dm.MappedBy, dm.MappedAt, dm.CrossBorder, dm.CreatedAt, dm.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create data mapping: %w", err)
	}
	return nil
}

// GetByID retrieves a data mapping by ID.
func (r *PostgresDataMappingRepository) GetByID(ctx context.Context, id types.ID) (*governance.DataMapping, error) {
	query := `
		SELECT id, tenant_id, classification_id, purpose_ids, retention_days, third_party_ids, notes, mapped_by, mapped_at, cross_border, created_at, updated_at
		FROM data_mappings
		WHERE id = $1
	`
	var dm governance.DataMapping
	row := r.db.QueryRow(ctx, query, id)
	if err := row.Scan(
		&dm.ID, &dm.TenantID, &dm.ClassificationID, &dm.PurposeIDs, &dm.RetentionDays, &dm.ThirdPartyIDs, &dm.Notes, &dm.MappedBy, &dm.MappedAt, &dm.CrossBorder, &dm.CreatedAt, &dm.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("data mapping not found", nil)
		}
		return nil, fmt.Errorf("get data mapping by id: %w", err)
	}
	return &dm, nil
}

// GetByClassification retrieves a data mapping for a specific classification.
func (r *PostgresDataMappingRepository) GetByClassification(ctx context.Context, classificationID types.ID) (*governance.DataMapping, error) {
	query := `
		SELECT id, tenant_id, classification_id, purpose_ids, retention_days, third_party_ids, notes, mapped_by, mapped_at, cross_border, created_at, updated_at
		FROM governance_data_mappings
		WHERE classification_id = $1
	`
	var dm governance.DataMapping
	row := r.db.QueryRow(ctx, query, classificationID)
	if err := row.Scan(
		&dm.ID, &dm.TenantID, &dm.ClassificationID, &dm.PurposeIDs, &dm.RetentionDays, &dm.ThirdPartyIDs, &dm.Notes, &dm.MappedBy, &dm.MappedAt, &dm.CrossBorder, &dm.CreatedAt, &dm.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("data mapping not found", nil)
		}
		return nil, fmt.Errorf("get data mapping by classification: %w", err)
	}
	return &dm, nil
}

// GetUnmapped retrieves classification IDs that do not have a corresponding mapping.
func (r *PostgresDataMappingRepository) GetUnmapped(ctx context.Context, tenantID types.ID) ([]types.ID, error) {
	// Query to find confirmed PII classifications that don't have an entry in data_mappings
	query := `
		SELECT c.id
		FROM pii_classifications c
		LEFT JOIN governance_data_mappings m ON c.id = m.classification_id
		WHERE c.tenant_id = $1 
		  AND c.status = 'VERIFIED'
		  AND m.id IS NULL
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get unmapped classifications: %w", err)
	}
	defer rows.Close()

	var ids []types.ID
	for rows.Next() {
		var id types.ID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan unmapped id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Update updates an existing data mapping.
func (r *PostgresDataMappingRepository) Update(ctx context.Context, dm *governance.DataMapping) error {
	query := `
		UPDATE governance_data_mappings
		SET purpose_ids = $1, retention_days = $2, third_party_ids = $3, notes = $4, mapped_by = $5, mapped_at = $6, cross_border = $7, updated_at = $8
		WHERE id = $9
	`
	cmdTag, err := r.db.Exec(ctx, query,
		dm.PurposeIDs, dm.RetentionDays, dm.ThirdPartyIDs, dm.Notes, dm.MappedBy, dm.MappedAt, dm.CrossBorder, dm.UpdatedAt, dm.ID,
	)
	if err != nil {
		return fmt.Errorf("update data mapping: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return types.NewNotFoundError("data mapping not found", nil)
	}
	return nil
}
