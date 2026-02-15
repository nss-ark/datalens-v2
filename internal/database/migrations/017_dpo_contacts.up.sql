CREATE TABLE IF NOT EXISTS dpo_contacts (
    tenant_id UUID PRIMARY KEY,
    org_name TEXT NOT NULL,
    dpo_name TEXT NOT NULL,
    dpo_email TEXT NOT NULL,
    dpo_phone TEXT,
    address TEXT,
    website_url TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for tenant lookups (primary key covers it, but good for explicit documentation/completeness check)
-- No additional index needed as tenant_id is PK.
