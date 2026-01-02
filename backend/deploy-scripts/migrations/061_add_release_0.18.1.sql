-- Migration: Add Release 0.18.1
-- Description: Adds release notes for version 0.18.1

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.18.1', '2026-01-02', '## What''s New in v0.18.1

### Minor
- Added contextual help tooltips throughout the application to explain maturity sections, investment priorities, and domain-specific terminology
- Added maturity distribution color legend explaining Genesis, Custom Build, Product, and Commodity stages

### Bugs
- Fixed bug in meta model configuration repository

### Removed
- Removed obsolete strategy pillar assignment from capability editing dialog', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
