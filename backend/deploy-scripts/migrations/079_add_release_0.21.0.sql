-- Migration: Add Release 0.21.0
-- Description: Adds release notes for version 0.21.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.21.0', '2026-01-21', '## What''s New in v0.21.0

### Major
- Added Portfolio Metadata Foundation with Origin Entities (Acquired Entities, Vendors, Internal Teams) for tracking application sources in the Architecture Modeling context
- Added Origin Relationships allowing you to connect applications to their acquisition source, vendor, or building team via canvas drawing
- Added Domain Architect assignment to Business Domains

### Bugs
- Fixed readmodel projection not updating correctly when deleting applications
- Fixed null returns in new API endpoints for origin entities

### Removed
- Removed "Unlinked Capabilities" feature from Enterprise Architecture view

### API
- New endpoints: `POST/GET/PUT/DELETE /api/v1/acquired-entities`
- New endpoints: `POST/GET/PUT/DELETE /api/v1/vendors`
- New endpoints: `POST/GET/PUT/DELETE /api/v1/internal-teams`
- New endpoints: `POST/GET/DELETE /api/v1/acquisition-relationships`
- New endpoints: `POST/GET/DELETE /api/v1/vendor-relationships`
- New endpoints: `POST/GET/DELETE /api/v1/team-relationships`', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
