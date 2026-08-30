-- Migration: Architecture Direction local caches (spec 207)
-- Description: Capability-node and enterprise-capability caches owned by Architecture Direction,
--              fed by Capability Mapping and Enterprise Architecture published events. They replace
--              the query-time reads of Enterprise Architecture's read models for composition,
--              source eligibility and maturity analysis.

CREATE TABLE IF NOT EXISTS architecturedirection.capability_node_cache (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    capability_name VARCHAR(500) NOT NULL,
    capability_level VARCHAR(2) NOT NULL,
    parent_id VARCHAR(255),
    l1_capability_id VARCHAR(255) NOT NULL,
    business_domain_id VARCHAR(255),
    business_domain_name VARCHAR(100),
    maturity_value INT,
    PRIMARY KEY (tenant_id, capability_id)
);

CREATE INDEX IF NOT EXISTS idx_ad_capability_node_cache_parent
    ON architecturedirection.capability_node_cache(tenant_id, parent_id);
CREATE INDEX IF NOT EXISTS idx_ad_capability_node_cache_l1
    ON architecturedirection.capability_node_cache(tenant_id, l1_capability_id);
CREATE INDEX IF NOT EXISTS idx_ad_capability_node_cache_business_domain
    ON architecturedirection.capability_node_cache(tenant_id, business_domain_id);

ALTER TABLE architecturedirection.capability_node_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.capability_node_cache;
CREATE POLICY tenant_isolation_policy ON architecturedirection.capability_node_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS architecturedirection.enterprise_capability_cache (
    tenant_id VARCHAR(50) NOT NULL,
    id VARCHAR(255) NOT NULL,
    name VARCHAR(200) NOT NULL,
    category VARCHAR(100),
    active BOOLEAN NOT NULL DEFAULT true,
    target_maturity INT,
    PRIMARY KEY (tenant_id, id)
);

ALTER TABLE architecturedirection.enterprise_capability_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.enterprise_capability_cache;
CREATE POLICY tenant_isolation_policy ON architecturedirection.enterprise_capability_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
