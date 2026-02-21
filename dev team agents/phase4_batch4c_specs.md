# Phase 4 — Batch 4C Task Specifications

**Sprint**: Phase 4 — Core Compliance Pages  
**Estimated Duration**: 2 days  
**Pre-requisites**: Batch 4A complete (audit log API live, retention migrations applied), Batch 4B complete (UI overhaul)

---

## Execution Order

**Parallel Group 1** (no dependencies):
- Task 4C-1 (Backend) + Task 4C-2 (Frontend) + Task 4C-3 (Backend) — can run in PARALLEL

**Parallel Group 2** (depends on 4C-1):
- Task 4C-4, 4C-5, 4C-6 (all Frontend) — can run in PARALLEL after 4C-1 completes

---

## Task 4C-1: Backend — Consent Sessions + Data Subjects + Retention CRUD APIs

**Agent**: Backend  
**Priority**: P0 (blocking — Frontend Tasks 4C-4, 4C-5, 4C-6 depend on this)  
**Depends On**: None  
**Estimated Effort**: Medium (3h)

### Objective

Expose three new APIs to the Control Centre: tenant-wide consent session listing, data subject search, and retention policy CRUD. Domain models and repos exist — create service + handler layers and wire into routes.

### Context — Read These Files First
- `internal/domain/compliance/retention.go` — RetentionPolicy, RetentionLog, RetentionPolicyRepository
- `internal/repository/postgres_retention.go` — CRUD repo (implemented), log stubs
- `internal/domain/consent/entities.go` — DataPrincipalProfile (L161), DataPrincipalProfileRepository (L319), ConsentSessionRepository (L293)
- `internal/repository/postgres_data_principal_profile.go` — ListByTenant exists
- `internal/handler/consent_handler.go` — listSessions (subject-only)
- `internal/handler/audit_handler.go` — pattern reference
- `cmd/api/routes.go` — mount new routes in mountCCRoutes()
- `cmd/api/main.go` — wire new handlers/services

### Requirements

#### 1. Consent Sessions — Tenant-Wide Listing
- Add `ConsentSessionFilters` struct and `ListByTenant` to `ConsentSessionRepository`
- Update `ConsentHandler.listSessions` to support tenant-wide mode when no `subject_id` provided
- API: `GET /api/v2/consent/sessions?page=1&page_size=20&status=GRANTED&purpose_id=uuid`

#### 2. Data Subjects — Search API
- Add `SearchByTenant(ctx, tenantID, query, pagination)` to `DataPrincipalProfileRepository`
- Create `internal/handler/data_subject_handler.go` — `GET /api/v2/subjects?search=&page=&page_size=`
- Mount at `/subjects` in `mountCCRoutes()`

#### 3. Retention Policy CRUD
- Create `internal/service/retention_service.go` — Create, GetByID, GetByTenant, Update, Delete, GetLogs
- Create `internal/handler/retention_handler.go` — full CRUD + logs endpoint
- Implement `CreateLog` + `GetLogs` stubs in `postgres_retention.go`
- Mount at `/retention` in `mountCCRoutes()`

### Acceptance Criteria
- [ ] `GET /api/v2/consent/sessions` returns tenant-wide paginated sessions
- [ ] `GET /api/v2/subjects?search=email@test.com` returns matching profiles
- [ ] Retention CRUD endpoints all work
- [ ] All endpoints tenant-scoped
- [ ] `go build ./...` passes
- [ ] All existing tests pass
- [ ] AGENT_COMMS.md updated with API contracts

---

## Task 4C-2: Frontend — Audit Logs Page

**Agent**: Frontend  
**Priority**: P1  
**Depends On**: None (API live from Batch 4A)  
**Estimated Effort**: Medium (3h)

### Objective
Replace placeholder at `/audit-logs` with audit log viewer. Filters: entity type, action, user, date range.

### Context
- Pattern: `pages/Compliance/GrievanceList.tsx`
- API: `GET /api/v2/audit-logs?page=1&page_size=20&entity_type=DSR&action=APPROVE&start_date=...&end_date=...`
- Response: `{ success, data: [AuditLog], meta: {page, page_size, total, total_pages} }`

### Requirements
1. Create `pages/AuditLogs.tsx` + `services/auditService.ts`
2. Filter bar: Entity Type dropdown, Action dropdown, Date range (start/end), Clear button
3. DataTable columns: Timestamp, Action (StatusBadge), Entity Type, Resource ID, User, IP, Details
4. Pagination controls
5. Update `App.tsx` route
6. Use KokonutUI components from `@datalens/shared`

### Acceptance Criteria
- [ ] All filters work
- [ ] Pagination works
- [ ] `npm run build -w @datalens/control-centre` passes

---

## Task 4C-3: Backend — Retention Scheduler (Cron Job)

**Agent**: Backend  
**Priority**: P1  
**Depends On**: None  
**Estimated Effort**: Large (4h)

### Objective
Add retention policy evaluation to the existing scheduler. Runs daily, checks all ACTIVE policies, logs retention actions.

### Context
- Pattern: `internal/service/scheduler.go` — existing scheduler loop
- `internal/domain/compliance/retention.go` — RetentionPolicy entity
- `internal/repository/postgres_retention.go` — repo with CreateLog/GetLogs stubs

### Requirements
1. Add `retentionRepo` to `SchedulerService`, wire in `NewSchedulerService()`
2. Implement `checkRetentionPolicies(ctx)` — run once/day via `lastRetentionCheck` timestamp
3. For each ACTIVE policy: evaluate if data exceeds `MaxRetentionDays`
4. If `AutoErase=true` → log `ERASED` action; else → log `RETENTION_EXCEEDED`
5. Implement `CreateLog` + `GetLogs` stubs in `postgres_retention.go`
6. Call from scheduler loop in `Start()`

### Acceptance Criteria
- [ ] `checkRetentionPolicies` runs once daily
- [ ] RetentionLog entries created for expired policies
- [ ] Stubs fully implemented
- [ ] `go build ./...` passes
- [ ] AGENT_COMMS.md updated

---

## Task 4C-4: Frontend — Consent Records Page

**Agent**: Frontend  
**Priority**: P1  
**Depends On**: Task 4C-1  
**Estimated Effort**: Medium (3h)

### Objective
Replace placeholder at `/consent` with consent session listing. Show purpose, subject, status, timestamp.

### Requirements
1. Create `pages/Consent/ConsentRecords.tsx` + `services/consentRecordService.ts`
2. Status filter buttons: All, Granted, Withdrawn, Expired
3. DataTable columns: Session ID, Subject, Status (StatusBadge), Purposes (count), Widget (link), Timestamp
4. Pagination + update `App.tsx`

### Acceptance Criteria
- [ ] Status filters work
- [ ] Widget links navigate correctly
- [ ] `npm run build -w @datalens/control-centre` passes

---

## Task 4C-5: Frontend — Data Subjects Page

**Agent**: Frontend  
**Priority**: P1  
**Depends On**: Task 4C-1  
**Estimated Effort**: Medium (3h)

### Objective
Replace placeholder at `/subjects` with data subject listing. Search by email/phone, link to DSRs and consent records.

### Requirements
1. Create `pages/DataSubjects.tsx` + `services/dataSubjectService.ts`
2. Search bar with 300ms debounce
3. DataTable columns: Email, Phone, Verification Status (StatusBadge), Method, Minor (badge), Last Access, Actions
4. Actions: "View DSRs" → `/dsr?subject_id=`, "View Consent" → `/consent?subject_id=`
5. Pagination + update `App.tsx`

### Acceptance Criteria
- [ ] Search works with debounce
- [ ] Action buttons navigate correctly
- [ ] `npm run build -w @datalens/control-centre` passes

---

## Task 4C-6: Frontend — Retention Policies Page

**Agent**: Frontend  
**Priority**: P1  
**Depends On**: Task 4C-1  
**Estimated Effort**: Medium (2h)

### Objective
Replace placeholder at `/retention` with CRUD page for retention policies.

### Requirements
1. Create `pages/RetentionPolicies.tsx` + `services/retentionService.ts`
2. DataTable: Description, Purpose, Days, Data Categories (badges), Status (StatusBadge), Auto-Erase, Actions
3. Create/Edit modal (Dialog): Description, Purpose, Days, Categories, Status, AutoErase
4. Delete confirmation modal
5. "Add Policy" header button
6. Update `App.tsx`

### Acceptance Criteria
- [ ] CRUD works (create, edit, delete)
- [ ] Status badges correct
- [ ] `npm run build -w @datalens/control-centre` passes
