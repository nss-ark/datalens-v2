CREATE TYPE retention_policy_status AS ENUM ('ACTIVE', 'PAUSED');

CREATE TABLE IF NOT EXISTS retention_policies (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    purpose_id UUID NOT NULL, -- Logical link to purpose (no FK enforced if purpose is in another service/table, but ideally should reference purposes)
    max_retention_days INTEGER NOT NULL,
    data_categories TEXT[] DEFAULT '{}',
    status retention_policy_status NOT NULL DEFAULT 'ACTIVE',
    auto_erase BOOLEAN NOT NULL DEFAULT false,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_retention_policies_tenant ON retention_policies(tenant_id);
