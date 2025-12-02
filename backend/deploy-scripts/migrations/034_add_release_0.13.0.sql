-- Migration: Add Release 0.13.0
-- Description: Adds release notes for version 0.13.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.13.0', '2025-12-02', '## What''s New in 0.13.0

### Major
- **Applications in Capability Grid**: View which applications realise each capability directly within the business domain grid view
  - Toggle "Show Applications" in the grid toolbar to display application chips inside capabilities
  - See realisation levels at a glance: Full (solid green), Partial (dashed yellow), Planned (dotted gray)
  - When a capability is hidden by the depth filter, its applications automatically bubble up to the nearest visible parent
  - Click any application chip to view its details

### API
- New endpoint: `GET /api/v1/business-domains/{domainId}/capability-realizations` for fetching applications within a business domain
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
