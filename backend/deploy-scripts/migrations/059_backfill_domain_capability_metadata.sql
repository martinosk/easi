-- Migration: Backfill Domain Capability Metadata
-- Description: Populates domain_capability_metadata for existing capabilities that were
--              created before the enterprise architecture feature was deployed.
--              This is a one-time data repair migration.

-- ============================================================================
-- Step 1: Insert metadata for all L1 capabilities with their business domain assignments
-- ============================================================================

INSERT INTO domain_capability_metadata (
    tenant_id,
    capability_id,
    capability_name,
    capability_level,
    parent_id,
    l1_capability_id,
    business_domain_id,
    business_domain_name
)
SELECT
    c.tenant_id,
    c.id AS capability_id,
    c.name AS capability_name,
    c.level AS capability_level,
    c.parent_id,
    c.id AS l1_capability_id,
    dca.business_domain_id,
    bd.name AS business_domain_name
FROM capabilities c
LEFT JOIN domain_capability_assignments dca
    ON c.id = dca.capability_id AND c.tenant_id = dca.tenant_id
LEFT JOIN business_domains bd
    ON dca.business_domain_id = bd.id AND dca.tenant_id = bd.tenant_id
WHERE c.level = 'L1'
  AND NOT EXISTS (
      SELECT 1 FROM domain_capability_metadata dcm
      WHERE dcm.capability_id = c.id AND dcm.tenant_id = c.tenant_id
  )
ON CONFLICT (tenant_id, capability_id) DO NOTHING;

-- ============================================================================
-- Step 2: Insert metadata for L2 capabilities (inherit L1 from parent)
-- ============================================================================

INSERT INTO domain_capability_metadata (
    tenant_id,
    capability_id,
    capability_name,
    capability_level,
    parent_id,
    l1_capability_id,
    business_domain_id,
    business_domain_name
)
SELECT
    c.tenant_id,
    c.id AS capability_id,
    c.name AS capability_name,
    c.level AS capability_level,
    c.parent_id,
    parent_meta.l1_capability_id,
    parent_meta.business_domain_id,
    parent_meta.business_domain_name
FROM capabilities c
INNER JOIN domain_capability_metadata parent_meta
    ON c.parent_id = parent_meta.capability_id AND c.tenant_id = parent_meta.tenant_id
WHERE c.level = 'L2'
  AND NOT EXISTS (
      SELECT 1 FROM domain_capability_metadata dcm
      WHERE dcm.capability_id = c.id AND dcm.tenant_id = c.tenant_id
  )
ON CONFLICT (tenant_id, capability_id) DO NOTHING;

-- ============================================================================
-- Step 3: Insert metadata for L3 capabilities (inherit L1 from parent which got it from L1)
-- ============================================================================

INSERT INTO domain_capability_metadata (
    tenant_id,
    capability_id,
    capability_name,
    capability_level,
    parent_id,
    l1_capability_id,
    business_domain_id,
    business_domain_name
)
SELECT
    c.tenant_id,
    c.id AS capability_id,
    c.name AS capability_name,
    c.level AS capability_level,
    c.parent_id,
    parent_meta.l1_capability_id,
    parent_meta.business_domain_id,
    parent_meta.business_domain_name
FROM capabilities c
INNER JOIN domain_capability_metadata parent_meta
    ON c.parent_id = parent_meta.capability_id AND c.tenant_id = parent_meta.tenant_id
WHERE c.level = 'L3'
  AND NOT EXISTS (
      SELECT 1 FROM domain_capability_metadata dcm
      WHERE dcm.capability_id = c.id AND dcm.tenant_id = c.tenant_id
  )
ON CONFLICT (tenant_id, capability_id) DO NOTHING;

-- ============================================================================
-- Step 4: Insert metadata for L4 capabilities (inherit L1 from parent chain)
-- ============================================================================

INSERT INTO domain_capability_metadata (
    tenant_id,
    capability_id,
    capability_name,
    capability_level,
    parent_id,
    l1_capability_id,
    business_domain_id,
    business_domain_name
)
SELECT
    c.tenant_id,
    c.id AS capability_id,
    c.name AS capability_name,
    c.level AS capability_level,
    c.parent_id,
    parent_meta.l1_capability_id,
    parent_meta.business_domain_id,
    parent_meta.business_domain_name
FROM capabilities c
INNER JOIN domain_capability_metadata parent_meta
    ON c.parent_id = parent_meta.capability_id AND c.tenant_id = parent_meta.tenant_id
WHERE c.level = 'L4'
  AND NOT EXISTS (
      SELECT 1 FROM domain_capability_metadata dcm
      WHERE dcm.capability_id = c.id AND dcm.tenant_id = c.tenant_id
  )
ON CONFLICT (tenant_id, capability_id) DO NOTHING;

-- ============================================================================
-- Migration complete
-- ============================================================================
