-- 024_third_party_dpa.sql
-- Add DPA tracking fields to third_parties table.

ALTER TABLE third_parties
    ADD COLUMN IF NOT EXISTS dpa_status VARCHAR(20) NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS dpa_signed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS dpa_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS dpa_notes TEXT,
    ADD COLUMN IF NOT EXISTS contact_name VARCHAR(255) DEFAULT '',
    ADD COLUMN IF NOT EXISTS contact_email VARCHAR(255) DEFAULT '';
