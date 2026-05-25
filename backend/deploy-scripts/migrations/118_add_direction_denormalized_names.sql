ALTER TABLE architecturedirection.direction_source_capabilities
    ADD COLUMN IF NOT EXISTS capability_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS business_domain_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS business_domain_name VARCHAR(255);

ALTER TABLE architecturedirection.standard_applications
    ADD COLUMN IF NOT EXISTS application_name VARCHAR(255);

ALTER TABLE architecturedirection.standard_application_history
    ADD COLUMN IF NOT EXISTS application_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS previous_application_name VARCHAR(255);

CREATE TABLE IF NOT EXISTS architecturedirection.reference_name_cache (
    tenant_id VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, entity_type, entity_id)
);

ALTER TABLE architecturedirection.reference_name_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.reference_name_cache;
CREATE POLICY tenant_isolation_policy ON architecturedirection.reference_name_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS architecturedirection.capability_domain_cache (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    business_domain_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, capability_id)
);

ALTER TABLE architecturedirection.capability_domain_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.capability_domain_cache;
CREATE POLICY tenant_isolation_policy ON architecturedirection.capability_domain_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.reference_name_cache TO easi_app';
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.capability_domain_cache TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.reference_name_cache TO easi_admin';
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.capability_domain_cache TO easi_admin';
    END IF;
END $$;
