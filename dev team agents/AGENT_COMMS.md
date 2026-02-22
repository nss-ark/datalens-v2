# Agent Communication Board

> **System Instructions**:
> - Post your status here when starting or finishing a task.
> - `[BLOCKED]` prefix if you are blocked. `[HANDOFF]` prefix for handoffs.
> - The Orchestrator reads this file at the start of every session.
> - **Archive**: Previous threads are in `AGENT_COMMS_archive_v1.md` (Batches 1–20B) and `AGENT_COMMS_archive_v2.md` (Phase 3A–4G).

---

## Project Status — Phase 5 Ready

> **Phase 4 is COMPLETE** (7/7 batches: 4A–4G). All backend + frontend builds pass.
> 53 Git commits across 12 working days (Feb 10–22, 2026).
> See `documentation/27_Development_History.md` for full timeline.

### What Exists
- **Backend**: 85 services, 36 handlers, 24 SQL migrations, ~80+ REST APIs
- **Frontend**: CC (28+ pages), Admin (7 pages), Portal (8 pages), Consent Widget SDK
- **Infrastructure**: Docker Compose, Nginx proxy, Prometheus/Grafana/Jaeger, NATS, Redis

### Priority Order for Phase 5+
1. **QA & Polish (C)** — End-to-end flow testing, fix broken integrations, UI/UX consistency
2. **Feature Completion (B)** — Close remaining gaps (OCR, widget playground, settings, user mgmt)
3. **Production Readiness (A)** — Docker/K8s, migration consolidation, deployment hardening

---

## Active Sprint: Phase 5 — QA, Integration & Feature Completion

**Status**: Plan approved. See `dev team agents/phase5_sprint_plan.md` for full details.

### Phase 5 Batch Order
| Batch | Focus | Priority |
|-------|-------|----------|
| **5A** | Misc QA — E2E Flow Testing & Fixing (SuperAdmin → CC → Portal) | P0 |
| **5B** | UI/UX Consistency Pass + Empty/Error/Loading States | P0 |
| **5C** | Feature Completion — Settings, User Management, Consent Widget Playground | P1 |
| **5D** | OCR Hardening + Advanced Features (Tesseract primary, Sarvam fallback) | P1 |
| **5E** | Infrastructure — Migration Consolidation, Docker/K8s, Deployment | P2 |

---

## Active Messages

### [HANDOFF] Orchestrator → All Agents — Batch 5A Tasks #1, #2 Complete
**Date**: 2026-02-22
**Status**: ✅ DONE

**Task #1: DataSource Type Normalization (Critical Fix)**
- Added `NormalizeDataSourceType()` to `pkg/types/types.go` — handles uppercase + alias mapping (`mssql`→`SQLSERVER`, `m365`→`MICROSOFT_365`, `local_file`→`FILE_UPLOAD`)
- Wired into `DataSourceService.Create()` in `internal/service/datasource_service.go`
- Added safety-net normalization in `ConnectorRegistry.GetConnector()` in `internal/infrastructure/connector/registry.go`
- Updated frontend: `datasource.ts` type union, `DataSources.tsx`, `DataSourceConfig.tsx`, `DataSourceDetail.tsx` — all type string comparisons now use uppercase canonical values
- **Root cause**: Scans failed because frontend sent `postgresql` but registry indexed by `POSTGRESQL`

**Task #2: Remove Global FAB**
- Removed `ActionToolbar` + `globalActions` array from `AppLayout.tsx`
- Each page retains its own relevant actions (e.g., "Add Data Source" on `/datasources`)

**Builds**: `go build ./...` ✅ | `npm run build -w @datalens/control-centre` ✅

**Next**: Task #3 (Discovery E2E pipeline validation) — requires spinning up the app and performing live scan tests.
