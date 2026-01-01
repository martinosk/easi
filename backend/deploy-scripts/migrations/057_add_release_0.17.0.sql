-- Migration: Add Release 0.17.0
-- Description: Adds release notes for version 0.17.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.17.0', '2026-01-01', '## What''s New in 0.17.0

### Major
- Added Enterprise Architecture module for cross-domain capability analysis
- Introduced Enterprise Capabilities to create canonical capability groupings and discover overlap across business domains
- Added Maturity Gap Analysis dashboard to identify standardization opportunities and prioritize investments
- Enabled customizable Strategy Pillars so organizations can define their own strategic initiatives
- Added Strategic Importance rating for domain capabilities against strategy pillars
- Introduced MetaModel bounded context for tenant-specific tool customization

### Bugs
- Fixed removing capabilities from business domains not working correctly

### API
- New endpoint: `GET /api/v1/enterprise-capabilities` - List enterprise capabilities with implementation counts
- New endpoint: `POST /api/v1/enterprise-capabilities/{id}/links` - Link domain capabilities to enterprise capabilities
- New endpoint: `GET /api/v1/enterprise-capabilities/maturity-analysis` - Get maturity gap candidates
- New endpoint: `GET /api/v1/enterprise-capabilities/{id}/maturity-gap` - Get maturity gap detail for an enterprise capability
- New endpoint: `GET /api/v1/domain-capabilities/unlinked` - List unlinked domain capabilities
- New endpoint: `PUT /api/v1/enterprise-capabilities/{id}/target-maturity` - Set target maturity for an enterprise capability
- New endpoint: `GET /api/v1/strategy-pillars` - List configurable strategy pillars
- New endpoint: `POST /api/v1/strategy-importance` - Rate capability strategic importance', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
