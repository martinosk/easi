-- Migration: Add Release 0.19.0
-- Description: Adds release notes for version 0.19.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.19.0', '2026-01-06', '## What''s New in v0.19.0

### Major
- Added **Private Views** - Create personal architecture views that only you can edit
  - New views default to private - only the creator can modify them
  - All users can still view private views in the navigation tree
  - Private views show a lock icon and display the owner''s name
  - Toggle visibility via right-click context menu (Make Public / Make Private)
  - Admins can make any private view public for cleanup when users leave

### API
- New endpoint: `PATCH /api/v1/views/{id}/visibility` - Toggle view between private and public
- Updated `GET /api/v1/views` - Returns `isPrivate`, `ownerUserId`, `ownerEmail` fields
- Updated `GET /api/v1/views/{id}` - Returns ownership and visibility info
- HATEOAS links now include `update`, `delete`, `changeVisibility` based on permissions
- All view write endpoints return 403 Forbidden for unauthorized access to private views', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
