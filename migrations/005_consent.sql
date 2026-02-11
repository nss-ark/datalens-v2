-- Add config to data_sources (missed in initial schema)
-- Add config to data_sources (missed in initial schema)
ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS config JSONB NOT NULL DEFAULT '{}';


-- Consent Widgets
CREATE TABLE IF NOT EXISTS consent_widgets (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    tenant_id UUID NOT NULL, -- references tenants(id)
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- BANNER, PREFERENCE_CENTER, PORTAL, INLINE_FORM
    domain VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL, -- DRAFT, ACTIVE, PAUSED
    config JSONB NOT NULL DEFAULT '{}',
    embed_code TEXT,
    api_key VARCHAR(255) UNIQUE, -- Public key
    allowed_origins TEXT[],
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_consent_widgets_tenant_id ON consent_widgets(tenant_id);
CREATE INDEX idx_consent_widgets_api_key ON consent_widgets(api_key);

-- Consent Sessions (Lightweight, high volume)
CREATE TABLE IF NOT EXISTS consent_sessions (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    tenant_id UUID NOT NULL,
    widget_id UUID NOT NULL, -- references consent_widgets(id)
    subject_id UUID, -- references data_subjects(id) in compliance, but nullable here
    decisions JSONB NOT NULL, -- Array of PurposeID + Granted
    ip_address VARCHAR(45),
    user_agent TEXT,
    page_url TEXT,
    widget_version INTEGER,
    notice_version VARCHAR(50),
    signature VARCHAR(255) -- HMAC integrity
);

CREATE INDEX idx_consent_sessions_tenant_id ON consent_sessions(tenant_id);
CREATE INDEX idx_consent_sessions_subject_id ON consent_sessions(subject_id);

-- Data Principal Profiles (Portal Identity)
CREATE TABLE IF NOT EXISTS data_principal_profiles (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    tenant_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    verification_status VARCHAR(50) NOT NULL, -- PENDING, VERIFIED, EXPIRED
    verified_at TIMESTAMP WITH TIME ZONE,
    verification_method VARCHAR(50),
    subject_id UUID, -- Link to compliance DSR subject
    last_access_at TIMESTAMP WITH TIME ZONE,
    preferred_lang VARCHAR(10) DEFAULT 'en'
    -- CONSTRAINT uq_principal_email_tenant UNIQUE (tenant_id, email) -- Optional, depends on multi-tenancy model
);

CREATE INDEX idx_data_principal_profiles_tenant_email ON data_principal_profiles(tenant_id, email);

-- Consent History (Audit Log)
CREATE TABLE IF NOT EXISTS consent_history (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    tenant_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    widget_id UUID,
    purpose_id UUID NOT NULL,
    purpose_name VARCHAR(255),
    previous_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL, -- GRANTED, WITHDRAWN, EXPIRED
    source VARCHAR(50), -- BANNER, PORTAL, API
    ip_address VARCHAR(45),
    user_agent TEXT,
    notice_version VARCHAR(50),
    signature VARCHAR(255)
);

CREATE INDEX idx_consent_history_subject_purpose ON consent_history(tenant_id, subject_id, purpose_id);

-- DPR Requests (Portal View)
CREATE TABLE IF NOT EXISTS dpr_requests (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    tenant_id UUID NOT NULL,
    profile_id UUID NOT NULL, -- references data_principal_profiles(id)
    dsr_id UUID, -- Link to internal compliance DSR
    type VARCHAR(50) NOT NULL, -- ACCESS, ERASURE, etc.
    description TEXT,
    status VARCHAR(50) NOT NULL, -- SUBMITTED, PENDING_VERIFICATION, ...
    submitted_at TIMESTAMP WITH TIME ZONE,
    deadline TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    verification_ref VARCHAR(255),
    is_minor BOOLEAN DEFAULT FALSE,
    guardian_name VARCHAR(255),
    guardian_email VARCHAR(255),
    guardian_relation VARCHAR(50),
    guardian_verified BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP WITH TIME ZONE,
    response_summary TEXT,
    download_url TEXT,
    appeal_of UUID,
    appeal_reason TEXT,
    is_escalated BOOLEAN DEFAULT FALSE,
    escalated_to VARCHAR(255)
);

CREATE INDEX idx_dpr_requests_profile_id ON dpr_requests(profile_id);
