-- Migration: 009_add_config_to_datasources
-- Description: Add missing config column to data_sources table

ALTER TABLE data_sources 
ADD COLUMN IF NOT EXISTS config TEXT DEFAULT '{}';

COMMENT ON COLUMN data_sources.config IS 'JSON configuration specific to the data source type (e.g., service account JSON)';
