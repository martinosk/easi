-- Migration: Add Release 0.20.2
-- Description: Adds release notes for version 0.20.2

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.20.2', '2026-01-14', '## What''s New in 0.20.2

### Major
- **Shareable deep-links** - Right-click any view or business domain and select "Share (copy URL)..." to copy a direct link that can be shared with colleagues

### Bugs
- Fixed dialog manager issues', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
