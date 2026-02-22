# Phase 4 — Batch 4G Task Specifications (FINAL BATCH)

**Sprint**: Phase 4 — Compliance Reporting Hub + Placeholder Resolution + QA  
**Estimated Duration**: 1.5 days  
**Pre-requisites**: Batch 4F complete

---

## Value Analysis

### Enterprise Customer (Compliance Officer / DPO)
1. **One-click audit readiness**: Generate a "Compliance Snapshot" PDF/JSON that aggregates DSR stats, consent rates, breach response times, RoPA status, retention compliance, department coverage — everything an auditor or Data Protection Authority asks for
2. **Export anything**: CSV export on every data table (DSRs, breaches, consent records, audit logs, departments, third-parties) — auditors always ask for raw data
3. **Compliance scorecard**: A single score (0-100) showing how well the org is doing across all DPDPA pillars, with drill-down into weaknesses

### End User (Data Principal via Portal)
1. **My Data Report**: One-click download of everything the enterprise holds about them — processed purposes, consent history, DSR status — DPDPA S11 (Right to Information)
2. **Transparency**: Know which departments process their data, which third-parties have access, and the lawful basis

---

## Execution Order

**Group 1** (Backend): Task 4G-1 + 4G-2 (parallel)  
**Group 2** (Frontend): Task 4G-3 (after Group 1 completes)

---

## Task 4G-1: Backend — Compliance Reporting Service + Export API

**Agent**: Backend  
**Priority**: P0  
**Effort**: Medium-High (4-5h)

### Objective

Create a `ReportService` that aggregates data from existing services to produce compliance reports. Add CSV/JSON export capabilities.

### Requirements

#### 1. Report Types

**A. Compliance Snapshot Report** (`GET /api/v2/reports/compliance-snapshot`)
Aggregates across all DPDPA pillars:
```json
{
  "generated_at": "2026-02-22T10:00:00Z",
  "period": { "from": "2026-01-01", "to": "2026-02-22" },
  "overall_score": 78,
  "pillars": {
    "consent_management": {
      "score": 85,
      "total_consents": 1200,
      "active_consents": 980,
      "withdrawal_rate": "3.2%",
      "avg_response_time_hours": 2.5,
      "notices_published": 5,
      "notices_compliant": 4
    },
    "dsr_compliance": {
      "score": 72,
      "total_requests": 45,
      "completed_on_time": 38,
      "overdue": 3,
      "avg_resolution_days": 8,
      "by_type": { "ACCESS": 20, "ERASURE": 15, "CORRECTION": 10 }
    },
    "breach_management": {
      "score": 90,
      "total_incidents": 2,
      "reported_to_cert_in": 2,
      "avg_notification_hours": 18,
      "principals_notified": true
    },
    "data_governance": {
      "score": 65,
      "departments_with_owners": 8,
      "departments_total": 12,
      "third_parties_with_dpa": 5,
      "third_parties_total": 8,
      "purposes_mapped": 15,
      "retention_policies_active": 3,
      "ropa_published": true,
      "ropa_version": "2.0"
    }
  },
  "recommendations": [
    "4 departments lack assigned owners — assign owners to ensure accountability",
    "3 third-parties have expired DPAs — renew to maintain compliance",
    "3 DSR requests are overdue — process immediately to avoid DPDPA penalties"
  ]
}
```

**B. Entity Export** (`GET /api/v2/reports/export/{entity}?format=csv|json`)
Entities: `dsr`, `breaches`, `consent-records`, `audit-logs`, `departments`, `third-parties`, `purposes`

CSV format with headers, tenant-scoped. JSON returns the same array as the list endpoints.

#### 2. New Files

- `internal/service/report_service.go` [NEW]
  - `GenerateComplianceSnapshot(ctx, tenantID, from, to)` — aggregates from existing repos/services
  - `ExportEntity(ctx, tenantID, entity, format)` — delegates to existing repos, formats as CSV or JSON
  - Dependencies: inject existing services (DashboardService, DSRService, BreachService, ConsentService, DepartmentService, ThirdPartyService, RoPAService, RetentionService, PurposeService, AuditService)

- `internal/handler/report_handler.go` [NEW]
  - `GET /reports/compliance-snapshot?from=&to=` — calls GenerateComplianceSnapshot
  - `GET /reports/export/{entity}?format=csv` — calls ExportEntity, sets Content-Type + Content-Disposition headers for download

- `cmd/api/routes.go` [MODIFY] — mount `/reports`
- `cmd/api/main.go` [MODIFY] — wire ReportService + ReportHandler

#### 3. Score Calculation

Simple weighted average:
- Consent: 25% weight — based on active consent rate + notice compliance
- DSR: 30% weight — based on on-time completion rate
- Breach: 20% weight — based on notification timeliness
- Governance: 25% weight — based on department ownership + DPA coverage + RoPA status

#### 4. CSV Generation

Use Go's `encoding/csv` — no external deps needed. Each entity maps to columns:
- DSR: ID, Type, Status, Subject, Created, Resolved, Overdue
- Breaches: ID, Title, Severity, Status, Reported, Principals Affected
- Departments: ID, Name, OwnerName, OwnerEmail, Responsibilities
- Third-Parties: ID, Name, DPAStatus, DPAExpiry, ContactName
- etc.

### Acceptance Criteria
- [ ] Compliance snapshot returns aggregated scores
- [ ] CSV export works for all 7 entity types
- [ ] JSON export works for all 7 entity types
- [ ] Content-Disposition header set for file downloads
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] AGENT_COMMS.md updated

---

## Task 4G-2: Backend — Nominations Filter + Placeholder Cleanup

**Agent**: Backend  
**Priority**: P1  
**Effort**: Small (1h)

### Objective

The `/nominations` route should display DSR requests of type `NOMINATION`. No new backend entity needed — just ensure the existing DSR list endpoint supports filtering by type so the frontend can call `GET /api/v2/dsr?type=NOMINATION`.

### Requirements

1. **Verify** that `GET /api/v2/dsr` already supports `?type=` query param filtering
2. If not, add type filtering to the DSR list handler/service
3. Ensure NOMINATION is a valid DSR type in the domain

### Acceptance Criteria
- [ ] `GET /api/v2/dsr?type=NOMINATION` returns only nomination-type requests
- [ ] `go build ./...` passes

---

## Task 4G-3: Frontend — Reports Page + Nominations Page + Final QA

**Agent**: Frontend  
**Priority**: P0  
**Effort**: Medium-High (4-5h)

### Objective

Build the Reports page (compliance dashboard + entity exports), wire up Nominations, resolve remaining placeholders, and do a final QA pass.

### Requirements

#### 1. Reports Page (`/reports`) — Replace Placeholder

Premium compliance dashboard with 3 sections:

**Section A: Compliance Scorecard**
- Large circular gauge (0-100) showing overall compliance score
- 4 pillar cards (Consent, DSR, Breach, Governance) with individual scores
- Color coding: 80+ green, 60-79 amber, <60 red
- "Last generated" timestamp

**Section B: Recommendations**
- Alert banner cards for each recommendation from the snapshot
- Priority-sorted (critical first)

**Section C: Data Exports**
- Grid of export cards (one per entity): DSRs, Breaches, Consent Records, Audit Logs, Departments, Third-Parties, Purposes
- Each card has: entity icon, count, "Export CSV" button, "Export JSON" button
- Clicking triggers download via `GET /api/v2/reports/export/{entity}?format=csv`

#### 2. Nominations Page (`/nominations`) — Replace Placeholder

Simple filtered DSR list:
- Reuse DSR list component pattern but filtered to `?type=NOMINATION`
- Add header: "Nominations — DPDPA Section 14" with brief explainer
- Table: Nominee Name, Nominator, Status, Created Date, Actions

#### 3. Placeholder Cleanup

- `/agents` → Keep as-is (Phase 5 scope — AI agent management)
- `/users` → Keep as-is (RBAC deferred to Batch 17+)
- Update both to show a clear "Coming Soon" message instead of generic "under construction"

#### 4. Final QA Pass

Full sidebar click-through — verify every page loads without crash:
- All 7 sidebar groups, every item
- Verify no console errors, no broken imports
- Run build for ALL 3 packages

### Acceptance Criteria
- [ ] Reports page with scorecard + recommendations + exports
- [ ] Nominations page with filtered DSR list
- [ ] "Coming Soon" for Agents and Users
- [ ] Full sidebar click-through passes
- [ ] `npm run build -w @datalens/control-centre` passes
- [ ] `npm run build -w @datalens/portal` passes (no regressions)

---

## Summary

| Task | Agent | Priority | Effort | Value |
|------|-------|----------|--------|-------|
| 4G-1: Compliance Reporting Service | Backend | P0 | 4-5h | Audit readiness, one-click compliance proof |
| 4G-2: Nominations Filter | Backend | P1 | 1h | DPDPA S14 compliance |
| 4G-3: Reports + Nominations UI + QA | Frontend | P0 | 4-5h | Enterprise dashboard, final polish |

**Parallelism**: 4G-1 + 4G-2 (parallel), then 4G-3 (depends on backend).
