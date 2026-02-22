# Phase 5 QA & Fixes Observations

> **Purpose**: This document captures detailed UI/UX and functionality observations from the user's manual testing of the DataLens 2.0 application (SuperAdmin → Control Centre → Portal). These notes will be used to enhance and restructure Batch 5A and subsequent Phase 5 plans.

---

## Observation Log

### 1. Global Floating Action Button (FAB) — ✅ RESOLVED (Batch 5A Task #2)
- **Issue**: The three-button pill (Add Data Source, New Request, Breach Reporting) is currently visible on *every single page* in the Control Centre.
- **Location**: Global layout (likely in `SidebarLayout` or `App.tsx` overlay).
- **Desired Fix**: Remove the global placement. Only display these actions on their relevant pages (e.g., "Add Data Source" on `/datasources`, "New Request" on `/dsr` or `/nominations`, "Breach Reporting" on `/breach`), preferably as primary action buttons in the page header rather than a floating pill.
- **Resolution**: Removed `ActionToolbar` + `globalActions` from `AppLayout.tsx`. Each page already has its own actions: DataSources (line 326), DSR (line 165 "New DSR"), Breach (line 76 "New Incident").

### 2. Dev Data Source Credentials Helper — ✅ ALREADY EXISTED
- **Issue**: In local development, the user has to constantly look up the connection strings (Postgres, MySQL, Mongo, etc.) spawned by Docker to test data sources.
- **Location**: `/datasources` (Add Data Source flow).
- **Desired Fix**: Add a "Dev Environment Connectors" helper popup or side-panel visible *only in development mode* (`import.meta.env.DEV`). This should display the ready-to-copy credentials for all locally running target databases (MySQL: 3307, Postgres: 5433, Mongo: 27018) so they can quickly be pasted into the configuration form without leaving the application.
- **Resolution**: Already implemented in `DataSources.tsx` lines 359-392 — dev helper shows inside the Add Data Source modal gated by `import.meta.env.DEV`.

### 3. Discovery Module (PII Inventory, Discovery, Lineage) Validation & UI — ⏳ UNBLOCKED
- **Issue**: These pages currently appear completely empty because no data sources have been connected. It's impossible to properly review or overhaul the UI without real data populating the charts (AI/Regex/Heuristic boxes), the inventory tables, and the lineage graph.
- **Location**: `/pii-inventory`, `/review-queue`, `/data-lineage`
- **Desired Approach**: Before attempting *any* frontend modernization of these pages, we must prove the entire backend discovery pipeline works. We need to connect the disparate local seeded databases, run real scans, and verify that the backend successfully detects PII and populates these frontend views. If the pipeline is broken, it's a **Batch 5A (P0)** fix. Once data flows correctly and rendering is proven, we can confidently upgrade the UI in **Batch 5B**.
- **Resolution**: Root cause fixed in Task #1 (type normalization). Needs live E2E test to confirm pipeline works end-to-end (Task #3).

### 4. Discovery Scans Failing for Local Databases — ✅ ROOT CAUSE FIXED (Batch 5A Task #1)
- **Issue**: After adding the Postgres, MySQL, and MongoDB data sources and clicking "Scan", the scans instantly fail. Checking the backend logs/database reveals the error: `resolve connector: unsupported data source type`.
- **Location**: `cmd/api/main.go` (Connector Registration) and Background Worker.
- **Desired Fix**: The actual connector code exists (`postgres.go`, `mysql.go`, `mongodb.go`), but they are currently **not registered** in the `ConnectorRegistry` inside `cmd/api/main.go` (only Microsoft 365 is registered). We need to wire up the existing connector constructors (`NewPostgresConnector()`, `NewMySQLConnector()`, `NewMongoDBConnector()`) into the registry so the background scanning worker can pick them up. This is a crucial **Batch 5A** fix to unblock the entire Discovery testing flow.
- **Resolution**: Connectors were already registered (lines 35-46 of `registry.go`). **Actual root cause**: Frontend sent lowercase types (`postgresql`) but backend `ConnectorRegistry` indexed by uppercase (`POSTGRESQL`) + value mismatches (`mssql`→`SQLSERVER`, `m365`→`MICROSOFT_365`, `local_file`→`FILE_UPLOAD`). Fixed with `NormalizeDataSourceType()` in `types.go` + wired into `DataSourceService.Create()` and `ConnectorRegistry.GetConnector()`.
