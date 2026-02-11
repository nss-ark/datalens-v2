CREATE TABLE IF NOT EXISTS data_flows (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    destination_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    data_type TEXT NOT NULL, -- e.g. "TABLE", "COLUMN", "FILE"
    data_path TEXT NOT NULL, -- e.g. "public.users" or "s3://bucket/path"
    purpose_id UUID REFERENCES purposes(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_data_flows_tenant_id ON data_flows(tenant_id);
CREATE INDEX idx_data_flows_source_id ON data_flows(source_id);
CREATE INDEX idx_data_flows_destination_id ON data_flows(destination_id);
