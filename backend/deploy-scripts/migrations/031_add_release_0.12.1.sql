-- Migration: Add Release 0.12.1
-- Description: Adds release notes for version 0.12.1

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.12.1', '2025-12-01', '## What''s New in 0.12.1

### Bug Fixes
- **Drag-and-Drop Reliability**: Fixed an issue where rapid position updates during drag-and-drop could fail or produce incorrect results
- **Import Memory Leak**: Fixed a memory leak that could occur during long-running import operations
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
