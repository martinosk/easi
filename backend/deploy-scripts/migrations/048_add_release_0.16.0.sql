-- Migration: Add Release 0.16.0
-- Description: Adds release notes for version 0.16.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.16.0', '2025-12-22', '## What''s New in 0.16.0

### Major
- Added user management for tenant administrators - view users, change roles, and disable/enable accounts
- Introduced role-based UI visibility - admins see full management menu, architects see architecture tools, stakeholders see read-only views

### Bugs
- Fixed layout issues in the UI
- Fixed security vulnerabilities in user management

### API
- Added user management API with support for listing, filtering, role changes, and account status management
- Added tenant information API for retrieving current organization details', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
