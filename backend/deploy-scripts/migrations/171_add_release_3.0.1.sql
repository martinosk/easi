-- Migration: Add Release 3.0.1
-- Description: Adds release notes for version 3.0.1

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('3.0.1', '2026-11-11', '## What''s New in v3.0.1

### Minor
- Realising applications can now be added, created and removed directly from the capability details panel. Inherited realisations no longer offer edit or delete.

### Bugs
- Long one-pagers can now be scrolled.
- Disabled users are now signed out, and their sessions and agent tokens are rejected on every request.
- Importing now requires write permission on the imported data, including business domains, and the Import button only shows for users who have it.
- Edit now appears in the business domain menu for users allowed to edit the domain.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
