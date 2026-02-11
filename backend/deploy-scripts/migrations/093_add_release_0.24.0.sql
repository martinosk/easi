-- Migration: Add Release 0.24.0
-- Description: Adds release notes for version 0.24.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.24.0', '2026-02-11', '## What''s New in v0.24.0

### Major
- Added "Created By" filter to the navigation tree view — filter artifacts by who created them to quickly find your own work or a colleague''s contributions

### Bugs
- Fixed Business Domains page showing context menu options (edit, delete) to users who are not authorized to perform those actions
- Fixed detail views for acquired entities, vendors, and internal teams incorrectly showing "Delete" button — changed to "Remove from View" to match the actual behavior
- Fixed long labels overflowing in context menus and buttons — added ellipses for truncation

### API
- New endpoint: `GET /api/v1/artifact-creators` — list the creator of each tree-relevant artifact (components, capabilities, vendors, internal teams, acquired entities)', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
