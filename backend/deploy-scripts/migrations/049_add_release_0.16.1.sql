-- Migration: Add Release 0.16.1
-- Description: Adds release notes for version 0.16.1

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.16.1', '2025-12-22', '## What''s New in 0.16.1

### Bugs
- Fixed frontend build failure caused by TypeScript import syntax

### Infrastructure
- Updated production URL configuration
- Enabled authentication in production environment', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
