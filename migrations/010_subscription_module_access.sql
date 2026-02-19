-- =============================================================================
-- DataLens 2.0 — Subscription & Module Access
-- Migration: 010_subscription_module_access
-- =============================================================================
-- Adds per-tenant subscription tracking and per-tenant module toggles.
-- =============================================================================

-- Subscription: one row per tenant, tracks plan + billing lifecycle
CREATE TABLE IF NOT EXISTS subscriptions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    plan            VARCHAR(50)  NOT NULL DEFAULT 'FREE',
    billing_start   TIMESTAMPTZ,
    billing_end     TIMESTAMPTZ,
    auto_revoke     BOOLEAN      NOT NULL DEFAULT TRUE,
    status          VARCHAR(50)  NOT NULL DEFAULT 'ACTIVE',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Module Access: per-tenant feature toggles
CREATE TABLE IF NOT EXISTS module_access (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    module_name     VARCHAR(100) NOT NULL,
    enabled         BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, module_name)
);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_module_access_tenant ON module_access(tenant_id);
