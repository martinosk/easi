-- Migration: Add Release 2.1.0
-- Description: Adds release notes for version 2.1.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('2.1.0', '2026-07-16', '## What''s New in v2.1.0

### Major
- Added capability journeys: record a capability''s change story — migration, consolidation, carve-out, or move — with status, progress, a target period, and milestones.
- The Domain Board gains Now / Journey / Target lenses: Journey overlays every capability''s change story with its progress, and Target projects the landscape as it will look when journeys land.
- Added One-Pager Quality: a global, sortable list tracking one-pager completeness across all subject types, with per-row invitations granting edit access to whoever should complete a one-pager.
- One-pager facts are now edited directly on the one-pager page, and relations (realizing applications, origins, included capabilities, and more) appear as built-in one-pager fields.

### Minor
- Built-in one-pager fields can be marked mandatory and count toward completeness alongside custom fields.
- Number fields on one-pagers support optional min/max bounds; values outside the range are flagged.
- An "Open capability/application/…" button on the one-pager opens the subject''s detail panel in a drawer, so missing details like description and experts can be fixed in place.
- Sped up Domain Board loading by deduplicating capability queries.

### Bugs
- Fixed the Domain Board to show every L1 capability as a collapsible card.
- Newly created applications are no longer automatically added to the active view.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
