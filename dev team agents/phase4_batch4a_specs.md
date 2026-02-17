# Phase 4 — Batch 4A Task Specifications

**Sprint**: Phase 4 — Foundation Fixes  
**Estimated Duration**: 1 day  
**Pre-requisites**: Phase 3A/3B/3C complete

---

## Task 4A-1: Consolidate DB Migrations

**Agent**: Backend  
**Effort**: 2h  
**Input**: Existing entity definitions from Phase 3A

### What to Build
Create a single consolidated migration file that includes all tables defined in Phase 3A but not yet migrated:

1. `dpo_contacts` — DPO Contact entity (from `dpo_service.go`)
2. `retention_policies` — Retention policy rules (from `internal/domain/governance/retention.go`)
3. `retention_logs` — Retention execution log
4. `breach_notifications` — User-facing breach notification inbox
5. `notice_schema_fields` — DPDP Schedule I compliance fields

### Reference Files
- `internal/domain/governance/retention.go` — RetentionPolicy, RetentionLog entities
- `internal/domain/breach/entities.go` — BreachNotification entity
- `internal/handler/dpo_handler.go` — DPO Contact entity shape
- `internal/database/migrations/` — existing migration numbering

### Acceptance Criteria
- [ ] New migration file follows append-only pattern with correct sequence number
- [ ] `go build ./...` passes
- [ ] All existing tests still pass
- [ ] Tables include proper foreign keys, indexes, and `tenant_id` columns

---

## Task 4A-2: Audit Log CC Handler

**Agent**: Backend  
**Effort**: 2h  
**Input**: Existing `AuditService` + `AuditLog` entity

### What to Build
Expose audit logs to the Control Centre via a new paginated API:

**Endpoint**: `GET /api/v2/audit-logs`  
**Query Params**: `page`, `page_size`, `entity_type`, `action`, `user_id`, `start_date`, `end_date`  
**Response**: `PaginatedResponse<AuditLog>`

### Implementation Steps
1. Create `audit_handler.go` in `internal/handler/` following the standard handler pattern
2. Add `ListByTenant(ctx, tenantID, filters, pagination)` to `AuditService` if not present
3. Add corresponding repository method with tenant-scoped query
4. Mount route in `cmd/api/routes.go` → `mountCCRoutes()`

### Reference Files
- `internal/domain/audit/entities.go` — AuditLog entity
- `internal/service/audit_service.go` — existing service
- `internal/handler/dsr_handler.go` — pagination pattern reference

### Acceptance Criteria
- [ ] `GET /api/v2/audit-logs` returns paginated results
- [ ] All filters work (entity_type, action, date range)
- [ ] Tenant-scoped (no cross-tenant leakage)
- [ ] `go build ./...` passes

---

## Task 4A-3: Route Dedup + Cleanup

**Agent**: Frontend  
**Effort**: 1h  
**Input**: Current `App.tsx` routes in Control Centre

### What to Build
Clean up duplicate/confusing routes in the CC frontend:

1. Remove standalone `/grievances` → users should use `/compliance/grievances`
2. Remove standalone `/users` → redirect to Settings section or remove if already in Admin
3. Remove standalone `/consent` placeholder (will be re-added in Batch 4C as real page)
4. Verify all sidebar nav links match actual routes
5. Remove `PlaceholderPage` component definition (will no longer be needed after Phase 4)

### Reference Files
- `frontend/packages/control-centre/src/App.tsx` — current routes
- `frontend/packages/shared/src/components/Layout/Sidebar.tsx` — nav structure

### Acceptance Criteria
- [ ] No duplicate routes for same functionality
- [ ] Sidebar links match routes
- [ ] `npm run build -w @datalens/control-centre` passes

---

## Task 4A-4: Build Verification

**Agent**: QA  
**Effort**: 0.5h

### What to Verify
After tasks 4A-1 through 4A-3 are complete:

```powershell
# Backend
go build ./...
go vet ./...
go test ./... -short -count=1

# Frontend (all 3 apps)
cd frontend
npm run build -w @datalens/control-centre
npm run build -w @datalens/admin
npm run build -w @datalens/portal
```

### Acceptance Criteria
- [ ] All builds pass with zero errors
- [ ] All existing tests pass
- [ ] Post results to `AGENT_COMMS.md`
