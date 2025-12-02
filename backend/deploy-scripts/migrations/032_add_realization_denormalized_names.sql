-- Migration: Add denormalized names to capability_realizations
-- Spec: 063_ShowApplications_CapabilityGrid_ongoing.md
-- Description: Adds component_name and source_capability_name columns to capability_realizations
--              for proper bounded context isolation (no cross-context JOINs)

ALTER TABLE capability_realizations
ADD COLUMN IF NOT EXISTS component_name VARCHAR(255);

ALTER TABLE capability_realizations
ADD COLUMN IF NOT EXISTS source_capability_name VARCHAR(255);

-- Backfill component names from application_components
UPDATE capability_realizations cr
SET component_name = ac.name
FROM application_components ac
WHERE cr.component_id = ac.id
  AND cr.tenant_id = ac.tenant_id
  AND ac.is_deleted = FALSE
  AND cr.component_name IS NULL;

-- Backfill source capability names for inherited realizations
UPDATE capability_realizations cr
SET source_capability_name = c.name
FROM capability_realizations source_r
JOIN capabilities c ON source_r.capability_id = c.id AND source_r.tenant_id = c.tenant_id
WHERE cr.source_realization_id = source_r.id
  AND cr.tenant_id = source_r.tenant_id
  AND cr.origin = 'Inherited'
  AND cr.source_capability_name IS NULL;
