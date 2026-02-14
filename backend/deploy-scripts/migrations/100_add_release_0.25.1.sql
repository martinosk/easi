-- Migration: Add Release 0.25.1
-- Description: Adds release notes for version 0.25.1

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.25.1', '2026-02-14', '## What''s New in v0.25.1

### Minor
- Added value stream support to ArchiMate import — value streams are now parsed, previewed, and imported alongside existing artifact types', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
