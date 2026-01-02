-- Migration: Add Release 0.18.0
-- Description: Adds release notes for version 0.18.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.18.0', '2026-01-02', '## What''s New in v0.18.0

### Major
- Added **Strategic Fit Analysis** - Score how well applications fit your organization''s strategy pillars to identify strategic liabilities in your IT landscape
  - Configure which strategy pillars should have fit scoring enabled
  - Rate applications (1-5) on how well they support each pillar
  - View fit scores vs capability importance to spot misalignments
  - New Strategic Fit dashboard tab shows all liabilities at a glance

### API
- New endpoint: `PUT /api/v1/strategy-pillars/{id}/fit-configuration` - Configure fit scoring for pillars
- New endpoint: `GET /api/v1/components/{id}/fit-scores` - Get all fit scores for an application
- New endpoint: `PUT /api/v1/components/{id}/fit-scores/{pillarId}` - Set fit score for a pillar
- New endpoint: `DELETE /api/v1/components/{id}/fit-scores/{pillarId}` - Remove fit score
- New endpoint: `GET /api/v1/strategic-fit-analysis` - Get strategic liability analysis', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
