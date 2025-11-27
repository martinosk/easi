-- Migration: Add Release 0.9.0
-- Description: Adds release notes for version 0.9.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.9.0', '2025-11-27', '## What''s New in 0.9.0

### Major Features
- **Custom Element Colors**: Assign custom colors to individual applications and capabilities in the "Custom" color scheme
- **Color Scheme Persistence**: Color scheme selection now persists per view in the backend
- **Improved Color Schemes**: Updated "Classic" color scheme with better contrast and removed deprecated "Modern" scheme

### Frontend Architecture Improvements
- **Component Directory Restructure**: Reorganized frontend code into feature-based structure for better maintainability
- **ComponentCanvas Decomposition**: Reduced main canvas component from 868 lines to 193 lines by extracting focused hooks and components
- **Better Code Organization**: Created dedicated folders for shared components, layout components, and feature modules

### UI Enhancements
- **Toolbar Alignment Fixed**: Edge Type, Color Scheme, and Layout Direction selectors now align properly
- **Modern Capability Icon**: Capabilities display with a recognizable 2x2 grid icon
- **Consistent Terminology**: Updated all UI text to use "Application" instead of "Component"

### Bug Fixes
- Fixed edge type selector display issues
- Removed broken auto-layout feature (to be reimplemented)
- Fixed frontend build issues
- Cleaned up obsolete tests for removed color schemes

### Technical Improvements
- Added color picker component using react-colorful
- Created custom hooks for canvas functionality (nodes, edges, selection, viewport, drag-drop, connections)
- Extracted context menu components for better code reusability
- Improved test coverage with 22 new color rendering tests
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
