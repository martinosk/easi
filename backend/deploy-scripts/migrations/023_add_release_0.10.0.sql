-- Migration: Add Release 0.10.0
-- Description: Adds release notes for version 0.10.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.10.0', '2025-11-30', '## What''s New in 0.10.0

### New Feature: Business Domains
- **Strategic Capability Groupings**: Organize L1 capabilities into business domains like "Finance", "Customer Experience", or "Operations"
- **Full Backend Support**: Aggregate with event sourcing, read models, and REST API
- **Management UI**: Create, edit, and delete business domains
- **Grid Visualization**: View domain capabilities in a visual grid layout
- **Drag-and-Drop Assignment**: Assign L1 capabilities to domains by dragging from the capability explorer

### UI Improvements
- **Centered Modal Dialogues**: Modal dialogs now center properly on screen
- **Better Domain Card Visuals**: Improved styling for business domain cards

### Code Quality
- Refactored view handlers and view models
- Improved integration test organization
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
