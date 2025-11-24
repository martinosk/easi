-- Migration: Add Release 0.8.0
-- Description: Adds release notes for version 0.8.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.8.0', '2025-11-24', '## What''s New in 0.8.0

### Major
- **Release Notes System**: Complete release notes infrastructure with version tracking and display

### Features
- Release notes overlay shown to users on first visit after version update
- Release notes browser to view full history of all releases
- "What''s New" button in toolbar for easy access
- Version endpoint at `/api/v1/version`
- Releases API at `/api/v1/releases` with HATEOAS links

### Infrastructure
- APP_VERSION automatically injected from git tags at build time
- Azure DevOps pipeline extracts version from latest git tag
- Release notes seeded via database migrations
- Claude commands for generating release notes and tagging releases

### Developer Experience
- `/generate-release-notes` command analyzes git history and specs
- `/tag-release` command creates migration and git tag in one step
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
