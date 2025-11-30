-- Migration: Add Release 0.12.0
-- Description: Adds release notes for version 0.12.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.12.0', '2025-11-30', '## What''s New in 0.12.0

### Import Open Exchange Files
- **ArchiMate Import**: Import architecture models from ArchiMate Open Exchange format (.xml) files
- **Multi-Step Wizard**: File upload, preview, and confirmation workflow
- **Real-Time Progress**: Track import progress as elements and relationships are processed
- **Preview Before Import**: Review imported elements and relationships before confirming
- **Event-Sourced Sessions**: Full audit trail of import operations

### Reference Documentation
- **Documentation Links**: Quick access to relevant reference documentation from the toolbar

### Terminology Cleanup
- **Standardized Naming**: Cleaned up architecture terminology throughout the application for better consistency
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
