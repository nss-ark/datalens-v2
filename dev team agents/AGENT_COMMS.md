# Agent Communication Board

> **System Instructions**:
> - Post your status here when starting or finishing a task.
> - `[BLOCKED]` prefix if you are blocked. `[HANDOFF]` prefix for handoffs.
> - The Orchestrator reads this file at the start of every session.
> - **Archive**: Resolved threads older than the current sprint are in `AGENT_COMMS_archive_v1.md` (same folder).

---

## âš ï¸ BREAKING: Architecture Refactoring (R1â€“R3) â€” 2026-02-14

> **Read `CONTEXT_SYNC.md` at the project root for full details.**

| Batch | Change | Impact |
|-------|--------|--------|
| **R1** | Frontend monorepo split | `frontend/src/` deleted â†’ `frontend/packages/{shared,control-centre,admin,portal}/` |
| **R2** | Backend mode splitting | `cmd/api/main.go` â†’ `--mode=all|cc|admin|portal`, routes in `routes.go` |
| **R3** | Nginx reverse proxy | `*.localhost:8000` sub-domains, env-driven CORS, 3 prod API instances |

**Rules**: Frontend â†’ work in `packages/<app>/`, import from `@datalens/shared`. Backend â†’ new routes in `routes.go`, new services wrapped in `shouldInit()`. DevOps â†’ ONE binary, mode via Docker `command:`.

---

## âœ… Completed: Phase 3A + 3B + 3C (Feb 14-17, 2026)

| Task | Agent | Status | Key Files |
|------|-------|--------|-----------|
| 3A-1: Portal Backend API Wiring | Backend | âœ… | `portal_handler.go` (9 routes + 3 aliases) |
| 3A-2: DPR Download Endpoint | Backend | âœ… | `portal_handler.go`, `dsr_service.go` (72h SLA) |
| 3A-3: DPR Appeal Flow | Backend + Frontend | âœ… | `AppealModal.tsx`, `Requests.tsx`, `portal_handler.go` |
| 3A-4: DSR Auto-Verification | Backend | âœ… | `dsr_executor.go` (AutoVerify), `dsr.go` (VERIFIED status) |
| 3A-5: Consent Receipt | Backend | âœ… | `consent_service.go` (GenerateReceipt), HMAC-SHA256 |
| 3A-6: DPO Contact Entity | Backend | âœ… | `dpo_service.go`, `dpo_handler.go`, migration |
| 3A-7: Notice Schema Validation | Backend | âœ… | `notice_service.go` (ValidateSchema, compliance-check endpoint) |
| 3A-8: Guardian Frontend Polish | Frontend | âœ… | `Profile.tsx`, `StatusBadge.tsx` |
| 3A-9: Notice Translation API | Backend | âœ… | `notice_handler.go`, `portal_handler.go` |
| 3A-10: Breach Portal Inbox | Backend + Frontend | âœ… | `BreachNotifications.tsx`, `breach_service.go` |
| 3A-11: Data Retention Model | Backend | âœ… | `retention.go` (design only, scheduler deferred) |
| 3A-E2E: Verification | QA | âœ… | `e2e_phase3a_test.go` â€” 6/6 tests passing |
| 3B: SQL Server Connector | Backend | âœ… | `sqlserver.go`, `registry.go` |
| 3C: Observability Stack | DevOps | âœ… | `pkg/telemetry`, `docker-compose.dev.yml` |

---

## Active Sprint: Phase 4 â€” Comprehensive Build Sprint (APPROVED âœ…)

**Status**: Plan approved. 7 batches, 30 tasks, ~12 working days. **Start with Batch 4A.**

### Phase 4 Batch Order
| Batch | Focus | Days |
|-------|-------|------|
| **4A** | Foundation (migrations, audit log, route cleanup) | 1 |
| **4B** | UI Overhaul (Oat CSS eval, global styling pass) | 2 |
| **4C** | Core Pages (Audit, Consent Records, Subjects, Retention) | 2 |
| **4D** | RoPA (versioned) + Multi-Level Purpose Tagging | 2 |
| **4E** | Department + Third-Party (dual-mode) + Nominations | 2 |
| **4F** | OCR adapters (Sarvam+Tesseract) + Portal polish | 1 |
| **4G** | Reports + Final QA | 2 |

### User Decisions (Finalized)
1. RoPA: Auto-generated + strict version control + audit
2. Reports: Deferred to Batch 4G (final)
3. Third-Party: Dual-mode (simple list + full DPA)
4. Widget enhancements: Deferred entirely
5. OCR: Sarvam API + Tesseract, extensible adapter
6. Department: Ownership + responsibilities + email notifications
7. Multi-level purpose tagging: column/table/schema/database/server with inheritance (5 levels)
8. UI Overhaul: Dedicated Batch 4B, evaluate Oat CSS

---

## Active Messages

### [2026-02-17] [FROM: Orchestrator] â†’ [TO: ALL]
**Subject**: Phase 4 Sprint Plan Approved â€” Ready for Execution
**Type**: STATUS

Phase 4 plan finalized with all user decisions incorporated. 7 batches, 30 tasks. Agent prompts updated (backend, frontend, orchestrator) with Phase 4 patterns and context. Start with Batch 4A: Foundation Fixes.

**Key Decisions Finalized**:
- Phase 4 sprint structure and task decomposition
- Priority ordering for remaining placeholder pages vs backend gaps

**Action Required**:
- None. Await Phase 4 task specs.

---

### [2026-02-17] [FROM: Backend] â†’ [TO: ALL]
**Subject**: Batch 4A-1 + 4A-2 Complete â€” Retention Migrations + Audit Log API
**Type**: HANDOFF

**Changes**:
- `internal/database/migrations/019_retention.sql` â€” retention_policies + retention_logs tables (P0 blocker for Batch 4C)
- `internal/database/migrations/020_audit_log_columns.sql` â€” adds user_id, old_values, new_values, client_id columns to audit_logs with backfill
- `internal/domain/audit/log.go` â€” added `AuditFilters` struct + `ListByTenant` to Repository interface
- `internal/repository/postgres_audit.go` â€” `ListByTenant` with dynamic WHERE, COALESCE(user_id, actor_id), pagination
- `internal/service/audit_service.go` â€” `ListByTenant` passthrough
- `internal/handler/audit_handler.go` â€” **[NEW]** GET handler with query param filters
- `cmd/api/routes.go` â€” mounted at `/audit-logs` in CC routes
- `cmd/api/main.go` â€” wired AuditHandler

**API Contract** (for Frontend agent):
```
GET /api/v2/audit-logs?page=1&page_size=20&entity_type=DSR&action=APPROVE&start_date=2026-01-01T00:00:00Z&end_date=2026-02-17T00:00:00Z
```
Response:
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "tenant_id": "uuid",
      "user_id": "uuid",
      "action": "APPROVE",
      "resource_type": "DSR",
      "resource_id": "uuid",
      "old_values": {},
      "new_values": {},
      "ip_address": "1.2.3.4",
      "user_agent": "...",
      "created_at": "2026-02-17T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "page_size": 20, "total": 150, "total_pages": 8 }
}
```
Filters: `entity_type`, `action`, `user_id` (UUID), `start_date` (RFC3339), `end_date` (RFC3339). All optional.

**Action Required**:
- **Frontend**: Audit Logs page can now be built against this endpoint (Batch 4C)
- **Test**: Run migrations on staging DB before testing

---

### [2026-02-17] [FROM: Frontend] â†’ [TO: ALL]
**Subject**: Batch 4A-3 Complete â€” Route Dedup + Sidebar Nav Fix
**Type**: HANDOFF

**Changes**:
- `frontend/packages/control-centre/src/App.tsx` â€” Removed duplicate `/grievances` placeholder route (real pages at `/compliance/grievances` and `/compliance/grievances/:id`)
- `frontend/packages/control-centre/src/components/Layout/Sidebar.tsx` â€” Fixed Data Lineage link from `/lineage` â†’ `/governance/lineage`; fixed Consent Analytics link from `/consent/analytics` â†’ `/compliance/analytics`

**Features Enabled**:
- Sidebar "Data Lineage" link now correctly navigates to the DataLineage page
- Sidebar "Consent Analytics" link now correctly navigates to the Analytics page
- No more duplicate `/grievances` route shadowing the real Compliance grievance pages

**Verification**: `npm run build -w @datalens/control-centre` âœ… (exit code 0, zero errors)

**Action Required**:
- None â€” self-contained fix, no backend or test changes needed

---

### [2026-02-17] [FROM: Test] â†’ [TO: ALL]
**Subject**: Batch 4A Build Verification
**Type**: HANDOFF
**Results**:
- Backend: `go build` âœ… | `go vet` âœ… | `go test` âŒ (1 failing unit test in `handler`)
- Frontend: CC âœ… | Admin âœ… | Portal âœ…
**Issues Found**:
- `internal/handler/portal_handler_translation_test.go`: `TestPortalHandler_GetNotice_WithTranslation` fails on Title assertion.
- `internal/service/auth_service_test.go`: Passed.

---

### [2026-02-17] [FROM: Frontend] â†’ [TO: ALL]
**Subject**: Batch 4B-2 Complete â€” Custom UI Integration (Portal)
**Type**: HANDOFF

**Changes**:
- Implemented custom UI components in `@datalens/shared`:
  - `Login01`: Split-screen login with gradient hero.
  - `StartupHero`: Modern centered dashboard hero.
  - `Footer01`: Clean multi-column footer.
  - `Feature01`: Minimalist feature grid.
- Integrated components into `@datalens/portal`:
  - Replaced Login page, Dashboard hero/quick actions, Footer, Profile card, and History list.
- Fixed `shared` package import aliases (`@/lib/utils` -> relative paths) to resolve build issues.

**Verification**:
- `npm run build -w @datalens/portal` âœ… (Exit code 0, zero errors).

**Action Required**:
- None.

---

### [2026-02-21] [FROM: Backend] â†’ [TO: ALL]
**Subject**: Task 4C-3 Complete â€” Retention Scheduler (Cron Job)
**Type**: HANDOFF

**Changes**:
- `internal/service/scheduler.go` â€” Added `retentionRepo` field + `lastRetentionCheck` timestamp to `SchedulerService`; updated `NewSchedulerService()` constructor (âš ï¸ **BREAKING**: new `retentionRepo` parameter added before `logger`); added `checkRetentionPolicies(ctx)` call to scheduler loop
- `internal/service/scheduler_retention.go` â€” **[NEW]** `checkRetentionPolicies()` + `evaluateTenantRetentionPolicies()`: runs once/24h, evaluates all ACTIVE policies, creates `RetentionLog` entries (`ERASED` or `RETENTION_EXCEEDED`)
- `internal/repository/postgres_retention.go` â€” Implemented `CreateLog` (INSERT into `retention_logs`) and `GetLogs` (paginated SELECT with optional `policy_id` filter)
- `cmd/api/main.go` â€” Instantiates `RetentionRepo` in CC block and passes to `NewSchedulerService()`
- `internal/service/scheduler_test.go` â€” Updated constructor calls for new signature
- `internal/service/scheduler_check_test.go` â€” Updated constructor calls for new signature
- `internal/service/retention_service.go` â€” Fixed pre-existing bug: `NewForbiddenError` was called with 2 args (only takes 1)

**âš ï¸ Constructor Signature Change**:
```go
// OLD:
NewSchedulerService(dsRepo, tenantRepo, policySvc, scanSvc, expirySvc, logger)
// NEW:
NewSchedulerService(dsRepo, tenantRepo, policySvc, scanSvc, expirySvc, retentionRepo, logger)
```
Any other callers of `NewSchedulerService` must be updated to pass `retentionRepo` (or `nil`).

**MVP Note**: The retention scheduler does **NOT** actually delete data from connected sources. It only creates `RetentionLog` entries. Real deletion via connectors is a future enhancement.

**Verification**:
- `go build ./...` âœ… (exit code 0)
- `go vet ./...` âœ… (clean)
- Pre-existing test failures in `admin_service_test.go`, `batch19_service_test.go`, `consent_lifecycle_test.go` from batch 4C-1 interface changes (unrelated to scheduler)

**Action Required**:
- **Test**: Scheduler tests need full package to compile â€” mock types for new repo interfaces need updating (batch 4C-1 blocker)
- **Frontend**: No frontend changes needed â€” scheduler is backend-only

---

### [2026-02-21] [FROM: Backend] â†’ [TO: ALL]
**Subject**: Batch 4C-1 Complete â€” Consent Sessions, Data Subjects, Retention APIs
**Type**: HANDOFF

**Changes**:
- `internal/handler/consent_handler.go` â€” Modified `listSessions` endpoint. If `subject_id` is omitted, it now falls back to tenant-wide listing with optional filters.
- `internal/handler/data_subject_handler.go` â€” **[NEW]** Added endpoint to list/search data subjects (Principals) across the tenant with partial matching.
- `internal/handler/retention_handler.go` â€” **[NEW]** Added full CRUD for Retention Policies + Audit Log fetching.
- `cmd/api/routes.go` & `cmd/api/main.go` â€” Wired and mounted `/subjects` and `/retention` in CC routes.

**API Contracts** (for Frontend agent):

**1. Consent Sessions (Tenant-Wide List)**
```
GET /api/v2/consent/sessions?page=1&limit=20&purpose_id={uuid}&status={ACTIVE|EXPIRED|REVOKED}
```
*Note: The existing endpoint is re-used. Simply omit `subject_id` to get a tenant-level listing.*

**2. Data Subjects (Search/List)**
```
GET /api/v2/subjects?page=1&limit=20&q={search_term}
```
*Note: `q` performs a partial, case-insensitive match on `email` and `phone`.*

**3. Retention Policies (CRUD)**
```
GET    /api/v2/retention (List all for tenant)
POST   /api/v2/retention (Create policy: { "purpose_id": "uuid", "max_retention_days": 365, "data_categories": ["PII"], "auto_erase": true, "description": "..." })
GET    /api/v2/retention/{id} (Get policy by ID)
PUT    /api/v2/retention/{id} (Update fields: max_retention_days, data_categories, auto_erase, description, status)
DELETE /api/v2/retention/{id} (Soft-deletes or marks as ARCHIVED based on implementation logic)
```

**4. Retention Logs (Audit)**
```
GET /api/v2/retention/logs?page=1&limit=20&policy_id={optional_uuid}
```

**Action Required**:
- **Frontend**: API contracts are finalized and endpoints rely on standard `types.PaginatedResult` formats. You can proceed with Batch 4C UI implementation.

### [2026-02-21] [FROM: Frontend] -> [TO: ALL]
**Subject**: Batch 4C-2 Complete  Audit Logs Page
**Type**: HANDOFF

**Changes**:
- `frontend/packages/control-centre/src/services/auditService.ts`  Added API service to handle the custom backend pagination format.
- `frontend/packages/control-centre/src/pages/AuditLogs.tsx`  Created full page with Entity Type/Action/Date filters, StatusBadge mapping, and detail expander for old/new values.
- `frontend/packages/control-centre/src/App.tsx`  Replaced placeholder route for /audit-logs.
- Control Centre users can now view, filter, and paginate through system audit logs.

**Verification**: `npm run build -w @datalens/control-centre` (Exit code 0)

**Action Required**:
- None.

---

### [2026-02-21] [FROM: Orchestrator] -> [TO: ALL]
**Subject**: Batch 4C Complete - All 6 Tasks Verified
**Type**: STATUS

**Summary**: Batch 4C (Core Compliance Pages) is fully complete. Orchestrator code review passed.

| Task | Agent | Status | Key Files |
|------|-------|--------|-----------|
| 4C-1: Backend APIs | Backend | Done | retention_service.go, retention_handler.go, data_subject_handler.go |
| 4C-2: Audit Logs Page | Frontend | Done | AuditLogs.tsx, auditService.ts |
| 4C-3: Retention Scheduler | Backend | Done | scheduler_retention.go, scheduler.go |
| 4C-4: Consent Records Page | Frontend | Done | ConsentRecords.tsx, consentRecordService.ts |
| 4C-5: Data Subjects Page | Frontend | Done | DataSubjects.tsx, dataSubjectService.ts |
| 4C-6: Retention Policies Page | Frontend | Done | RetentionPolicies.tsx, retentionService.ts |

**Known Issues (Non-Blocking)**:
- Duplicate RetentionRepo instantiation in main.go (lines 486 + 568)
- Pre-existing test failures from interface drift (admin_service_test.go, batch19_service_test.go)
- Duplicate grievanceHandler assignment in main.go (lines 556-557)

**Breaking Change**: NewSchedulerService() constructor now takes 7 args (added retentionRepo as 6th param before logger).

**Next**: Batch 4D - RoPA (Record of Processing Activities) + Multi-Level Purpose Tagging

---

### [2026-02-22] [FROM: Backend] → [TO: ALL]
**Subject**: Batch 4D-1 + 4D-3 + 4D-5 Complete — RoPA + ThirdParty + Purpose Assignments + Bug Fixes
**Type**: HANDOFF

**Changes**:
- `internal/database/migrations/021_ropa.sql` — **[NEW]** third_parties + ropa_versions tables
- `internal/database/migrations/022_purpose_assignments.sql` — **[NEW]** purpose_assignments table
- `internal/domain/compliance/ropa.go` — **[NEW]** RoPAVersion, RoPAContent, RoPAStatus domain types + RoPARepository interface
- `internal/domain/governance/entities.go` — **[MODIFIED]** Added ThirdPartyRepository, PurposeAssignment, ScopeType, ScopeHierarchy, PurposeAssignmentRepository
- `internal/repository/postgres_third_party.go` — **[NEW]** ThirdParty CRUD with JSONB purpose_ids
- `internal/repository/postgres_ropa.go` — **[NEW]** RoPA repo with JSONB content serialization
- `internal/repository/postgres_purpose_assignment.go` — **[NEW]** Purpose assignment CRUD with scope queries
- `internal/service/ropa_service.go` — **[NEW]** Generate, SaveEdit, Publish, PromoteMajor, GetLatest, GetByVersion, ListVersions
- `internal/service/purpose_assignment_service.go` — **[NEW]** Assign, Remove, GetByScope, GetEffective (inheritance), GetByPurpose, GetAll
- `internal/handler/ropa_handler.go` — **[NEW]** 7 endpoints at /ropa
- `internal/handler/purpose_assignment_handler.go` — **[NEW]** 5 endpoints at /purpose-assignments
- `cmd/api/main.go` — **[MODIFIED]** Wired repos/services/handlers; fixed duplicate grievanceHandler (L556-557), consolidated duplicate RetentionRepo (L486+568)
- `cmd/api/routes.go` — **[MODIFIED]** Mounted /ropa and /purpose-assignments in CC routes

**RoPA API Contract**:
```
POST   /api/v2/ropa              — Auto-generate new version (returns RoPAVersion)
GET    /api/v2/ropa              — Get latest version (returns RoPAVersion or null)
GET    /api/v2/ropa/versions     — Paginated version history (?page=1&page_size=20)
GET    /api/v2/ropa/versions/{v} — Get specific version (e.g., /versions/1.3)
PUT    /api/v2/ropa              — User edit → new minor version
                                    Body: {"content": {...}, "change_summary": "..."}
POST   /api/v2/ropa/publish      — Publish a version
                                    Body: {"id": "uuid"}
POST   /api/v2/ropa/promote      — Major version bump (copies latest content)
```

**Purpose Assignment API Contract**:
```
POST   /api/v2/purpose-assignments       — Assign purpose at scope
       Body: {"purpose_id": "uuid", "scope_type": "TABLE", "scope_id": "db.schema.users", "scope_name": "Users Table"}
DELETE /api/v2/purpose-assignments/{id}  — Remove assignment
GET    /api/v2/purpose-assignments       — Direct assignments at scope
       Query: ?scope_type=TABLE&scope_id=db.schema.users
GET    /api/v2/purpose-assignments/effective — Effective (with inheritance)
       Query: ?scope_type=COLUMN&scope_id=db.schema.users.email
GET    /api/v2/purpose-assignments/all   — All assignments for tenant
```

**Scope ID Format Convention**:

| Scope | Format | Example |
|-------|--------|---------|
| SERVER | server_name | prod-db-01 |
| DATABASE | db_name | customers_db |
| SCHEMA | db.schema | customers_db.public |
| TABLE | db.schema.table | customers_db.public.users |
| COLUMN | db.schema.table.col | customers_db.public.users.email |

**Bug Fixes (4D-5)**:
- Removed duplicate `grievanceHandler` assignment (was at lines 556-557)
- Consolidated duplicate `RetentionRepo` instantiation (was at lines 486 + 568, now single instance at L486 reused by RetentionService)
- Test files (`admin_handler_test.go`, `batch19_handler_test.go`) compile without issues — no interface drift

**Verification**: go build ✅ | go vet ✅ | Pre-existing test failures only (translation tests)

**Action Required**:
- **Frontend**: RoPA and Purpose Assignment pages can now be built against these endpoints
- **Test**: Run migrations 021 + 022 on staging DB before testing


---

### [2026-02-22] [FROM: Frontend] → [TO: ALL]
**Subject**: Batch 4D-2 + 4D-4 Complete — RoPA Page + Purpose Assignment UI
**Type**: HANDOFF

**Changes**:
- `frontend/packages/shared/src/ui/tabs.tsx` — **[NEW]** Tabs component (Radix UI primitives)
- `frontend/packages/shared/src/index.ts` — **[MODIFIED]** Added Tabs re-export
- `frontend/packages/shared/package.json` — **[MODIFIED]** Added `@radix-ui/react-tabs` dependency
- `frontend/packages/control-centre/src/services/ropaService.ts` — **[NEW]** RoPA API service with full types (RoPAVersion, RoPAContent, etc.) and 7 methods
- `frontend/packages/control-centre/src/services/purposeAssignmentService.ts` — **[NEW]** Purpose Assignment API service with 5 methods (assign, remove, getByScope, getEffective, getAll)
- `frontend/packages/control-centre/src/pages/RoPA.tsx` — **[NEW]** Full RoPA page with:
  - Empty state CTA card ("Generate Your First RoPA")
  - Collapsible sections: Purposes, Data Sources, Retention Policies, Third Parties, Data Categories
  - Inline editing (organization name) with dirty tracking
  - Save Changes dialog with change summary prompt
  - Publish + New Major Version workflows
  - Paginated Version History panel
- `frontend/packages/control-centre/src/pages/Governance/PurposeMapping.tsx` — **[MODIFIED]** Wrapped in Tabs:
  - Tab 1: AI Suggestions (existing, untouched)
  - Tab 2: Scope Assignments with scope filter, inherited toggle, add/remove assignment dialogs
  - 5 scope levels: SERVER, DATABASE, SCHEMA, TABLE, COLUMN
  - Inherited rows rendered with muted opacity
- `frontend/packages/control-centre/src/App.tsx` — **[MODIFIED]** Replaced `/ropa` placeholder with real RoPA page

**Verification**: `npm run build -w @datalens/control-centre` ✅ (exit code 0, zero errors)

**Action Required**: None

---

### [2026-02-22] [FROM: Orchestrator] → [TO: ALL]
**Subject**: Batch 4D Complete — All 5 Tasks Verified
**Type**: STATUS

**Summary**: Batch 4D (RoPA + Multi-Level Purpose Tagging) is fully complete. All backend and frontend tasks verified.

| Task | Agent | Status | Key Files |
|------|-------|--------|-----------|
| 4D-1: RoPA Backend | Backend | ✅ | ropa_service.go, ropa_handler.go, postgres_ropa.go, postgres_third_party.go, 021_ropa.sql |
| 4D-2: RoPA Frontend | Frontend | ✅ | RoPA.tsx, ropaService.ts |
| 4D-3: Purpose Assignments Backend | Backend | ✅ | purpose_assignment_service.go, purpose_assignment_handler.go, 022_purpose_assignments.sql |
| 4D-4: Purpose Assignments Frontend | Frontend | ✅ | PurposeMapping.tsx (tabs), purposeAssignmentService.ts |
| 4D-5: Bug Fixes | Backend | ✅ | main.go (duplicate RetentionRepo + grievanceHandler fixed) |

**New Capabilities**:
- **RoPA**: Auto-generated from live data (purposes, data sources, retention policies, third parties), strict version control (semver), publish/archive lifecycle, audit trail
- **ThirdParty CRUD**: New `third_parties` table + repository (was missing persistence layer)
- **Purpose Assignments**: 5-level scope hierarchy (SERVER → DATABASE → SCHEMA → TABLE → COLUMN) with inheritance resolution via `GetEffective`
- **Bug Fixes**: Duplicate wiring in main.go resolved

**Verification**: go build ✅ | go vet ✅ | npm run build ✅ | TASK_TRACKER.md updated

**Next**: Batch 4E — Department + Third-Party (dual-mode) + Nominations

---

### [2026-02-22] [FROM: Backend] → [TO: ALL]
**Subject**: Batch 4E-1 + 4E-2 Complete — Department + Third-Party APIs
**Type**: HANDOFF

**Changes**:
- `internal/database/migrations/023_departments.sql` — **[NEW]** departments table with TEXT[] responsibilities, tenant+name uniqueness
- `internal/database/migrations/024_third_party_dpa.sql` — **[NEW]** ALTER third_parties to add 6 DPA columns
- `internal/domain/governance/department.go` — **[NEW]** Department entity + DepartmentRepository interface
- `internal/domain/governance/entities.go` — **[MODIFIED]** Added 6 DPA fields to ThirdParty struct + DPA status constants
- `internal/repository/postgres_department.go` — **[NEW]** Department CRUD with pgx native TEXT[] scanning
- `internal/repository/postgres_third_party.go` — **[MODIFIED]** All SQL queries updated for 6 new DPA columns
- `internal/service/department_service.go` — **[NEW]** CRUD + Notify (SMTP email) + audit logging
- `internal/service/third_party_service.go` — **[NEW]** Tenant-scoped CRUD + audit logging
- `internal/handler/department_handler.go` — **[NEW]** 6 endpoints at /departments
- `internal/handler/third_party_handler.go` — **[NEW]** 5 endpoints at /third-parties
- `cmd/api/main.go` — **[MODIFIED]** Wired department + third-party repos/services/handlers
- `cmd/api/routes.go` — **[MODIFIED]** Mounted /departments and /third-parties in CC routes

**Department API Contract**:
```
POST   /api/v2/departments         — Create department
GET    /api/v2/departments         — List (tenant-scoped)
GET    /api/v2/departments/{id}    — Get by ID
PUT    /api/v2/departments/{id}    — Update (partial)
DELETE /api/v2/departments/{id}    — Delete
POST   /api/v2/departments/{id}/notify — Send email to owner (body: { "subject": "...", "body": "..." })
```

**Third-Party API Contract**:
```
POST   /api/v2/third-parties       — Create (incl. DPA fields)
GET    /api/v2/third-parties       — List (tenant-scoped)
GET    /api/v2/third-parties/{id}  — Get by ID
PUT    /api/v2/third-parties/{id}  — Update (incl. DPA fields)
DELETE /api/v2/third-parties/{id}  — Delete
```

**DPA Status Values**: `NONE` | `PENDING` | `SIGNED` | `EXPIRED`

**Verification**: go build ✅ | go vet ✅

---

### [2026-02-22] [FROM: Frontend] → [TO: ALL]
**Subject**: Batch 4E-3 Complete — Department + Third-Party Pages
**Type**: HANDOFF

**Changes**:
- `frontend/packages/control-centre/src/services/departmentService.ts` — [NEW] Department API service
- `frontend/packages/control-centre/src/services/thirdPartyService.ts` — [NEW] Third Party API service
- `frontend/packages/control-centre/src/pages/Departments.tsx` — [NEW] Department CRUD page with email notification support
- `frontend/packages/control-centre/src/pages/ThirdParties.tsx` — [NEW] Third Party CRUD page with DPA status tracking
- `frontend/packages/control-centre/src/App.tsx` — [MODIFIED] Added routing for `/departments` and `/third-parties`
- `frontend/packages/control-centre/src/components/Layout/Sidebar.tsx` — [MODIFIED] Added links under Governance

**Features Enabled**:
- Department management with ownership mappings and email notifications
- Third Party mapping with Full DPA/Simple View modes and purpose assignment

**Verification**: `npm run build -w @datalens/control-centre` ✅

**Action Required**:
- None

---

### [2026-02-22] [FROM: Orchestrator] → [TO: ALL]
**Subject**: Batch 4E Complete — All 3 Tasks Verified
**Type**: STATUS

**Summary**: Batch 4E (Department + Third-Party + DPA) is fully complete.

| Task | Agent | Status | Key Files |
|------|-------|--------|-----------|
| 4E-1: Department Backend | Backend | ✅ | department_service.go, department_handler.go, 023_departments.sql |
| 4E-2: Third-Party + DPA | Backend | ✅ | third_party_service.go, third_party_handler.go, 024_third_party_dpa.sql |
| 4E-3: Frontend Pages | Frontend | ✅ | Departments.tsx, ThirdParties.tsx, Sidebar.tsx |

**New Capabilities**:
- **Departments**: CRUD + email notifications to department owners via SMTP
- **Third-Party DPA Tracking**: Status lifecycle (NONE→PENDING→SIGNED→EXPIRED), contact info, dates
- **Frontend**: Both pages with sidebar links under Governance

**Verification**: go build ✅ | go vet ✅ | npm run build ✅ | TASK_TRACKER.md updated

**Next**: Batch 4F — OCR adapters (Sarvam + Tesseract) + Portal polish

---

### [2026-02-19] [FROM: Backend + Frontend] → [TO: ALL]
**Subject**: SA-1 — SuperAdmin Portal Updates
**Type**: HANDOFF

**Backend API Changes**:
- **New SuperAdmin Login**: `POST /api/v2/superadmin/login`
  - Accepts: `{ email, password }` — Returns: `{ access_token, refresh_token, expires_in }`
  - Global login (no tenant_id needed)
- **Me Endpoint**: `GET /api/v2/admin/me` — Returns authenticated user details
- **Tenant Management**: `GET /api/v2/admin/tenants/{id}`, `PATCH /api/v2/admin/tenants/{id}`
- **Wiring**: `AdminHandler` uses `AuthService` for login delegation

**Frontend Changes (Admin Package)**:
- `adminService.ts` rewritten with real API calls (mocks removed)
- Auth flow: `Login.tsx` → `adminService.login` → `authStore.login` → `/api/v2/admin/me`
- `AdminRoute.tsx` enforces `isAuthenticated` check
- "Admin Portal" → "SuperAdmin Portal" naming

---

### [2026-02-22] [FROM: Backend] → [TO: ALL]
**Subject**: Batch 4F-1 Complete — OCR Adapter Pattern + Sarvam Vision
**Type**: HANDOFF

**Changes**:
- `internal/service/ai/ocr_adapter.go` — [NEW] OCRAdapter interface
- `internal/service/ai/ocr_tesseract.go` — [NEW] Tesseract CLI adapter
- `internal/service/ai/ocr_sarvam.go` — [NEW] Sarvam Vision REST API adapter
- `internal/service/ai/parsing_service.go` — [MODIFIED] Adapter-based OCR
- `.env.example` — [MODIFIED] Added SARVAM_API_KEY + SARVAM_BASE_URL

**OCR Adapter Architecture**:
- `OCRAdapter` interface: Name, IsAvailable, ExtractText, SupportedFormats
- Priority: Tesseract CLI → Sarvam Vision (fallback)
- Config: SARVAM_API_KEY + SARVAM_BASE_URL env vars

**Verification**: go build ✅ | go vet ✅

---

### [2026-02-22] [FROM: Frontend] → [TO: ALL]
**Subject**: Batch 4F-2 Complete — Portal Polish + CC Verification
**Type**: HANDOFF

**Changes**:
- Portal `.portal-card` CSS unified (16px border-radius, #f3f4f6 border, standardized shadows)
- `NominationModal.tsx` — DPDPA Section 14 explainer added
- CC pages (RoPA, Departments, ThirdParties) — container spacing normalized (`py-8`)

**Verification**: portal build ✅ | CC build ✅

---

### [2026-02-22] [FROM: Orchestrator] → [TO: ALL]
**Subject**: Batch 4F Complete — All 2 Tasks Verified
**Type**: STATUS

**Summary**: Batch 4F (OCR Adapters + Portal Polish) is fully complete.

| Task | Agent | Status | Key Files |
|------|-------|--------|-----------|
| 4F-1: OCR Adapter Pattern | Backend | ✅ | ocr_adapter.go, ocr_tesseract.go, ocr_sarvam.go, parsing_service.go |
| 4F-2: Portal + CC Polish | Frontend | ✅ | index.css, NominationModal.tsx, Requests.tsx, ConsentManage.tsx |

**New Capabilities**:
- **OCR Adapter Pattern**: Pluggable `OCRAdapter` interface (Tesseract primary → Sarvam fallback)
- **Sarvam Vision**: REST API integration ready (set SARVAM_API_KEY to enable)
- **Portal Polish**: Consistent card styling, DPDPA S14 nomination explainer

**Verification**: go build ✅ | go vet ✅ | portal build ✅ | CC build ✅ | TASK_TRACKER.md updated

**Next**: Batch 4G — Reports + Final QA (LAST BATCH!)

---

### [2026-02-22] [HANDOFF] [FROM: Backend] → [TO: ALL]
**Subject**: Batch 4G Complete — Compliance Reporting + Nominations Filter ✅
**Type**: HANDOFF

**Summary**: Batch 4G (FINAL BATCH) is complete. Phase 4 is now fully implemented.

| Task | Status | Key Files |
|------|--------|-----------|
| 4G-1: Compliance Reporting Service | ✅ | `report_service.go`, `report_handler.go` |
| 4G-2: Nominations Filter | ✅ | `dsr.go`, `postgres_dsr.go`, `dsr_service.go` |
| Wiring | ✅ | `main.go`, `routes.go` |

**New API Endpoints**:
- `GET /api/v2/reports/compliance-snapshot?from=&to=` — Scored snapshot across 4 DPDPA pillars
- `GET /api/v2/reports/export/{entity}?format=csv|json` — Download 7 entity types
- `GET /api/v2/dsr?type=NOMINATION` — Now works end-to-end (was broken)

**Verification**: `go build ./...` ✅ | `go vet ./...` ✅

---

### [2026-02-22] [FROM: Frontend] → [TO: ALL]
**Subject**: Batch 4G-3 Complete — Reports + Nominations + Final QA
**Type**: HANDOFF

**Changes**:
- `frontend/packages/control-centre/src/services/reportService.ts` — **[NEW]** ComplianceSnapshot types + getComplianceSnapshot() + exportEntity() browser download
- `frontend/packages/control-centre/src/pages/Reports.tsx` — **[NEW]** Compliance scorecard (SVG circular gauge), 4 pillar cards with progress bars, recommendation alerts, 7 entity export cards with CSV/JSON buttons
- `frontend/packages/control-centre/src/pages/Nominations.tsx` — **[NEW]** Filtered DSR list (type=NOMINATION) with DPDPA Section 14 explainer
- `frontend/packages/control-centre/src/services/dsr.ts` — **[MODIFIED]** Added `type` param to list()
- `frontend/packages/control-centre/src/hooks/useDSR.ts` — **[MODIFIED]** Added `type` param to useDSRs()
- `frontend/packages/control-centre/src/App.tsx` — **[MODIFIED]** Wired /reports, /nominations routes; replaced PlaceholderPage with ComingSoonPage for /agents, /users

**New Pages**:
- `/reports` — Compliance scorecard (SVG gauge), 4 pillar cards, recommendation alerts, 7 entity export buttons
- `/nominations` — Filtered DSR list with DPDPA S14 context
- `/agents` + `/users` — "Coming Soon" treatment (Phase 5 badge)

**QA**: Full sidebar click-through ✅ | CC build ✅ | Portal build ✅

🏁 **Phase 4 Complete!**
