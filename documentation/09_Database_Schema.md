# 09. Database Schema

## Overview

DataLens uses PostgreSQL for both the Control Centre and Agent databases. This document covers the complete schema design.

---

## Control Centre Database Schema

### Core Tables

#### Clients (Multi-Tenant)

```sql
clients (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            VARCHAR(255) NOT NULL,
  domain          VARCHAR(255),
  industry        VARCHAR(100),
  status          VARCHAR(50) DEFAULT 'ACTIVE',   -- ACTIVE, SUSPENDED, DELETED
  settings        JSONB,
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)
```

#### Client Users

```sql
client_users (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  email           VARCHAR(255) NOT NULL UNIQUE,
  password_hash   VARCHAR(255) NOT NULL,
  name            VARCHAR(255),
  role            VARCHAR(50),          -- ADMIN, DPO, ANALYST, VIEWER
  status          VARCHAR(50) DEFAULT 'ACTIVE',
  last_login_at   TIMESTAMP,
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)
```

### Agent Management

```sql
agents (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  name            VARCHAR(255) NOT NULL,
  api_key_hash    VARCHAR(255) NOT NULL,
  status          VARCHAR(50) DEFAULT 'ACTIVE',
  last_heartbeat  TIMESTAMP,
  version         VARCHAR(20),
  metadata        JSONB,                -- IP, location, deployment info
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)

agent_metadata_uploads (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_id        UUID REFERENCES agents(id),
  client_id       UUID REFERENCES clients(id),
  upload_type     VARCHAR(50),          -- PII_DISCOVERY, DATA_SUBJECT, SCAN_RUN
  payload_hash    VARCHAR(64),
  record_count    INTEGER,
  status          VARCHAR(50),
  created_at      TIMESTAMP DEFAULT NOW()
)
```

### PII Inventory

```sql
-- Standard PII categories
pii_category_master (
  id              SERIAL PRIMARY KEY,
  code            VARCHAR(50) UNIQUE,   -- EMAIL_ADDRESS, PHONE_NUMBER, etc.
  name            VARCHAR(100),
  description     TEXT,
  sensitivity     VARCHAR(20),          -- LOW, MEDIUM, HIGH, CRITICAL
  is_active       BOOLEAN DEFAULT true
)

-- Verified PII inventory (after human review)
client_pii_column_mapping (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  agent_id        UUID REFERENCES agents(id),
  data_source_id  INTEGER,
  object_identifier VARCHAR(500),        -- schema.table.column
  pii_category_code VARCHAR(50),
  sensitivity     VARCHAR(20),
  purpose_ids     UUID[],               -- Linked purposes
  lawful_basis    VARCHAR(50),
  is_verified     BOOLEAN DEFAULT false,
  verified_by     UUID REFERENCES client_users(id),
  verified_at     TIMESTAMP,
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)

-- PII candidates pending review
pii_discovery_queue (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  agent_id        UUID REFERENCES agents(id),
  data_source_id  INTEGER,
  object_identifier VARCHAR(500),
  detected_pii_category VARCHAR(50),
  confidence_score DECIMAL(3,2),
  detection_method VARCHAR(50),
  sample_data     TEXT,                 -- Masked/anonymized
  status          VARCHAR(50),          -- PENDING, VERIFIED, REJECTED
  reviewed_by     UUID REFERENCES client_users(id),
  reviewed_at     TIMESTAMP,
  created_at      TIMESTAMP DEFAULT NOW()
)
```

### Purpose & Consent

```sql
-- Processing purposes (template)
purpose_master (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code            VARCHAR(50),
  name            VARCHAR(255),
  description     TEXT,
  lawful_basis    VARCHAR(50),          -- CONSENT, CONTRACT, LEGAL_OBLIGATION
  is_system       BOOLEAN DEFAULT false,
  created_at      TIMESTAMP DEFAULT NOW()
)

-- Client-specific purposes
client_purpose_definitions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  purpose_master_id UUID REFERENCES purpose_master(id),
  name            VARCHAR(255),
  description     TEXT,
  lawful_basis    VARCHAR(50),
  is_required     BOOLEAN DEFAULT false,
  default_consent BOOLEAN DEFAULT false,
  is_active       BOOLEAN DEFAULT true,
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)

-- PII to purpose mapping
pii_purpose_mapping (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  pii_mapping_id  UUID REFERENCES client_pii_column_mapping(id),
  purpose_id      UUID REFERENCES client_purpose_definitions(id),
  lawful_basis    VARCHAR(50),
  created_at      TIMESTAMP DEFAULT NOW()
)

-- Consent records
consent_records (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  data_subject_id UUID,
  purpose_id      UUID REFERENCES client_purpose_definitions(id),
  consent_status  VARCHAR(20),          -- GRANTED, DENIED, WITHDRAWN
  consent_method  VARCHAR(50),          -- BANNER, FORM, API
  ip_address      INET,
  user_agent      TEXT,
  notice_version  VARCHAR(20),
  granted_at      TIMESTAMP,
  expires_at      TIMESTAMP,
  withdrawn_at    TIMESTAMP,
  created_at      TIMESTAMP DEFAULT NOW()
)
```

### Recipients

```sql
-- Internal departments
client_departments (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  name            VARCHAR(255),
  description     TEXT,
  head_user_id    UUID REFERENCES client_users(id),
  is_active       BOOLEAN DEFAULT true,
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)

-- External third parties
client_third_parties (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  name            VARCHAR(255),
  type            VARCHAR(50),          -- PROCESSOR, CONTROLLER, PARTNER
  country         VARCHAR(100),
  contract_ref    VARCHAR(255),
  contract_expiry DATE,
  dpa_signed      BOOLEAN DEFAULT false,
  is_active       BOOLEAN DEFAULT true,
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)

-- PII sharing
pii_recipient_mapping (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  pii_mapping_id  UUID REFERENCES client_pii_column_mapping(id),
  recipient_type  VARCHAR(20),          -- DEPARTMENT, THIRD_PARTY
  department_id   UUID REFERENCES client_departments(id),
  third_party_id  UUID REFERENCES client_third_parties(id),
  purpose_id      UUID REFERENCES client_purpose_definitions(id),
  created_at      TIMESTAMP DEFAULT NOW()
)
```

### DSR Management

```sql
dsr_requests (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  data_subject_id UUID,
  request_type    VARCHAR(20),          -- ACCESS, ERASURE, RECTIFICATION
  status          VARCHAR(50),          -- SUBMITTED, VERIFIED, IN_PROGRESS, COMPLETED
  priority        INTEGER DEFAULT 3,
  submitted_at    TIMESTAMP DEFAULT NOW(),
  due_date        TIMESTAMP,
  verified_at     TIMESTAMP,
  completed_at    TIMESTAMP,
  assigned_to     UUID REFERENCES client_users(id),
  notes           TEXT,
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)

dsr_affected_data (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  dsr_request_id  UUID REFERENCES dsr_requests(id),
  pii_mapping_id  UUID REFERENCES client_pii_column_mapping(id),
  action_taken    VARCHAR(50),          -- DELETED, ANONYMIZED, EXPORTED
  record_count    INTEGER,
  executed_at     TIMESTAMP,
  agent_id        UUID REFERENCES agents(id),
  created_at      TIMESTAMP DEFAULT NOW()
)

dsr_activity_log (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  dsr_request_id  UUID REFERENCES dsr_requests(id),
  action          VARCHAR(100),
  performed_by    UUID REFERENCES client_users(id),
  details         JSONB,
  created_at      TIMESTAMP DEFAULT NOW()
)
```

### Grievance Management

```sql
grievances (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  data_subject_id UUID,
  subject         VARCHAR(255),
  description     TEXT,
  status          VARCHAR(50),          -- OPEN, IN_PROGRESS, RESOLVED, ESCALATED
  priority        INTEGER DEFAULT 3,
  assigned_to     UUID REFERENCES client_users(id),
  resolution      TEXT,
  submitted_at    TIMESTAMP DEFAULT NOW(),
  due_date        TIMESTAMP,
  resolved_at     TIMESTAMP,
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)
```

### Auditing

```sql
audit_logs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  user_id         UUID REFERENCES client_users(id),
  action          VARCHAR(100),         -- LOGIN, CREATE_DSR, VERIFY_PII, etc.
  resource_type   VARCHAR(50),          -- DSR, PII, USER, AGENT
  resource_id     UUID,
  old_values      JSONB,
  new_values      JSONB,
  ip_address      INET,
  user_agent      TEXT,
  created_at      TIMESTAMP DEFAULT NOW()
)

-- Index for audit log queries
CREATE INDEX idx_audit_logs_client_date ON audit_logs(client_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
```

### Compliance Reports

```sql
compliance_reports (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id       UUID REFERENCES clients(id),
  report_type     VARCHAR(50),          -- ROPA, DATA_INVENTORY, DSR_SUMMARY
  title           VARCHAR(255),
  generated_by    UUID REFERENCES client_users(id),
  parameters      JSONB,
  content         JSONB,
  file_path       VARCHAR(500),
  generated_at    TIMESTAMP DEFAULT NOW(),
  created_at      TIMESTAMP DEFAULT NOW()
)
```

---

## Agent Database Schema

### Configuration

```sql
agent_configuration (
  id              SERIAL PRIMARY KEY,
  key             VARCHAR(100) UNIQUE,
  value           TEXT,
  encrypted       BOOLEAN DEFAULT false,
  updated_at      TIMESTAMP DEFAULT NOW()
)
```

### Data Sources

```sql
data_sources (
  id              SERIAL PRIMARY KEY,
  name            VARCHAR(255) NOT NULL,
  type            VARCHAR(50),          -- postgresql, mysql, filesystem, s3
  connection_details BYTEA,             -- Encrypted JSON
  status          VARCHAR(50),
  last_scan_at    TIMESTAMP,
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)
```

### PII Discovery

```sql
pii_detection_candidates (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  data_source_id  INTEGER REFERENCES data_sources(id),
  object_identifier VARCHAR(500),
  pii_category    VARCHAR(100),
  confidence_score DECIMAL(3,2),
  detection_method VARCHAR(50),
  sample_data     TEXT,
  status          VARCHAR(50),          -- pending, synced, rejected
  synced_at       TIMESTAMP,
  created_at      TIMESTAMP DEFAULT NOW()
)

-- Verified mappings from CONTROL CENTRE
verified_pii_mappings (
  id              UUID PRIMARY KEY,     -- From CONTROL CENTRE
  data_source_id  INTEGER REFERENCES data_sources(id),
  object_identifier VARCHAR(500),
  pii_category    VARCHAR(100),
  purpose_ids     UUID[],
  lawful_basis    VARCHAR(50),
  synced_at       TIMESTAMP DEFAULT NOW()
)
```

### Data Subjects

```sql
data_subjects (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  global_subject_id UUID,               -- Cross-agent identifier
  type            VARCHAR(50),          -- employee, customer, vendor
  primary_identifier VARCHAR(255),
  name            VARCHAR(255),
  status          VARCHAR(50),
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
)

data_subject_identifiers (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id      UUID REFERENCES data_subjects(id),
  identifier_type VARCHAR(50),          -- email, phone, employee_id
  identifier_value VARCHAR(255),
  is_primary      BOOLEAN DEFAULT false,
  created_at      TIMESTAMP DEFAULT NOW()
)

data_subject_pii_locations (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id      UUID REFERENCES data_subjects(id),
  data_source_id  INTEGER REFERENCES data_sources(id),
  pii_category    VARCHAR(100),
  object_identifier VARCHAR(500),
  created_at      TIMESTAMP DEFAULT NOW()
)
```

### Multi-Agent

```sql
peer_agents (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_id        UUID,
  name            VARCHAR(255),
  endpoint        VARCHAR(500),
  status          VARCHAR(50),
  last_seen_at    TIMESTAMP,
  created_at      TIMESTAMP DEFAULT NOW()
)

global_pii_registry (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  fingerprint     VARCHAR(64),          -- SHA256 hash
  pii_category    VARCHAR(100),
  first_seen_agent UUID,
  global_id       UUID,
  created_at      TIMESTAMP DEFAULT NOW()
)

global_data_subject_registry (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  global_subject_id UUID,
  first_seen_agent UUID,
  identifier_hash VARCHAR(64),
  created_at      TIMESTAMP DEFAULT NOW()
)
```

### Data Lineage

```sql
data_lineage_edges (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_node_type VARCHAR(50),         -- DATA_SOURCE, FILE, TABLE
  source_node_id  VARCHAR(500),
  target_node_type VARCHAR(50),
  target_node_id  VARCHAR(500),
  relationship_type VARCHAR(50),        -- COPY, SHARE, DERIVE
  confidence      DECIMAL(3,2),
  created_at      TIMESTAMP DEFAULT NOW()
)
```

### Scan History

```sql
scan_runs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  data_source_id  INTEGER REFERENCES data_sources(id),
  status          VARCHAR(50),
  tables_scanned  INTEGER,
  columns_scanned INTEGER,
  pii_found       INTEGER,
  error_message   TEXT,
  started_at      TIMESTAMP,
  completed_at    TIMESTAMP,
  created_at      TIMESTAMP DEFAULT NOW()
)
```

### DSR Tasks

```sql
dsr_tasks (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  CONTROL CENTRE_id         UUID,                 -- ID from CONTROL CENTRE
  request_type    VARCHAR(20),
  global_subject_id UUID,
  scope           JSONB,
  status          VARCHAR(50),
  priority        INTEGER,
  due_date        TIMESTAMP,
  received_at     TIMESTAMP DEFAULT NOW(),
  started_at      TIMESTAMP,
  completed_at    TIMESTAMP,
  result          JSONB,
  created_at      TIMESTAMP DEFAULT NOW()
)
```

---

## Indexing Strategy

### Common Indexes

```sql
-- Multi-tenant isolation
CREATE INDEX idx_<table>_client_id ON <table>(client_id);

-- Time-based queries
CREATE INDEX idx_<table>_created_at ON <table>(created_at DESC);

-- Status filtering
CREATE INDEX idx_<table>_status ON <table>(status);

-- Combined indexes for common queries
CREATE INDEX idx_pii_queue_client_status 
  ON pii_discovery_queue(client_id, status);

CREATE INDEX idx_dsr_client_status_due 
  ON dsr_requests(client_id, status, due_date);
```
