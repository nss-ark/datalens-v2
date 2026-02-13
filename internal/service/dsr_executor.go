package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// DSRExecutor orchestrates the execution of DSR tasks across data sources.
type DSRExecutor struct {
	dsrRepo        compliance.DSRRepository
	dsRepo         discovery.DataSourceRepository
	piiRepo        discovery.PIIClassificationRepository
	connRegistry   *connector.ConnectorRegistry
	eventBus       eventbus.EventBus
	logger         *slog.Logger
	maxConcurrency int
}

// NewDSRExecutor creates a new DSRExecutor.
func NewDSRExecutor(
	dsrRepo compliance.DSRRepository,
	dsRepo discovery.DataSourceRepository,
	piiRepo discovery.PIIClassificationRepository,
	connRegistry *connector.ConnectorRegistry,
	eventBus eventbus.EventBus,
	logger *slog.Logger,
) *DSRExecutor {
	return &DSRExecutor{
		dsrRepo:        dsrRepo,
		dsRepo:         dsRepo,
		piiRepo:        piiRepo,
		connRegistry:   connRegistry,
		eventBus:       eventBus,
		logger:         logger.With("service", "dsr_executor"),
		maxConcurrency: 5, // Default concurrency limit
	}
}

// ExecuteDSR executes all tasks for a DSR request.
func (e *DSRExecutor) ExecuteDSR(ctx context.Context, dsrID types.ID) error {
	// 1. Fetch DSR
	dsr, err := e.dsrRepo.GetByID(ctx, dsrID)
	if err != nil {
		return fmt.Errorf("fetch dsr: %w", err)
	}

	e.logger.InfoContext(ctx, "starting dsr execution", "dsr_id", dsrID, "type", dsr.RequestType)

	// 2. Transition to IN_PROGRESS
	if err := dsr.ValidateTransition(compliance.DSRStatusInProgress); err != nil {
		return fmt.Errorf("invalid transition: %w", err)
	}
	dsr.Status = compliance.DSRStatusInProgress
	if err := e.dsrRepo.Update(ctx, dsr); err != nil {
		return fmt.Errorf("update dsr status: %w", err)
	}

	// 3. Fetch tasks
	tasks, err := e.dsrRepo.GetTasksByDSR(ctx, dsrID)
	if err != nil {
		return fmt.Errorf("fetch tasks: %w", err)
	}

	// 4. Execute tasks concurrently with semaphore
	taskCount := len(tasks)
	sem := make(chan struct{}, e.maxConcurrency)
	var wg sync.WaitGroup
	errorsCh := make(chan error, taskCount)

	for _, task := range tasks {
		wg.Add(1)
		go func(t compliance.DSRTask) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			if err := e.executeTask(ctx, dsr, &t); err != nil {
				e.logger.ErrorContext(ctx, "task failed", "task_id", t.ID, "error", err)
				errorsCh <- err
			}
		}(task)
	}

	// Wait for all tasks
	wg.Wait()
	close(errorsCh)

	// 5. Check for errors
	var taskErrors []error
	for err := range errorsCh {
		taskErrors = append(taskErrors, err)
	}

	// 6. Update DSR status based on results
	if len(taskErrors) > 0 {
		dsr.Status = compliance.DSRStatusFailed
		dsr.Reason = fmt.Sprintf("%d task(s) failed", len(taskErrors))
		e.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventDSRFailed, "dsr_executor", dsr.TenantID, map[string]any{
			"dsr_id": dsr.ID,
			"errors": len(taskErrors),
		}))
	} else {
		dsr.Status = compliance.DSRStatusCompleted
		completedAt := time.Now().UTC()
		dsr.CompletedAt = &completedAt
		e.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventDSRCompleted, "dsr_executor", dsr.TenantID, map[string]any{
			"dsr_id": dsr.ID,
			"tasks":  taskCount,
		}))
	}

	if err := e.dsrRepo.Update(ctx, dsr); err != nil {
		return fmt.Errorf("update final dsr status: %w", err)
	}

	e.logger.InfoContext(ctx, "dsr execution completed", "dsr_id", dsrID, "status", dsr.Status)
	return nil
}

// executeTask executes a single DSR task against a data source.
func (e *DSRExecutor) executeTask(ctx context.Context, dsr *compliance.DSR, task *compliance.DSRTask) error {
	task.Status = compliance.TaskStatusRunning
	if err := e.dsrRepo.UpdateTask(ctx, task); err != nil {
		return err
	}

	var result interface{}
	var execErr error

	switch task.TaskType {
	case compliance.RequestTypeAccess:
		result, execErr = e.executeAccessRequest(ctx, dsr, task)
	case compliance.RequestTypeErasure:
		result, execErr = e.executeErasureRequest(ctx, dsr, task)
	case compliance.RequestTypeCorrection:
		result, execErr = e.executeCorrectionRequest(ctx, dsr, task)
	case compliance.RequestTypePortability:
		result, execErr = e.executeAccessRequest(ctx, dsr, task) // Same as ACCESS for MVP
	default:
		execErr = fmt.Errorf("unsupported task type: %s", task.TaskType)
	}

	// Update task result
	if execErr != nil {
		task.Status = compliance.TaskStatusFailed
		task.Error = execErr.Error()
	} else if task.Status == compliance.TaskStatusRunning {
		// Only mark completed if still running (sub-functions might set other statuses like MANUAL_ACTION_REQUIRED)
		task.Status = compliance.TaskStatusCompleted
		task.Result = result
		completedAt := time.Now().UTC()
		task.CompletedAt = &completedAt
	} else {
		// Result might be partial if status was changed
		task.Result = result
	}

	if err := e.dsrRepo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("update task status: %w", err)
	}

	return execErr
}

// executeAccessRequest collects and exports all data for the subject.
func (e *DSRExecutor) executeAccessRequest(ctx context.Context, dsr *compliance.DSR, task *compliance.DSRTask) (interface{}, error) {
	// 1. Get data source
	ds, err := e.dsRepo.GetByID(ctx, task.DataSourceID)
	if err != nil {
		return nil, fmt.Errorf("fetch data source: %w", err)
	}

	// 2. Get connector
	conn, err := e.connRegistry.GetConnector(ds.Type)
	if err != nil {
		return nil, fmt.Errorf("get connector: %w", err)
	}

	// 3. Connect
	if err := conn.Connect(ctx, ds); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	// 4. Find PII fields for retrieval
	pagination := types.Pagination{Page: 1, PageSize: 1000}
	piiResult, err := e.piiRepo.GetByDataSource(ctx, ds.ID, pagination)
	if err != nil {
		return nil, fmt.Errorf("fetch pii classifications: %w", err)
	}

	// 5. Group by entity
	entityFields := make(map[string][]string)
	for _, pii := range piiResult.Items {
		entityFields[pii.EntityName] = append(entityFields[pii.EntityName], pii.FieldName)
	}

	// 6. Execute Export
	accessResults := make([]map[string]interface{}, 0)
	var totalRecords int64

	for entityName, fields := range entityFields {
		filter := make(map[string]string)
		for _, field := range fields {
			for idKey, idVal := range dsr.SubjectIdentifiers {
				if strings.EqualFold(idKey, field) {
					filter[field] = idVal
				}
			}
		}

		if len(filter) == 0 {
			// e.logger.WarnContext(ctx, "no matching identifiers", "entity", entityName)
			continue
		}

		records, err := conn.Export(ctx, entityName, filter)
		if err != nil {
			e.logger.ErrorContext(ctx, "export failed", "entity", entityName, "error", err)
			continue
		}

		if len(records) > 0 {
			totalRecords += int64(len(records))
			accessResults = append(accessResults, map[string]interface{}{
				"entity":  entityName,
				"records": records,
			})
		}
	}

	// Emit event
	e.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventDSRDataAccessed, "dsr_executor", dsr.TenantID, map[string]any{
		"dsr_id":         dsr.ID,
		"data_source_id": ds.ID,
		"entities_count": len(accessResults),
		"total_records":  totalRecords,
	}))

	result := map[string]interface{}{
		"data_source_id": ds.ID,
		"data_source":    ds.Name,
		"accessed_at":    time.Now().UTC(),
		"data":           accessResults,
		"total_records":  totalRecords,
	}

	return result, nil
}

// executeErasureRequest deletes all data for the subject.
func (e *DSRExecutor) executeErasureRequest(ctx context.Context, dsr *compliance.DSR, task *compliance.DSRTask) (interface{}, error) {
	// 1. Get data source
	ds, err := e.dsRepo.GetByID(ctx, task.DataSourceID)
	if err != nil {
		return nil, fmt.Errorf("fetch data source: %w", err)
	}

	// CHECK DELETION MODE
	if ds.DeletionMode == discovery.DeletionModeManual {
		e.logger.InfoContext(ctx, "manual deletion required", "dsr_id", dsr.ID, "data_source_id", ds.ID)

		// Emit event
		e.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventDSRManualDeletionRequired, "dsr_executor", dsr.TenantID, map[string]any{
			"dsr_id":         dsr.ID,
			"data_source_id": ds.ID,
			"reason":         "Data Source configured for manual deletion",
		}))

		task.Status = compliance.TaskStatusManualActionRequired
		return map[string]interface{}{
			"status":  "MANUAL_ACTION_REQUIRED",
			"message": "Manual deletion verification required by configuration.",
		}, nil
	}

	// 2. Get connector
	conn, err := e.connRegistry.GetConnector(ds.Type)
	if err != nil {
		return nil, fmt.Errorf("get connector: %w", err)
	}

	// 3. Connect
	if err := conn.Connect(ctx, ds); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	// 4. Find PII fields for deletion
	pagination := types.Pagination{Page: 1, PageSize: 1000}
	piiResult, err := e.piiRepo.GetByDataSource(ctx, ds.ID, pagination)
	if err != nil {
		return nil, fmt.Errorf("fetch pii classifications: %w", err)
	}

	// 5. Group by entity for deletion
	entityFields := make(map[string][]string)
	for _, pii := range piiResult.Items {
		entityFields[pii.EntityName] = append(entityFields[pii.EntityName], pii.FieldName)
	}

	// 6. Execute deletion
	deletionLog := make([]map[string]interface{}, 0)
	var totalDeleted int64

	for entityName, fields := range entityFields {
		filter := make(map[string]string)
		for _, field := range fields {
			for idKey, idVal := range dsr.SubjectIdentifiers {
				if strings.EqualFold(idKey, field) {
					filter[field] = idVal
				}
			}
		}

		if len(filter) == 0 {
			e.logger.WarnContext(ctx, "no matching identifiers for entity deletion", "entity", entityName)
			continue
		}

		count, err := conn.Delete(ctx, entityName, filter)
		if err != nil {
			e.logger.ErrorContext(ctx, "failed to delete entity", "entity", entityName, "error", err)
			deletionLog = append(deletionLog, map[string]interface{}{
				"entity": entityName,
				"status": "FAILED",
				"error":  err.Error(),
			})
			continue
		}

		totalDeleted += count
		deletionLog = append(deletionLog, map[string]interface{}{
			"entity":  entityName,
			"status":  "DELETED",
			"count":   count,
			"filters": filter,
		})
	}

	// Emit deletion event
	e.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventDSRDataDeleted, "dsr_executor", dsr.TenantID, map[string]any{
		"dsr_id":         dsr.ID,
		"data_source_id": ds.ID,
		"entities_count": len(entityFields),
		"total_deleted":  totalDeleted,
	}))

	result := map[string]interface{}{
		"data_source_id": ds.ID,
		"data_source":    ds.Name,
		"deleted_at":     time.Now().UTC(),
		"deletions":      deletionLog,
		"total_deleted":  totalDeleted,
	}

	return result, nil
}

// executeCorrectionRequest updates specified data for the subject.
func (e *DSRExecutor) executeCorrectionRequest(ctx context.Context, dsr *compliance.DSR, task *compliance.DSRTask) (interface{}, error) {
	// 1. Get data source
	ds, err := e.dsRepo.GetByID(ctx, task.DataSourceID)
	if err != nil {
		return nil, fmt.Errorf("fetch data source: %w", err)
	}

	// 2. For MVP, log the correction request
	// Real implementation would need:
	// - Correction payload in DSR
	// - Connector.Update() capability
	// - Before/after snapshot

	result := map[string]interface{}{
		"data_source_id": ds.ID,
		"data_source":    ds.Name,
		"corrected_at":   time.Now().UTC(),
		"note":           "Correction capability requires connector Update() method",
	}

	return result, nil
}

// filterSamplesBySubject filters data samples by subject identifiers.
func (e *DSRExecutor) filterSamplesBySubject(samples []string, identifiers map[string]string) []string {
	// Simple implementation: check if any identifier value appears in sample
	filtered := make([]string, 0)
	for _, sample := range samples {
		for _, idValue := range identifiers {
			// Case-insensitive contains check
			sampleLower := strings.ToLower(sample)
			idLower := strings.ToLower(idValue)
			if strings.Contains(sampleLower, idLower) {
				filtered = append(filtered, sample)
				break
			}
		}
	}
	return filtered
}

// GetExecutionResult retrieves the execution result for a DSR.
func (e *DSRExecutor) GetExecutionResult(ctx context.Context, dsrID types.ID) (interface{}, error) {
	tasks, err := e.dsrRepo.GetTasksByDSR(ctx, dsrID)
	if err != nil {
		return nil, fmt.Errorf("fetch tasks: %w", err)
	}

	results := make([]map[string]interface{}, 0, len(tasks))
	for _, task := range tasks {
		if task.Result != nil {
			// Marshal and unmarshal to ensure proper structure
			resultBytes, _ := json.Marshal(task.Result)
			var resultMap map[string]interface{}
			json.Unmarshal(resultBytes, &resultMap)

			results = append(results, map[string]interface{}{
				"task_id":        task.ID,
				"data_source_id": task.DataSourceID,
				"status":         task.Status,
				"result":         resultMap,
				"completed_at":   task.CompletedAt,
			})
		}
	}

	return map[string]interface{}{
		"dsr_id": dsrID,
		"tasks":  results,
		"total":  len(tasks),
	}, nil
}
