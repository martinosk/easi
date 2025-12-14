-- Migration: Add Release 0.14.2
-- Description: Adds release notes for version 0.14.2

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.14.2', '2025-12-14', '## What''s New in 0.14.2

### User Invitations
- **Invitation management UI**: Administrators can now invite users directly from the web interface
- **Client-side filtering**: Invitation list supports filtering for better usability
- **Backend invitation system**: Complete API support for creating, listing, and managing user invitations

### Bug Fixes
- **Import session**: Fixed an issue where import sessions were not working correctly
- **Capability reparenting**: Fixed inherited realizations not displaying after reparenting capabilities
- **Development mode**: Fixed bypass mode configuration issues

### Documentation
- Expanded OpenAPI documentation for invitation endpoints
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
