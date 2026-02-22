# Phase 4 — Batch 4D Task Specifications

**Sprint**: Phase 4 — RoPA (Record of Processing Activities) + Multi-Level Purpose Tagging  
**Estimated Duration**: 2 days  
**Pre-requisites**: Batch 4C complete (retention CRUD live, scheduler running, all CC pages built)

> **v2 Corrections**: SCHEMA scope level added (5-level hierarchy). ThirdPartyRepository + table included in 4D-1 (not skipped). See `batch4d_agent_prompts.md` for authoritative sub-agent prompts.

---

## Execution Order

**Parallel Group 1** (no dependencies):
- Task 4D-1 (Backend) + Task 4D-3 (Backend) — can run in PARALLEL

**Sequential Group 2** (depends on 4D-1):
- Task 4D-2 (Frontend) — depends on 4D-1 (RoPA API)

**Sequential Group 3** (depends on 4D-3):
- Task 4D-4 (Frontend) — depends on 4D-3 (Purpose Assignment API)

**Standalone (anytime):**
- Task 4D-5 (Backend) — Bug fixes, can run in PARALLEL with anything

---

## Task 4D-1: Backend — RoPA Domain + Auto-Generation + Version Control API

**Agent**: Backend  
**Priority**: P0 (blocking — Frontend Task 4D-2 depends on this)  
**Depends On**: None  
**Estimated Effort**: Large (4-5h)

### Objective

Implement the RoPA (Record of Processing Activities) backend: domain entities, database migration, repository, service with auto-generation logic, and HTTP handler. The RoPA is auto-generated from existing data (purposes, data sources, data mappings, retention policies) and supports strict version control where every save creates a new version with an audit trail.

**Scope boundaries**: This task covers ONLY the backend. The frontend is Task 4D-2. Does NOT include PDF/export — that's Batch 4G.

### Context — Read These Files First
- `internal/domain/governance/entities.go` — Purpose, DataMapping, ThirdParty entities (data sources for RoPA)
- `internal/domain/compliance/retention.go` — RetentionPolicy entity (data source for RoPA)
- `internal/handler/retention_handler.go` — Pattern reference for CRUD handler with chi router
- `internal/service/retention_service.go` — Pattern reference for service with tenant context
- `internal/repository/postgres_retention.go` — Pattern reference for PostgreSQL repository
- `internal/handler/audit_handler.go` — Pattern reference for audit logging
- `cmd/api/routes.go` — Route mounting in `mountCCRoutes()` (line 33-131)
- `cmd/api/main.go` — Service/handler wiring
- `pkg/types/types.go` — `TenantEntity`, `ID`, `Pagination` types

### Requirements

#### 1. DB Migration — `internal/database/migrations/021_ropa.sql`

```sql
-- RoPA Versions table
CREATE TABLE IF NOT EXISTS ropa_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    version VARCHAR(20) NOT NULL,           -- semver: "1.0", "1.1", "2.0"
    generated_by VARCHAR(100) NOT NULL,     -- "auto" or user_id string
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT', -- DRAFT, PUBLISHED, ARCHIVED
    content JSONB NOT NULL,                 -- full RoPA content snapshot
    change_summary TEXT,                    -- what changed from previous version
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, version)
);

CREATE INDEX idx_ropa_versions_tenant ON ropa_versions(tenant_id);
CREATE INDEX idx_ropa_versions_tenant_version ON ropa_versions(tenant_id, version DESC);
```

#### 2. Domain Entity — `internal/domain/compliance/ropa.go`

```go
type RoPAVersion struct {
    ID            types.ID  `json:"id"`
    TenantID      types.ID  `json:"tenant_id"`
    Version       string    `json:"version"`       // semver string
    GeneratedBy   string    `json:"generated_by"`  // "auto" | user_id
    Status        string    `json:"status"`        // DRAFT, PUBLISHED, ARCHIVED
    Content       RoPAContent `json:"content"`
    ChangeSummary string    `json:"change_summary,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
}

type RoPAContent struct {
    OrganizationName string               `json:"organization_name"`
    GeneratedAt      time.Time            `json:"generated_at"`
    Purposes         []RoPAPurpose        `json:"purposes"`
    DataSources      []RoPADataSource     `json:"data_sources"`
    RetentionPolicies []RoPARetention     `json:"retention_policies"`
    ThirdParties     []RoPAThirdParty     `json:"third_parties"`
    DataCategories   []string             `json:"data_categories"`
}

type RoPAPurpose struct {
    ID          types.ID `json:"id"`
    Name        string   `json:"name"`
    Code        string   `json:"code"`
    LegalBasis  string   `json:"legal_basis"`
    Description string   `json:"description"`
    IsActive    bool     `json:"is_active"`
}

type RoPADataSource struct {
    ID       types.ID `json:"id"`
    Name     string   `json:"name"`
    Type     string   `json:"type"`
    IsActive bool     `json:"is_active"`
}

type RoPARetention struct {
    ID               types.ID `json:"id"`
    PurposeName      string   `json:"purpose_name"`
    MaxRetentionDays int      `json:"max_retention_days"`
    DataCategories   []string `json:"data_categories"`
    AutoErase        bool     `json:"auto_erase"`
}

type RoPAThirdParty struct {
    ID      types.ID `json:"id"`
    Name    string   `json:"name"`
    Type    string   `json:"type"`
    Country string   `json:"country"`
}

type RoPARepository interface {
    Create(ctx context.Context, version *RoPAVersion) error
    GetLatest(ctx context.Context, tenantID types.ID) (*RoPAVersion, error)
    GetByVersion(ctx context.Context, tenantID types.ID, version string) (*RoPAVersion, error)
    ListVersions(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[RoPAVersion], error)
    UpdateStatus(ctx context.Context, id types.ID, status string) error
}
```

#### 3. Repository — `internal/repository/postgres_ropa.go`
- Implement `RoPARepository` interface against PostgreSQL
- `Content` field stored as JSONB — use `json.Marshal/Unmarshal`
- `GetLatest` should `ORDER BY created_at DESC LIMIT 1`
- `ListVersions` should return versions ordered newest-first with pagination

#### 4. Service — `internal/service/ropa_service.go`

```go
type RoPAService struct {
    ropaRepo     compliance.RoPARepository
    purposeRepo  governance.PurposeRepository
    dsRepo       types.DataSourceRepository // or whatever interface exists
    retentionRepo compliance.RetentionPolicyRepository
    auditSvc     *AuditService
    logger       *slog.Logger
}
```

Key methods:
- `Generate(ctx context.Context) (*RoPAVersion, error)` — Auto-generate v1.0 (or next minor) from live data:
  1. Query all Purposes, DataSources, RetentionPolicies for tenant
  2. Build `RoPAContent` struct
  3. Determine version: if no existing versions → "1.0"; else bump minor (e.g., "1.3" → "1.4")
  4. Create RoPAVersion with `GeneratedBy: "auto"`, `Status: "DRAFT"`
  5. Create AuditLog entry
- `SaveEdit(ctx context.Context, content RoPAContent, changeSummary string) (*RoPAVersion, error)` — User edited fields:
  1. Bump minor version (e.g., "1.4" → "1.5")
  2. `GeneratedBy: userID` from context
  3. Create AuditLog entry
- `Publish(ctx context.Context, id types.ID) error` — Mark version as PUBLISHED, archive previous published
- `PromoteMajor(ctx context.Context) (*RoPAVersion, error)` — User-chosen major version bump (e.g., "1.5" → "2.0")
- `GetLatest(ctx)`, `GetByVersion(ctx, version)`, `ListVersions(ctx, pagination)` — passthrough to repo

**Version Logic:**
- Parse version string with `strings.Split(version, ".")` — major.minor
- Minor bump: increment minor
- Major bump: increment major, reset minor to 0

#### 5. Handler — `internal/handler/ropa_handler.go`

```go
type RoPAHandler struct {
    service *service.RoPAService
}
```

Routes (mounted at `/ropa`):
- `POST /` → Generate (auto-generate new version)
- `GET /` → GetLatest (latest version)
- `GET /versions` → ListVersions (paginated)
- `GET /versions/{version}` → GetByVersion
- `PUT /` → SaveEdit (body: `{ content: RoPAContent, change_summary: string }`)
- `POST /publish` → Publish (body: `{ id: "uuid" }`)
- `POST /promote` → PromoteMajor

Follow the handler pattern from `retention_handler.go`: chi router, `httputil.JSON()` responses, `httputil.ErrorFromDomain()` for errors.

#### 6. Wiring
- Wire `RoPARepo`, `RoPAService`, `RoPAHandler` in `cmd/api/main.go` (CC block)
- Add `ropaHandler *handler.RoPAHandler` param to `mountCCRoutes()` in `cmd/api/routes.go`
- Mount at `r.Mount("/ropa", ropaHandler.Routes())`

### Acceptance Criteria
- [ ] Migration `021_ropa.sql` creates `ropa_versions` table
- [ ] `POST /api/v2/ropa` auto-generates RoPA from live data
- [ ] `GET /api/v2/ropa` returns latest version
- [ ] `GET /api/v2/ropa/versions` returns paginated version history
- [ ] `PUT /api/v2/ropa` saves user edits as new minor version
- [ ] `POST /api/v2/ropa/publish` marks version as PUBLISHED
- [ ] `POST /api/v2/ropa/promote` creates major version bump
- [ ] Every version change creates an AuditLog entry
- [ ] All endpoints tenant-scoped
- [ ] `go build ./...` passes
- [ ] AGENT_COMMS.md updated with API contracts

### Integration Notes
- **Frontend** (Task 4D-2) will consume these APIs to build the RoPA page
- **Reports** (Batch 4G) will add PDF export on top of this data
- **Data sources for auto-generation**: Purposes (`purposeRepo.GetByTenant`), DataSources (`dsRepo.GetByTenant`), RetentionPolicies (`retentionRepo.GetByTenant`). ThirdParty repo may not exist yet — if no `ThirdPartyRepository` exists, leave that section empty in auto-generated content.

### Known Gotchas
- Use `types.TenantIDFromContext()` or `middleware.TenantIDFromContext()` for tenant extraction — check which one the handler context uses (recent handlers use `middleware.TenantIDFromContext`)
- Use `httputil.ParsePagination(r)` for pagination parsing — see `retention_handler.go` for pattern
- `Content` must be stored as JSONB — use `json.RawMessage` or encode/decode in repository layer
- Duplicate `RetentionRepo` instantiation exists in main.go (lines ~486 + ~568) — avoid creating another duplicate when wiring RoPAService

---

## Task 4D-2: Frontend — RoPA Page

**Agent**: Frontend  
**Priority**: P1  
**Depends On**: Task 4D-1 (RoPA API must be live)  
**Estimated Effort**: Medium (3-4h)

### Objective

Build the RoPA page at `/ropa` in Control Centre. Replace the existing placeholder. The page shows the auto-generated Record of Processing Activities with inline editing, version history, and publish workflow.

### Context — Read These Files First
- `frontend/packages/control-centre/src/pages/RetentionPolicies.tsx` — Pattern reference for CRUD page
- `frontend/packages/control-centre/src/pages/AuditLogs.tsx` — Pattern reference for data table + filters
- `frontend/packages/control-centre/src/services/retentionService.ts` — Pattern reference for API service
- `frontend/packages/control-centre/src/App.tsx` — Route replacement (line 122 has `/ropa` placeholder)
- `frontend/packages/control-centre/src/components/Layout/Sidebar.tsx` — Sidebar link (line 83, already exists)

### Reference Documentation
- `documentation/04_DataLens_SaaS_Application.md` — Control Centre modules

### Existing Code to Extend
- Replace the placeholder `<PlaceholderPage title="RoPA" />` at `/ropa` in `App.tsx` (line 122)
- Sidebar link already exists at line 83

### Requirements

#### 1. Service — `services/ropaService.ts`
```typescript
export const ropaService = {
  getLatest: () => api.get('/ropa'),
  generate: () => api.post('/ropa'),
  listVersions: (page: number, pageSize: number) =>
    api.get(`/ropa/versions?page=${page}&page_size=${pageSize}`),
  getVersion: (version: string) => api.get(`/ropa/versions/${version}`),
  saveEdit: (content: RoPAContent, changeSummary: string) =>
    api.put('/ropa', { content, change_summary: changeSummary }),
  publish: (id: string) => api.post('/ropa/publish', { id }),
  promote: () => api.post('/ropa/promote'),
};
```

#### 2. Page — `pages/RoPA.tsx`

**Layout**:
- Header with title "Record of Processing Activities"
- Version indicator badge showing current version + status (DRAFT/PUBLISHED/ARCHIVED)
- Action buttons: "Regenerate", "Save Changes", "Publish", "New Major Version"
- Main content area with collapsible sections for each RoPA section

**Sections** (each as a collapsible Card):
1. **Purposes** — Table: Name, Code, Legal Basis, Description, Active status
2. **Data Sources** — Table: Name, Type, Active status
3. **Retention Policies** — Table: Purpose, Max Days, Categories, Auto-Erase
4. **Third Parties** — Table: Name, Type, Country (may be empty)
5. **Data Categories** — Badge list

**Features**:
- "Regenerate" button calls `POST /ropa` and refreshes content
- Inline editing of text fields (Organization Name, descriptions) — track edited state
- "Save Changes" button calls `PUT /ropa` with edited content + change summary (Dialog prompt)
- "Publish" button calls `POST /ropa/publish` with confirmation dialog
- "New Major Version" button calls `POST /ropa/promote` with confirmation dialog
- Version History sidebar/drawer: list of all versions with timestamps, click to view any version

**States**:
- No versions yet → Show "Generate your first RoPA" CTA button
- DRAFT → Show "Publish" and "Save Changes" buttons
- PUBLISHED → Show "Regenerate" (creates new draft), "New Major Version"
- Loading → Spinner
- Error → Error state

#### 3. Update Route
- Replace placeholder import + route in `App.tsx`

#### 4. Use KokonutUI components
- `Card`, `Badge`, `Button`, `Dialog`, `Table`, `Collapsible` from `@datalens/shared`
- `toast` for success/error notifications
- Use `useQuery` + `useMutation` from `@tanstack/react-query`

### Acceptance Criteria
- [ ] `/ropa` shows auto-generated RoPA content
- [ ] "Regenerate" creates a new version
- [ ] Inline edits save as new minor version
- [ ] "Publish" marks version as PUBLISHED
- [ ] Version history shows all versions
- [ ] Placeholder replaced in App.tsx
- [ ] `npm run build -w @datalens/control-centre` passes

### Known Gotchas
- The API may return an empty/404 state when no RoPA versions exist — handle gracefully with CTA
- RoPAContent sections may be empty arrays — render appropriate empty states
- Use KokonutUI components from `@datalens/shared` — don't create ad-hoc styled divs

---

## Task 4D-3: Backend — Multi-Level Purpose Assignment (Domain + API)

**Agent**: Backend  
**Priority**: P0 (blocking — Frontend Task 4D-4 depends on this)  
**Depends On**: None  
**Estimated Effort**: Medium (3-4h)

### Objective

Implement multi-level purpose tagging: purposes can be assigned at COLUMN, TABLE, DATABASE, or SERVER scope, with inheritance (server-level cascades down unless overridden at lower level). This adds a `PurposeAssignment` entity and API endpoints alongside the existing `Purpose` CRUD.

**Scope boundaries**: This task covers the backend domain, migration, repository, service, and handler. The frontend enhancement is Task 4D-4. This does NOT modify existing Purpose CRUD — it adds a parallel assignment system.

### Context — Read These Files First
- `internal/domain/governance/entities.go` — Purpose entity (L21-30), PurposeRepository (L162-169)
- `internal/handler/purpose_handler.go` — Existing purpose CRUD handler (extend with assignment routes)
- `internal/service/purpose_service.go` — Existing service (add assignment methods)
- `internal/repository/postgres_purpose.go` — Existing repo (add assignment queries)
- `cmd/api/routes.go` — Purposes mounted at `/purposes` (line 70)

### Requirements

#### 1. DB Migration — `internal/database/migrations/022_purpose_assignments.sql`

```sql
-- Purpose Assignments (multi-level scope)
CREATE TABLE IF NOT EXISTS purpose_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    purpose_id UUID NOT NULL REFERENCES purposes(id) ON DELETE CASCADE,
    scope_type VARCHAR(20) NOT NULL,        -- COLUMN, TABLE, DATABASE, SERVER
    scope_id VARCHAR(500) NOT NULL,         -- identifier for the scope target
    scope_name VARCHAR(500),                -- human-readable name
    inherited BOOLEAN NOT NULL DEFAULT false,
    overridden_by UUID REFERENCES purpose_assignments(id),
    assigned_by UUID,                       -- user who assigned
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, purpose_id, scope_type, scope_id)
);

CREATE INDEX idx_purpose_assignments_tenant ON purpose_assignments(tenant_id);
CREATE INDEX idx_purpose_assignments_scope ON purpose_assignments(tenant_id, scope_type, scope_id);
CREATE INDEX idx_purpose_assignments_purpose ON purpose_assignments(purpose_id);
```

#### 2. Domain Entity — Add to `internal/domain/governance/entities.go`

```go
// PurposeAssignment represents a purpose assigned at a specific scope level.
// Supports inheritance: SERVER → DATABASE → TABLE → COLUMN
type PurposeAssignment struct {
    ID           types.ID  `json:"id"`
    TenantID     types.ID  `json:"tenant_id"`
    PurposeID    types.ID  `json:"purpose_id"`
    ScopeType    ScopeType `json:"scope_type"`
    ScopeID      string    `json:"scope_id"`
    ScopeName    string    `json:"scope_name,omitempty"`
    Inherited    bool      `json:"inherited"`
    OverriddenBy *types.ID `json:"overridden_by,omitempty"`
    AssignedBy   *types.ID `json:"assigned_by,omitempty"`
    AssignedAt   time.Time `json:"assigned_at"`
}

// ScopeType defines the level at which a purpose is assigned.
type ScopeType string

const (
    ScopeTypeColumn   ScopeType = "COLUMN"
    ScopeTypeTable    ScopeType = "TABLE"
    ScopeTypeDatabase ScopeType = "DATABASE"
    ScopeTypeServer   ScopeType = "SERVER"
)

// ScopeHierarchy defines the inheritance order (higher overrides lower).
var ScopeHierarchy = map[ScopeType]int{
    ScopeTypeServer:   0,
    ScopeTypeDatabase: 1,
    ScopeTypeTable:    2,
    ScopeTypeColumn:   3,
}

// PurposeAssignmentRepository defines persistence for purpose assignments.
type PurposeAssignmentRepository interface {
    Create(ctx context.Context, a *PurposeAssignment) error
    Delete(ctx context.Context, id types.ID) error
    GetByScope(ctx context.Context, tenantID types.ID, scopeType ScopeType, scopeID string) ([]PurposeAssignment, error)
    GetByPurpose(ctx context.Context, tenantID types.ID, purposeID types.ID) ([]PurposeAssignment, error)
    GetByTenant(ctx context.Context, tenantID types.ID) ([]PurposeAssignment, error)
    GetEffective(ctx context.Context, tenantID types.ID, scopeType ScopeType, scopeID string) ([]PurposeAssignment, error)
}
```

#### 3. Repository — `internal/repository/postgres_purpose_assignment.go`
- Implement `PurposeAssignmentRepository` interface
- `GetEffective` is the key method — it resolves inheritance:
  1. Check for direct assignments at the given scope
  2. If none (or partial), walk up the hierarchy (COLUMN → TABLE → DATABASE → SERVER)
  3. Mark inherited assignments with `Inherited: true`
  4. Exclude assignments that are overridden at a lower level
  
  **Implementation approach** (simpler for MVP):
  - Query all assignments for the tenant
  - Filter in Go by walking up the scope hierarchy
  - The `scope_id` hierarchy uses dot-delimited format: SERVER=`name`, DATABASE=`db`, SCHEMA=`db.schema`, TABLE=`db.schema.table`, COLUMN=`db.schema.table.col`

#### 4. Service — Add methods to `internal/service/purpose_service.go` (or create `internal/service/purpose_assignment_service.go`)

```go
type PurposeAssignmentService struct {
    assignmentRepo governance.PurposeAssignmentRepository
    purposeRepo    governance.PurposeRepository
    auditSvc       *AuditService
    logger         *slog.Logger
}
```

Methods:
- `Assign(ctx, input AssignPurposeInput) (*PurposeAssignment, error)` — Assign a purpose at a scope level. Create audit log.
- `Remove(ctx, id types.ID) error` — Remove an assignment. Create audit log.
- `GetByScope(ctx, scopeType, scopeID) ([]PurposeAssignment, error)` — Direct assignments at scope.
- `GetEffective(ctx, scopeType, scopeID) ([]PurposeAssignment, error)` — Resolved (inherited + direct).
- `GetByPurpose(ctx, purposeID) ([]PurposeAssignment, error)` — All scopes for a purpose.
- `GetAll(ctx) ([]PurposeAssignment, error)` — All assignments for tenant.

```go
type AssignPurposeInput struct {
    PurposeID types.ID
    ScopeType governance.ScopeType
    ScopeID   string
    ScopeName string
}
```

#### 5. Handler — Add routes to existing `purpose_handler.go` or create `internal/handler/purpose_assignment_handler.go`

New endpoints (mounted under `/purposes/assignments`):
- `POST /purposes/assignments` → Assign purpose to scope
- `DELETE /purposes/assignments/{id}` → Remove assignment
- `GET /purposes/assignments?scope_type=TABLE&scope_id=mydb.users` → Get by scope
- `GET /purposes/assignments/effective?scope_type=COLUMN&scope_id=mydb.users.email` → Get effective (resolved)
- `GET /purposes/assignments/all` → Get all for tenant

Or alternatively, create a new handler and mount at `/purpose-assignments`.

#### 6. Wiring
- Wire `PurposeAssignmentRepo`, `PurposeAssignmentService`, handler in `main.go`
- Mount assignment routes in `routes.go` — add to `mountCCRoutes()` params and mount

### Acceptance Criteria
- [ ] Migration `022_purpose_assignments.sql` creates table + indexes
- [ ] `POST /api/v2/purposes/assignments` assigns a purpose at a scope
- [ ] `GET /api/v2/purposes/assignments?scope_type=TABLE&scope_id=x` returns direct assignments
- [ ] `GET /api/v2/purposes/assignments/effective?scope_type=COLUMN&scope_id=x` resolves inheritance
- [ ] Server-level purpose cascades to database/table/column unless overridden
- [ ] Overriding at lower scope works correctly
- [ ] `DELETE` removes an assignment
- [ ] All endpoints tenant-scoped
- [ ] `go build ./...` passes
- [ ] AGENT_COMMS.md updated with API contracts

### Integration Notes
- **Frontend** (Task 4D-4) will build a scope selector UI on the Purpose Mapping page
- The `scope_id` format convention must be documented in AGENT_COMMS.md for the Frontend agent:
  - SERVER: `server_name`
  - DATABASE: `db_name`
  - SCHEMA: `db_name.schema_name`
  - TABLE: `db_name.schema_name.table_name`
  - COLUMN: `db_name.schema_name.table_name.column_name`

### Known Gotchas
- Do NOT modify the existing Purpose CRUD API — PurposeAssignment is a parallel system
- The inheritance resolution in `GetEffective` should be efficient — for MVP, in-memory resolution is fine
- Use `types.NewValidationError()` for invalid scope types
- The `overridden_by` column is nullable — set it when a lower-level scope explicitly overrides a higher-level assignment

---

## Task 4D-4: Frontend — Multi-Level Purpose Assignment UI

**Agent**: Frontend  
**Priority**: P1  
**Depends On**: Task 4D-3 (Purpose Assignment API must be live)  
**Estimated Effort**: Medium (3-4h)

### Objective

Enhance the existing Purpose Mapping page (`/governance/purposes`) to support multi-level purpose assignment. Add a scope selector UI and a visual hierarchy showing how purposes are assigned at SERVER → DATABASE → TABLE → COLUMN levels with inheritance indicators.

### Context — Read These Files First
- `frontend/packages/control-centre/src/pages/Governance/PurposeMapping.tsx` — Current page (AI suggestions only)
- `frontend/packages/control-centre/src/services/governance.ts` — Current governance service
- `frontend/packages/control-centre/src/types/governance.ts` — Current types
- `frontend/packages/control-centre/src/App.tsx` — Current route (line 106)

### Requirements

#### 1. Service — Add to `services/governance.ts` or create `services/purposeAssignmentService.ts`
```typescript
export const purposeAssignmentService = {
  assign: (data: { purpose_id: string; scope_type: string; scope_id: string; scope_name?: string }) =>
    api.post('/purposes/assignments', data),
  remove: (id: string) => api.delete(`/purposes/assignments/${id}`),
  getByScope: (scopeType: string, scopeId: string) =>
    api.get(`/purposes/assignments?scope_type=${scopeType}&scope_id=${scopeId}`),
  getEffective: (scopeType: string, scopeId: string) =>
    api.get(`/purposes/assignments/effective?scope_type=${scopeType}&scope_id=${scopeId}`),
  getAll: () => api.get('/purposes/assignments/all'),
};
```

#### 2. Enhance Page — `pages/Governance/PurposeMapping.tsx`

Add a **tabbed view** or **new section** below the existing AI suggestions:

**Tab 1: AI Suggestions** (existing content — keep as-is)
**Tab 2: Scope Assignments** (new):
- **Scope Selector**: Dropdown for scope type (Server, Database, Table, Column) + text input for scope ID
- **Assignment List**: Table showing current assignments at selected scope
  - Columns: Purpose Name, Scope Type, Scope ID, Inherited (badge), Assigned By, Date, Actions (Remove)
  - Inherited rows shown with a different visual style (muted + "↓ Inherited from [parent scope]")
- **Add Assignment**: Button opens a Dialog to pick a Purpose from existing purposes list + scope type/ID
- **Effective View Toggle**: Button/switch to show "effective" (inherited + direct) vs "direct only"
- **Scope Hierarchy Visualization**: Optional — show a simple breadcrumb or tree:
  ```
  🔵 Server: production-db
    └─ 🟢 Database: customers_db
       └─ 🟡 Table: users
          └─ 🟠 Column: email
  ```

#### 3. Types — Add to `types/governance.ts`
```typescript
export interface PurposeAssignment {
  id: string;
  tenant_id: string;
  purpose_id: string;
  scope_type: 'COLUMN' | 'TABLE' | 'SCHEMA' | 'DATABASE' | 'SERVER';

  scope_id: string;
  scope_name?: string;
  inherited: boolean;
  overridden_by?: string;
  assigned_by?: string;
  assigned_at: string;
}
```

#### 4. Use KokonutUI components
- `Tabs`, `TabsList`, `TabsTrigger`, `TabsContent` from `@datalens/shared`
- `Dialog`, `Select`, `Card`, `Badge`, `Button`, `Table` from `@datalens/shared`
- Install additional components if needed: `npx shadcn@latest add tabs`

### Acceptance Criteria
- [ ] Scope Assignments tab shows all assignments for a selected scope
- [ ] Assign a purpose to a scope type/ID works
- [ ] Remove an assignment works
- [ ] Effective view shows inherited assignments from parent scopes
- [ ] Inherited assignments visually distinct from direct assignments
- [ ] Existing AI Suggestions tab unchanged
- [ ] `npm run build -w @datalens/control-centre` passes

### Known Gotchas
- Don't break the existing AI Suggestions functionality
- The scope_id format is: SERVER=`name`, DATABASE=`db_name`, TABLE=`db_name.table_name`, COLUMN=`db_name.table_name.col_name`
- Use `toast` from `@datalens/shared` for success/error notifications
- If `tabs` component doesn't exist in shared, install it: `npx shadcn@latest add tabs`

---

## Task 4D-5: Backend — Bug Fixes (Pre-existing Issues)

**Agent**: Backend  
**Priority**: P2 (nice-to-have, non-blocking)  
**Depends On**: None  
**Estimated Effort**: Small (1h)

### Objective

Fix the three known issues carried forward from Batch 4C that affect code quality and may cause problems in future batches.

### Context — Read These Files First
- `cmd/api/main.go` — Lines ~486 + ~568 (duplicate RetentionRepo), lines ~556-557 (duplicate grievanceHandler)
- `internal/handler/admin_handler_test.go` — Test compilation issues
- `internal/handler/batch19_handler_test.go` — Test compilation issues

### Requirements

#### 1. Remove duplicate `RetentionRepo` instantiation in `main.go`
- Find the two locations where `RetentionRepo` is created (approximately lines 486 and 568)
- Remove the duplicate, keep the one in the correct scope
- Ensure all references use the single instance

#### 2. Remove duplicate `grievanceHandler` assignment in `main.go`
- Find the duplicate assignment at approximately lines 556-557
- Remove the duplicate line

#### 3. Fix test compilation errors (if time permits)
- `admin_handler_test.go` and `batch19_handler_test.go` may have interface drift from 4C-1 changes
- Update mock types to include new repository methods added in Batch 4C-1
- These are test-only fixes — no production code changes needed

### Acceptance Criteria
- [ ] No duplicate `RetentionRepo` instantiation
- [ ] No duplicate `grievanceHandler` assignment
- [ ] `go build ./...` passes
- [ ] `go vet ./...` clean
- [ ] Test compilation fixed (best effort)

---

## Summary

| Task | Agent | Priority | Depends On | Effort |
|------|-------|----------|------------|--------|
| 4D-1: RoPA Backend | Backend | P0 | None | Large (4-5h) |
| 4D-2: RoPA Frontend | Frontend | P1 | 4D-1 | Medium (3-4h) |
| 4D-3: Purpose Assignment Backend | Backend | P0 | None | Medium (3-4h) |
| 4D-4: Purpose Assignment Frontend | Frontend | P1 | 4D-3 | Medium (3-4h) |
| 4D-5: Bug Fixes | Backend | P2 | None | Small (1h) |

**Parallelism:**
- Group 1: 4D-1 + 4D-3 + 4D-5 (all Backend, can run PARALLEL)
- Group 2: 4D-2 (depends on 4D-1) + 4D-4 (depends on 4D-3) — can run PARALLEL with each other
