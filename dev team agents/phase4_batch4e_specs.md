# Phase 4 — Batch 4E Task Specifications

**Sprint**: Phase 4 — Department Ownership + Third-Party Management + Nominations  
**Estimated Duration**: 2 days  
**Pre-requisites**: Batch 4D complete (ThirdParty repo/table exist, RoPA live)

---

## Execution Order

**Parallel Group 1** (no dependencies):
- Task 4E-1 (Backend) + Task 4E-2 (Backend) — can run in PARALLEL

**Sequential Group 2** (depends on 4E-1 + 4E-2):
- Task 4E-3 (Frontend) — depends on 4E-1 + 4E-2 (all APIs)

---

## Task 4E-1: Backend — Department Ownership + Notifications

**Agent**: Backend  
**Priority**: P0 (blocking)  
**Depends On**: None  
**Estimated Effort**: Large (4-5h)

### Objective

Implement a Department entity system for organizational ownership of data and compliance tasks. Each department has an owner, responsibilities, and can receive email notifications for relevant events (e.g., new DSR assigned, retention policy expiring, policy violation).

### Context — Read These Files First
- `internal/domain/governance/entities.go` — Pattern for domain entities with TenantEntity
- `internal/service/notification_service.go` — Has `DispatchNotification()` and `sendEmail()` via SMTP
- `internal/handler/retention_handler.go` — Pattern for CRUD handler
- `internal/service/retention_service.go` — Pattern for tenant-scoped service
- `internal/service/audit_service.go` — AuditService.Log() signature
- `cmd/api/routes.go` — Route mounting
- `cmd/api/main.go` — Service wiring

### Requirements

#### 1. Migration: `internal/database/migrations/023_departments.sql`

```sql
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID REFERENCES users(id),
    owner_name VARCHAR(255),
    owner_email VARCHAR(255),
    responsibilities TEXT[],           -- Array of responsibility strings
    notification_enabled BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);
CREATE INDEX idx_departments_tenant ON departments(tenant_id);
CREATE INDEX idx_departments_owner ON departments(owner_id);
```

#### 2. Domain Entity: `internal/domain/governance/department.go` [NEW]

```go
type Department struct {
    types.TenantEntity
    Name                string    `json:"name" db:"name"`
    Description         string    `json:"description,omitempty" db:"description"`
    OwnerID             *types.ID `json:"owner_id,omitempty" db:"owner_id"`
    OwnerName           string    `json:"owner_name,omitempty" db:"owner_name"`
    OwnerEmail          string    `json:"owner_email,omitempty" db:"owner_email"`
    Responsibilities    []string  `json:"responsibilities" db:"responsibilities"`
    NotificationEnabled bool      `json:"notification_enabled" db:"notification_enabled"`
    IsActive            bool      `json:"is_active" db:"is_active"`
}

type DepartmentRepository interface {
    Create(ctx context.Context, d *Department) error
    GetByID(ctx context.Context, id types.ID) (*Department, error)
    GetByTenant(ctx context.Context, tenantID types.ID) ([]Department, error)
    GetByOwner(ctx context.Context, ownerID types.ID) ([]Department, error)
    Update(ctx context.Context, d *Department) error
    Delete(ctx context.Context, id types.ID) error
}
```

#### 3. Repository: `internal/repository/postgres_department.go` [NEW]
- Standard CRUD implementation
- `Responsibilities` stored as PostgreSQL TEXT array (`pq.Array` or `pgx` array)
- `GetByOwner` filters by `owner_id`

#### 4. Service: `internal/service/department_service.go` [NEW]

```go
type DepartmentService struct {
    repo          governance.DepartmentRepository
    notifSvc      *NotificationService
    auditSvc      *AuditService
    logger        *slog.Logger
}
```

Methods:
- `Create`, `GetByID`, `List` (by tenant), `Update`, `Delete` — standard CRUD with tenant context
- `NotifyDepartment(ctx, departmentID, subject, body)` — Send email to department owner if notification_enabled is true. Uses `notifSvc.sendEmail()` or `DispatchNotification()` pattern.
- Audit log on create/update/delete

#### 5. Handler: `internal/handler/department_handler.go` [NEW]

Routes (mounted at `/departments`):
- `POST /` → Create
- `GET /` → List (by tenant)
- `GET /{id}` → GetByID
- `PUT /{id}` → Update
- `DELETE /{id}` → Delete
- `POST /{id}/notify` → NotifyDepartment (body: `{ subject, body }`)

#### 6. Wiring: `cmd/api/main.go` + `cmd/api/routes.go`
- Wire DepartmentRepo, DepartmentService, DepartmentHandler
- Mount at `r.Mount("/departments", departmentHandler.Routes())`

### Acceptance Criteria
- [ ] Migration `023_departments.sql` creates table + indexes
- [ ] CRUD endpoints functional (6 endpoints at `/departments`)
- [ ] Notify endpoint sends email to department owner
- [ ] Tenant-scoped, audit-logged
- [ ] `go build ./...` passes

---

## Task 4E-2: Backend — Third-Party Service + Handler + DPA Tracking

**Agent**: Backend  
**Priority**: P0 (blocking)  
**Depends On**: None (repo already exists from 4D-1)  
**Estimated Effort**: Medium (3h)

### Objective

Build a service and handler for the existing ThirdParty repository (created in Batch 4D-1). Add DPA (Data Processing Agreement) tracking with status, dates, and document path. The entity struct already has `DPADocPath` — extend it to support full DPA tracking lifecycle.

### Context — Read These Files First
- `internal/domain/governance/entities.go` — ThirdParty entity (L138-155), ThirdPartyRepository (added in 4D-1)
- `internal/repository/postgres_third_party.go` — Already built in 4D-1 (CRUD complete)
- `internal/handler/retention_handler.go` — Pattern for CRUD handler
- `cmd/api/routes.go` — Route mounting

### Requirements

#### 1. Migration: `internal/database/migrations/024_third_party_dpa.sql` [NEW]

Add DPA tracking columns to existing `third_parties` table:
```sql
ALTER TABLE third_parties
    ADD COLUMN IF NOT EXISTS dpa_status VARCHAR(20) DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS dpa_signed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS dpa_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS dpa_notes TEXT,
    ADD COLUMN IF NOT EXISTS contact_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS contact_email VARCHAR(255);
```

#### 2. Extend Domain: Modify `internal/domain/governance/entities.go`

Add fields to `ThirdParty` struct:
```go
DPAStatus    string     `json:"dpa_status" db:"dpa_status"`        // NONE, PENDING, SIGNED, EXPIRED
DPASignedAt  *time.Time `json:"dpa_signed_at,omitempty" db:"dpa_signed_at"`
DPAExpiresAt *time.Time `json:"dpa_expires_at,omitempty" db:"dpa_expires_at"`
DPANotes     string     `json:"dpa_notes,omitempty" db:"dpa_notes"`
ContactName  string     `json:"contact_name,omitempty" db:"contact_name"`
ContactEmail string     `json:"contact_email,omitempty" db:"contact_email"`
```

DPA Status constants:
```go
const (
    DPAStatusNone    = "NONE"
    DPAStatusPending = "PENDING"
    DPAStatusSigned  = "SIGNED"
    DPAStatusExpired = "EXPIRED"
)
```

#### 3. Update Repository: `internal/repository/postgres_third_party.go` [MODIFY]
- Update all SQL queries to include new columns
- Update `Create`, `GetByID`, `GetByTenant`, `Update` scan/insert lists

#### 4. Service: `internal/service/third_party_service.go` [NEW]

```go
type ThirdPartyService struct {
    repo     governance.ThirdPartyRepository
    auditSvc *AuditService
    logger   *slog.Logger
}
```

Methods: `Create`, `GetByID`, `List`, `Update`, `Delete` — standard tenant-scoped CRUD with audit logging.

#### 5. Handler: `internal/handler/third_party_handler.go` [NEW]

Routes (mounted at `/third-parties`):
- `POST /` → Create
- `GET /` → List (by tenant)
- `GET /{id}` → GetByID
- `PUT /{id}` → Update
- `DELETE /{id}` → Delete

#### 6. Wiring
- ThirdPartyRepo was already wired in main.go for RoPAService (from 4D-1). Reuse it.
- Wire ThirdPartyService + ThirdPartyHandler
- Mount at `r.Mount("/third-parties", thirdPartyHandler.Routes())`

### Acceptance Criteria
- [ ] Migration `024_third_party_dpa.sql` adds DPA columns
- [ ] CRUD endpoints functional (5 endpoints at `/third-parties`)
- [ ] DPA tracking fields included in create/update/response
- [ ] Existing ThirdParty repo updated for new columns
- [ ] `go build ./...` passes

---

## Task 4E-3: Frontend — Department + Third-Party Pages + Nomination Link

**Agent**: Frontend  
**Priority**: P1  
**Depends On**: Tasks 4E-1 + 4E-2  
**Estimated Effort**: Large (4-5h)

### Objective

Build CRUD pages for Departments and Third-Party management in Control Centre. Add sidebar links and routes.

### Context — Read These Files First
- `frontend/packages/control-centre/src/pages/RetentionPolicies.tsx` — Pattern for CRUD page
- `frontend/packages/control-centre/src/services/retentionService.ts` — Pattern for service
- `frontend/packages/control-centre/src/App.tsx` — Add routes
- `frontend/packages/control-centre/src/components/Layout/Sidebar.tsx` — Add sidebar links

### Requirements

#### 1. Department Service + Page
- `services/departmentService.ts` [NEW] — CRUD + notify endpoint
- `pages/Departments.tsx` [NEW]:
  - Table: Name, Owner, Email, Responsibilities (badge list), Notifications (toggle), Actions
  - Create/Edit modal: Name, Description, Owner Name, Owner Email, Responsibilities (tag input), Notification toggle
  - Notify button → opens Dialog: Subject + Body → POST `/departments/{id}/notify`
  - Empty state if no departments

#### 2. Third-Party Service + Page
- `services/thirdPartyService.ts` [NEW] — CRUD
- `pages/ThirdParties.tsx` [NEW]:
  - Table: Name, Type (PROCESSOR/CONTROLLER/VENDOR badge), Country, DPA Status badge, Contact, Actions
  - DPA Status badges: NONE (gray), PENDING (yellow), SIGNED (green), EXPIRED (red)
  - Create/Edit modal: Name, Type select, Country, DPA fields (status, signed date, expiry, doc path, notes), Contact name/email, Purpose IDs (multi-select from purposes API)
  - Toggle between "Simple View" (name/type/country) and "Full DPA View" (all fields)

#### 3. Routes + Sidebar
- `App.tsx`: Add routes for `/departments` and `/third-parties`
- `Sidebar.tsx`: Add links under appropriate sections:
  - "Departments" link (Building2 icon) under Governance or Organization
  - "Third Parties" link (Globe icon) under Governance

#### 4. Nomination Link (Quick Win)
- Add a "Nominations" sidebar link → routes to existing DSR page filtered by type=NOMINATION
- Or show a small info card on the DSR page explaining nomination rights (DPDPA S14)

### Acceptance Criteria
- [ ] `/departments` page with full CRUD
- [ ] `/third-parties` page with full CRUD + DPA tracking
- [ ] Sidebar links for both pages
- [ ] DPA status badges with correct colors
- [ ] Notify department sends email notification
- [ ] `npm run build -w @datalens/control-centre` passes

---

## Summary

| Task | Agent | Priority | Depends On | Effort |
|------|-------|----------|------------|--------|
| 4E-1: Department Backend | Backend | P0 | None | Large (4-5h) |
| 4E-2: Third-Party Service + DPA | Backend | P0 | None | Medium (3h) |
| 4E-3: Frontend Pages | Frontend | P1 | 4E-1 + 4E-2 | Large (4-5h) |

**Parallelism:**
- Group 1: 4E-1 + 4E-2 (both Backend, run PARALLEL)
- Group 2: 4E-3 (after both backend tasks complete)
