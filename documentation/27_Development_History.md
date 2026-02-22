# DataLens 2.0 — Complete Development History

> **Period**: Feb 10 – Feb 22, 2026 | **53 Git Commits** | **~12 Working Days**

---

## Timeline & Structure Overview

The project followed an evolving methodology:

| Phase | Structure | Date Range | Commits |
|-------|-----------|------------|---------|
| **Foundation** | Sprints 0–6 (numbered batches 1–8) | Feb 10–12 | 877cda8 → e277f6e |
| **Enterprise** | Numbered Batches 9–20B | Feb 12–14 | 539b098 → ac3c0eb |
| **Architecture** | Refactoring Batches R1–R4 | Feb 14 | 046b2ef → 3e47b90 |
| **Miscellaneous** | Fix Batches 1A/1B/1C, 2A/2B | Feb 14–15 | fc85763 → 9c1a48b |
| **DPDPA Compliance** | Phases 3A, 3B, 3C | Feb 15–17 | 06f6f10 → 6597853 |
| **UX & Portal** | Portal Vibe-Coding, SA Sprint | Feb 17–19 | fac5566 → 597fa6a |
| **Polish & Pages** | Phases 4A–4G | Feb 17–22 | 45980ce → present |

---

## 1. Core Foundation — Sprints 0–6 / Batches 1–8A (Feb 10–12)

### Batch 1 (Feb 10) — `877cda8`, `1ca27e4`
- Initial commit: Go monorepo, PostgreSQL schema, migrations
- API key management, authentication, data discovery services & handlers

### Batch 2 (Feb 10) — `2ef1a44`, `516353d`, `182777e`, `8d8d8fb`  
- AI & pattern-based PII detection strategies
- AI Gateway integration (OpenAI/Anthropic/Generic)
- Data source management, discovery, scanning, dashboard
- DSR (Data Subject Rights) initial management

### Batch 3 (Feb 10–11) — `d646175`
- Authentication system: backend middleware, frontend services, login tests

### Batch 4 (Feb 11) — `90a9d40`, `068d1ef`
- DSR full-stack: state machine, SLA engine, NATS queue, scheduled scanning
- Consent management system (entities, capture API, enforcement)
- CI/CD: GitHub Actions, Docker builds, test coverage

### Batch 5 (Feb 11) — `fc06f26`
- Consent lifecycle, embeddable widget CRUD
- Public consent session API, consent enforcement middleware
- Data Principal Portal foundation, DPR submission

### Batch 6 (Feb 11) — Included in `fc06f26`
- Portal OTP verification (Email + Phone)
- DataPrincipalProfile CRUD, consent dashboard, history timeline
- Portal session management (short-lived JWT)

### Batch 7 (Feb 11) — `f14a6a5`, `a9313d1`
- Governance module: policy management, violation detection
- Purpose mapping automation: context analysis, sector templates, AI suggestions
- DSR & Governance Violations E2E

### Batch 8 & 8A (Feb 12) — `e277f6e`, `b928bb6`, `d72a420`
- Enterprise Audit Logging (hash-chained)
- Data lineage visualization API
- S3/RDS/DynamoDB/Azure Blob connector hardening
- Local dev setup: Dockerized target DBs, seeder tool (10k+ rows), `setup_local_dev.ps1`

---

## 2. Enterprise Features — Batches 9–20B (Feb 12–14)

### Batch 9 (Feb 12) — `2ba0949`
- Breach management: entity lifecycle, CERT-In notification rules
- M365 OAuth2 authentication flow, cryptographic utilities

### Batch 10 (Feb 12) — `539b098`
- M365 scanning: OneDrive, SharePoint, Outlook, Teams

### Batch 11 (Feb 12) — `3d48a1c`
- Google Workspace: Drive + Gmail scanning, OAuth2
- M365 Discovery APIs for UI (Users/Sites list)

### Batch 12 (Feb 12) — `6b4dda1`
- Identity Assurance Domain (IAL1/IAL2)
- DigiLocker integration (OAuth2 + PKCE + HMAC)
- Identity Verification Policy UI

### Batch 13 (Feb 12) — `36a3222`, `04423e7`
- DSR Automation (auto-execute export/delete across connectors)
- Field-level lineage tracking (column A → column B)
- Smart Discovery scheduler logic

### Batch 14 (Feb 12) — `c3bb42b`
- Consent Analytics: conversion stats, purpose stats
- Dark Pattern Detector (13 patterns per India Guidelines 2023)

### Batch 15 (Feb 13) — `40d2a4a`
- Consent Module completion: Notice CRUD, widget binding, renewal scheduler
- Public APIs: consent check, withdraw, widget config
- Widget auth middleware + CORS validation

### Batch 16 (Feb 13) — `e7cca1d`
- Translation Pipeline: IndicTrans2/HuggingFace, 22 Indian languages
- Consent Notifications: event-driven Email/Webhook/SMS(stub)
- Grievance Redressal: DPDPA-compliant, 30-day SLA, DPO escalation

### Batch 17 (Feb 13) — `22c6d8a`
- Consent SDK: vanilla JS, script blocking engine, i18n, UI styles

### Batch 17A (Feb 13) — `04da3c7`
- SuperAdmin Portal: `PLATFORM_ADMIN` role, tenant CRUD
- Admin Dashboard with StatCards, Tenant List, Create Tenant Modal
- Admin seed script (`cmd/admin-setup/main.go`)

### Batch 17B (Feb 13) — `3bacd4e`
- Admin User CRUD API (search, suspend, assign roles)
- Redis Consent Cache (<50ms, pub/sub invalidation)
- User Management UI + Admin Dashboard live stats

### Batch 18 (Feb 13) — `fcb9b3a`
- Guardian Consent Flow (DPDPA Section 9): OTP-based guardian verification for minors
- Admin DSR Workflow API: approve/reject with DPR sync
- Portal UI: RequestNew, GuardianVerifyModal

### Batch 18.1 (Feb 13) — `9ed5e84`
- Admin DSR Patch: cross-tenant `GET /admin/dsr` + IDOR security fix

### Batch 19 (Feb 13) — `671b316`
- Breach UI routes + Sidebar integration
- Breach Detail: SLA countdown, CERT-In report, status transitions
- Cloud Data Source Config UI (M365 + Google Workspace OAuth)
- Breach §28 notifications: auto-trigger on HIGH/CRITICAL

### Batch 20 (Feb 13) — `1e9712e`
- Vanilla JS Consent Widget SDK (~11.5KB, framework-agnostic)
- Event Mesh refactoring: 15 typed constants, 9 services + 2 subscribers

### UX Review Agent (Feb 13) — `1e30ca9`
- New agent created: `ux-review-agent.md` for systematic UI auditing

### Batch 20A (Feb 14) — `ac3c0eb`
- 5-session UX review (Auth, Compliance, Governance, Admin, Cross-cutting)
- High + Medium priority fix sprint

### Batch 20B (Feb 14) — Included in `ac3c0eb`
- Logout API, File Upload API (`POST /datasources/upload`)
- Document parsing: PDF, DOCX, XLSX text extraction
- File Upload drag-and-drop UI, Delete Data Source action

---

## 3. Architecture Refactoring — Batches R1–R4 (Feb 14)

> [!IMPORTANT]
> These were **breaking changes** that restructured the entire codebase.

### R1: Frontend Monorepo Split — `f960b88`
- `frontend/src/` deleted → `frontend/packages/{shared, control-centre, admin, portal}/`
- npm workspace with `@datalens/shared` shared library

### R2: Backend Mode Splitting — `698346f`
- `cmd/api/main.go` → `--mode=all|cc|admin|portal`
- Routes extracted to `cmd/api/routes.go` (4 mount functions)
- Conditional service init via `shouldInit()` helper

### R3: Nginx Reverse Proxy — `5d1282e`
- `*.localhost:8000` sub-domains for dev
- `docker-compose.prod.yml` with 3 API instances + nginx gateway
- Env-driven CORS (`CORS_ALLOWED_ORIGINS`)

### R4: Misc cleanup — `3e47b90`
- Agent prompt updates for post-R1-R3 architecture

---

## 4. Miscellaneous Fix Batches — 1A/1B/1C, 2A/2B (Feb 14–15)

> [!NOTE]
> These were targeted fix sprints to address issues discovered after the architecture refactoring.

### Batch 1A (Feb 14) — `fc85763`
- All 7 agent prompts updated with post-R1-R3 architecture context
- Build fixes across the new monorepo structure

### Batches 1B, 1C, 2A, 2B (Feb 15) — `9c1a48b`
- Miscellaneous fixes spanning frontend packages and backend wiring
- **2B**: `ParsingService` → `ScanService` wiring for `FILE_UPLOAD` sources

---

## 5. DPDPA Compliance — Phases 3A, 3B, 3C (Feb 15–17)

### Phase 3A (Feb 15) — `06f6f10`, `f1098fe` — 11 sub-tasks
| Task | Description |
|------|-------------|
| 3A-1 | Portal Backend API Wiring — 9 routes + 3 aliases |
| 3A-2 | DPR Download Endpoint — 72h ACCESS SLA |
| 3A-3 | DPR Appeal Flow (DPDPA §18) |
| 3A-4 | DSR Auto-Verification — VERIFIED/VERIFICATION_FAILED |
| 3A-5 | Consent Receipt — HMAC-SHA256 signed |
| 3A-6 | DPO Contact Entity — tenant-level CRUD |
| 3A-7 | Notice Schema Validation — DPDP R3(1) Schedule I |
| 3A-8 | Guardian Frontend Polish |
| 3A-9 | Notice Translation API wiring |
| 3A-10 | Breach Portal Inbox |
| 3A-11 | Data Retention Model (entities only) |
| 3A-E2E | 6/6 integration tests passing |

### Phase 3B (Feb 17) — `6597853`
- SQL Server connector (`sqlserver.go`)

### Phase 3C (Feb 17) — `6597853`
- Observability stack: Prometheus + Grafana + Jaeger (`pkg/telemetry`)

---

## 6. SA Sprint — SuperAdmin Portal Build-Out (Feb 18–19)

> [!NOTE]
> This was a standalone sprint focused entirely on making the SuperAdmin Portal production-ready.

### Commit `597fa6a` — SA-1 to SA-4:
| Task | Description |
|------|-------------|
| **SA-1** | SuperAdmin Login: `POST /api/v2/superadmin/login`, global (no tenant_id) |
| **SA-2** | Admin Me endpoint: `GET /api/v2/admin/me` |
| **SA-3** | Tenant Detail + PATCH: `GET/PATCH /api/v2/admin/tenants/{id}` |
| **SA-4** | Frontend rewrite: mocks removed from `adminService.ts`, real API calls, AdminRoute guard |

### Portal Vibe-Coding Sessions (Feb 18–19) — `1b8d40e`, `c70c102`, `31b02e9`
| Commit | Work Done |
|--------|-----------|
| `1b8d40e` | Portal Login and Dashboard vibe-coded |
| `c70c102` | My Requests fixed: modals, request intake flows |
| `31b02e9` | Portal UI general fixes |

### Subscription Billing (Separate conversation, Feb 17–19)
- Billing date editing in `TenantDetail.tsx`
- `CheckSubscriptionExpiry` service + daily background worker
- Expiry warning notification stub

---

## 7. Phase 4 — Comprehensive Build Sprint (Feb 17–22)

### Phase 4 Prep (Feb 17) — `45980ce`
- Agent prompts revamped for Phase 4 patterns
- Sprint plan: 7 batches, 30 tasks

### Batch 4A: Foundation Fixes (Feb 17) — `fac5566`
- `019_retention.sql`, `020_audit_log_columns.sql` migrations
- Audit Log CC Handler (`GET /api/v2/audit-logs`)
- Route dedup + sidebar nav fixes
- Build verification: backend + 3 frontend apps ✅

### Batch 4B: UI Overhaul (Feb 17–18) — `f1148b9`
- KokonutUI + shadcn/ui integration across CC, Admin, Portal
- Dashboard, Tenant List, Settings, Profile polish
- Portal spacing, padding, notifications redesign

### Batch 4C: Core Compliance Pages (Feb 19–21) — `3dcb141`
| Task | Page/Feature |
|------|-------------|
| 4C-1 | Backend: Consent Sessions listing, Data Subjects search, Retention CRUD APIs |
| 4C-2 | Audit Logs Page (`/audit-logs`) — entity/action/date filters |
| 4C-3 | Retention Scheduler — daily cron, 24h throttle |
| 4C-4 | Consent Records Page (`/consent`) — status filters |
| 4C-5 | Data Subjects Page (`/subjects`) — debounced search |
| 4C-6 | Retention Policies Page (`/retention`) — CRUD modals |

### Batch 4D: RoPA + Multi-Level Purpose Tagging (Feb 21–22)
| Task | Description |
|------|-------------|
| 4D-1 | RoPA Backend: auto-generation, version control (semver), 7 endpoints |
| 4D-2 | RoPA Frontend: 612-line page, collapsible sections, inline editing, version history |
| 4D-3 | Purpose Assignments Backend: 5-level scope hierarchy (SERVER→COLUMN) |
| 4D-4 | Purpose Assignments Frontend: tabs on Purpose Mapping, inherited rows |
| 4D-5 | Bug fixes: duplicate wiring in `main.go` |

### Batch 4E: Department + Third-Party (Feb 22)
| Task | Description |
|------|-------------|
| 4E-1 | Department Backend: CRUD + email notifications via SMTP |
| 4E-2 | Third-Party + DPA: 6 DPA columns, status lifecycle (NONE→PENDING→SIGNED→EXPIRED) |
| 4E-3 | Frontend: Departments + ThirdParties pages, sidebar links |

### Batch 4F: OCR + Portal Polish (Feb 22)
| Task | Description |
|------|-------------|
| 4F-1 | OCR Adapter Pattern: `OCRAdapter` interface, Tesseract + Sarvam Vision adapters |
| 4F-2 | Portal polish: card CSS, NominationModal DPDPA S14 explainer, CC spacing |

### Batch 4G: Reports + Nominations Filter (Feb 22) ✅
| Task | Description |
|------|-------------|
| 4G-1 | Compliance Reporting Service: 4 DPDPA pillars, CSV/JSON export for 7 entities |
| 4G-2 | DSR Nominations type filter: `GetByTenant` + 5 test mock updates |
| 4G-3 | Frontend: Reports page (SVG gauge, pillar cards), Nominations page |

---

## Summary Statistics

| Metric | Value |
|--------|-------|
| Total Git Commits | 53+ |
| Calendar Days | 12 (Feb 10–22) |
| Numbered Batches | 1–20B (27 batches) |
| Architecture Refactors | R1–R4 |
| Miscellaneous Fix Batches | 1A, 1B, 1C, 2A, 2B |
| DPDPA Gap Batches | 3A (11 sub-tasks), 3B, 3C |
| SA Sprint Tasks | SA-1 to SA-4 |
| Phase 4 Batches | 4A, 4B, 4C, 4D, 4E, 4F, 4G ✅ |
| Portal Vibe-Coding Sessions | 3 commits |
| Database Migrations | 001–024 |
| Frontend Apps | 3 (Control Centre, Admin, Portal) + Widget SDK |
| Backend Endpoints | ~80+ REST APIs |
| Backend Services | 85 |
| Backend Handlers | 36 |

