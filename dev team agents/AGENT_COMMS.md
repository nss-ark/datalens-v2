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
| **R2** | Backend mode splitting | `cmd/api/main.go` → `--mode=all\|cc\|admin\|portal`, routes in `routes.go` |
| **R3** | Nginx reverse proxy | `*.localhost:8000` sub-domains, env-driven CORS, 3 prod API instances |

**Rules**: Frontend → work in `packages/<app>/`, import from `@datalens/shared`. Backend → new routes in `routes.go`, new services wrapped in `shouldInit()`. DevOps → ONE binary, mode via Docker `command:`.

---

## ✅ Completed: Batch 20B + Batch 2 (Architecture Stabilization)

| Goal | Owner | Status | Details |
|------|-------|--------|---------|
| Logout API + UI | Backend/Frontend | ✅ | `POST /api/v2/auth/logout`, sidebar button |
| File Upload API + UI | Backend/Frontend | ✅ | Upload, cleanup on delete, drag & drop UI |
| Document Parsing | AI/ML | ✅ | PDF/DOCX/XLSX extraction done |
| ScanService Wiring | Backend | ✅ | `ParsingService` → `ScanService` for `FILE_UPLOAD` data sources |
| Stale Test Signatures | Backend | ✅ | `NewConnectorRegistry` calls fixed in all test files |
| UX Re-Review (20A) | UX/Frontend | ✅ | PII Inventory + Settings implemented, Dashboard polish done |
| OCR (Tesseract) | AI/ML | ⏸️ Deferred | Missing C libs. Decision pending: Tesseract vs Cloud Vision API |

---

## Active Sprint: Phase 3A — DPDPA Compliance Gaps

**Next Tasks** (see `batch_3_task_specs.md` in orchestrator brain):

| Task | Agent | Priority | Effort |
|------|-------|----------|--------|
| 3A-1: DPR Download endpoint | Backend | P1 | 2h |
| 3A-2: DPR Appeal flow | Backend + Frontend | P1 | 5h |
| 3A-3: Guardian consent (frontend) | Frontend | P1 | 2h |
| 3A-4: DSR Auto-Verification | Backend | P2 | 4h |
| 3A-5: Consent Receipt Generation | Backend | P2 | 3h |
| 3B: Missing Connectors (SQL Server) | Backend | P3 | Sprint |
| 3C: Observability Stack | DevOps | P3 | Sprint |

---

## Active Messages

### [2026-02-15] [FROM: Orchestrator] → [TO: ALL]
**Subject**: Batch 2 Complete — Moving to Phase 3
**Type**: STATUS

All stabilization work is done:
- ✅ E2E smoke test passed (login, proxy routing, CORS)
- ✅ Admin seeder fixed (port 5433)
- ✅ Test signatures fixed (`NewConnectorRegistry` + `parsingSvc`)
- ✅ ScanService wired for file uploads
- ✅ UX re-review complete (PII Inventory, Settings, Dashboard polish)
- **Next**: Phase 3A DPDPA compliance gaps

### [2026-02-14] [FROM: Frontend] → [TO: ALL]
**Subject**: Batch 20A Fixes Complete
**Type**: STATUS

- PII Inventory page: DataTable with confidence badges, sensitivity indicators
- Settings page: Profile display with role badges
- Sidebar: Truncation for long names/emails
- Dashboard: Polished empty states
- TypeScript build passing ✅
