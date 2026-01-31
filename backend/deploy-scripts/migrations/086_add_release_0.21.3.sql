-- Migration: Add Release 0.21.3
-- Description: Adds release notes for version 0.21.3

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.21.3', '2026-01-31', '## What''s New in v0.21.3

### Major
- Added Pillar Fit Type configuration to strategy pillars, allowing each pillar to be tagged as assessing Technical or Functional fit
- Added TIME Suggestions preview, showing automated Tolerate/Invest/Migrate/Eliminate classifications for application realizations based on Strategic Fit gap analysis
- Added strategic importance rationale display in Strategic Fit analysis

### Bugs
- Fixed TIME suggestions not showing when domain capability metadata was incomplete

### API
- New endpoint: `GET /time-suggestions` - retrieve calculated TIME suggestions with optional filtering by capability or component
- Updated endpoint: `PUT /strategy-pillars/{id}` - now includes `fitType` field (TECHNICAL / FUNCTIONAL)', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
