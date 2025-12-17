-- Migration: Add Release 0.14.4
-- Description: Adds release notes for version 0.14.4

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.14.4', '2025-12-17', '## What''s New in 0.14.4

### Bug Fixes
- **Fixed "Remove from Business Domain"**: Resolved an issue where removing capabilities from a business domain would fail with "Dissociate link not available" error
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
