CREATE TABLE IF NOT EXISTS breach_incidents (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    type VARCHAR(255) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
    reported_to_cert_in_at TIMESTAMP WITH TIME ZONE,
    reported_to_dpb_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    
    affected_systems TEXT[], -- Array of strings
    affected_data_subject_count INTEGER DEFAULT 0,
    pii_categories TEXT[], -- Array of strings
    
    is_reportable_cert_in BOOLEAN DEFAULT FALSE,
    is_reportable_dpb BOOLEAN DEFAULT FALSE,
    
    poc_name VARCHAR(255),
    poc_role VARCHAR(255),
    poc_email VARCHAR(255),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_breach_incidents_tenant_id ON breach_incidents(tenant_id);
CREATE INDEX idx_breach_incidents_status ON breach_incidents(status);
CREATE INDEX idx_breach_incidents_severity ON breach_incidents(severity);
