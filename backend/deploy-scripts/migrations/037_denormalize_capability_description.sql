-- Migration: Denormalize Capability Description in Assignments
-- Description: Adds capability_description column to domain_capability_assignments
-- and backfills from capabilities table to eliminate JOIN queries

-- Add the new column
ALTER TABLE domain_capability_assignments
ADD COLUMN IF NOT EXISTS capability_description TEXT NOT NULL DEFAULT '';

-- Backfill existing assignments with capability descriptions
UPDATE domain_capability_assignments a
SET capability_description = COALESCE(c.description, '')
FROM capabilities c
WHERE c.id = a.capability_id
  AND c.tenant_id = a.tenant_id;