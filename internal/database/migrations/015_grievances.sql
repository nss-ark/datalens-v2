-- Grievance Redressal Module
-- DPDPA Requirement: Mechanism for data principals to lodge complaints.

CREATE TABLE IF NOT EXISTS grievances (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    data_subject_id UUID NOT NULL,
    subject TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL,
    status TEXT NOT NULL,
    priority INT NOT NULL DEFAULT 0,
    assigned_to UUID,
    resolution TEXT,
    submitted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    escalated_to TEXT,
    feedback_rating INT,
    feedback_comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_grievances_tenant_status ON grievances (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_grievances_tenant_subject ON grievances (tenant_id, data_subject_id);
CREATE INDEX IF NOT EXISTS idx_grievances_assigned_to ON grievances (assigned_to);
