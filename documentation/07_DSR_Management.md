# 07. DSR (Data Subject Rights) Management

## Overview

DSR (Data Subject Rights) Management handles requests from data subjects to exercise their privacy rights under DPDPA. The system supports the complete lifecycle from request submission to execution and reporting.

---

## DSR Request Types

| Type | DPDPA Section | Description | Agent Action |
|------|---------------|-------------|--------------|
| **ACCESS** | Section 11 | Right to access personal data | Collect and export data |
| **ERASURE** | Section 13 | Right to delete personal data | Delete from all sources |
| **RECTIFICATION** | Section 12 | Right to correct inaccurate data | Update specified records |
| **PORTABILITY** | Section 11 | Right to data in portable format | Export in JSON/CSV |
| **RESTRICTION** | Section 6 | Restrict processing of data | Mark data as restricted |

---

## DSR Workflow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            DSR WORKFLOW                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌───────────────┐    ┌───────────────┐    ┌───────────────┐               │
│  │ Data Subject  │───►│ Data Principal│───►│   Control Centre DSR    │               │
│  │ (Person)      │    │    Portal     │    │   Manager     │               │
│  └───────────────┘    └───────────────┘    └───────┬───────┘               │
│                                                     │                        │
│                              ┌──────────────────────┼──────────────────────┐│
│                              ▼                      ▼                      ▼│
│                     ┌────────────────┐    ┌────────────────┐    ┌─────────┐│
│                     │ Identity       │    │ Locate PII     │    │ Create  ││
│                     │ Verification   │    │ Across Sources │    │ Task    ││
│                     └────────────────┘    └────────────────┘    └─────────┘│
│                                                                      │      │
│                                                                      ▼      │
│                              ┌───────────────────────────────────────────┐  │
│                              │     AGENT RECEIVES DSR TASK               │  │
│                              ├───────────────────────────────────────────┤  │
│                              │                                           │  │
│                              │  ┌─────────────────────────────────────┐  │  │
│                              │  │  DSR Executor Service               │  │  │
│                              │  │  ├── executeAccessRequest()         │  │  │
│                              │  │  ├── executeErasureRequest()        │  │  │
│                              │  │  └── executeRectificationRequest()  │  │  │
│                              │  └─────────────────────────────────────┘  │  │
│                              │                    │                      │  │
│                              │                    ▼                      │  │
│                              │  ┌─────────────────────────────────────┐  │  │
│                              │  │  Execute on Data Sources            │  │  │
│                              │  │  • DELETE FROM table WHERE id = ?   │  │  │
│                              │  │  • Remove files                      │  │  │
│                              │  │  • Anonymize records                 │  │  │
│                              │  └─────────────────────────────────────┘  │  │
│                              └───────────────────────────────────────────┘  │
│                                                     │                        │
│                                                     ▼                        │
│                              ┌───────────────────────────────────────────┐  │
│                              │  Report Back to CONTROL CENTRE                      │  │
│                              │  • Records affected                       │  │
│                              │  • Errors encountered                     │  │
│                              │  • Completion status                      │  │
│                              └───────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## DSR Status Flow

```
SUBMITTED → PENDING_VERIFICATION → VERIFIED → IN_PROGRESS → COMPLETED
                    │                  │             │
                    ▼                  ▼             ▼
               REJECTED          ON_HOLD        PARTIAL/FAILED
```

| Status | Description |
|--------|-------------|
| `SUBMITTED` | Request received from data subject |
| `PENDING_VERIFICATION` | Awaiting identity verification |
| `VERIFIED` | Identity confirmed, ready to process |
| `IN_PROGRESS` | Agents are executing the request |
| `COMPLETED` | All tasks finished successfully |
| `PARTIAL` | Some tasks failed, partial completion |
| `FAILED` | Critical failure, could not process |
| `ON_HOLD` | Paused pending legal review |
| `REJECTED` | Identity verification failed |

---

## Agent DSR Executor

**File**: [`services/dsr_executor.go`](file:///e:/Comply%20Ark/Technical/Data%20Lens%20Application/DataLensApplication/DataLensAgent/backend/services/dsr_executor.go) (623 lines)

### Data Structures

```go
// DSR Task received from CONTROL CENTRE
type DSRTask struct {
    ID              string
    CONTROL CENTREID          string                 // DSR ID in CONTROL CENTRE
    GlobalSubjectID string                 // Data subject identifier
    RequestType     DSRRequestType         // ACCESS, ERASURE, etc.
    Scope           map[string]interface{} // What to include/exclude
    DueDate         time.Time              // SLA deadline
    Priority        int                    // 1-5 priority level
    Status          DSRStatus
    ReceivedAt      time.Time
}

// Result of DSR execution
type DSRResult struct {
    TaskID               string
    CONTROL CENTREID               string
    RequestType          DSRRequestType
    Status               DSRStatus
    AffectedDataSources  int
    AffectedRecords      int
    ExecutionLogs        []DSRExecutionLog
    AccessData           *AccessRequestData  // For ACCESS requests
    Error                string
    StartedAt            time.Time
    CompletedAt          time.Time
}

// Log entry for each action taken
type DSRExecutionLog struct {
    Timestamp    time.Time
    DataSourceID int
    Action       string  // DELETE, ANONYMIZE, EXPORT
    TableName    string
    ColumnName   string
    RecordCount  int
    Status       string
    Error        string
}
```

### Key Methods

```go
// QueueTask queues a DSR task for execution
func (e *DSRExecutor) QueueTask(ctx context.Context, task sync.DSRTask) error

// ProcessPendingTasks processes all pending DSR tasks
func (e *DSRExecutor) ProcessPendingTasks(ctx context.Context) error

// executeTask executes a single DSR task
func (e *DSRExecutor) executeTask(ctx context.Context, task DSRTask)

// executeAccessRequest collects all data for a subject
func (e *DSRExecutor) executeAccessRequest(ctx context.Context, task DSRTask) (*AccessRequestData, error)

// executeErasureRequest deletes all data for a subject
func (e *DSRExecutor) executeErasureRequest(ctx context.Context, task DSRTask) ([]DSRExecutionLog, error)

// executeRectificationRequest updates specified data
func (e *DSRExecutor) executeRectificationRequest(ctx context.Context, task DSRTask) ([]DSRExecutionLog, error)
```

---

## ACCESS Request Data

For ACCESS requests, the agent collects:

```go
type AccessRequestData struct {
    SubjectInfo  SubjectInfo            // Who the data belongs to
    PIILocations []PIILocationData      // Where their data is stored
    Summary      map[string]interface{} // Statistics
}

type SubjectInfo struct {
    GlobalSubjectID  string
    ClientSpecificID string
    Type             string  // employee, customer, vendor
    Name             string
    Email            string
}

type PIILocationData struct {
    DataSourceID      int
    DataSourceName    string
    DataSourceType    string
    ObjectIdentifier  string    // e.g., "hr.employees.email"
    PIICategory       string
    RecordCount       int
    SampleData        string    // Masked
    ProcessingPurpose string
    LawfulBasis       string
}
```

---

## ERASURE Request Execution

For ERASURE requests, the agent:

1. **Locates all PII** for the subject across data sources
2. **Applies retention rules** - some data may need to be kept
3. **Executes deletion** using appropriate method:
   - Database: `DELETE` or `UPDATE` (anonymization)
   - Files: File deletion
   - External: API calls to delete
4. **Logs every action** for audit trail
5. **Reports back** to CONTROL CENTRE

### Execution Logic

```go
// For each data source where subject has PII
for _, location := range piiLocations {
    switch location.DataSourceType {
    case "postgresql", "mysql":
        err := e.deleteFromDatabase(ctx, location, subjectIdentifiers)
    case "filesystem":
        err := e.deleteFiles(ctx, location, subjectIdentifiers)
    case "s3":
        err := e.deleteFromS3(ctx, location, subjectIdentifiers)
    }
    
    // Log the action
    logs = append(logs, DSRExecutionLog{
        Timestamp:    time.Now(),
        DataSourceID: location.DataSourceID,
        Action:       "DELETE",
        TableName:    location.TableName,
        RecordCount:  affectedCount,
        Status:       status,
    })
}
```

---

## SLA Tracking

DPDPA requires responses within specific timeframes:

| Request Type | SLA | Default in DataLens |
|--------------|-----|---------------------|
| ACCESS | Reasonable time | 30 days |
| ERASURE | Without delay | 30 days |
| RECTIFICATION | Without delay | 30 days |

### SLA Dashboard

```
DSR Requests Status:
┌─────────────────────────────────────────────────────────────────┐
│  Total: 45  │  On Time: 38  │  At Risk: 5  │  Overdue: 2       │
├─────────────────────────────────────────────────────────────────┤
│ ID      │ Type    │ Subject     │ Due Date   │ Status          │
├─────────────────────────────────────────────────────────────────┤
│ DSR-001 │ ACCESS  │ John Doe    │ 2026-02-20 │ ✅ COMPLETED    │
│ DSR-002 │ ERASURE │ Jane Smith  │ 2026-02-15 │ 🔄 IN_PROGRESS  │
│ DSR-003 │ ACCESS  │ Bob Wilson  │ 2026-02-12 │ ⚠️ AT_RISK      │
│ DSR-004 │ ERASURE │ Alice Chen  │ 2026-02-10 │ 🔴 OVERDUE      │
└─────────────────────────────────────────────────────────────────┘
```

---

## Audit Trail

Every DSR action is logged:

```sql
dsr_activity_log (
  id              UUID PRIMARY KEY,
  dsr_request_id  UUID,
  action          VARCHAR(100),   -- SUBMITTED, VERIFIED, EXECUTED, etc.
  performed_by    UUID,           -- User who performed action
  details         JSONB,          -- Action-specific details
  created_at      TIMESTAMP
)
```

---

## Control Centre DSR Components

### Backend

| Handler/Service | Purpose |
|-----------------|---------|
| `dsr_handler.go` | API endpoints for DSR management |
| `dsr_service.go` | DSR business logic |
| `dsr_repository.go` | Database operations |

### Frontend Pages

| Page | Purpose |
|------|---------|
| `DSRRequests.tsx` | List all DSR requests |
| `DSRDetails.tsx` | View/manage single request |
| `DSRCreate.tsx` | Create new DSR request |
| `DSRTemplates.tsx` | Response templates |

### Data Principal Portal

| Page | Purpose |
|------|---------|
| `SubmitRequest.tsx` | Data subjects submit requests |
| `TrackRequest.tsx` | Track request status |
| `VerifyIdentity.tsx` | Identity verification flow |
