ALTER TABLE dsr_requests ADD COLUMN metadata JSONB;
COMMENT ON COLUMN dsr_requests.metadata IS 'Flexible storage for request-specific details (e.g., Nominee info, Portability formats)';
