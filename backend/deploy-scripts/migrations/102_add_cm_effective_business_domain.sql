CREATE TABLE IF NOT EXISTS cm_effective_business_domain (
    tenant_id VARCHAR(255) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    business_domain_id VARCHAR(255),
    business_domain_name VARCHAR(500),
    l1_capability_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, capability_id)
);

ALTER TABLE cm_effective_business_domain ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON cm_effective_business_domain
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Backfill cm_effective_business_domain from CM's own source tables
WITH RECURSIVE capability_tree AS (
    SELECT id, tenant_id, parent_id, id AS root_l1_id, level
    FROM capabilities
    WHERE level = 'L1'

    UNION ALL

    SELECT c.id, c.tenant_id, c.parent_id, ct.root_l1_id, c.level
    FROM capabilities c
    JOIN capability_tree ct ON c.parent_id = ct.id AND c.tenant_id = ct.tenant_id
    WHERE c.level != 'L1'
)
INSERT INTO cm_effective_business_domain (tenant_id, capability_id, business_domain_id, business_domain_name, l1_capability_id)
SELECT
    ct.tenant_id,
    ct.id,
    dca.business_domain_id,
    dca.business_domain_name,
    ct.root_l1_id
FROM capability_tree ct
LEFT JOIN domain_capability_assignments dca
    ON dca.capability_id = ct.root_l1_id AND dca.tenant_id = ct.tenant_id
ON CONFLICT DO NOTHING;
