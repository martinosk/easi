-- Migration: Add Release 1.0.1
-- Description: Adds release notes for version 1.0.1

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('1.0.1', '2026-04-26', '## What''s New in v1.0.1

### Bugs
- Fixed "Remove from View" cascading to related entities. The action now removes only the entity you clicked (or, when multi-selecting, exactly the entities you selected) — entities that lose their only connection as a result are left in the view.
- Fixed cascade of 409 errors when saving a view after a partial failure. Adds and removes are now treated as idempotent: re-saving entities that are already in (or already out of) the view succeeds instead of erroring.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
