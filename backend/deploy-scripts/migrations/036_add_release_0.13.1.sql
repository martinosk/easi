-- Migration: Add Release 0.13.1
-- Description: Adds release notes for version 0.13.1

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.13.1', '2025-12-02', '## What''s New in 0.13.1

### Performance
- **Optimized capability realisation loading**: Realisations are now fetched per business domain instead of per capability, significantly reducing API calls when viewing applications in the grid

### Fixes
- Fixed race condition when toggling application visibility in the capability grid

### Maintenance
- Removed unused capability_code column from domain assignments
- Internal cleanup of unused query bus code
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
