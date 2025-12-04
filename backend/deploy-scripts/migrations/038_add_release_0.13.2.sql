-- Migration: Add Release 0.13.2
-- Description: Adds release notes for version 0.13.2

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.13.2', '2025-12-04', '## What''s New in 0.13.2

### Performance
- **Fixed N+1 query performance issue**: Business domain capabilities endpoint now completes in <50ms instead of >1000ms for domains with 100+ capabilities
- **Fully denormalized read model**: Assignment queries now use a single table scan with no JOINs required

### Technical Improvements
- Added `capability_description` column to domain assignment read model for complete denormalization
- Capability updates now automatically propagate to all related domain assignments via event handlers
- Eliminated cross-table JOINs in read models maintaining clean bounded context boundaries
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
