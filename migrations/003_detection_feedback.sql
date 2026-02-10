-- =============================================================================
-- DataLens 2.0 — Detection Feedback Table
-- Migration: 003_detection_feedback
-- =============================================================================
-- Stores human feedback (verify/correct/reject) on PII classification results.
-- Feeds into the learning loop to improve detection accuracy over time.
-- =============================================================================

CREATE TABLE detection_feedback (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    classification_id   UUID NOT NULL REFERENCES pii_classifications(id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    feedback_type       VARCHAR(50) NOT NULL,  -- VERIFIED, CORRECTED, REJECTED

    -- Original detection snapshot (for learning comparison)
    original_category   VARCHAR(50) NOT NULL,
    original_type       VARCHAR(50) NOT NULL,
    original_confidence DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    original_method     VARCHAR(50) NOT NULL,

    -- Corrected values (populated only for CORRECTED feedback)
    corrected_category  VARCHAR(50),
    corrected_type      VARCHAR(50),

    -- Metadata
    corrected_by        UUID NOT NULL REFERENCES users(id),
    corrected_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes               TEXT NOT NULL DEFAULT '',

    -- Context for learning — column/table info for pattern extraction
    column_name         VARCHAR(255) NOT NULL DEFAULT '',
    table_name          VARCHAR(255) NOT NULL DEFAULT '',
    data_type           VARCHAR(100) NOT NULL DEFAULT '',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance indexes
CREATE INDEX idx_detection_feedback_tenant ON detection_feedback(tenant_id);
CREATE INDEX idx_detection_feedback_classification ON detection_feedback(classification_id);
CREATE INDEX idx_detection_feedback_type ON detection_feedback(feedback_type);
CREATE INDEX idx_detection_feedback_column_pattern ON detection_feedback(tenant_id, column_name);
