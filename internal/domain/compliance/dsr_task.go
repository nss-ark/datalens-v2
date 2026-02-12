package compliance

import (
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// DSRTaskStatus represents the status of a sub-task for a specific data source.
type DSRTaskStatus string

const (
	TaskStatusPending              DSRTaskStatus = "PENDING"
	TaskStatusRunning              DSRTaskStatus = "RUNNING"
	TaskStatusCompleted            DSRTaskStatus = "COMPLETED"
	TaskStatusVerified             DSRTaskStatus = "VERIFIED"
	TaskStatusFailed               DSRTaskStatus = "FAILED"
	TaskStatusManualActionRequired DSRTaskStatus = "MANUAL_ACTION_REQUIRED"
)

// DSRTask represents a unit of work for a DSR against a specific Data Source.
type DSRTask struct {
	ID           types.ID       `json:"id"`
	DSRID        types.ID       `json:"dsr_id"`
	DataSourceID types.ID       `json:"data_source_id"`
	TenantID     types.ID       `json:"tenant_id"`
	TaskType     DSRRequestType `json:"task_type"` // Usually matches DSR type
	Status       DSRTaskStatus  `json:"status"`
	Result       any            `json:"result,omitempty"` // JSONB payload of findings/actions
	Error        string         `json:"error,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
}
