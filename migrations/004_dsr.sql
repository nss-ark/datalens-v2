CREATE TABLE IF NOT EXISTS dsr_requests (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    request_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    subject_name TEXT NOT NULL,
    subject_email TEXT NOT NULL,
    subject_identifiers JSONB NOT NULL DEFAULT '{}',
    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',
    sla_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    assigned_to UUID,
    reason TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_dsr_tenant_id ON dsr_requests(tenant_id);
CREATE INDEX idx_dsr_status ON dsr_requests(status);
CREATE INDEX idx_dsr_sla_deadline ON dsr_requests(sla_deadline);
CREATE INDEX idx_dsr_subject_email ON dsr_requests(subject_email);

CREATE TABLE IF NOT EXISTS dsr_tasks (
    id UUID PRIMARY KEY,
    dsr_id UUID NOT NULL REFERENCES dsr_requests(id) ON DELETE CASCADE,
    data_source_id UUID NOT NULL, -- Logical FK to data_sources
    tenant_id UUID NOT NULL,
    task_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    result JSONB,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Fix compatibility with 001 if exists
ALTER TABLE dsr_tasks ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE dsr_tasks ADD COLUMN IF NOT EXISTS task_type VARCHAR(50) NOT NULL DEFAULT 'ACCESS';
ALTER TABLE dsr_tasks ADD COLUMN IF NOT EXISTS error TEXT;
ALTER TABLE dsr_tasks ALTER COLUMN result TYPE JSONB USING result::JSONB;
ALTER TABLE dsr_tasks ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE dsr_tasks DROP COLUMN IF EXISTS action;

-- Drop old FK to dsrs and add new FK to dsr_requests
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'dsr_tasks_dsr_id_fkey') THEN
        ALTER TABLE dsr_tasks DROP CONSTRAINT dsr_tasks_dsr_id_fkey;
    END IF;
END $$;

ALTER TABLE dsr_tasks ADD CONSTRAINT dsr_tasks_dsr_id_fkey FOREIGN KEY (dsr_id) REFERENCES dsr_requests(id) ON DELETE CASCADE;

CREATE INDEX idx_dsr_task_dsr_id ON dsr_tasks(dsr_id);
CREATE INDEX idx_dsr_task_tenant_id ON dsr_tasks(tenant_id);
CREATE INDEX idx_dsr_task_datasource_id ON dsr_tasks(data_source_id);
