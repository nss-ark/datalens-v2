-- Migration 022: Purpose Assignments (Multi-Level Scope Tagging)
CREATE TABLE IF NOT EXISTS purpose_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    purpose_id UUID NOT NULL REFERENCES purposes(id) ON DELETE CASCADE,
    scope_type VARCHAR(20) NOT NULL,        -- SERVER, DATABASE, SCHEMA, TABLE, COLUMN
    scope_id VARCHAR(500) NOT NULL,         -- identifier for the scope target
    scope_name VARCHAR(500),                -- human-readable name
    inherited BOOLEAN NOT NULL DEFAULT false,
    overridden_by UUID REFERENCES purpose_assignments(id),
    assigned_by UUID,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, purpose_id, scope_type, scope_id)
);
CREATE INDEX IF NOT EXISTS idx_purpose_assignments_tenant ON purpose_assignments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_purpose_assignments_scope ON purpose_assignments(tenant_id, scope_type, scope_id);
CREATE INDEX IF NOT EXISTS idx_purpose_assignments_purpose ON purpose_assignments(purpose_id);
