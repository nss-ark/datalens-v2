-- Migration 020: Add missing audit_logs columns
-- The Go entity (audit.AuditLog) uses UserID, OldValues, NewValues,
-- and the repository (postgres_audit.go) maps them to user_id, old_values, new_values.
-- The original migration 009 used actor_id and changes columns.
-- Also adds client_id alias for tenant_id (repo uses client_id for tenant scoping).

-- Add client_id column (alias for tenant_id used by the repository layer)
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS client_id UUID;
UPDATE audit_logs SET client_id = tenant_id WHERE client_id IS NULL;

-- Add user_id column (maps to Go entity's UserID field)
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS user_id UUID;

-- Add old_values and new_values JSONB columns (maps to Go entity's OldValues/NewValues)
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS old_values JSONB;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS new_values JSONB;

-- Backfill from existing columns
UPDATE audit_logs SET user_id = actor_id WHERE user_id IS NULL;
UPDATE audit_logs SET old_values = changes WHERE old_values IS NULL AND changes IS NOT NULL;
