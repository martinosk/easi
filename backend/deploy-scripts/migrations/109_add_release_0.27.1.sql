-- Migration: Add Release 0.27.1
-- Description: Adds release notes for version 0.27.1

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('0.27.1', '2026-03-27', '## What''s New in v0.27.1

### Major
- Added cascade delete for capabilities — delete an entire capability subtree (children, realizations, dependencies) in one action, with impact analysis preview before confirming
- Added search/filter to the capabilities list for quick lookup

### Minor
- Increased rationale max length from 500 to 2,000 characters

### Bugs
- Fixed context menu incorrectly appearing for non-Level 1 capabilities in the business domain grid
- Added confirmation dialog before auto-layout to prevent accidental rearrangement
- Fixed performance issue where realizations were fetched one-by-one instead of in batch
- Fixed race condition when adding elements to a view in parallel

### API
- New endpoint: `GET /api/v1/capabilities/{id}/delete-impact` — preview what a cascade delete would affect
- Updated endpoint: `DELETE /api/v1/capabilities/{id}` — now supports `cascade: true` and `deleteRealisingApplications: true` parameters', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
