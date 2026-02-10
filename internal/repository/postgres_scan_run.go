package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// PostgresScanRunRepo implements discovery.ScanRunRepository
type PostgresScanRunRepo struct {
	pool *pgxpool.Pool
}

// NewScanRunRepo creates a new PostgresScanRunRepo
func NewScanRunRepo(pool *pgxpool.Pool) *PostgresScanRunRepo {
	return &PostgresScanRunRepo{pool: pool}
}

// Create persists a new scan run
func (r *PostgresScanRunRepo) Create(ctx context.Context, run *discovery.ScanRun) error {
	query := `
		INSERT INTO scan_runs (
			id, data_source_id, tenant_id, type, status, progress, 
			started_at, completed_at, error_message, 
			entities_scanned, fields_scanned, pii_detected, duration_ms, bytes_processed,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, 
			$7, $8, $9, 
			$10, $11, $12, $13, $14,
			NOW(), NOW()
		)`

	// Handle optional timestamps
	var startedAt, completedAt *time.Time
	if run.StartedAt != nil {
		startedAt = run.StartedAt
	}
	if run.CompletedAt != nil {
		completedAt = run.CompletedAt
	}

	_, err := r.pool.Exec(ctx, query,
		run.ID, run.DataSourceID, run.TenantID, run.Type, run.Status, run.Progress,
		startedAt, completedAt, run.ErrorMessage,
		run.Stats.EntitiesScanned, run.Stats.FieldsScanned, run.Stats.PIIDetected, run.Stats.Duration.Milliseconds(), run.Stats.BytesProcessed,
	)
	if err != nil {
		return fmt.Errorf("create scan run: %w", err)
	}

	return nil
}

// GetByID retrieves a scan run by ID
func (r *PostgresScanRunRepo) GetByID(ctx context.Context, id types.ID) (*discovery.ScanRun, error) {
	query := `
		SELECT 
			id, data_source_id, tenant_id, type, status, progress, 
			started_at, completed_at, error_message, 
			entities_scanned, fields_scanned, pii_detected, duration_ms, bytes_processed
		FROM scan_runs
		WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)
	return scanRowToScanRun(row)
}

// GetByDataSource retrieves all scan runs for a data source
func (r *PostgresScanRunRepo) GetByDataSource(ctx context.Context, dataSourceID types.ID) ([]discovery.ScanRun, error) {
	query := `
		SELECT 
			id, data_source_id, tenant_id, type, status, progress, 
			started_at, completed_at, error_message, 
			entities_scanned, fields_scanned, pii_detected, duration_ms, bytes_processed
		FROM scan_runs
		WHERE data_source_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("query scan runs: %w", err)
	}
	defer rows.Close()

	var runs []discovery.ScanRun
	for rows.Next() {
		run, err := scanRowToScanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *run)
	}
	return runs, nil
}

// GetActive retrieves all non-terminal scan runs for a tenant
func (r *PostgresScanRunRepo) GetActive(ctx context.Context, tenantID types.ID) ([]discovery.ScanRun, error) {
	query := `
		SELECT 
			id, data_source_id, tenant_id, type, status, progress, 
			started_at, completed_at, error_message, 
			entities_scanned, fields_scanned, pii_detected, duration_ms, bytes_processed
		FROM scan_runs
		WHERE tenant_id = $1 AND status IN ('PENDING', 'RUNNING')
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query active scan runs: %w", err)
	}
	defer rows.Close()

	var runs []discovery.ScanRun
	for rows.Next() {
		run, err := scanRowToScanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *run)
	}
	return runs, nil
}

// GetRecent retrieves a limited number of recent scan runs for a tenant
func (r *PostgresScanRunRepo) GetRecent(ctx context.Context, tenantID types.ID, limit int) ([]discovery.ScanRun, error) {
	query := `
		SELECT 
			id, data_source_id, tenant_id, type, status, progress, 
			started_at, completed_at, error_message, 
			entities_scanned, fields_scanned, pii_detected, duration_ms, bytes_processed
		FROM scan_runs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent scans: %w", err)
	}
	defer rows.Close()

	var runs []discovery.ScanRun
	for rows.Next() {
		run, err := scanRowToScanRun(rows)
		if err != nil {
			return nil, fmt.Errorf("scan run row: %w", err)
		}
		runs = append(runs, *run)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return runs, nil
}

// Update persists changes to an existing scan run
func (r *PostgresScanRunRepo) Update(ctx context.Context, run *discovery.ScanRun) error {
	query := `
		UPDATE scan_runs
		SET 
			status = $1, progress = $2, started_at = $3, completed_at = $4, error_message = $5,
			entities_scanned = $6, fields_scanned = $7, pii_detected = $8, duration_ms = $9, bytes_processed = $10,
			updated_at = NOW()
		WHERE id = $11`

	var startedAt, completedAt *time.Time
	if run.StartedAt != nil {
		startedAt = run.StartedAt
	}
	if run.CompletedAt != nil {
		completedAt = run.CompletedAt
	}

	cmd, err := r.pool.Exec(ctx, query,
		run.Status, run.Progress, startedAt, completedAt, run.ErrorMessage,
		run.Stats.EntitiesScanned, run.Stats.FieldsScanned, run.Stats.PIIDetected, run.Stats.Duration.Milliseconds(), run.Stats.BytesProcessed,
		run.ID,
	)
	if err != nil {
		return fmt.Errorf("update scan run: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return types.NewNotFoundError("scan run", run.ID)
	}

	return nil
}

// Helper to map row to ScanRun
func scanRowToScanRun(row pgx.Row) (*discovery.ScanRun, error) {
	var run discovery.ScanRun
	var durationMs int64
	var startedAt, completedAt *time.Time

	err := row.Scan(
		&run.ID, &run.DataSourceID, &run.TenantID, &run.Type, &run.Status, &run.Progress,
		&startedAt, &completedAt, &run.ErrorMessage,
		&run.Stats.EntitiesScanned, &run.Stats.FieldsScanned, &run.Stats.PIIDetected, &durationMs, &run.Stats.BytesProcessed,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("scan run", "")
		}
		return nil, fmt.Errorf("scan row: %w", err)
	}

	run.StartedAt = startedAt
	run.CompletedAt = completedAt
	run.Stats.Duration = time.Duration(durationMs) * time.Millisecond

	return &run, nil
}
