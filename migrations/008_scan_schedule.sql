-- Migration: 006_scan_schedule
-- Description: Add scan scheduling support to data sources

ALTER TABLE data_sources 
ADD COLUMN IF NOT EXISTS scan_schedule TEXT DEFAULT NULL;

COMMENT ON COLUMN data_sources.scan_schedule IS 'Cron expression for automated scans (e.g., "0 2 * * *" for daily at 2 AM)';
