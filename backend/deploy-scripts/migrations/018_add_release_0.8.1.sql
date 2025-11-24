-- Migration: Add Release 0.8.1
-- Description: Adds release notes for version 0.8.1

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.8.1', '2025-11-24', '## What''s New in 0.8.1

### Bug Fixes
- Fixed "no component selected" error in edit component dialog
- Fixed frontend test warnings for cleaner test output
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
