-- Migration: 009_align_grievances
-- Description: Align grievances table schema with application code and drop FK constraint on subject_id

-- 1. Add missing columns that the application expects
ALTER TABLE grievances
  ADD COLUMN IF NOT EXISTS subject TEXT,
  ADD COLUMN IF NOT EXISTS priority INT DEFAULT 0,
  ADD COLUMN IF NOT EXISTS feedback_rating INT,
  ADD COLUMN IF NOT EXISTS feedback_comment TEXT,
  ADD COLUMN IF NOT EXISTS escalated_to TEXT,
  ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ DEFAULT NOW();

-- 2. Rename columns to match application code (and intended schema)
DO $$
BEGIN
  IF EXISTS(SELECT * FROM information_schema.columns WHERE table_name='grievances' AND column_name='type') THEN
    ALTER TABLE grievances RENAME COLUMN type TO category;
  END IF;
  IF EXISTS(SELECT * FROM information_schema.columns WHERE table_name='grievances' AND column_name='deadline') THEN
    ALTER TABLE grievances RENAME COLUMN deadline TO due_date;
  END IF;
  IF EXISTS(SELECT * FROM information_schema.columns WHERE table_name='grievances' AND column_name='response') THEN
    ALTER TABLE grievances RENAME COLUMN response TO resolution;
  END IF;
  IF EXISTS(SELECT * FROM information_schema.columns WHERE table_name='grievances' AND column_name='received_at') THEN
    ALTER TABLE grievances RENAME COLUMN received_at TO submitted_at;
  END IF;
END $$;

-- 3. Drop Foreign Key constraint on subject_id
-- This allows Data Principals (Profiles) who are not yet linked to a Data Subject to submit grievances.
-- The application uses the ProfileID as the data_subject_id in these cases.
ALTER TABLE grievances DROP CONSTRAINT IF EXISTS grievances_subject_id_fkey;
