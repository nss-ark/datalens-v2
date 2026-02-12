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
