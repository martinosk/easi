-- Migration: Add Release 2.2.0
-- Description: Adds release notes for version 2.2.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('2.2.0', '2026-08-24', '## What''s New in v2.2.0

### Major
- Added a Timeline view to the Business Domains page: every active journey as a row on a quarter axis, with milestones and target periods placed at their quarters, behind-schedule work highlighted, and click-through to the capability.
- Journeys now compose across the capability hierarchy: a capability''s journey section shows the journeys running beneath its sub-capabilities, and a sub-capability shows the ancestor journey it is part of.
- The AI assistant is now available to stakeholders with read-only access — grounded answers without write affordances.

### Minor
- The capability map view is back on the Business Domains page.
- Consolidation journeys derive their source applications automatically from the capability''s current realisations, so a forgotten realiser can no longer silently narrow the recorded story.

### Bugs
- Invited users whose identity provider reports their email address with uppercase letters can now log in.
- Duplicate checks and automatic invitations for edit grants now match email addresses case-insensitively.
- A move journey''s target can now be an application that already realises the capability — the most common move scenario.
- Canvas elements no longer appear at outdated positions after having been moved.
- The journey capture form no longer shows a stale validation error after the input is corrected.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
