-- Migration: Add Release 3.0.0
-- Description: Adds release notes for version 3.0.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('3.0.0', '2026-10-01', '## What''s New in v3.0.0

### Major
- EASI now opens on a personal Home at `/`, composed for the signed-in user. It shows which parts of the landscape are yours, with tiles for the capabilities, applications and domains in your scope and the TIME distribution of their realisations, and a My Work list of the capabilities you are EA owner of and the applications you own, are nominated to own or hold an edit grant on.
- Your scope comes from what you answer for: domains you steward or architect, capabilities you are EA owner of, applications you own and active edit grants. Admins and architects without any of these see the whole tenant; stakeholders without any are told how to get scope.
- The Architecture Canvas has moved to `/canvas`.
- Added domain stewardship: one user can be made accountable for each concern in a business domain (ownership, assessment, documentation, planning, structure). Stewards are assigned, changed and released from the new Stewards dialog on the domain board, any active user can be a steward regardless of role, and unassigned concerns show the domain architect as fallback.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
