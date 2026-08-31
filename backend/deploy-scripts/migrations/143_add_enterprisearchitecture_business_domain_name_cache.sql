-- Migration: Enterprise Architecture business domain name cache (spec 209 WP4)
-- Description: Business domain name cache owned by Enterprise Architecture, fed by
--              Capability Mapping's BusinessDomainCreated/Updated/Deleted published events.
--              Replaces the composition-root BusinessDomainNameLookup bridge into
--              Capability Mapping's business domain read model.

CREATE TABLE IF NOT EXISTS enterprisearchitecture.business_domain_name_cache (
    tenant_id VARCHAR(50) NOT NULL,
    business_domain_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (tenant_id, business_domain_id)
);

ALTER TABLE enterprisearchitecture.business_domain_name_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON enterprisearchitecture.business_domain_name_cache;
CREATE POLICY tenant_isolation_policy ON enterprisearchitecture.business_domain_name_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON enterprisearchitecture.business_domain_name_cache TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON enterprisearchitecture.business_domain_name_cache TO easi_admin';
    END IF;
END $$;
