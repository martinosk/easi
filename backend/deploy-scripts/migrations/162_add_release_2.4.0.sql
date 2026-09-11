-- Migration: Add Release 2.4.0
-- Description: Adds release notes for version 2.4.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('2.4.0', '2026-09-11', '## What''s New in v2.4.0

### Major
- Added maturity journeys: a capability''s maturity ambition is now planned as a journey — the level to reach, by when, and the milestones to get there — and the journey shows the current maturity, the target, and the remaining gap as the capability matures.
- Retired the enterprise capability, together with directions, standard applications and composition. Strategic fit analysis now has its own page in the main navigation.
- Added application ownership: every application carries an ownership state (Unknown, Nominated, Owned or Managed); an architect nominates a user or internal team as owner and confirms it, and ownership statistics show the orphan count.
- Added application hosting classification (on-premises, cloud, SaaS, third-party hosted) so the landscape can be filtered by where applications run.
- Custom fields are now defined in the meta model as the tenant''s attribute vocabulary per subject type; one-pager configuration picks from them, and existing fields, options, bounds and recorded facts carry over unchanged.

### Minor
- Applications, capabilities, origin entities (acquired entities, vendors, internal teams) and canvas edges are now edited in place: every field is edited where it is shown, edit controls appear only when permitted, and the whole-record Edit dialogs are gone. The Architecture Canvas, the Business Domains drawer and the one-pager subject drawer show the same sections in the same order.
- The computed TIME suggestion, with its confidence, is shown beside the grade choices while recording a TIME assessment.
- Application list filters for ownership and hosting sit behind a filter icon, with a count of active filters.

### Bugs
- Planning a Move journey no longer fails with a generic 404 when the target parent lies outside the target domain; the parent picker only offers capabilities in the chosen domain.
- Read-only viewers no longer see Edit and Delete controls on relation and realization edges that would be rejected on click.
- The Business Domains application drawer now shows the experts section and reflects ownership and hosting edits immediately instead of rendering stale data.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
