# 03. DataLens Agent (v2)

## Overview

The DataLens Agent is a lightweight application deployed **inside client infrastructure** to discover and manage personal data. It communicates with the CONTROL CENTRE platform but keeps all actual PII within the client's environment.

---

## Agent Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          DATALENS AGENT (V2)                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                           HANDLERS LAYER                             │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐  │    │
│  │  │  Config  │ │DataSource│ │   PII    │ │  Data    │ │   Peer    │  │    │
│  │  │ Handler  │ │ Handler  │ │ Handler  │ │ Subject  │ │  Handler  │  │    │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └───────────┘  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                     │                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                          SERVICES LAYER                              │    │
│  │                                                                       │    │
│  │  ┌─────────────────┐   ┌─────────────────┐   ┌──────────────────┐   │    │
│  │  │  PII Detection  │   │   File Scanner  │   │ Database Scanner │   │    │
│  │  │     Service     │   │     Service     │   │     Service      │   │    │
│  │  └─────────────────┘   └─────────────────┘   └──────────────────┘   │    │
│  │                                                                       │    │
│  │  ┌─────────────────┐   ┌─────────────────┐   ┌──────────────────┐   │    │
│  │  │ S3 Scanner      │   │ Salesforce      │   │  IMAP Scanner    │   │    │
│  │  │    Service      │   │ Scanner Service │   │     Service      │   │    │
│  │  └─────────────────┘   └─────────────────┘   └──────────────────┘   │    │
│  │                                                                       │    │
│  │  ┌─────────────────┐   ┌─────────────────┐   ┌──────────────────┐   │    │
│  │  │  Deduplication  │   │    Lineage      │   │   DSR Executor   │   │    │
│  │  │    Service      │   │    Service      │   │     Service      │   │    │
│  │  └─────────────────┘   └─────────────────┘   └──────────────────┘   │    │
│  │                                                                       │    │
│  │  ┌─────────────────┐   ┌─────────────────┐   ┌──────────────────┐   │    │
│  │  │  Encryption     │   │  Audit Logger   │   │   NLP Client     │   │    │
│  │  │    Service      │   │    Service      │   │    Service       │   │    │
│  │  └─────────────────┘   └─────────────────┘   └──────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                     │                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         REPOSITORY LAYER                             │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐  │    │
│  │  │DataSource│ │   PII    │ │   Data   │ │ ScanRun  │ │   DSR     │  │    │
│  │  │   Repo   │ │Candidate │ │ Subject  │ │   Repo   │ │  Repo     │  │    │
│  │  │          │ │   Repo   │ │   Repo   │ │          │ │           │  │    │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └───────────┘  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                     │                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         LOCAL DATABASE                               │    │
│  │                          (PostgreSQL)                                │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         EXTERNAL SERVICES                            │    │
│  │  ┌───────────────────┐     ┌────────────────────────────────────┐   │    │
│  │  │  NLP Microservice │     │        CONTROL CENTRE Sync Client            │   │    │
│  │  │  (Python/spaCy)   │     │     (HTTPS to DataLens Cloud)      │   │    │
│  │  └───────────────────┘     └────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
DataLensAgent/
├── backend/                      # Go backend application
│   ├── main.go                   # Entry point (690 lines)
│   ├── config/                   # Configuration management
│   ├── database/                 # Database connection
│   ├── handlers/                 # API endpoint handlers
│   │   ├── config_handler.go
│   │   ├── datasource_handler.go
│   │   ├── pii_handler.go
│   │   ├── data_subject_handler.go
│   │   └── peer_handler.go
│   ├── models/                   # Data models
│   ├── peer/                     # Agent-to-agent communication
│   ├── repository/               # Database access layer
│   ├── services/                 # Business logic (26+ files)
│   └── sync/                     # CONTROL CENTRE synchronization
├── frontend/                     # React admin console
├── nlp-service/                  # Python NLP microservice
│   ├── app.py                    # Flask application
│   └── requirements.txt
├── kubernetes/                   # K8s deployment configs
├── terraform/                    # Infrastructure as Code
└── docker-compose.yml
```

---

## Handlers (API Endpoints)

| Handler | Endpoints | Purpose |
|---------|-----------|---------|
| `config_handler.go` | `/api/status`, `/api/config` | Agent configuration, health checks |
| `datasource_handler.go` | `/api/datasources/*` | CRUD for data sources, connection testing |
| `pii_handler.go` | `/api/pii/*` | PII scan operations, results |
| `data_subject_handler.go` | `/api/subjects/*` | Data subject management |
| `peer_handler.go` | `/api/peer/*` | Agent-to-agent communication |

---

## Services (Business Logic)

### Core Services

| Service | File | Purpose |
|---------|------|---------|
| PII Detection | `pii_detection.go` | Pattern matching, column heuristics, NLP integration |
| Database Scanner | `database_scanner.go` | PostgreSQL, MySQL scanning |
| File Scanner | `file_scanner.go` | File system scanning, text extraction |
| S3 Scanner | `s3_scanner.go` | Amazon S3 bucket scanning |
| Salesforce Scanner | `salesforce_scanner.go` | Salesforce CRM scanning |
| IMAP Scanner | `imap_scanner.go` | Email server scanning |
| MongoDB Scanner | `mongodb_scanner.go` | MongoDB document scanning |

### Support Services

| Service | File | Purpose |
|---------|------|---------|
| Deduplication | `deduplication_service.go` | Prevent duplicate PII detections |
| Lineage | `lineage_service.go` | Track data flow between sources |
| DSR Executor | `dsr_executor.go` | Execute deletion/access requests locally |
| Encryption | `encryption_service.go` | Encrypt connection credentials |
| Audit Logger | `audit_logger.go` | Log all operations |
| NLP Client | `nlp_client.go` | Interface to Python NLP service |
| LLM Client | `llm_client.go` | AI/LLM integration |
| OCR Service | `ocr_service.go` | Extract text from images |

---

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `CONTROL CENTRE_ENDPOINT` | Yes | URL of DataLens CONTROL CENTRE |
| `AGENT_API_KEY` | Yes | Authentication key |
| `CLIENT_ID` | Yes | Client identifier |
| `AGENT_ID` | Yes | This agent's identifier |
| `DATABASE_URL` | Yes | Local PostgreSQL connection |
| `NLP_SERVICE_URL` | No | Python NLP service URL |
| `LLM_API_KEY` | No | OpenAI-compatible API key |
| `ENCRYPTION_KEY` | Yes | Key for encrypting credentials |

### Configuration File (agent-config.yaml)

```yaml
agent:
  id: "agent-001"
  name: "HR Department Agent"
  
CONTROL CENTRE:
  endpoint: "https://datalens.complyark.com"
  sync_interval: "5m"
  
database:
  host: "localhost"
  port: 5432
  name: "datalens_agent"
  
nlp:
  enabled: true
  service_url: "http://localhost:5000"
  
scanning:
  sample_size: 100
  max_concurrent_scans: 3
  
peers:
  enabled: true
  discovery: "manual"
```

---

## Local Database Schema

### Core Tables

```sql
-- Data sources configured for scanning
data_sources (
  id                SERIAL PRIMARY KEY,
  name              VARCHAR(255),
  type              VARCHAR(50),      -- postgresql, mysql, filesystem, etc.
  connection_details JSONB,           -- Encrypted connection info
  status            VARCHAR(50),
  last_scan_at      TIMESTAMP,
  created_at        TIMESTAMP
)

-- Discovered PII candidates
pii_detection_candidates (
  id                UUID PRIMARY KEY,
  data_source_id    INTEGER REFERENCES data_sources(id),
  object_identifier VARCHAR(500),     -- e.g., "schema.table.column"
  pii_category      VARCHAR(100),
  confidence_score  DECIMAL(3,2),
  detection_method  VARCHAR(50),
  sample_data       TEXT,             -- Masked/anonymized
  status            VARCHAR(50),      -- pending, verified, false_positive
  created_at        TIMESTAMP
)

-- Discovered data subjects
data_subjects (
  id                UUID PRIMARY KEY,
  global_subject_id UUID,             -- Cross-agent identifier
  type              VARCHAR(50),      -- employee, customer, vendor
  primary_identifier VARCHAR(255),
  name              VARCHAR(255),
  status            VARCHAR(50),
  created_at        TIMESTAMP
)

-- PII locations for each subject
data_subject_pii_locations (
  id                UUID PRIMARY KEY,
  subject_id        UUID REFERENCES data_subjects(id),
  data_source_id    INTEGER REFERENCES data_sources(id),
  pii_category      VARCHAR(100),
  object_identifier VARCHAR(500),
  created_at        TIMESTAMP
)

-- Scan run history
scan_runs (
  id                UUID PRIMARY KEY,
  data_source_id    INTEGER REFERENCES data_sources(id),
  status            VARCHAR(50),
  tables_scanned    INTEGER,
  pii_found         INTEGER,
  started_at        TIMESTAMP,
  completed_at      TIMESTAMP
)
```

---

## Multi-Agent Communication

### Peer Discovery

Agents discover each other through:
1. **Manual Configuration** - List peer agents in config
2. **CONTROL CENTRE Registry** - Query CONTROL CENTRE for registered agents
3. **Multicast Discovery** - UDP multicast on local network

### Message Types

| Message | Direction | Purpose |
|---------|-----------|---------|
| `HEARTBEAT` | Agent → Peers | Health check, online status |
| `PII_DISCOVERED` | Agent → Peers | Notify about new PII detection |
| `DUPLICATE_CHECK` | Agent → Peers | Ask "Have you seen this?" |
| `DATA_SUBJECT_FOUND` | Agent → Peers | Notify about new data subject |
| `LINEAGE_UPDATE` | Agent → Peers | Share data flow information |

### Deduplication Flow

```
Agent A finds email "john@example.com":
1. Check local registry → Not found
2. Query peer agents → Agent B already found it
3. Link to Agent B's discovery (don't create duplicate)
4. Report to CONTROL CENTRE with cross-agent reference
```

---

## CONTROL CENTRE Synchronization

### Upload Operations

| Operation | Frequency | Data Sent |
|-----------|-----------|-----------|
| PII Discovery Upload | On discovery | Object identifier, category, confidence |
| Data Subject Upload | On discovery | Subject type, encrypted identifiers |
| Scan Run Report | On completion | Statistics, duration, errors |
| Agent Heartbeat | Every 30s | Status, resource usage |

### Download Operations

| Operation | Frequency | Data Received |
|-----------|-----------|---------------|
| Verified PII Mappings | On change | Verified fields, purposes |
| DSR Tasks | Polling (1m) | Tasks to execute |
| Configuration Updates | On change | Scan schedules, settings |
| Peer Agent List | On change | Other agents to sync with |
