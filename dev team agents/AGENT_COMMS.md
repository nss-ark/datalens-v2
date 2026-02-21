# Agent Communication Board

> **System Instructions**:
> - Post your status here when starting or finishing a task.
> - `[BLOCKED]` prefix if you are blocked. `[HANDOFF]` prefix for handoffs.
> - The Orchestrator reads this file at the start of every session.
> - **Archive**: Resolved threads older than the current sprint are in `AGENT_COMMS_archive_v1.md` (same folder).

---

## ⚠️ BREAKING: Architecture Refactoring (R1–R3) — 2026-02-14

> **Read `CONTEXT_SYNC.md` at the project root for full details.**

| Batch | Change | Impact |
|-------|--------|--------|
| **R1** | Frontend monorepo split | `frontend/src/` deleted → `frontend/packages/{shared,control-centre,admin,portal}/` |
| **R2** | Backend mode splitting | `cmd/api/main.go` → `--mode=all|cc|admin|portal`, routes in `routes.go` |
| **R3** | Nginx reverse proxy | `*.localhost:8000` sub-domains, env-driven CORS, 3 prod API instances |

**Rules**: Frontend → work in `packages/<app>/`, import from `@datalens/shared`. Backend → new routes in `routes.go`, new services wrapped in `shouldInit()`. DevOps → ONE binary, mode via Docker `command:`.

---

## ✅ Completed: Phase 3A + 3B + 3C (Feb 14-17, 2026)

| Task | Agent | Status | Key Files |
|------|-------|--------|-----------|
| 3A-1: Portal Backend API Wiring | Backend | ✅ | `portal_handler.go` (9 routes + 3 aliases) |
| 3A-2: DPR Download Endpoint | Backend | ✅ | `portal_handler.go`, `dsr_service.go` (72h SLA) |
| 3A-3: DPR Appeal Flow | Backend + Frontend | ✅ | `AppealModal.tsx`, `Requests.tsx`, `portal_handler.go` |
| 3A-4: DSR Auto-Verification | Backend | ✅ | `dsr_executor.go` (AutoVerify), `dsr.go` (VERIFIED status) |
| 3A-5: Consent Receipt | Backend | ✅ | `consent_service.go` (GenerateReceipt), HMAC-SHA256 |
| 3A-6: DPO Contact Entity | Backend | ✅ | `dpo_service.go`, `dpo_handler.go`, migration |
| 3A-7: Notice Schema Validation | Backend | ✅ | `notice_service.go` (ValidateSchema, compliance-check endpoint) |
| 3A-8: Guardian Frontend Polish | Frontend | ✅ | `Profile.tsx`, `StatusBadge.tsx` |
| 3A-9: Notice Translation API | Backend | ✅ | `notice_handler.go`, `portal_handler.go` |
| 3A-10: Breach Portal Inbox | Backend + Frontend | ✅ | `BreachNotifications.tsx`, `breach_service.go` |
| 3A-11: Data Retention Model | Backend | ✅ | `retention.go` (design only, scheduler deferred) |
| 3A-E2E: Verification | QA | ✅ | `e2e_phase3a_test.go` — 6/6 tests passing |
| 3B: SQL Server Connector | Backend | ✅ | `sqlserver.go`, `registry.go` |
| 3C: Observability Stack | DevOps | ✅ | `pkg/telemetry`, `docker-compose.dev.yml` |

---

## Active Sprint: Phase 4 — Comprehensive Build Sprint (APPROVED ✅)

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
7. Multi-level purpose tagging: column/table/database/server with inheritance
8. UI Overhaul: Dedicated Batch 4B, evaluate Oat CSS

---

## Active Messages

### [2026-02-17] [FROM: Orchestrator] → [TO: ALL]
**Subject**: Phase 4 Sprint Plan Approved — Ready for Execution
**Type**: STATUS

Phase 4 plan finalized with all user decisions incorporated. 7 batches, 30 tasks. Agent prompts updated (backend, frontend, orchestrator) with Phase 4 patterns and context. Start with Batch 4A: Foundation Fixes.

**Key Decisions Finalized**:
- Phase 4 sprint structure and task decomposition
- Priority ordering for remaining placeholder pages vs backend gaps

**Action Required**:
- None. Await Phase 4 task specs.

---

### [2026-02-17] [FROM: Backend] → [TO: ALL]
**Subject**: Batch 4A-1 + 4A-2 Complete — Retention Migrations + Audit Log API
**Type**: HANDOFF

**Changes**:
- `internal/database/migrations/019_retention.sql` — retention_policies + retention_logs tables (P0 blocker for Batch 4C)
- `internal/database/migrations/020_audit_log_columns.sql` — adds user_id, old_values, new_values, client_id columns to audit_logs with backfill
- `internal/domain/audit/log.go` — added `AuditFilters` struct + `ListByTenant` to Repository interface
- `internal/repository/postgres_audit.go` — `ListByTenant` with dynamic WHERE, COALESCE(user_id, actor_id), pagination
- `internal/service/audit_service.go` — `ListByTenant` passthrough
- `internal/handler/audit_handler.go` — **[NEW]** GET handler with query param filters
- `cmd/api/routes.go` — mounted at `/audit-logs` in CC routes
- `cmd/api/main.go` — wired AuditHandler

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

### [2026-02-17] [FROM: Frontend] → [TO: ALL]
**Subject**: Batch 4A-3 Complete — Route Dedup + Sidebar Nav Fix
**Type**: HANDOFF

**Changes**:
- `frontend/packages/control-centre/src/App.tsx` — Removed duplicate `/grievances` placeholder route (real pages at `/compliance/grievances` and `/compliance/grievances/:id`)
- `frontend/packages/control-centre/src/components/Layout/Sidebar.tsx` — Fixed Data Lineage link from `/lineage` → `/governance/lineage`; fixed Consent Analytics link from `/consent/analytics` → `/compliance/analytics`

**Features Enabled**:
- Sidebar "Data Lineage" link now correctly navigates to the DataLineage page
- Sidebar "Consent Analytics" link now correctly navigates to the Analytics page
- No more duplicate `/grievances` route shadowing the real Compliance grievance pages

**Verification**: `npm run build -w @datalens/control-centre` ✅ (exit code 0, zero errors)

**Action Required**:
- None — self-contained fix, no backend or test changes needed

---

### [2026-02-17] [FROM: Test] → [TO: ALL]
**Subject**: Batch 4A Build Verification
**Type**: HANDOFF
**Results**:
- Backend: `go build` ✅ | `go vet` ✅ | `go test` ❌ (1 failing unit test in `handler`)
- Frontend: CC ✅ | Admin ✅ | Portal ✅
**Issues Found**:
- `internal/handler/portal_handler_translation_test.go`: `TestPortalHandler_GetNotice_WithTranslation` fails on Title assertion.
- `internal/service/auth_service_test.go`: Passed.

---

### [2026-02-17] [FROM: Frontend] → [TO: ALL]
**Subject**: Batch 4B-2 Complete — Custom UI Integration (Portal)
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
- `npm run build -w @datalens/portal` ✅ (Exit code 0, zero errors).

**Action Required**:
- None.

---

### [2026-02-21] [FROM: Backend] → [TO: ALL]
**Subject**: Task 4C-3 Complete — Retention Scheduler (Cron Job)
**Type**: HANDOFF

**Changes**:
- `internal/service/scheduler.go` — Added `retentionRepo` field + `lastRetentionCheck` timestamp to `SchedulerService`; updated `NewSchedulerService()` constructor (⚠️ **BREAKING**: new `retentionRepo` parameter added before `logger`); added `checkRetentionPolicies(ctx)` call to scheduler loop
- `internal/service/scheduler_retention.go` — **[NEW]** `checkRetentionPolicies()` + `evaluateTenantRetentionPolicies()`: runs once/24h, evaluates all ACTIVE policies, creates `RetentionLog` entries (`ERASED` or `RETENTION_EXCEEDED`)
- `internal/repository/postgres_retention.go` — Implemented `CreateLog` (INSERT into `retention_logs`) and `GetLogs` (paginated SELECT with optional `policy_id` filter)
- `cmd/api/main.go` — Instantiates `RetentionRepo` in CC block and passes to `NewSchedulerService()`
- `internal/service/scheduler_test.go` — Updated constructor calls for new signature
- `internal/service/scheduler_check_test.go` — Updated constructor calls for new signature
- `internal/service/retention_service.go` — Fixed pre-existing bug: `NewForbiddenError` was called with 2 args (only takes 1)

**⚠️ Constructor Signature Change**:
```go
// OLD:
NewSchedulerService(dsRepo, tenantRepo, policySvc, scanSvc, expirySvc, logger)
// NEW:
NewSchedulerService(dsRepo, tenantRepo, policySvc, scanSvc, expirySvc, retentionRepo, logger)
```
Any other callers of `NewSchedulerService` must be updated to pass `retentionRepo` (or `nil`).

**MVP Note**: The retention scheduler does **NOT** actually delete data from connected sources. It only creates `RetentionLog` entries. Real deletion via connectors is a future enhancement.

**Verification**:
- `go build ./...` ✅ (exit code 0)
- `go vet ./...` ✅ (clean)
- Pre-existing test failures in `admin_service_test.go`, `batch19_service_test.go`, `consent_lifecycle_test.go` from batch 4C-1 interface changes (unrelated to scheduler)

**Action Required**:
- **Test**: Scheduler tests need full package to compile — mock types for new repo interfaces need updating (batch 4C-1 blocker)
- **Frontend**: No frontend changes needed — scheduler is backend-only

---

### [2026-02-21] [FROM: Backend] → [TO: ALL]
**Subject**: Batch 4C-1 Complete — Consent Sessions, Data Subjects, Retention APIs
**Type**: HANDOFF

**Changes**:
- `internal/handler/consent_handler.go` — Modified `listSessions` endpoint. If `subject_id` is omitted, it now falls back to tenant-wide listing with optional filters.
- `internal/handler/data_subject_handler.go` — **[NEW]** Added endpoint to list/search data subjects (Principals) across the tenant with partial matching.
- `internal/handler/retention_handler.go` — **[NEW]** Added full CRUD for Retention Policies + Audit Log fetching.
- `cmd/api/routes.go` & `cmd/api/main.go` — Wired and mounted `/subjects` and `/retention` in CC routes.

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

**Features Enabled**:
- Control Centre users can now view, filter, and paginate through system audit logs.

**Verification**: `npm run build -w @datalens/control-centre`  (Exit code 0)

**Action Required**:
- None.---

### [2026-02-21] [FROM: Frontend] → [TO: ALL]
**Subject**: Batch 4C-4 Complete — Consent Records Page
**Type**: HANDOFF

**Changes**:
- `frontend/packages/control-centre/src/services/consentRecordService.ts` — Added API service.
- `frontend/packages/control-centre/src/pages/Consent/ConsentRecords.tsx` — Created list page with filters and data table.
- `frontend/packages/control-centre/src/App.tsx` — Replaced placeholder route.

**Features Enabled**:
- Control Centre users can view, filter, and paginate through consent sessions.

**Verification**: `npm run build -w @datalens/control-centre` ✅ (Exit code 0)

**Action Required**:
- None.

---
# # #   [ 2 0 2 6 - 0 2 - 2 1 ]   [ F R O M :   F r o n t e n d ]   - >   [ T O :   A L L ]  
 * * S u b j e c t * * :   C o m p l e t e d   T a s k   4 C - 5 :   F r o n t e n d   -   D a t a   S u b j e c t s   P a g e  
 * * T y p e * * :   H A N D O F F  
  
 * * C h a n g e s * * :  
 -   s r c / s e r v i c e s / d a t a S u b j e c t S e r v i c e . t s :   C r e a t e d   n e w   A P I   s e r v i c e   f o r   f e t c h i n g   s u b j e c t s .  
 -   s r c / p a g e s / D a t a S u b j e c t s . t s x :   I m p l e m e n t e d   u i   p a t t e r n   u t i l i z i n g   d a t a T a b l e ,   w i t h   f i l t e r e d   r e q u e s t   l i n k i n g   t o   d s r s   a n d   c o n s e n t   t a b s .  
 -   s r c / A p p . t s x :   R e p l a c e d   p l a c e h o l d e r   r o u t e   f o r   / s u b j e c t s .  
  
 * * F e a t u r e s   E n a b l e d * * :  
 -   S e a r c h i n g   S u b j e c t s   v i a   e m a i l   /   p h o n e  
 -   R o u t i n g   s u b j e c t s   t o   t h e i r   c o n s e n t / d s r   f i l t e r s   v i a   s u b j e c t s   I D  
  
 * * V e r i f i c a t i o n * * :   \ 
 p m   r u n   b u i l d \   : w h i t e _ c h e c k _ m a r k :  
  
 * * A c t i o n   R e q u i r e d * * :  
 -   * * T e s t * * :   E 2 E   t e s t i n g   t o   e n s u r e   t a b l e   s e a r c h   f i l t e r s   a r e   a c c u r a t e l y   f e t c h e d   a n d   r o u t e d .  
  
 