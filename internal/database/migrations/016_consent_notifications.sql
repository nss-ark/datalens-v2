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
