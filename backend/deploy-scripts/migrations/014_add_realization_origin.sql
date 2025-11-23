-- Migration: Add Realization Origin and Source columns
-- Spec: 038_Link_Capabilities_Applications_pending.md
-- Description: Adds origin tracking for capability realizations (Direct vs Inherited)

-- Add origin column to track whether realization was explicit (Direct) or derived (Inherited)
ALTER TABLE capability_realizations
ADD COLUMN IF NOT EXISTS origin VARCHAR(20) NOT NULL DEFAULT 'Direct' CHECK (origin IN ('Direct', 'Inherited'));

-- Add source_realization_id to track which direct realization triggered an inherited one
ALTER TABLE capability_realizations
ADD COLUMN IF NOT EXISTS source_realization_id VARCHAR(255);

-- Index for querying inherited realizations by their source
CREATE INDEX IF NOT EXISTS idx_capability_realizations_source ON capability_realizations(tenant_id, source_realization_id)
WHERE source_realization_id IS NOT NULL;

-- Index for filtering by origin
CREATE INDEX IF NOT EXISTS idx_capability_realizations_origin ON capability_realizations(tenant_id, origin);
