package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// PIIClassificationRepo implements discovery.PIIClassificationRepository.
type PIIClassificationRepo struct {
	pool *pgxpool.Pool
}

// NewPIIClassificationRepo creates a new PIIClassificationRepo.
func NewPIIClassificationRepo(pool *pgxpool.Pool) *PIIClassificationRepo {
	return &PIIClassificationRepo{pool: pool}
}

func (r *PIIClassificationRepo) Create(ctx context.Context, c *discovery.PIIClassification) error {
	c.ID = types.NewID()
	query := `
		INSERT INTO pii_classifications (id, field_id, data_source_id, entity_name, field_name,
		    category, type, sensitivity, confidence, detection_method, status, reasoning)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		c.ID, c.FieldID, c.DataSourceID, c.EntityName, c.FieldName,
		c.Category, c.Type, c.Sensitivity, c.Confidence, c.DetectionMethod, c.Status, c.Reasoning,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
}

func (r *PIIClassificationRepo) GetByID(ctx context.Context, id types.ID) (*discovery.PIIClassification, error) {
	query := `
		SELECT id, field_id, data_source_id, entity_name, field_name, category, type,
		       sensitivity, confidence, detection_method, status, verified_by, verified_at,
		       reasoning, created_at, updated_at
		FROM pii_classifications
		WHERE id = $1`

	c := &discovery.PIIClassification{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.FieldID, &c.DataSourceID, &c.EntityName, &c.FieldName, &c.Category, &c.Type,
		&c.Sensitivity, &c.Confidence, &c.DetectionMethod, &c.Status, &c.VerifiedBy, &c.VerifiedAt,
		&c.Reasoning, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("PIIClassification", id)
		}
		return nil, fmt.Errorf("get pii classification: %w", err)
	}
	return c, nil
}

func (r *PIIClassificationRepo) GetByDataSource(ctx context.Context, dataSourceID types.ID, pagination types.Pagination) (*types.PaginatedResult[discovery.PIIClassification], error) {
	countQuery := `SELECT COUNT(*) FROM pii_classifications WHERE data_source_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, dataSourceID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count pii classifications: %w", err)
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query := `
		SELECT id, field_id, data_source_id, entity_name, field_name, category, type,
		       sensitivity, confidence, detection_method, status, verified_by, verified_at,
		       reasoning, created_at, updated_at
		FROM pii_classifications
		WHERE data_source_id = $1
		ORDER BY confidence DESC, created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, dataSourceID, pagination.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list pii classifications: %w", err)
	}
	defer rows.Close()

	var items []discovery.PIIClassification
	for rows.Next() {
		var c discovery.PIIClassification
		if err := rows.Scan(
			&c.ID, &c.FieldID, &c.DataSourceID, &c.EntityName, &c.FieldName, &c.Category, &c.Type,
			&c.Sensitivity, &c.Confidence, &c.DetectionMethod, &c.Status, &c.VerifiedBy, &c.VerifiedAt,
			&c.Reasoning, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pii classification: %w", err)
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := total / pagination.PageSize
	if total%pagination.PageSize > 0 {
		totalPages++
	}

	return &types.PaginatedResult[discovery.PIIClassification]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *PIIClassificationRepo) GetPending(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[discovery.PIIClassification], error) {
	countQuery := `
		SELECT COUNT(*) FROM pii_classifications pc
		JOIN data_sources ds ON ds.id = pc.data_source_id
		WHERE ds.tenant_id = $1 AND pc.status = 'PENDING'`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count pending classifications: %w", err)
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query := `
		SELECT pc.id, pc.field_id, pc.data_source_id, pc.entity_name, pc.field_name, pc.category,
		       pc.type, pc.sensitivity, pc.confidence, pc.detection_method, pc.status,
		       pc.verified_by, pc.verified_at, pc.reasoning, pc.created_at, pc.updated_at
		FROM pii_classifications pc
		JOIN data_sources ds ON ds.id = pc.data_source_id
		WHERE ds.tenant_id = $1 AND pc.status = 'PENDING'
		ORDER BY pc.confidence DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, tenantID, pagination.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list pending classifications: %w", err)
	}
	defer rows.Close()

	var items []discovery.PIIClassification
	for rows.Next() {
		var c discovery.PIIClassification
		if err := rows.Scan(
			&c.ID, &c.FieldID, &c.DataSourceID, &c.EntityName, &c.FieldName, &c.Category, &c.Type,
			&c.Sensitivity, &c.Confidence, &c.DetectionMethod, &c.Status, &c.VerifiedBy, &c.VerifiedAt,
			&c.Reasoning, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pending classification: %w", err)
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := total / pagination.PageSize
	if total%pagination.PageSize > 0 {
		totalPages++
	}

	return &types.PaginatedResult[discovery.PIIClassification]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *PIIClassificationRepo) GetClassifications(ctx context.Context, tenantID types.ID, filter discovery.ClassificationFilter) (*types.PaginatedResult[discovery.PIIClassification], error) {
	// Build base query
	baseQuery := `
		FROM pii_classifications pc
		JOIN data_sources ds ON ds.id = pc.data_source_id
		WHERE ds.tenant_id = $1`

	args := []interface{}{tenantID}
	argIdx := 2

	// Helper to add condition
	addCondition := func(condition string, val interface{}) {
		baseQuery += fmt.Sprintf(" AND %s = $%d", condition, argIdx)
		args = append(args, val)
		argIdx++
	}

	if filter.DataSourceID != nil {
		addCondition("pc.data_source_id", *filter.DataSourceID)
	}
	if filter.Status != nil {
		addCondition("pc.status", *filter.Status)
	}
	if filter.DetectionMethod != nil {
		addCondition("pc.detection_method", *filter.DetectionMethod)
	}

	// Count total
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count classifications: %w", err)
	}

	// Fetch items
	offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
	query := `
		SELECT pc.id, pc.field_id, pc.data_source_id, pc.entity_name, pc.field_name, pc.category,
			   pc.type, pc.sensitivity, pc.confidence, pc.detection_method, pc.status,
			   pc.verified_by, pc.verified_at, pc.reasoning, pc.created_at, pc.updated_at ` +
		baseQuery +
		fmt.Sprintf(" ORDER BY pc.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)

	args = append(args, filter.Pagination.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list classifications: %w", err)
	}
	defer rows.Close()

	var items []discovery.PIIClassification
	for rows.Next() {
		var c discovery.PIIClassification
		if err := rows.Scan(
			&c.ID, &c.FieldID, &c.DataSourceID, &c.EntityName, &c.FieldName, &c.Category, &c.Type,
			&c.Sensitivity, &c.Confidence, &c.DetectionMethod, &c.Status, &c.VerifiedBy, &c.VerifiedAt,
			&c.Reasoning, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan classification: %w", err)
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := total / filter.Pagination.PageSize
	if total%filter.Pagination.PageSize > 0 {
		totalPages++
	}

	return &types.PaginatedResult[discovery.PIIClassification]{
		Items:      items,
		Total:      total,
		Page:       filter.Pagination.Page,
		PageSize:   filter.Pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *PIIClassificationRepo) GetCounts(ctx context.Context, tenantID types.ID) (*discovery.PIICounts, error) {
	// 1. Total PII for Tenant
	// Join with data_sources to filter by tenant_id
	totalQuery := `
		SELECT COUNT(*) 
		FROM pii_classifications pc
		JOIN data_sources ds ON ds.id = pc.data_source_id
		WHERE ds.tenant_id = $1`

	counts := &discovery.PIICounts{
		ByCategory: make(map[string]int),
	}

	if err := r.pool.QueryRow(ctx, totalQuery, tenantID).Scan(&counts.Total); err != nil {
		return nil, fmt.Errorf("count total pii: %w", err)
	}

	// 2. Group by Category
	categoryQuery := `
		SELECT pc.category, COUNT(*)
		FROM pii_classifications pc
		JOIN data_sources ds ON ds.id = pc.data_source_id
		WHERE ds.tenant_id = $1
		GROUP BY pc.category`

	rows, err := r.pool.Query(ctx, categoryQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("count pii by category: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("scan category count: %w", err)
		}
		counts.ByCategory[category] = count
	}

	return counts, nil
}

func (r *PIIClassificationRepo) Update(ctx context.Context, c *discovery.PIIClassification) error {
	query := `
		UPDATE pii_classifications
		SET status = $2, verified_by = $3, verified_at = $4, reasoning = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query, c.ID, c.Status, c.VerifiedBy, c.VerifiedAt, c.Reasoning).
		Scan(&c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.NewNotFoundError("PIIClassification", c.ID)
		}
		return fmt.Errorf("update pii classification: %w", err)
	}
	return nil
}

func (r *PIIClassificationRepo) BulkCreate(ctx context.Context, classifications []discovery.PIIClassification) error {
	if len(classifications) == 0 {
		return nil
	}

	// Build batch insert
	var sb strings.Builder
	sb.WriteString(`INSERT INTO pii_classifications (id, field_id, data_source_id, entity_name, field_name,
		category, type, sensitivity, confidence, detection_method, status, reasoning) VALUES `)

	args := make([]any, 0, len(classifications)*12)
	for i, c := range classifications {
		if i > 0 {
			sb.WriteString(", ")
		}
		base := i * 12
		sb.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6,
			base+7, base+8, base+9, base+10, base+11, base+12))

		id := types.NewID()
		classifications[i].ID = id
		args = append(args, id, c.FieldID, c.DataSourceID, c.EntityName, c.FieldName,
			c.Category, c.Type, c.Sensitivity, c.Confidence, c.DetectionMethod, c.Status, c.Reasoning)
	}

	_, err := r.pool.Exec(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("bulk create pii classifications: %w", err)
	}
	return nil
}

// Compile-time check.
var _ discovery.PIIClassificationRepository = (*PIIClassificationRepo)(nil)
