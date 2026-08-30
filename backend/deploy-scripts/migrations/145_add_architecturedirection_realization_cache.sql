-- Migration: Architecture Direction realization cache and non-null capability maturity (spec 209)
-- Description: Adds the realization cache Architecture Direction uses to resolve the direct
--              realization behind a time assessment or a realization role, replacing the
--              request-time read of Capability Mapping's realization read model. Also repairs
--              capability_node_cache.maturity_value, which was nullable and never written on
--              CapabilityCreated, so maturity analysis failed to scan rows created after
--              migration 137. Capability Mapping's true default maturity is Genesis (12), so the
--              NULL rows left behind by the old projector are backfilled to 12, not 0.

UPDATE architecturedirection.capability_node_cache SET maturity_value = 12 WHERE maturity_value IS NULL;
ALTER TABLE architecturedirection.capability_node_cache ALTER COLUMN maturity_value SET DEFAULT 12;
ALTER TABLE architecturedirection.capability_node_cache ALTER COLUMN maturity_value SET NOT NULL;

CREATE TABLE IF NOT EXISTS architecturedirection.realization_cache (
    tenant_id VARCHAR(50) NOT NULL,
    realization_id VARCHAR(255) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, realization_id)
);

CREATE INDEX IF NOT EXISTS idx_ad_realization_cache_pair
    ON architecturedirection.realization_cache(tenant_id, capability_id, component_id);
CREATE INDEX IF NOT EXISTS idx_ad_realization_cache_component
    ON architecturedirection.realization_cache(tenant_id, component_id);

ALTER TABLE architecturedirection.realization_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.realization_cache;
CREATE POLICY tenant_isolation_policy ON architecturedirection.realization_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.realization_cache TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.realization_cache TO easi_admin';
    END IF;
END $$;
