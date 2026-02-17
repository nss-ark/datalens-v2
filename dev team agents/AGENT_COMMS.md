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
