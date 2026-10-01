-- Migration: 168_add_stewardship_home_caches
-- Purpose: spec 228 — the Stewardship read side behind GET /home. Six caches
-- mirror supplier facts (capabilities, L1 domain assignments, applications,
-- direct realizations, TIME grades, edit grants) so the home is composed from
-- this schema alone. Fed by published-language events and seeded by migration 169.
-- edit_grant_cache holds grantee e-mails (personal data): a user-erasure flow
-- must purge it together with user_cache.

CREATE TABLE IF NOT EXISTS stewardship.capability_cache (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL DEFAULT '',
    level VARCHAR(2) NOT NULL DEFAULT '',
    parent_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'Active',
    ea_owner TEXT,
    PRIMARY KEY (tenant_id, capability_id)
);

CREATE INDEX IF NOT EXISTS idx_stewardship_capability_cache_parent
    ON stewardship.capability_cache(tenant_id, parent_id);

CREATE INDEX IF NOT EXISTS idx_stewardship_capability_cache_ea_owner
    ON stewardship.capability_cache(tenant_id, ea_owner);

CREATE TABLE IF NOT EXISTS stewardship.domain_assignment_cache (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    domain_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, capability_id, domain_id)
);

CREATE INDEX IF NOT EXISTS idx_stewardship_domain_assignment_cache_domain
    ON stewardship.domain_assignment_cache(tenant_id, domain_id);

CREATE TABLE IF NOT EXISTS stewardship.application_cache (
    tenant_id VARCHAR(50) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL DEFAULT '',
    ownership_state VARCHAR(20) NOT NULL DEFAULT 'unknown',
    owner_kind VARCHAR(10),
    owner_id VARCHAR(255),
    PRIMARY KEY (tenant_id, component_id)
);

CREATE INDEX IF NOT EXISTS idx_stewardship_application_cache_owner
    ON stewardship.application_cache(tenant_id, owner_kind, owner_id);

CREATE TABLE IF NOT EXISTS stewardship.realization_cache (
    tenant_id VARCHAR(50) NOT NULL,
    realization_id VARCHAR(255) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, realization_id)
);

CREATE INDEX IF NOT EXISTS idx_stewardship_realization_cache_capability
    ON stewardship.realization_cache(tenant_id, capability_id);

CREATE INDEX IF NOT EXISTS idx_stewardship_realization_cache_component
    ON stewardship.realization_cache(tenant_id, component_id);

CREATE TABLE IF NOT EXISTS stewardship.time_assessment_cache (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    grade VARCHAR(20) NOT NULL,
    assessed_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, capability_id, component_id)
);

CREATE TABLE IF NOT EXISTS stewardship.edit_grant_cache (
    tenant_id VARCHAR(50) NOT NULL,
    grant_id VARCHAR(255) NOT NULL,
    artifact_type VARCHAR(20) NOT NULL,
    artifact_id VARCHAR(255) NOT NULL,
    grantee_email VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, grant_id)
);

CREATE INDEX IF NOT EXISTS idx_stewardship_edit_grant_cache_grantee
    ON stewardship.edit_grant_cache(tenant_id, lower(grantee_email));

DO $$
DECLARE
    cache TEXT;
BEGIN
    FOREACH cache IN ARRAY ARRAY['capability_cache', 'domain_assignment_cache', 'application_cache',
                                 'realization_cache', 'time_assessment_cache', 'edit_grant_cache']
    LOOP
        EXECUTE format('ALTER TABLE stewardship.%I ENABLE ROW LEVEL SECURITY', cache);
        EXECUTE format('DROP POLICY IF EXISTS tenant_isolation_policy ON stewardship.%I', cache);
        EXECUTE format('CREATE POLICY tenant_isolation_policy ON stewardship.%I FOR ALL TO easi_app '
                       'USING (tenant_id = current_setting(''app.current_tenant'', true)) '
                       'WITH CHECK (tenant_id = current_setting(''app.current_tenant'', true))', cache);
        IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
            EXECUTE format('GRANT SELECT, INSERT, UPDATE, DELETE ON stewardship.%I TO easi_app', cache);
        END IF;
        IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
            EXECUTE format('GRANT ALL PRIVILEGES ON stewardship.%I TO easi_admin', cache);
        END IF;
    END LOOP;
END $$;
