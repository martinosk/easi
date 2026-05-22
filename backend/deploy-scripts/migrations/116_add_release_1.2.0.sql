-- Migration: Add Release 1.2.0
-- Description: Adds release notes for version 1.2.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('1.2.0', '2026-05-22', '## What''s New in v1.2.0

### Major
- Capture and manage strategic direction (keep, retire, replace, move) on enterprise capabilities, with source-capability linkage and graceful handling of stale references when source capabilities change.

### Minor
- Comprehensive UI redesign: the frontend has been unified onto a single Mantine v8 design vocabulary across business domains, components, capabilities, origin entities, relations, invitations, users, importing, edit-grants, auth, and enterprise-architecture surfaces for visual consistency and improved accessibility.
- Refined capability colour scheme.

### Bugs
- The Created-by filter in the navigation tree now also filters the Views section, and newly created views appear in the filter immediately without requiring a page reload.
- Cancelling an import no longer leaves the dialog stuck open if the cancel request fails.
- Business domain name and description no longer falsely reject values whose length only exceeds the limit due to surrounding whitespace.
- The inline view rename input now automatically receives focus when entering edit mode.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
