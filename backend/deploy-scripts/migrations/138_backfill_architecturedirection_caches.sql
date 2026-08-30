-- Migration: Backfill Architecture Direction caches (spec 207)
-- Description: Populates architecturedirection.capability_node_cache and
--              architecturedirection.enterprise_capability_cache from the upstream tables so that
--              composition, source eligibility and maturity analysis are complete immediately after
--              deployment. Projectors keep both caches current from this point on. Idempotent.

INSERT INTO architecturedirection.enterprise_capability_cache (tenant_id, id, name, category, active, target_maturity)
SELECT tenant_id, id, name, category, active, target_maturity
FROM enterprisearchitecture.enterprise_capabilities
ON CONFLICT (tenant_id, id) DO UPDATE SET
    name = EXCLUDED.name,
    category = EXCLUDED.category,
    active = EXCLUDED.active,
    target_maturity = EXCLUDED.target_maturity;

INSERT INTO architecturedirection.capability_node_cache (
    tenant_id, capability_id, capability_name, capability_level, parent_id,
    l1_capability_id, business_domain_id, business_domain_name, maturity_value
)
WITH RECURSIVE tree AS (
    SELECT c.tenant_id, c.id, c.name, c.level, c.parent_id, c.id AS l1_id, c.maturity_value, 1 AS depth
    FROM capabilitymapping.capabilities c
    WHERE c.level = 'L1'
    UNION ALL
    SELECT c.tenant_id, c.id, c.name, c.level, c.parent_id, t.l1_id, c.maturity_value, t.depth + 1
    FROM capabilitymapping.capabilities c
    INNER JOIN tree t ON c.parent_id = t.id AND c.tenant_id = t.tenant_id
    WHERE t.depth < 10
),
l1_domain AS (
    SELECT DISTINCT ON (dca.tenant_id, dca.capability_id)
        dca.tenant_id, dca.capability_id, dca.business_domain_id, bd.name AS business_domain_name
    FROM capabilitymapping.domain_capability_assignments dca
    LEFT JOIN capabilitymapping.business_domains bd
        ON bd.id = dca.business_domain_id AND bd.tenant_id = dca.tenant_id
    ORDER BY dca.tenant_id, dca.capability_id, dca.business_domain_id
)
SELECT t.tenant_id, t.id, t.name, t.level, t.parent_id, t.l1_id,
       d.business_domain_id, d.business_domain_name, t.maturity_value
FROM tree t
LEFT JOIN l1_domain d ON d.tenant_id = t.tenant_id AND d.capability_id = t.l1_id
ON CONFLICT (tenant_id, capability_id) DO UPDATE SET
    capability_name = EXCLUDED.capability_name,
    capability_level = EXCLUDED.capability_level,
    parent_id = EXCLUDED.parent_id,
    l1_capability_id = EXCLUDED.l1_capability_id,
    business_domain_id = EXCLUDED.business_domain_id,
    business_domain_name = EXCLUDED.business_domain_name,
    maturity_value = EXCLUDED.maturity_value;
