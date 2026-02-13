-- Consent Notice Translations
CREATE TABLE IF NOT EXISTS consent_notice_translations (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    notice_id UUID NOT NULL, -- references consent_notices(id) - assuming logic, not FK for flexibility if needed, but best practice is FK
    notice_version INTEGER NOT NULL,
    language_code VARCHAR(10) NOT NULL,
    translated_text TEXT,
    translation_source VARCHAR(50) NOT NULL, -- INDICTRANS2, MANUAL, UNSUPPORTED
    is_rtl BOOLEAN DEFAULT FALSE,
    translated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    reviewed_by UUID, -- references users(id) or similar
    reviewed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_consent_translations_notice_ver ON consent_notice_translations(notice_id, notice_version);
CREATE UNIQUE INDEX idx_consent_translations_unique_lang ON consent_notice_translations(notice_id, notice_version, language_code);
