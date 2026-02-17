-- Migration 019: Retention tables
-- Creates retention_policies and retention_logs to support DPDP Rule R8 (data retention and erasure).

CREATE TABLE IF NOT EXISTS retention_policies (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    purpose_id UUID NOT NULL,
    max_retention_days INT NOT NULL,
    data_categories TEXT[] NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    auto_erase BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_retention_policies_tenant ON retention_policies(tenant_id);

CREATE TABLE IF NOT EXISTS retention_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    policy_id UUID NOT NULL REFERENCES retention_policies(id),
    action VARCHAR(50) NOT NULL,
    target TEXT NOT NULL,
    details TEXT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_retention_logs_tenant_policy ON retention_logs(tenant_id, policy_id);
