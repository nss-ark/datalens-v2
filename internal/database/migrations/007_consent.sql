-- 007_consent.sql
-- Consent Engine tables: widgets, sessions, history
-- Sessions and history are APPEND-ONLY (compliance requirement — no UPDATE/DELETE)

-- =============================================================================
-- consent_widgets — Embeddable consent widget configurations
-- =============================================================================
CREATE TABLE IF NOT EXISTS consent_widgets (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    name            VARCHAR(255) NOT NULL,
    type            VARCHAR(50) NOT NULL DEFAULT 'BANNER',
    domain          VARCHAR(255) NOT NULL DEFAULT '',
    status          VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    config          JSONB NOT NULL DEFAULT '{}',
    embed_code      TEXT NOT NULL DEFAULT '',
    api_key         VARCHAR(128) NOT NULL,
    allowed_origins TEXT[] NOT NULL DEFAULT '{}',
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_consent_widgets_tenant ON consent_widgets(tenant_id);
CREATE UNIQUE INDEX idx_consent_widgets_api_key ON consent_widgets(api_key);
CREATE INDEX idx_consent_widgets_domain ON consent_widgets(domain);

-- =============================================================================
-- consent_sessions — Immutable consent interaction records
-- =============================================================================
CREATE TABLE IF NOT EXISTS consent_sessions (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    widget_id       UUID NOT NULL REFERENCES consent_widgets(id),
    subject_id      UUID,
    decisions       JSONB NOT NULL DEFAULT '[]',
    ip_address      INET,
    user_agent      TEXT NOT NULL DEFAULT '',
    page_url        TEXT NOT NULL DEFAULT '',
    widget_version  INT NOT NULL DEFAULT 1,
    notice_version  VARCHAR(100) NOT NULL DEFAULT '',
    signature       VARCHAR(256) NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_consent_sessions_tenant_widget ON consent_sessions(tenant_id, widget_id);
CREATE INDEX idx_consent_sessions_tenant_subject ON consent_sessions(tenant_id, subject_id);

-- =============================================================================
-- consent_history — Immutable consent timeline entries
-- =============================================================================
CREATE TABLE IF NOT EXISTS consent_history (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    subject_id      UUID NOT NULL,
    widget_id       UUID,
    purpose_id      UUID NOT NULL,
    purpose_name    VARCHAR(255) NOT NULL DEFAULT '',
    previous_status VARCHAR(20),
    new_status      VARCHAR(20) NOT NULL,
    source          VARCHAR(50) NOT NULL DEFAULT 'BANNER',
    ip_address      INET,
    user_agent      TEXT NOT NULL DEFAULT '',
    notice_version  VARCHAR(100) NOT NULL DEFAULT '',
    signature       VARCHAR(256) NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_consent_history_tenant_subject ON consent_history(tenant_id, subject_id);
CREATE INDEX idx_consent_history_tenant_purpose ON consent_history(tenant_id, purpose_id);
