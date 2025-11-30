-- Migration: Add Release 0.11.1
-- Description: Adds release notes for version 0.11.1

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.11.1', '2025-11-30', '## What''s New in 0.11.1

### Persistent Depth Selector
- **Remember Depth Setting**: Your selected depth level (L1, L1-L2, L1-L3, L1-L4) in Business Domains is now remembered across page navigations and browser sessions

### Architecture Improvements
- **ViewLayouts Bounded Context**: Separated presentation concerns (element positions, layout preferences) from domain concerns (view membership) for better maintainability
- **Cleaner Data Model**: Business Domain layouts now managed independently from Architecture Canvas views
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
