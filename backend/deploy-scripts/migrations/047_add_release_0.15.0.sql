-- Migration: Add Release 0.15.0
-- Description: Adds release notes for version 0.15.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.15.0', '2025-12-21', '## What''s New in 0.15.0

### Major
- Added customizable dockable panel layouts with drag, resize, and tabbing - users can now personalize their workspace
- Added application details viewing and editing directly from the Business Domains view
- Improved mobile and tablet support with responsive layouts and better touch interactions
- Introduced React Query for faster data loading with automatic caching and background refresh
- Added code splitting for faster initial page loads

### Improvements
- Added error boundaries to prevent crashes from taking down the entire application
- Replaced hash-based URLs with standard path-based routing for better deep linking
- Migrated all dialogs to Mantine for improved accessibility (keyboard navigation, screen reader support)
- Added View menu to toggle panel visibility

### API
- Split API client into domain-specific modules for better maintainability', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
