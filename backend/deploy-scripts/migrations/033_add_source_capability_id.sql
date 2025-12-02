-- Migration: Add source_capability_id to capability_realizations
-- Description: Stores the source capability ID directly for inherited realizations
--              to avoid JOINs in the read model

ALTER TABLE capability_realizations
ADD COLUMN IF NOT EXISTS source_capability_id VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_capability_realizations_source_capability
ON capability_realizations(tenant_id, source_capability_id)
WHERE source_capability_id IS NOT NULL;

-- Backfill from source_realization_id
UPDATE capability_realizations cr
SET source_capability_id = source_r.capability_id
FROM capability_realizations source_r
WHERE cr.source_realization_id = source_r.id
  AND cr.tenant_id = source_r.tenant_id
  AND cr.origin = 'Inherited'
  AND cr.source_capability_id IS NULL;
