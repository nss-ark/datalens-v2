CREATE TABLE IF NOT EXISTS platform_settings (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID -- Optional: ID of the admin who last updated it
);

-- Seed default settings
INSERT INTO platform_settings (key, value) VALUES 
('branding', '{"logo_url": "/logo.png", "primary_color": "#0F172A", "company_name": "DataLens"}'),
('maintenance', '{"enabled": false, "message": "System is under maintenance."}'),
('security', '{"mfa_required": false, "session_timeout_minutes": 60}')
ON CONFLICT (key) DO NOTHING;
