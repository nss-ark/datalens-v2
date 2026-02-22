-- Migration 021: RoPA (Record of Processing Activities) + Third Parties
-- Third-party processors/controllers
CREATE TABLE IF NOT EXISTS third_parties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,              -- PROCESSOR, CONTROLLER, VENDOR
    country VARCHAR(100) NOT NULL DEFAULT '',
    dpa_doc_path TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    purpose_ids JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_third_parties_tenant ON third_parties(tenant_id);

-- RoPA Versions (version-controlled snapshots)
CREATE TABLE IF NOT EXISTS ropa_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    version VARCHAR(20) NOT NULL,           -- semver: "1.0", "1.1", "2.0"
    generated_by VARCHAR(100) NOT NULL,     -- "auto" or user_id UUID string
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT', -- DRAFT, PUBLISHED, ARCHIVED
    content JSONB NOT NULL,                 -- full RoPA content snapshot
    change_summary TEXT,                    -- what changed from previous version
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, version)
);
CREATE INDEX IF NOT EXISTS idx_ropa_versions_tenant ON ropa_versions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ropa_versions_tenant_created ON ropa_versions(tenant_id, created_at DESC);
