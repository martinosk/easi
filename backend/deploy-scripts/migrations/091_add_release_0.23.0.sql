-- Migration: Add Release 0.23.0
-- Description: Adds release notes for version 0.23.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.23.0', '2026-02-08', '## What''s New in v0.23.0

### Major
- Added "Invite to Edit" — architects and admins can now grant stakeholders temporary edit access to specific capabilities, components, views, and business domains, with automatic 30-day expiration and revocation support
- Added "My Edit Access" page where stakeholders can discover and navigate to artifacts they''ve been granted edit access to, accessible from the user menu with a live count badge
- Added automatic platform invitation when granting edit access to a non-user email, so new collaborators can join and immediately start editing
- Made the canvas minimap interactive — click or drag on the minimap to navigate the canvas

### Bugs
- Fixed strategy pillars not showing in application details
- Fixed HATEOAS links missing for strategic importance endpoints
- Fixed domain assignment not propagating to capability metadata

### API
- New endpoint: `POST /api/v1/edit-grants` — create an edit grant for a user on a specific artifact
- New endpoint: `GET /api/v1/edit-grants` — list edit grants (grantor or grantee view)
- New endpoint: `GET /api/v1/edit-grants/{id}` — get a single edit grant
- New endpoint: `DELETE /api/v1/edit-grants/{id}` — revoke an edit grant
- New endpoint: `GET /api/v1/edit-grants/artifact/{artifactType}/{artifactId}` — list grants for an artifact', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
