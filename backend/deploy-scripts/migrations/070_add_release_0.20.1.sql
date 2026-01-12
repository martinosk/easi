-- Migration: Add Release 0.20.1
-- Description: Adds release notes for version 0.20.1

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.20.1', '2026-01-12', '## What''s New in 0.20.1

### Bugs
- Fixed release notes API returning 404/500 errors due to invalid version format in database', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
