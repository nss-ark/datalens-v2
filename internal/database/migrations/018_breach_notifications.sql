CREATE TABLE IF NOT EXISTS breach_notifications (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    incident_id TEXT NOT NULL REFERENCES breach_incidents(id),
    data_principal_id TEXT NOT NULL, -- Logical link to data_principal_profiles
    title TEXT NOT NULL,
    severity TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    description TEXT NOT NULL,
    affected_data TEXT[] NOT NULL,
    what_we_are_doing TEXT NOT NULL,
    contact_email TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_breach_notifications_tenant_principal ON breach_notifications(tenant_id, data_principal_id);
