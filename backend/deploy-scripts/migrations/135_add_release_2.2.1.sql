-- Migration: Add Release 2.2.1
-- Description: Adds release notes for version 2.2.1

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('2.2.1', '2026-08-26', '## What''s New in v2.2.1

### Minor
- The top bar is now responsive: on narrower screens it collapses to icon-only controls with an overflow menu, Invitations moved under Users as a tab, Settings moved into the user menu, and every header control now has a tooltip.
- Enterprise Capability owners are now shown by name instead of by ID.

### Bugs
- Fixed the overflow menu appearing hidden behind the explorer tree on the canvas page.
- Fixed the user initials avatar being illegible on the dark top bar.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
