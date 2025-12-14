-- Migration: Add Release 0.14.1
-- Description: Adds release notes for version 0.14.1

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.14.1', '2025-12-14', '## What''s New in 0.14.1

### Session Management
- **Session introspection**: Users can now check their current session status via the `/api/v1/session` endpoint
- **Logout functionality**: Users can log out and invalidate their session via `POST /api/v1/session/logout`

### Bug Fixes
- **Auth route handling**: Fixed an issue where auth routes were incorrectly set up when authentication was bypassed in development mode

### Code Quality
- Improved code health scores across integration tests
- Refactored test helpers to reduce code duplication
- Linting and formatting improvements
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
