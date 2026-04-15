-- Migration: Add Release 0.27.2
-- Description: Adds release notes for version 0.27.2

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('0.27.2', '2026-04-15', '## What''s New in v0.27.2

### Major
- Added OpenAI Responses API support for the AI architecture assistant, with full Azure OpenAI endpoint compatibility
- Added AI tool for retrieving all capabilities for a business domain, enabling richer architecture analysis

### Minor
- Origin entities present in the active view are now visually highlighted for better context
- Migrated frontend linting from ESLint to Biome for faster and more consistent code quality checks
- Updated API documentation with improved authentication details

### Bugs
- Fixed Azure OpenAI tool connectivity issues
- Fixed origin entity highlighting not appearing for entities in the active view
- Security vulnerability fixes across frontend dependencies', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
