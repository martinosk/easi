-- Migration: Add Release 0.9.0
-- Description: Adds release notes for version 0.9.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.9.0', '2025-11-25', '## What''s New in 0.9.0

### Major
- Added color scheme selector with 4 options: Maturity (default), ArchiMate, ArchiMate (classic), and Custom
- Introduced custom color assignment for individual capabilities and components when using the Custom scheme
- Added view-specific color persistence - your color choices are saved and restored when you reload the view
- Improved auto-layout to include all canvas elements (capabilities, applications, and their relationships)
- Added capability icon to capability nodes for better visual recognition

### Enhancements
- Color scheme selection now persists across sessions
- Custom colors are view-specific - the same element can have different colors in different views
- Custom colors are preserved when switching between schemes, making it easy to experiment
- Added color indicators in the navigation tree showing which elements have custom colors
- Elements without custom colors automatically use a neutral gray when Custom scheme is active

### API
- New endpoint: `PATCH /api/v1/views/{id}/color-scheme` - Update color scheme for a view
- New endpoint: `PATCH /api/v1/views/{id}/components/{componentId}/color` - Set custom color for a component
- New endpoint: `PATCH /api/v1/views/{id}/capabilities/{capabilityId}/color` - Set custom color for a capability
- New endpoint: `DELETE /api/v1/views/{id}/components/{componentId}/color` - Clear custom color from a component
- New endpoint: `DELETE /api/v1/views/{id}/capabilities/{capabilityId}/color` - Clear custom color from a capability', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
