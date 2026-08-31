-- Migration: Backfill Architecture Direction caches (spec 207)
-- Description: Populates architecturedirection.capability_node_cache and
--              architecturedirection.enterprise_capability_cache from the upstream tables so that
--              composition, source eligibility and maturity analysis are complete immediately after
--              deployment. Projectors keep both caches current from this point on. Idempotent.
--              capability_node_cache is seeded by walking every capability up to its L1 ancestor
--              (mirroring the projector's l1AncestorOf semantics), not by walking down from L1 roots,
--              so orphans, chains rooted at a non-'L1' level and broken parent chains still get a row,
--              with the unreachable node's own id as its l1_capability_id and no business domain.

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
WITH RECURSIVE ancestor_walk AS (
    SELECT
        c.tenant_id, c.id AS origin_id, c.id AS current_id, c.level AS current_level,
        c.parent_id AS current_parent_id, FALSE AS broken, 1 AS depth
    FROM capabilitymapping.capabilities c
    UNION ALL
    SELECT
        w.tenant_id, w.origin_id, p.id AS current_id, p.level AS current_level,
        p.parent_id AS current_parent_id, (p.id IS NULL) AS broken, w.depth + 1
    FROM ancestor_walk w
    LEFT JOIN capabilitymapping.capabilities p
        ON p.tenant_id = w.tenant_id AND p.id = w.current_parent_id
    WHERE w.broken = FALSE
      AND w.current_level <> 'L1'
      AND w.current_parent_id IS NOT NULL
      AND w.current_parent_id <> ''
      AND w.depth < 10
),
terminal AS (
    SELECT DISTINCT ON (tenant_id, origin_id)
        tenant_id, origin_id, current_id, current_level, current_parent_id, broken
    FROM ancestor_walk
    ORDER BY tenant_id, origin_id, depth DESC
),
l1_per_capability AS (
    SELECT
        tenant_id, origin_id AS capability_id,
        CASE
            WHEN broken THEN origin_id
            WHEN current_level = 'L1' OR current_parent_id IS NULL OR current_parent_id = '' THEN current_id
            ELSE origin_id
        END AS l1_capability_id
    FROM terminal
),
l1_domain AS (
    SELECT DISTINCT ON (dca.tenant_id, dca.capability_id)
        dca.tenant_id, dca.capability_id, dca.business_domain_id, bd.name AS business_domain_name
    FROM capabilitymapping.domain_capability_assignments dca
    LEFT JOIN capabilitymapping.business_domains bd
        ON bd.id = dca.business_domain_id AND bd.tenant_id = dca.tenant_id
    ORDER BY dca.tenant_id, dca.capability_id, dca.business_domain_id
)
SELECT c.tenant_id, c.id, c.name, c.level, c.parent_id,
       l1.l1_capability_id, d.business_domain_id, d.business_domain_name, c.maturity_value
FROM capabilitymapping.capabilities c
JOIN l1_per_capability l1 ON l1.tenant_id = c.tenant_id AND l1.capability_id = c.id
LEFT JOIN l1_domain d ON d.tenant_id = l1.tenant_id AND d.capability_id = l1.l1_capability_id
ON CONFLICT (tenant_id, capability_id) DO UPDATE SET
    capability_name = EXCLUDED.capability_name,
    capability_level = EXCLUDED.capability_level,
    parent_id = EXCLUDED.parent_id,
    l1_capability_id = EXCLUDED.l1_capability_id,
    business_domain_id = EXCLUDED.business_domain_id,
    business_domain_name = EXCLUDED.business_domain_name,
    maturity_value = EXCLUDED.maturity_value;
