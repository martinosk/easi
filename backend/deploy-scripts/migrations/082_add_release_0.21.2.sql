-- Migration: Add Release 0.21.2
-- Description: Adds release notes for version 0.21.2

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.21.2', '2026-01-25', '## What''s New in v0.21.2

### Bugs
- Fixed origin entities reappearing at center of canvas after using "Remove from View" - they now stay removed like components and capabilities

### Major
- Completed Portfolio Metadata Foundation: origin entities (Acquired Entities, Vendors, Internal Teams) now have full canvas context menu support with "Remove from View" and "Delete from Model" options', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
