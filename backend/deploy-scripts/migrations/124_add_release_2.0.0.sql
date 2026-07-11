-- Migration: Add Release 2.0.0
-- Description: Adds release notes for version 2.0.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('2.0.0', '2026-07-11', '## What''s New in v2.0.0

### Major
- Redesigned the application with a new design system, tenant-specific visual themes, a Domain Board overview, and a canvas-based workspace.
- An enterprise capability''s composition is now automatically derived from its active Direction, replacing manual capability-to-capability linking.
- Added One-Pagers: configurable one-pager fields, a settings UI, fact capture, and a dedicated one-pager view page.

### Minor
- Added Standard Application tracking on enterprise capabilities, including setting/changing the standard application and viewing its history.
- Unified capability list views on a shared, searchable capability tree for consistent behavior across the app.
- Improved page load performance via gzip compression of static assets and API responses.
- Sped up the user listing page by eliminating redundant admin-count queries.
- Sped up domain filtering by deferring per-domain queries until a filter is selected.

### Bugs
- Fixed the capability context menu not appearing at all levels in the business domain view.
- Fixed multi-select "Remove from view" not working when the view is unsaved.
- Fixed capability, direction, and standard application names not updating when the referenced entity was renamed.
- Fixed source capability names rendering as raw UUIDs when unlinked.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
