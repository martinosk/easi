-- Migration: Add Release 0.25.0
-- Description: Adds release notes for version 0.25.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.25.0', '2026-02-14', '## What''s New in v0.25.0

### Major Features
- **Value Streams — Phase 1**: Model how your organization delivers value end-to-end. Create value streams with sequential stages, visually represent the flow, and map capabilities to each stage. Interactive flow diagram with drag-and-drop stage reordering and capability mapping.
- **Tree-View Filters**: Filter the navigation tree by creator and business domain assignment. Select "Created by" to show only artifacts created by specific users, or "Assigned to domain" to see artifacts linked to particular domains. Both filters can be combined for precise navigation.

### Bugs Fixed
- Fixed applications not displaying when L1-L4 capability level filters were active
- Fixed inherited realizations incorrectly persisting when reparenting capabilities to different parents

### API
- New endpoint: `GET /api/v1/artifact-creators` — Returns creator information for tree-relevant artifacts, enabling the "Created by" filter
- New endpoints: `GET /api/v1/value-streams`, `POST /api/v1/value-streams` — Create and list value streams
- New endpoints: Stage and capability mapping endpoints for modeling value stream flows

**Note:** Value Streams Slices 3–4 (sidebar integration and cross-capability impact analysis) are in development and will be included in a future release.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
