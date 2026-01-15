-- Migration: Add Release 0.20.3
-- Description: Adds release notes for version 0.20.3

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.20.3', '2025-01-14', '## What''s New in v0.20.3

### Major
- Added subject matter experts to application components - you can now track who has domain knowledge about specific components with name, role, and contact information
- Added role autocomplete for capability experts - previously used roles are now suggested when adding experts to capabilities
- Added expert removal - you can now remove experts from both capabilities and application components
- Added EA Owner pre-fill during import - when importing ArchiMate models, you can now select an EA Owner to apply to all imported capabilities

### API
- New endpoint: `POST /api/v1/components/{id}/experts` - Add expert to application component
- New endpoint: `DELETE /api/v1/components/{id}/experts/{name}` - Remove expert from application component
- New endpoint: `GET /api/v1/components/expert-roles` - Get distinct roles for autocomplete
- New endpoint: `GET /api/v1/capabilities/expert-roles` - Get distinct roles for autocomplete
- New endpoint: `DELETE /api/v1/capabilities/{id}/experts/{name}` - Remove expert from capability
- Updated endpoint: `POST /api/v1/imports` - Now accepts optional `capabilityEAOwner` field for pre-filling metadata', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
