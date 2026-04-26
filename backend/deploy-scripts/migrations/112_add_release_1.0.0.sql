-- Migration: Add Release 1.0.0
-- Description: Adds release notes for version 1.0.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('1.0.0', '2026-04-26', '## What''s New in v1.0.0

### Major
- Dynamic View Mode replaces the one-shot "Generate View" flow. Editable views automatically enter draft mode where you can expand neighbours via on-node badges, drag-drop entities into the view, reposition nodes, and remove entities with cascade-on-delete — all visualised on the canvas immediately and saved or cancelled atomically.
- Open multiple views as tabs, each with its own independent draft. Tabs show a dirty indicator and a close button (with confirmation when there are unsaved changes), and clicking a view in the navigation tree or creating a new dynamic view opens it as a tab.

### Minor
- The "Generate View for X" context menu item has been replaced with "Create dynamic view from X" on both the canvas and the navigation tree.
- Auto Layout now applies immediately on click; the prior confirmation dialog was removed.
- Browsers now warn before navigating away when any open view has unsaved draft changes.

### Bugs
- Entities added to a view by dragging from the navigation tree no longer stay grayed-out in the tree, and clicking them now correctly centres the canvas on the entity.
- Fixed an intermittent issue where dragging an entity onto a freshly loaded view (most often the default view) silently failed to add it; drops are now reliably applied.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
