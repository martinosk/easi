-- Migration: Add Release 0.22.0
-- Description: Adds release notes for version 0.22.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.22.0', '2026-02-01', '## What''s New in v0.22.0

### Major
- Added multi-select on the canvas via Shift+drag rectangle selection and Ctrl/Shift+click, with a bulk context menu for "Remove from View" and "Delete from Model" actions that respects HATEOAS permissions across all selected items
- Added multi-select in the navigation treeview via Ctrl+click (toggle) and Shift+click (range), with bulk "Delete from Model" context menu and drag-and-drop of multiple selected items onto the canvas', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
