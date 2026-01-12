-- Migration: Add Release v0.20.0
-- Description: Adds release notes for version v0.20.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('v0.20.0', '2026-01-12', '## What''s New in v0.20.0

### Major
- **Strategic analysis now considers capability hierarchy** - When evaluating strategic fit gaps, the analysis uses the most specific capability rating available, inheriting from parent capabilities when needed. Applications realizing multiple capabilities in the same hierarchy now show separate gap analyses for each.
- **Improved permission-based UI controls** - The frontend now consistently uses HATEOAS links from API responses to determine which actions are available. Users will only see buttons and menu options for actions they can actually perform.
- **Private view ownership display** - Private views now show the owner''s name, making it easier to identify who created each view.

### Bugs
- Fixed hardcoded maturity values that weren''t using the configured model
- Improved error handling in event deserialization with proper error messages including aggregate ID, event type, and field name
- Fixed RBAC issues with HATEOAS link generation

### API
- Enhanced HATEOAS links across all endpoints to include permission-aware action links (`update`, `delete`, `removeFromView`, etc.)
- Strategic fit analysis responses now include `importanceSource` and `isInherited` fields indicating where capability ratings originate', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
