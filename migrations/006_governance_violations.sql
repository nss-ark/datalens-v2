-- =============================================================================
-- DataLens 2.0 — Governance Updates
-- Migration: 006_governance_violations
-- =============================================================================

CREATE TABLE violations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    policy_id       UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    data_source_id  UUID NOT NULL REFERENCES data_sources(id), -- No cascade to keep history
    entity_name     VARCHAR(255) NOT NULL,
    field_name      VARCHAR(255) NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    severity        VARCHAR(50) NOT NULL DEFAULT 'WARNING',
    detected_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ,
    resolved_by     UUID REFERENCES users(id),
    resolution      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_violations_tenant_status ON violations(tenant_id, status);
CREATE INDEX idx_violations_policy ON violations(policy_id);
