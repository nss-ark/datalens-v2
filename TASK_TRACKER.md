# DataLens 2.0 — Development Task Tracker

> **Start Date**:10 February 2026  
> **Target GA**: Q4 2026  
> **Methodology**: 2-week sprints, quarterly releases

---

## Sprint 0: Foundation Setup (Weeks 1-2)

### 0.1 Monorepo & Project Structure
- [x] Create monorepo structure with Go workspace
- [x] Set up `/cmd` (entrypoints), `/internal` (domain), `/pkg` (shared libs)
- [x] Define module boundaries: `discovery`, `compliance`, `governance`, `evidence`, `identity`, `notification`
- [x] Create shared types package (`/pkg/types`) with universal entities
- [x] Set up `/api` with OpenAPI/Swagger specs
- [x] Create `Makefile` with standard targets (`build`, `test`, `lint`, `dev`, `migrate`)

### 0.2 Database & Infrastructure
- [x] Design PostgreSQL schema from Domain Model (Doc 21)
- [x] Create migration system (golang-migrate or Atlas)
- [x] Write initial migration: tenants, users, roles, permissions + all contexts
- [x] Set up Redis configuration (caching, rate limiting, pub/sub)
- [x] Set up NATS JetStream for event bus
- [x] Create `docker-compose.dev.yml` for local development
- [x] Create seed scripts for development data

### 0.3 CI/CD Pipeline
- [x] Configure GitHub Actions: lint → test → build — `.github/workflows/ci.yml` (Batch 4)
- [ ] Set up `golangci-lint` with project rules
- [x] Configure test coverage reporting (target: 80%) — race detection + coverage upload (Batch 4)
- [x] Set up Docker image builds — backend + frontend Dockerfiles, `docker-compose.prod.yml` (Batch 4)
- [ ] Create staging deployment workflow

### 0.4 Observability
- [ ] Set up Prometheus metrics collection
- [ ] Create Grafana dashboards (basic)
- [x] Configure structured logging (slog)
- [ ] Set up Jaeger for distributed tracing

### 0.5 Developer Experience
- [ ] Write `CONTRIBUTING.md` with coding standards
- [x] Create `.env.example` with all config variables
- [x] Verify `make dev` → full stack running locally
- [ ] Write onboarding guide for new developers

---

## Phase 1: Core Foundation (Q1 2026 — Sprints 1-6)

### Sprint 1-2: Core Domain & API Gateway (Weeks 3-6)

#### 1.1 Core Domain Entities
- [x] Implement `DataSource` entity + repository
- [x] Implement `DataInventory`, `DataEntity`, `DataField` entities
- [x] Implement `PIIClassification` entity + repository
- [x] Implement `Purpose` entity + repository
- [x] Implement `DataMapping` entity + repository
- [x] Write unit tests for all entities (validation, invariants)
- [x] Write integration tests for all repositories

#### 1.2 Event Bus Integration
- [x] Create `EventBus` interface and NATS implementation
- [x] Define event types (see Doc 20, Event System)
- [x] Wire repositories to publish events on create/update/delete
- [x] Create event subscriber framework
- [x] Write audit log subscriber (first subscriber)
- [ ] Test event delivery and replay

#### 1.3 API Gateway
- [x] Create unified HTTP router (chi or gin)
- [x] Implement JWT authentication middleware
- [x] Implement tenant isolation middleware
- [x] Implement rate limiting middleware (Redis-backed)
- [x] Create standard error response format
- [x] Implement request/response logging
- [x] Write CRUD API endpoints for DataSource
- [x] Write CRUD API endpoints for Purpose
- [x] Generate OpenAPI docs

#### 1.4 Multi-Tenant Auth
- [x] Implement user registration + login
- [x] Implement role-based access (ADMIN, DPO, ANALYST, VIEWER)
- [x] Implement permission checks per endpoint
- [x] Create API key system for agent authentication
- [x] Write auth integration tests

### Sprint 3-4: PII Detection Engine (Weeks 7-10)

#### 1.5 AI Gateway
- [x] Create `AIGateway` interface (see Doc 22) — `gateway.go`
- [x] Implement OpenAI-compatible provider (covers OpenAI, Ollama, vLLM, Groq, etc.) — `openai.go`
- [x] Implement Anthropic provider — `anthropic.go`
- [x] Implement Generic HTTP provider (Hugging Face, custom endpoints) — `generic_http.go`
- [x] Implement provider registry + factory — `registry.go`
- [x] Implement provider selection logic + fallback chain — `selector.go`
- [x] Add Redis caching for AI responses
- [x] Implement token budget & cost tracking
- [x] Write PII detection prompt templates — `prompts.go`
- [x] Write purpose suggestion prompt templates — `prompts.go`
- [ ] Test with real LLM providers

#### 1.6 Detection Strategies
- [x] Create `DetectionStrategy` interface — `strategy.go`
- [x] Implement `AIStrategy` (LLM-based contextual detection) — `ai_strategy.go`
- [x] Implement `PatternStrategy` (regex patterns) — `pattern.go`
- [x] Implement `HeuristicStrategy` (column name heuristics) — `heuristic.go`
- [x] Implement `IndustryStrategy` (sector-specific patterns) — `industry_strategy.go`
- [x] Create `ComposablePIIDetector` that chains strategies — `detector.go`
- [x] Implement confidence scoring and merging logic — `detector.go`
- [x] Implement PII sanitizer (never send real PII to cloud AI) — `sanitizer.go`
- [x] Write comprehensive tests per strategy
- [ ] Benchmark detection speed and accuracy

#### 1.7 Feedback & Learning Loop
- [x] Create `DetectionFeedback` entity + repository — `feedback.go`
- [x] Implement verify/correct/reject workflow
- [ ] Create rule extraction from feedback patterns
- [ ] Implement cache invalidation on corrections
- [ ] Track accuracy metrics per strategy

#### Frontend Foundation (Parallel)
- [x] Scaffold React + TypeScript + Vite application
- [x] Implement Authentication flows (Login, Register)
- [x] Implement App Shell (Sidebar, Header, Layout)
- [x] DataSources list page (DataTable, Pagination, Add/Scan actions)
- [x] PII Discovery / Review Queue page (feedback verify/correct/reject UI)
- [x] Dashboard with real metrics — `StatCard`, `PIIChart`, scan polling
- [x] DSR Management page — list, detail, create modal, SLA countdown, action buttons (Batch 4)

### Sprint 5-6: Data Source Connectors & Scanning (Weeks 11-14)

#### 1.8 Connector Framework
- [x] Create `DataSourceConnector` interface (see Doc 20) — `internal/domain/discovery`
- [x] Create `ConnectorCapabilities` system — `internal/infrastructure/connector`
- [ ] Implement connection pooling manager
- [x] Create connector factory/registry — `internal/infrastructure/connector/registry.go`

#### 1.9 Database Connectors
- [x] Implement PostgreSQL connector (parallel column scanning) — `internal/infrastructure/connector/postgres.go`
- [x] Implement MySQL connector (parallel column scanning) — `internal/infrastructure/connector/mysql.go`
- [x] Implement MongoDB connector — `internal/infrastructure/connector/mongodb.go`
- [ ] Implement SQL Server connector (port from v1)
- [x] Write integration tests per connector (testcontainers) — registry + MySQL tests

#### 1.10 File & Cloud Connectors
- [ ] Implement file system connector (local + network drives)
- [x] Implement S3 connector (with streaming) — `internal/infrastructure/connector/s3.go`, CSV/JSON/JSONL parsing (Batch 4)
- [ ] Implement Azure Blob connector
- [ ] Support file types: PDF, DOCX, XLSX, CSV, JSON, images (OCR)
- [ ] Write tests for each file type

#### 1.11 Scan Orchestrator
- [x] Create scan job queue (NATS-backed) — `internal/infrastructure/queue/scan_queue.go`
- [x] Implement async scan worker — `internal/service/scan_service.go`
- [ ] Implement parallel table/column scanning
- [x] Implement incremental scanning (only changes since last scan) — `DiscoveryInput.ChangedSince`
- [x] Create progress tracking (polling) — `GET /data-sources/{id}/scan/status`
- [ ] Create real-time progress tracking (WebSocket)
- [x] Implement scan scheduling (cron-like) — `internal/service/scheduler.go`, cron via `robfig/cron/v3` (Batch 4)
- [ ] Write scan performance benchmarks (target: 5x v1 speed)

### Phase 1 Milestone: `v2.1-alpha`
- [ ] All core entities working
- [ ] PII detection with AI achieving >85% accuracy
- [ ] 4+ database connectors working
- [x] File and S3 scanning working (Batch 4/5)
- [ ] Event bus operational
- [ ] API gateway with auth
- [ ] **Tag release v2.1-alpha**

---

## Phase 2: Compliance Features (Q2 2026 — Sprints 7-12)

### Sprint 7-8: DSR Engine (Weeks 15-18)

#### 2.1 DSR Workflow
- [x] Implement `DSR` entity with state machine (PENDING → IN_PROGRESS → COMPLETED) — `internal/domain/compliance`
- [x] Implement `DSRTask` decomposition per data source
- [x] Create SLA engine (auto-compute deadlines from regulation) — 30-day default
- [ ] Implement task assignment and routing

#### 2.2 DSR Execution
- [x] Implement access request execution (data export) — `dsr_executor.go`, samples PII fields per data source (Batch 4)
- [x] Implement erasure request execution (data deletion) — identifies PII locations, emits `dsr.data_deleted` event (Batch 4)
- [x] Implement correction request execution (data update) — stub for MVP, needs connector `Update()` method (Batch 4)
- [ ] Implement portability request (structured export)
- [/] Implement nomination request execution (DPDPA Requirement) (Batch 5)
- [x] Execute across multiple data sources in parallel — semaphore-bounded (default 5), NATS queue (Batch 4)

#### 2.3 DSR Auto-Verification (User Feedback P0)
- [ ] Implement post-execution re-query verification
- [ ] Auto-close on verification success
- [ ] Alert + retry on verification failure
- [ ] Generate evidence package on completion

#### 2.4 DSR Identity Verification
- [ ] Create identity matching service
- [ ] Implement document verification (Aadhaar/PAN matching)
- [ ] AI-assisted verification for ambiguous cases
- [ ] Manual override workflow

### Sprint 9-10: Consent Manager (Weeks 19-22)

#### 2.5 Consent Engine
- [x] Implement `Consent` entity with lifecycle (Batch 5)
- [x] Create consent capture API (with proof recording) (Batch 5 - partial)
- [ ] Implement consent withdrawal flow
- [ ] Implement consent expiry management with notifications
- [ ] Create consent receipt generation
- [x] Implement consent enforcement (check before data processing) (Batch 5)

#### 2.6 Embeddable Consent Widget (CMS)
- [x] Implement `ConsentWidget` CRUD service (Batch 5)
- [ ] Implement widget API key generation and validation
- [x] Build public API: `POST /api/public/consent/sessions` (record decisions) (Batch 5)
- [ ] Build public API: `GET /api/public/consent/check` (check consent status)
- [ ] Build public API: `POST /api/public/consent/withdraw` (withdraw consent)
- [ ] Build public API: `GET /api/public/consent/widget/{id}/config` (fetch config)
- [ ] Implement CORS validation against `allowed_origins`
- [ ] Build vanilla JS consent snippet (~15 KB gzipped, framework-agnostic)
- [ ] Support widget types: BANNER, PREFERENCE_CENTER, INLINE_FORM, PORTAL
- [ ] Support layouts: BOTTOM_BAR, TOP_BAR, MODAL, SIDEBAR, FULL_PAGE
- [ ] Implement widget theming (colors, fonts, logo, border radius)
- [ ] Implement custom CSS injection for widgets
- [ ] Support multi-language translations in widget config
- [ ] Implement `block_until_consent` mode
- [ ] Implement granular per-purpose toggle switches
- [ ] Auto-generate embed code snippet for each widget
- [ ] Implement widget version auto-increment on config change

#### 2.7 Data Principal Portal
- [x] Build portal page served by Control Centre (standalone + iframe-embeddable) (Batch 6)
- [x] Implement OTP-based identity verification (Email + Phone) (Batch 6)
- [x] Build `DataPrincipalProfile` CRUD service (Batch 6)
- [x] Link verified profile to `compliance.DataSubject` (Batch 6)
- [x] Build consent dashboard (current status per purpose, toggle on/off) (Batch 6)
- [x] Build consent history timeline (immutable, chronological, paginated) (Batch 6)
- [x] Implement digital signature for consent history entries (Batch 5/6)
- [x] Build portal public API: `POST /api/public/portal/verify` (Batch 6)
- [x] Build portal public API: `GET /api/public/portal/profile` (Batch 6)
- [x] Build portal public API: `GET /api/public/portal/consent-history` (Batch 6)
- [x] Implement portal session management (short-lived JWT) (Batch 6)

#### 2.8 DPR (Data Principal Rights) Flows
- [x] Build DPR submission: `POST /api/public/portal/dpr` (Batch 6)
- [x] Build DPR tracking: `GET /api/public/portal/dpr/{id}` (Batch 6)
- [ ] Build DPR download: `GET /api/public/portal/dpr/{id}/download` (ACCESS)
- [x] Link DPR request to internal `compliance.DSR` on creation (Batch 6)
- [ ] Implement DPR status flow: SUBMITTED → PENDING_VERIFY → VERIFIED → IN_PROGRESS → COMPLETED
- [ ] Implement guardian consent for minors (DPDPA Section 9)
  - [ ] Guardian name, email, relation fields
  - [ ] Guardian OTP verification flow
  - [ ] Block request until guardian verifies
- [ ] Implement appeal flow (DPDPA Section 18)
  - [ ] `POST /api/public/portal/dpr/{id}/appeal`
  - [ ] Appeal links to original DPR
  - [ ] Escalation to DPA authority flag
- [x] Implement SLA deadline tracking for DPR requests (Mocked in E2E)
- [x] Write E2E tests for consent + DPR flows (Batch 7A - Portal E2E)


### Sprint 11-12: Purpose Mapping & Governance (Weeks 23-26)

#### 2.9 Purpose Mapping Automation (User Feedback P0)
- [x] Implement context analysis engine (table + column patterns) (Batch 7)
- [x] Create sector template framework (Batch 7)
- [x] Build 6 sector templates: Hospitality, Airlines, E-commerce (1/6), Healthcare, BFSI, HR (Batch 7 - Partial)
- [x] Implement AI-powered purpose suggestion (Batch 7)
- [x] Create one-click confirm UI for suggestions (Batch 7)
- [ ] Implement batch second-round review for low-confidence
- [ ] Target: 70% auto-fill rate

#### 2.10 Governance Policy Engine
- [x] Implement `Policy` entity with rule evaluation (Batch 7)
- [x] Create policy templates (retention, access, transfer) (UI support in Batch 7)
- [x] Implement violation detection (scheduled job) (Batch 7)
- [x] Create alert system for policy violations (UI Dashboard in Batch 7)
- [ ] Implement auto-remediation for simple cases

#### 2.11 Data Lineage
- [ ] Implement data flow tracking
- [ ] Create data lineage visualization API
- [ ] Track purpose across data flows
- [ ] Cross-border transfer documentation

### Phase 2 Milestone: `v2.2-beta`
- [x] DSR end-to-end with auto-verification (Verified in Batch 7A)
- [ ] Consent portal deployed & white-labeled (Partially done in Batch 6)
- [x] Purpose mapping with 70% auto-fill (Batch 7)
- [x] Governance policies enforced (Batch 7 + E2E in 7A)
- [ ] **Tag release v2.2-beta**

---

## Phase 3: Enterprise Features (Q3 2026 — Sprints 13-18)

### Sprint 13-14: Cloud Integrations (Weeks 27-30)

#### 3.1 Microsoft 365 Connector
- [ ] OneDrive file scanning
- [ ] SharePoint document scanning
- [ ] Outlook email scanning
- [ ] Teams message scanning (if applicable)
- [ ] OAuth2 authentication flow

#### 3.2 Google Workspace Connector
- [ ] Google Drive scanning
- [ ] Gmail scanning
- [ ] Google Calendar (PII in events)
- [ ] OAuth2 authentication flow

#### 3.3 Additional Connectors
- [ ] Snowflake data warehouse connector
- [ ] Enhanced Salesforce connector (full CRM)
- [ ] SAP connector (basic)

#### 3.4 Webhook & Integration System
- [ ] Outbound webhook framework
- [ ] Configurable event triggers
- [ ] Retry logic with exponential backoff
- [ ] Webhook management UI
- [ ] Pre-built integrations (Slack, Teams, Jira)

### Sprint 15-16: Breach Management (Weeks 31-34)

#### 3.5 Breach Module
- [ ] Implement `Breach` entity with lifecycle
- [ ] Create breach detection (manual + automated triggers)
- [ ] AI-powered impact assessment
- [ ] CERT-In incident checklists (21 categories)
- [ ] Implement response workflow (detect → contain → investigate → resolve)

#### 3.6 Breach Notifications
- [ ] Authority notification system (CERT-In, DPA)
- [ ] Subject notification system (email/portal)
- [ ] Notification templates per regulation
- [ ] Evidence package for breach response
- [ ] SLA tracking for notification deadlines

### Sprint 17-18: Security Enhancements (Weeks 35-38)

#### 3.7 Enterprise Authentication
- [ ] SSO/SAML integration
- [ ] Multi-factor authentication
- [ ] Device fingerprinting for agents
- [ ] Session management enhancements

#### 3.8 Audit & Evidence
- [ ] Implement hash-chained audit log (tamper-proof)
- [ ] Digital signature for evidence records
- [ ] Evidence package export (PDF, JSON)
- [ ] Evidence retention and archival

#### 3.9 Advanced Security
- [ ] ML-based anomaly detection for access patterns
- [ ] HashiCorp Vault integration for secrets
- [ ] Penetration test remediation
- [ ] Security audit documentation

### Phase 3 Milestone: `v2.3-rc`
- [ ] Cloud integrations live (M365, Google)
- [ ] Breach management complete
- [ ] Enterprise security (SSO, audit chain)
- [ ] **Tag release v2.3-rc**

---

## Phase 4: Scale & Polish (Q4 2026 — Sprints 19-24)

### Sprint 19-20: Performance (Weeks 39-42)

#### 4.1 Scaling
- [ ] Multi-pod Kubernetes deployment
- [ ] Horizontal auto-scaling configuration
- [ ] Database connection pooling optimization
- [ ] Cache hit rate optimization (target: >80%)

#### 4.2 Performance
- [ ] Scan speed: 10x improvement over v1
- [ ] API response time: p95 < 200ms
- [ ] Load test: 100k+ records/sec throughput
- [ ] Database query optimization pass
- [ ] Implement database read replicas

### Sprint 21-22: UX & Frontend (Weeks 43-46)

#### 4.3 Bulk Operations
- [ ] Multi-select framework
- [ ] Bulk verify/reject PII classifications
- [ ] Bulk purpose assignment
- [ ] Saved views and filters
- [ ] Keyboard shortcuts for power users

#### 4.4 Mobile & UX
- [ ] Mobile responsive design
- [ ] Dark mode toggle
- [ ] Onboarding flow for new users
- [ ] Contextual help system
- [ ] Notification center in UI

#### 4.5 Analytics Dashboard
- [ ] Compliance score dashboard
- [ ] PII discovery trends
- [ ] DSR completion metrics
- [ ] Consent health metrics
- [ ] Exportable reports (PDF, CSV)

### Sprint 23-24: Release Prep (Weeks 47-50)

#### 4.6 Migration & Docs
- [ ] v1 → v2 data migration tools
- [ ] API documentation (complete)
- [ ] User guides and tutorials
- [ ] Admin guide

#### 4.7 Release Readiness
- [ ] Full E2E test suite passing
- [ ] Third-party security audit passed
- [ ] Performance SLA validation
- [ ] Disaster recovery (backup/restore) tested
- [ ] Rollback procedure validated
- [ ] **Tag release v2.0 GA** 🚀

---

## Consent Module (Sprint — MeITY BRD Alignment)

### Notice Management
- [ ] Implement `consent_notices` and `consent_notice_translations` DB migrations
- [ ] Notice CRUD API (create, read, update, publish, archive)
- [ ] Notice versioning logic (version increment on publish)
- [ ] Notice-to-widget binding API
- [ ] Notice management UI in Control Centre

### Translation Pipeline (HuggingFace)
- [ ] HuggingFace API integration service
- [ ] Translate endpoint (`POST /consent/notices/:id/translate`)
- [ ] Translation storage and retrieval
- [ ] Manual translation override endpoint
- [ ] Translation status tracking (per-language progress)

### Consent Notifications
- [ ] `consent_notifications` DB migration
- [ ] Notification template management API
- [ ] Event-driven notification triggers (consent granted/withdrawn/expiring)
- [ ] Email, SMS, Webhook delivery channels
- [ ] In-app notification component

### Consent Renewal
- [ ] `consent_renewal_logs` DB migration
- [ ] Renewal reminder scheduler (30/15/7 days)
- [ ] Renewal API and UI flow
- [ ] Expiry handling (mark as LAPSED)

### DigiLocker Integration
- [ ] OAuth 2.0 + PKCE flow implementation
- [ ] Identity verification via DigiLocker User API
- [ ] Age verification for parental consent (DPDPA § 9)
- [ ] Consent artifact push to DigiLocker
- [ ] Fallback to OTP on DigiLocker unavailability

### Consent Enforcement Middleware (Planned)
- [ ] Consent check endpoint optimization (< 50ms p99)
- [ ] Redis-backed consent cache
- [ ] Cache invalidation on consent withdrawal (pub/sub)
- [ ] Language SDKs (Go, Python, Node.js)

---

## Compliance Adapter Backlog (Post-GA)

### DPDPA Adapter (Ships with GA)
- [ ] All DPDPA DSR types configured
- [ ] DPDPA-specific timelines (30 days)
- [ ] DPDPA notice requirements
- [ ] CERT-In breach notification rules
- [ ] Data fiduciary obligations

### GDPR Adapter (Q1 2027)
- [ ] GDPR DSR types (incl. objection, restriction)
- [ ] GDPR timelines (30 days, extendable)
- [ ] DPO role and obligations
- [ ] DPIA (Data Protection Impact Assessment)
- [ ] Cross-border transfer mechanisms (SCCs, BCR)

### CCPA/CPRA Adapter (Q2 2027)
- [ ] CCPA rights configuration
- [ ] "Do Not Sell" implementation
- [ ] Financial incentives tracking
- [ ] CPRA-specific enhancements

---

## Legend

| Symbol | Meaning |
|--------|---------|
| `[ ]` | Not started |
| `[/]` | In progress |
| `[x]` | Completed |
| `[!]` | Blocked |

> **Last Updated**: February 11, 2026 — Consent module documentation and sprint tasks added (MeITY BRD alignment); Nomination request planning
