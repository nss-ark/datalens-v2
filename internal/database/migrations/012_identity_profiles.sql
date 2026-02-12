-- Create identity_profiles table
-- This stores the assurance level and verification documents for a subject

CREATE TABLE IF NOT EXISTS identity_profiles (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id),
    subject_id TEXT NOT NULL REFERENCES compliance_data_subjects(id),
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
