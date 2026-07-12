CREATE TABLE IF NOT EXISTS architecturedirection.realization_roles (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    realization_id VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL,
    assigned_by VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    capability_name VARCHAR(255),
    component_name VARCHAR(255),
    aggregate_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, capability_id, component_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_realization_roles_single_standard
    ON architecturedirection.realization_roles(tenant_id, capability_id) WHERE role = 'standard';

CREATE INDEX IF NOT EXISTS idx_realization_roles_component
    ON architecturedirection.realization_roles(tenant_id, component_id);

CREATE INDEX IF NOT EXISTS idx_realization_roles_realization
    ON architecturedirection.realization_roles(tenant_id, realization_id);

ALTER TABLE architecturedirection.realization_roles ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.realization_roles;
CREATE POLICY tenant_isolation_policy ON architecturedirection.realization_roles
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS architecturedirection.realization_role_aggregates (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, capability_id)
);

ALTER TABLE architecturedirection.realization_role_aggregates ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.realization_role_aggregates;
CREATE POLICY tenant_isolation_policy ON architecturedirection.realization_role_aggregates
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.realization_roles TO easi_app';
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.realization_role_aggregates TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.realization_roles TO easi_admin';
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.realization_role_aggregates TO easi_admin';
    END IF;
END $$;
