-- =============================================================================
-- DataLens 2.0 — Initial Database Schema
-- Migration: 001_initial_schema
-- =============================================================================
-- This migration creates the core tables for all bounded contexts.
-- All tables use UUID primary keys and include tenant isolation.
-- =============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- IDENTITY CONTEXT
-- =============================================================================

CREATE TABLE tenants (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    domain          VARCHAR(255) UNIQUE NOT NULL,
    industry        VARCHAR(100) NOT NULL DEFAULT 'GENERAL',
    country         VARCHAR(10) NOT NULL DEFAULT 'IN',
    plan            VARCHAR(50) NOT NULL DEFAULT 'FREE',
    status          VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    settings        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE roles (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    permissions     JSONB NOT NULL DEFAULT '[]',
    is_system       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email           VARCHAR(255) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    password        VARCHAR(255) NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    mfa_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE(tenant_id, email)
);

CREATE TABLE user_roles (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id         UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- =============================================================================
-- DISCOVERY CONTEXT
-- =============================================================================

CREATE TABLE data_sources (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    type            VARCHAR(50) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    host            TEXT,               -- Encrypted
    port            INT,
    database_name   VARCHAR(255),
    credentials     TEXT,               -- Encrypted
    status          VARCHAR(50) NOT NULL DEFAULT 'DISCONNECTED',
    last_sync_at    TIMESTAMPTZ,
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE data_inventories (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    data_source_id  UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    total_entities  INT NOT NULL DEFAULT 0,
    total_fields    INT NOT NULL DEFAULT 0,
    pii_fields_count INT NOT NULL DEFAULT 0,
    last_scanned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    schema_version  VARCHAR(50) NOT NULL DEFAULT '1',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE data_entities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    inventory_id    UUID NOT NULL REFERENCES data_inventories(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    schema_name     VARCHAR(255) NOT NULL DEFAULT 'public',
    type            VARCHAR(50) NOT NULL DEFAULT 'TABLE',
    row_count       BIGINT,
    pii_confidence  DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE data_fields (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id       UUID NOT NULL REFERENCES data_entities(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    data_type       VARCHAR(100) NOT NULL,
    nullable        BOOLEAN NOT NULL DEFAULT TRUE,
    is_primary_key  BOOLEAN NOT NULL DEFAULT FALSE,
    is_foreign_key  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE pii_classifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    field_id        UUID REFERENCES data_fields(id) ON DELETE SET NULL,
    data_source_id  UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    entity_name     VARCHAR(255) NOT NULL,
    field_name      VARCHAR(255) NOT NULL,
    category        VARCHAR(50) NOT NULL,
    type            VARCHAR(50) NOT NULL,
    sensitivity     VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    confidence      DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    detection_method VARCHAR(50) NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    verified_by     UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at     TIMESTAMPTZ,
    reasoning       TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE scan_runs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    data_source_id  UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type            VARCHAR(50) NOT NULL DEFAULT 'FULL',
    status          VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    progress        INT NOT NULL DEFAULT 0,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    stats           JSONB NOT NULL DEFAULT '{}',
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- GOVERNANCE CONTEXT
-- =============================================================================

CREATE TABLE purposes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(100) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    legal_basis     VARCHAR(50) NOT NULL DEFAULT 'CONSENT',
    retention_days  INT NOT NULL DEFAULT 365,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    requires_consent BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, code)
);

CREATE TABLE data_mappings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    classification_id UUID NOT NULL REFERENCES pii_classifications(id) ON DELETE CASCADE,
    purpose_ids     UUID[] NOT NULL DEFAULT '{}',
    retention_days  INT NOT NULL DEFAULT 365,
    third_party_ids UUID[] NOT NULL DEFAULT '{}',
    notes           TEXT NOT NULL DEFAULT '',
    mapped_by       UUID NOT NULL REFERENCES users(id),
    mapped_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cross_border    JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE policies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    type            VARCHAR(50) NOT NULL,
    rules           JSONB NOT NULL DEFAULT '[]',
    severity        VARCHAR(50) NOT NULL DEFAULT 'WARNING',
    actions         JSONB NOT NULL DEFAULT '[]',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE third_parties (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    type            VARCHAR(50) NOT NULL DEFAULT 'PROCESSOR',
    country         VARCHAR(10) NOT NULL,
    dpa_doc_path    TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    purpose_ids     UUID[] NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- COMPLIANCE CONTEXT
-- =============================================================================

CREATE TABLE data_subjects (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    identifier      VARCHAR(500) NOT NULL,
    identifier_type VARCHAR(50) NOT NULL DEFAULT 'EMAIL',
    display_name    VARCHAR(255),
    first_seen_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, identifier, identifier_type)
);

CREATE TABLE dsrs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subject_id      UUID NOT NULL REFERENCES data_subjects(id),
    type            VARCHAR(50) NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    priority        VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    regulation_ref  VARCHAR(50) NOT NULL DEFAULT 'DPDPA',
    requested_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deadline        TIMESTAMPTZ NOT NULL,
    requested_by    VARCHAR(255) NOT NULL,
    request_notes   TEXT NOT NULL DEFAULT '',
    assigned_to     UUID REFERENCES users(id),
    completed_at    TIMESTAMPTZ,
    completion_notes TEXT NOT NULL DEFAULT '',
    evidence_id     UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE dsr_tasks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dsr_id          UUID NOT NULL REFERENCES dsrs(id) ON DELETE CASCADE,
    data_source_id  UUID NOT NULL REFERENCES data_sources(id),
    action          VARCHAR(50) NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    executed_at     TIMESTAMPTZ,
    result          TEXT,
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE consents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subject_id      UUID NOT NULL REFERENCES data_subjects(id),
    purpose_ids     UUID[] NOT NULL,
    granted_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ,
    mechanism       VARCHAR(50) NOT NULL DEFAULT 'EXPLICIT',
    status          VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    withdrawn_at    TIMESTAMPTZ,
    withdrawal_reason TEXT,
    regulation_ref  VARCHAR(50) NOT NULL DEFAULT 'DPDPA',
    proof           JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE breaches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    detected_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    detected_by     VARCHAR(255) NOT NULL,
    type            VARCHAR(50) NOT NULL,
    severity        VARCHAR(50) NOT NULL DEFAULT 'WARNING',
    description     TEXT NOT NULL,
    affected_records BIGINT,
    status          VARCHAR(50) NOT NULL DEFAULT 'DETECTED',
    contained_at    TIMESTAMPTZ,
    resolved_at     TIMESTAMPTZ,
    authority_notified_at TIMESTAMPTZ,
    subjects_notified_at  TIMESTAMPTZ,
    root_cause      TEXT NOT NULL DEFAULT '',
    remediation     TEXT NOT NULL DEFAULT '',
    evidence_id     UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE grievances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subject_id      UUID REFERENCES data_subjects(id),
    type            VARCHAR(50) NOT NULL DEFAULT 'OTHER',
    description     TEXT NOT NULL,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    received_via    VARCHAR(100) NOT NULL DEFAULT 'PORTAL',
    status          VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    assigned_to     UUID REFERENCES users(id),
    response        TEXT,
    resolved_at     TIMESTAMPTZ,
    deadline        TIMESTAMPTZ NOT NULL,
    regulation_ref  VARCHAR(50) NOT NULL DEFAULT 'DPDPA',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- EVIDENCE CONTEXT
-- =============================================================================

CREATE TABLE audit_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_type      VARCHAR(100) NOT NULL,
    actor_id        UUID NOT NULL,
    actor_type      VARCHAR(50) NOT NULL DEFAULT 'USER',
    resource_type   VARCHAR(100) NOT NULL,
    resource_id     UUID NOT NULL,
    action          VARCHAR(100) NOT NULL,
    before_state    JSONB,
    after_state     JSONB,
    metadata        JSONB DEFAULT '{}',
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    previous_hash   VARCHAR(64) NOT NULL DEFAULT '',
    hash            VARCHAR(64) NOT NULL,
    signature       TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit events are append-only, never updated
-- Create a partial index for performance
CREATE INDEX idx_audit_events_tenant ON audit_events(tenant_id, created_at DESC);
CREATE INDEX idx_audit_events_resource ON audit_events(resource_type, resource_id);

CREATE TABLE evidence_packages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type            VARCHAR(50) NOT NULL,
    title           VARCHAR(500) NOT NULL,
    summary         TEXT NOT NULL DEFAULT '',
    event_ids       UUID[] NOT NULL DEFAULT '{}',
    documents       JSONB NOT NULL DEFAULT '[]',
    generated_for   VARCHAR(255) NOT NULL,
    generated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ,
    hash            VARCHAR(64) NOT NULL,
    signature       TEXT NOT NULL DEFAULT '',
    storage_path    TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- NOTIFICATION CONTEXT
-- =============================================================================

CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type            VARCHAR(50) NOT NULL,
    channel         VARCHAR(50) NOT NULL DEFAULT 'EMAIL',
    recipient       VARCHAR(500) NOT NULL,
    subject         VARCHAR(500) NOT NULL,
    body            TEXT NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    sent_at         TIMESTAMPTZ,
    error           TEXT,
    retry_count     INT NOT NULL DEFAULT 0,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE webhook_configs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    url             TEXT NOT NULL,
    secret          TEXT NOT NULL,
    events          TEXT[] NOT NULL DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    max_retries     INT NOT NULL DEFAULT 3,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- PERFORMANCE INDEXES
-- =============================================================================

CREATE INDEX idx_data_sources_tenant ON data_sources(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_pii_classifications_source ON pii_classifications(data_source_id);
CREATE INDEX idx_pii_classifications_status ON pii_classifications(status) WHERE status = 'PENDING';
CREATE INDEX idx_dsrs_tenant ON dsrs(tenant_id, status);
CREATE INDEX idx_dsrs_deadline ON dsrs(deadline) WHERE status NOT IN ('COMPLETED', 'REJECTED');
CREATE INDEX idx_consents_subject ON consents(subject_id, status);
CREATE INDEX idx_notifications_pending ON notifications(status) WHERE status = 'PENDING';

-- =============================================================================
-- SEED: System Roles
-- =============================================================================

INSERT INTO roles (id, name, description, permissions, is_system) VALUES
    (uuid_generate_v4(), 'ADMIN', 'Full system access', '[{"resource":"*","actions":["*"]}]', TRUE),
    (uuid_generate_v4(), 'DPO', 'Data Protection Officer', '[{"resource":"DSR","actions":["*"]},{"resource":"CONSENT","actions":["*"]},{"resource":"BREACH","actions":["*"]},{"resource":"PII","actions":["READ","VERIFY"]},{"resource":"AUDIT","actions":["READ"]}]', TRUE),
    (uuid_generate_v4(), 'ANALYST', 'Data analyst', '[{"resource":"PII","actions":["READ","VERIFY","REJECT"]},{"resource":"DATA_SOURCE","actions":["READ"]},{"resource":"AUDIT","actions":["READ"]}]', TRUE),
    (uuid_generate_v4(), 'VIEWER', 'Read-only access', '[{"resource":"*","actions":["READ"]}]', TRUE);
