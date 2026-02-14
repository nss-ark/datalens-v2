-- Migration: 006_scan_schedule
-- Description: Add scan scheduling support to data sources

ALTER TABLE data_sources 
ADD COLUMN IF NOT EXISTS scan_schedule TEXT DEFAULT NULL;

COMMENT ON COLUMN data_sources.scan_schedule IS 'Cron expression for automated scans (e.g., "0 2 * * *" for daily at 2 AM)';
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
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    actor_id UUID NOT NULL,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255) NOT NULL,
    resource_id UUID NOT NULL,
    changes JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
ALTER TABLE dsr_requests ADD COLUMN metadata JSONB;
COMMENT ON COLUMN dsr_requests.metadata IS 'Flexible storage for request-specific details (e.g., Nominee info, Portability formats)';
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
-- Create identity_profiles table
-- This stores the assurance level and verification documents for a subject

CREATE TABLE IF NOT EXISTS identity_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES data_subjects(id),
    assurance_level TEXT NOT NULL,
    verification_status TEXT NOT NULL,
    documents JSONB,
    last_verified_at TIMESTAMPTZ,
    next_verification_due TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Indexes for efficient lookup
CREATE UNIQUE INDEX idx_identity_profiles_subject ON identity_profiles(tenant_id, subject_id);
CREATE INDEX idx_identity_profiles_status ON identity_profiles(verification_status, tenant_id);
ALTER TABLE data_flows ADD COLUMN transformation TEXT;
ALTER TABLE data_flows ADD COLUMN confidence DOUBLE PRECISION DEFAULT 1.0;
-- 014_consent_notices.sql

CREATE TABLE IF NOT EXISTS consent_notices (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    series_id UUID NOT NULL, -- Groups versions of the same notice
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT', -- DRAFT, PUBLISHED, ARCHIVED
    purposes UUID[], -- Array of linked purpose IDs
    widget_ids UUID[], -- Array of bound widget IDs
    regulation VARCHAR(100),
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_consent_notices_tenant ON consent_notices(tenant_id);
CREATE INDEX idx_consent_notices_status ON consent_notices(status);

CREATE TABLE IF NOT EXISTS consent_notice_translations (
    id UUID PRIMARY KEY,
    notice_id UUID NOT NULL REFERENCES consent_notices(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- ISO 639-1
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(notice_id, language)
);

CREATE INDEX idx_consent_notice_translations_notice ON consent_notice_translations(notice_id);
-- Grievance Redressal Module
-- DPDPA Requirement: Mechanism for data principals to lodge complaints.

CREATE TABLE IF NOT EXISTS grievances (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    subject_id UUID NOT NULL,
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
CREATE INDEX IF NOT EXISTS idx_grievances_tenant_subject ON grievances (tenant_id, subject_id);
CREATE INDEX IF NOT EXISTS idx_grievances_assigned_to ON grievances (assigned_to);
-- Up Migration

-- Create clients table if it doesn't exist (as per requirements for branding)
-- This is a minimal definition based on usage requirements
CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    logo_url TEXT,
    primary_color VARCHAR(50),
    support_email VARCHAR(255),
    portal_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Notification Templates
CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    channel VARCHAR(50) NOT NULL,
    subject TEXT,
    body_template TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, event_type, channel)
);

CREATE INDEX IF NOT EXISTS idx_notification_templates_tenant ON notification_templates(tenant_id);

-- Consent Notifications
CREATE TABLE IF NOT EXISTS consent_notifications (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    recipient_type VARCHAR(50) NOT NULL,
    recipient_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    channel VARCHAR(50) NOT NULL,
    template_id UUID REFERENCES notification_templates(id),
    payload JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    sent_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consent_notifications_tenant ON consent_notifications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_consent_notifications_recipient ON consent_notifications(recipient_id);
CREATE INDEX IF NOT EXISTS idx_consent_notifications_status ON consent_notifications(tenant_id, status);

-- Down Migration
-- DROP TABLE IF EXISTS consent_notifications;
-- DROP TABLE IF EXISTS notification_templates;
-- DROP TABLE IF EXISTS clients;
