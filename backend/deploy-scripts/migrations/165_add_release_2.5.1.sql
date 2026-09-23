-- Migration: Add Release 2.5.1
-- Description: Adds release notes for version 2.5.1

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('2.5.1', '2026-09-23', '## What''s New in v2.5.1

### Minor
- Capability, application and origin entity (acquired entity, vendor, internal team) details are now arranged in named, collapsible groups instead of one long list, with the name heading the panel above the groups.
- Groups can be moved up and down from their headers so the sections you use most sit at the top. The order and collapsed groups are remembered in the browser and shared by the Architecture Canvas details pane, the Business Domains drawer and the one-pager subject drawer.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
