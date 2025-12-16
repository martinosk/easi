-- Migration: Add Release 0.14.3
-- Description: Adds release notes for version 0.14.3

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.14.3', '2025-12-16', '## What''s New in 0.14.3

### Drag and Drop Improvements
- **Migrated to native HTML5 drag and drop**: Replaced dnd-kit library with browser-native drag and drop functionality
- **Improved reliability**: Fixed intermittent drag initialization issues that required hiding/showing the capability explorer sidebar
- **Reduced bundle size**: Removed ~50KB of dnd-kit dependencies for faster load times
- **Simplified architecture**: Cleaner implementation using standard browser APIs
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
