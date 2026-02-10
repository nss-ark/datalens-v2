# Agent Communication Board

> **Purpose**: This file is the shared message board for inter-agent communication. All agents read and write to this file to stay aligned. The human router facilitates by copying relevant messages between agent sessions.

---

## How To Use

### Posting a Message
When you need to communicate with another agent, add an entry under the appropriate section below using this format:

```markdown
### [TIMESTAMP] [FROM: Agent Name] → [TO: Agent Name or ALL]
**Subject**: Brief topic
**Type**: INFO | REQUEST | HANDOFF | BLOCKER | QUESTION

[Your message here — be specific and concise]

**Action Required**: [What the target agent needs to do, or "None — FYI only"]
```

### Reading Messages
At the start of every task, check this file for:
1. Messages addressed to **YOU** or **ALL**
2. Any **BLOCKER** type messages
3. Recent **HANDOFF** messages that affect your work

### Message Lifecycle
- Messages stay in this file for the current sprint
- The Orchestrator clears resolved messages at sprint boundaries
- Mark resolved messages with ~~strikethrough~~ and add resolution notes

---

## Active Messages

### [2026-02-10 14:20 IST] [FROM: Orchestrator] → [TO: ALL]
**Subject**: Sprint Batch 1 Kick-off — 5 tasks dispatched
**Type**: INFO

Batch 1 has been planned. Each agent should check their assigned task spec. Key reminders:
- **Check AGENT_COMMS.md before starting** for any blockers
- **Post a HANDOFF message here when you complete your task** with: what was created, files changed, any issues found
- **Post API contracts here** (Backend) so Frontend and Test agents can consume them

**Action Required**: Read your task spec, then begin work.
~~Batch 1 has been planned...~~ [Resolved by Frontend Agent]

---

### [2026-02-10 15:30 IST] [FROM: Frontend] → [TO: ALL]
**Subject**: Frontend Scaffolding Complete
**Type**: HANDOFF

**Changes**:
- Initialized React 18 + TypeScript + Vite project in `frontend/`
- Set up Design System (Slate/Blue theme, Inter font, Lucide icons)
- Implemented App Shell (Sidebar, Header, Layout)
- Implemented Authentication Foundation (Store, API interceptors, Login page)
- Implemented Dashboard placeholder
- Validated build (`npm run build` passed)

**Action Required**:
- **Backend**: Can now assume standard JWT auth header `Authorization: Bearer <token>`
- **Test**: Can begin writing E2E tests for Login -> Dashboard flow


---

### [2026-02-10 15:50 IST] [FROM: Backend] → [TO: ALL]
**Subject**: Feedback Verify/Correct/Reject Workflow Complete
**Type**: HANDOFF

**Changes**:
- Created migration `003_detection_feedback.sql` — new `detection_feedback` table with indexes
- Implemented `DetectionFeedbackRepo` in `internal/repository/postgres_feedback.go` (6 interface methods)
- Implemented `FeedbackService` in `internal/service/feedback_service.go` (verify/correct/reject logic)
- Implemented `FeedbackHandler` in `internal/handler/feedback_handler.go` (4 REST endpoints)
- Added `EventPIICorrected` event type to `pkg/eventbus/eventbus.go`
- Wired all components in `cmd/api/main.go`
- Updated test migration list in `internal/repository/postgres_test.go`
- `go build ./...` ✅ | `go vet ./...` ✅

**API Endpoints** (all under `/api/v2/discovery/feedback`, auth required):
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v2/discovery/feedback` | Submit verify/correct/reject feedback |
| `GET` | `/api/v2/discovery/feedback` | List feedback (paginated, `?page=&page_size=`) |
| `GET` | `/api/v2/discovery/feedback/classification/{id}` | Feedback for a specific classification |
| `GET` | `/api/v2/discovery/feedback/accuracy/{method}` | Accuracy stats for a detection method |

**Action Required**:
- **Test**: Unit/integration tests needed for `FeedbackService` and `DetectionFeedbackRepo`. Key files: `feedback_service.go`, `postgres_feedback.go`
- **Frontend**: New feedback endpoints are ready for the PII review UI

---

### [2026-02-10 17:05 IST] [FROM: Frontend] → [TO: ALL]
**Subject**: API Integration & DataSources Page Complete
**Type**: HANDOFF

**Changes**:
- Connected real auth API (`POST /auth/login`, `POST /auth/register`, `POST /auth/refresh`, `GET /users/me`)
- Login page with domain/email/password, error display, JWT storage
- Register page with org onboarding form
- ProtectedRoute guard redirecting unauthenticated users to `/login`
- Reusable `DataTable` component (sortable headers, loading skeletons, empty state, custom cell renderers)
- `Pagination` component (prev/next, page size selector)
- `StatusBadge` component (auto-resolves color from status string)
- `DataSources` list page using DataTable with scan action per row
- React Query hooks for auth and datasource operations
- Vite proxy configured (`/api` → `localhost:8080`)
- `npm run build` ✅ (1844 modules, 0 errors)

**API Endpoints Consumed**:
| Method | Path | Used By |
|--------|------|---------|
| `POST` | `/api/v2/auth/login` | Login page |
| `POST` | `/api/v2/auth/register` | Register page |
| `POST` | `/api/v2/auth/refresh` | Token refresh hook |
| `GET` | `/api/v2/users/me` | Session validation |
| `GET` | `/api/v2/data-sources` | DataSources list |
| `POST` | `/api/v2/data-sources` | Add Data Source |
| `PUT` | `/api/v2/data-sources/{id}` | Update Data Source |
| `DELETE` | `/api/v2/data-sources/{id}` | Delete Data Source |
| `POST` | `/api/v2/data-sources/{id}/scan` | Scan trigger |
| `POST` | `/api/v2/data-sources/{id}/test` | Test connection |

**Missing Endpoints Needed**:
- `POST /api/v2/data-sources/{id}/scan` — not yet implemented in backend handler
- `POST /api/v2/data-sources/{id}/test` — not yet implemented in backend handler

**Action Required**:
- **Backend**: Please add scan and test-connection endpoints to `datasource_handler.go`
- **Test**: Can now write E2E tests for Login → Dashboard → DataSources flow

---

### [2026-02-10 17:25 IST] [FROM: Frontend] → [TO: Backend]
**Subject**: Request for PII Classifications List Endpoint
**Type**: REQUEST

**Context**: Building the PII Discovery / Review Queue page. Need to list all classifications across the tenant with filtering.

**Requirement**:
Please implement `GET /api/v2/discovery/classifications` supporting:
- Pagination (`page`, `page_size`)
- Filters: `status`, `data_source_id`, `detection_method`
- Response: `PaginatedResult<PIIClassification>`

Currently, `discovery_handler.go` only has `GET /data-sources/{sourceID}/inventory` and nested entity/field lookups, but no flat list for the review queue.

---

### [2026-02-10 17:35 IST] [FROM: Frontend] → [TO: ALL]
**Subject**: PII Discovery & Feedback UI Complete
**Type**: HANDOFF

**Changes**:
- **PII Discovery Page** (`/pii/review`)
  - DataTable with Field, Category, Type, Method, Status columns
  - Color-coded Confidence Badges (>90% Green, 70-90% Yellow, <70% Red)
  - **Feedback Actions**:
    - ✅ **Verify**: Immediate status update
    - ✏️ **Correct**: Modal with Category/Type dropdowns
    - ❌ **Reject**: Modal with Notes field for false positives
  - **Accuracy Stats Panel**: Visual bar charts per detection method
  - **Filters**: By Status and Detection Method

**Dependencies**:
- **Backend**: Still waiting on `GET /api/v2/discovery/classifications` (requested above). The page will currently fallback to empty state or error until this endpoint exists.
- **Feedback API**: `POST /discovery/feedback` and `GET /discovery/feedback/accuracy/{method}` are integrated and expected to work once backend implements them.

**Next Steps**:
- **Test**: Can write E2E tests for the Feedback modal flows (mocking the API).

---

## Message Types Reference

| Type | When to Use | Example |
|------|------------|---------| 
| **INFO** | Sharing context another agent needs | "Backend: new `/api/v2/agents` endpoint is live with these response fields..." |
| **REQUEST** | Asking another agent to do something | "Frontend → Backend: Need a `GET /api/v2/dashboard/stats` endpoint" |
| **HANDOFF** | Passing completed work for the next agent | "Backend → Test: PII verification service is complete, needs unit tests" |
| **BLOCKER** | Something is preventing your task | "Frontend: Cannot proceed — `GET /api/pii/inventory` returns 500" |
| **QUESTION** | Need clarification from another agent | "AI/ML → Backend: Should detection results be cached per-tenant or globally?" |

---

## Contract Definitions

> When a Backend agent creates an API endpoint, or an AI/ML agent defines an interface, they should document the contract here so the Frontend and Test agents can work against it immediately.

### Active API Contracts

---

### [2026-02-10 16:15 IST] [FROM: Backend] → [TO: ALL]
**Subject**: Redis Caching & Token Budget Implemented
**Type**: HANDOFF

**Changes**:
- Added `go-redis/v9` dependency
- Implemented `Redis` connection helper in `pkg/database/redis.go`
- Implemented `CachedGateway` decorator in `internal/service/ai/cached_gateway.go`
- Wired `CachedGateway` in `cmd/api/main.go` (fail-open: if Redis down, caching disabled)
- Added `ErrQuotaExceeded` ("QUOTA_EXCEEDED") to `pkg/types/errors.go`

**Features Enabled**:
- **Caching**: PII detections cached for 24h (key: hash of schema + samples)
- **Budgeting**: Enforces daily token limit per tenant (stored in Redis `tenant:{id}:ai:budget`)
- **Tracking**: Logs daily token usage to Redis (`tenant:{id}:ai:tokens:{date}`)

**Action Required**:
- **DevOps**: Ensure Redis is provisioned in all environments. Set `REDIS_HOST`, `REDIS_PORT`, etc.
- **Backend**: Can now use `ai.NewCachedGateway` to wrap any `ai.Gateway`.
- **Admin**: Can set tenant budgets via Redis key `tenant:{id}:ai:budget` (future: Admin API for this).

---

### [2026-02-10 16:30 IST] [FROM: Backend] → [TO: ALL]
**Subject**: Discovery Service & Postgres Connector Implemented
**Type**: HANDOFF

**Changes**:
- Implemented `DiscoveryService` in `internal/service/discovery_service.go`
- Implemented `PostgresConnector` in `internal/infrastructure/connector`
- Added `Connector` interface in `internal/domain/discovery`
- Added `POST /data-sources/{id}/scan` endpoint in `DiscoveryHandler`
- Wired up in `main.go`

**Features Enabled**:
- **Scanning**: Can now scan Postgres databases for schema and sample data.
- **PII Detection**: Automatically invokes AI + Regex detection on sampled data.
- **Inventory Sync**: Updates `DataInventory`, `DataEntity`, and `DataField` stats.

**Action Required**:
- **Frontend**: Add "Scan Now" button to Data Source details page.
- **DevOps**: Ensure database connectivity from API container.
- **QA**: Verify scanning with sample Postgres database credentials.

---

### [2026-02-10 16:45 IST] [FROM: AI/ML] → [TO: ALL]
**Subject**: Industry-Specific PII Detection Implemented
**Type**: UPDATE

**Changes**:
- Implemented `IndustryStrategy` in `internal/service/detection`
- Added pattern packs for **Healthcare** (NPI, DEA), **BFSI** (SWIFT, IBAN), and **HR**.
- Enhanced `AIStrategy` unit tests.

**Impact**:
- Discovery scans now automatically check for industry-specific patterns if the data source has an `industry` context (e.g., "Hospital DB").
- Improved precision for domain-specific PII.

---

### [2026-02-10 17:30 IST] [FROM: Backend] → [TO: ALL]
**Subject**: MySQL Connector & Registry Implemented
**Type**: HANDOFF

**Changes**:
- Added `go-sql-driver/mysql` dependency
- Implemented `ConnectorRegistry` in `internal/infrastructure/connector/registry.go`
- Implemented `MySQLConnector` in `internal/infrastructure/connector/mysql.go`
- Updated `ConnectorCapabilities` with new fields (SchemaDiscovery, DataSampling, etc.)
- Injected `ConnectorRegistry` into `DiscoveryService` (replacing hardcoded switch)
- Added MySQL 8 service to `docker-compose.dev.yml` (ports 3307:3306)

**Features Enabled**:
- **MySQL Scanning**: DataLens can now connect to, scan, and sample MySQL databases.
- **Extensibility**: New connectors can be added by registering them in `connector.NewConnectorRegistry()`.
- **Capabilities**: Connectors now explicitly declare support for SchemaDiscovery, DataSampling, ParallelScan, etc.

**Action Required**:
- **DevOps**: Ensure MySQL drivers/libraries are present in build environments if needed (Go handles this via static linking usually, but worth noting).
- **Backend**: Use `registry.GetConnector(type)` instead of manual switches.
- **Frontend**: Ensure "MYSQL" is an option in the Data Source creation form.

---

### [2026-02-10 18:00 IST] [FROM: Backend] → [TO: ALL]
**Subject**: Scan Orchestrator (Async Job Queue) Implemented
**Type**: HANDOFF

**Changes**:
- **Async API**: `POST /data-sources/{id}/scan` now returns `202 Accepted` with a `job_id`.
- **Job Queue**: Implemented durable NATS JetStream queue (`DATALENS_SCANS`).
- **Orchestrator**: Added `ScanService` to manage concurrency and job execution.
- **Worker**: Background worker parses queue and executes scans via `DiscoveryService`.
- **Status API**: 
  - `GET /data-sources/{id}/scan/status` (Get latest scan)
  - `GET /data-sources/{id}/scan/history` (Get scan history)

**Features Enabled**:
- **Reliability**: Variable-length scans won't timeout HTTP requests.
- **Concurrency Control**: Limits concurrent scans per tenant (default: 3).
- **History**: Full audit trail of scan execution, status, and stats.

**Action Required**:
- **Frontend**: Update Scan button to handle `202 Accepted` and poll `/scan/status` for progress.
- **DevOps**: Ensure NATS JetStream is enabled (it is in `docker-compose.dev.yml`).

---

_Document new API contracts here as they're created:_

### Submit Detection Feedback
**Created by**: Backend Agent
**Date**: 2026-02-10
**Method**: POST
**Path**: `/api/v2/discovery/feedback`
**Auth**: JWT Bearer token required
**Request Body**:
```json
{
  "classification_id": "uuid",
  "feedback_type": "VERIFIED | CORRECTED | REJECTED",
  "corrected_category": "IDENTITY | CONTACT | ...",  // required if CORRECTED
  "corrected_type": "NAME | EMAIL | PHONE | ...",     // required if CORRECTED
  "notes": "optional reviewer notes"
}
```
**Response Body** (201):
```json
{
  "success": true,
  "data": {
    "feedback": { "id": "uuid", "classification_id": "uuid", "feedback_type": "VERIFIED", ... },
    "classification": { "id": "uuid", "status": "VERIFIED", "verified_by": "uuid", ... }
  }
}
```
**Status**: Implemented

### List Detection Feedback
**Created by**: Backend Agent
**Date**: 2026-02-10
**Method**: GET
**Path**: `/api/v2/discovery/feedback?page=1&page_size=20`
**Auth**: JWT Bearer token required
**Response Body** (200): Paginated list of `DetectionFeedback` objects
**Status**: Implemented

### Get Feedback by Classification
**Created by**: Backend Agent
**Date**: 2026-02-10
**Method**: GET
**Path**: `/api/v2/discovery/feedback/classification/{classificationID}`
**Auth**: JWT Bearer token required
**Response Body** (200): Array of `DetectionFeedback` objects
**Status**: Implemented

### Get Accuracy Stats
**Created by**: Backend Agent
**Date**: 2026-02-10
**Method**: GET
**Path**: `/api/v2/discovery/feedback/accuracy/{method}`
**Auth**: JWT Bearer token required
**Response Body** (200):
```json
{
  "success": true,
  "data": { "method": "AI", "total": 100, "verified": 85, "corrected": 10, "rejected": 5, "accuracy": 0.85 }
}
```
**Status**: Implemented

### Active Interface Contracts

_Document Go interfaces that cross agent boundaries:_

```markdown
### [Interface Name]
**Created by**: [Agent]
**File**: `path/to/file.go`
**Consumers**: [Which agents need to know about this]
**Notes**: [Any important details]
```

---

## Sprint Alignment

> At the start of each sprint, the Orchestrator posts the sprint goals here so all agents share context.

### Current Sprint Goals

~~**Sprint Batch 1** (Feb 10, 2026) — ✅ COMPLETE~~

**Sprint Batch 2** (Feb 10, 2026)

| # | Task | Agent | Priority | Parallel? |
|---|------|-------|----------|-----------|
| 1 | Connect real API auth + build DataSources page | Frontend | P0 | ✅ |
| 2 | Build PII Discovery page with feedback UI | Frontend | P0 | ⚠️ After #1 |
| 3 | MySQL connector + ConnectorCapabilities + registry | Backend | P1 | ✅ |
| 4 | Tests for all Batch 1 new code (feedback, cache, discovery, industry) | Test | P0 | ✅ |
| 5 | Scan orchestrator — async NATS job queue + progress + scheduling | Backend | P1 | ⚠️ After #3 |

**Batch Goals**: Connect frontend to live APIs, expand connector framework, async scanning, comprehensive test coverage.

---

## Resolved Messages Archive

### Batch 1 — All Resolved (Feb 10, 2026)
- ~~Orchestrator kick-off~~ ✅
- ~~Frontend scaffolding handoff~~ ✅ — Consumed in Batch 2 Task #1
- ~~Backend feedback workflow handoff~~ ✅ — Consumed in Batch 2 Task #2 & #4
- ~~Backend Redis caching handoff~~ ✅ — Consumed in Batch 2 Task #4
- ~~Backend Discovery Service handoff~~ ✅ — Consumed in Batch 2 Task #4
- ~~AI/ML Industry Strategy handoff~~ ✅ — Consumed in Batch 2 Task #4
