CREATE TABLE IF NOT EXISTS architecturedirection.standard_applications (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    enterprise_capability_id VARCHAR(255) NOT NULL,
    application_id VARCHAR(255) NOT NULL,
    narrative TEXT NOT NULL,
    application_stale BOOLEAN NOT NULL DEFAULT FALSE,
    set_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_standard_applications_per_ec
    ON architecturedirection.standard_applications(tenant_id, enterprise_capability_id);

CREATE INDEX IF NOT EXISTS idx_standard_applications_application
    ON architecturedirection.standard_applications(tenant_id, application_id);

ALTER TABLE architecturedirection.standard_applications ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.standard_applications;
CREATE POLICY tenant_isolation_policy ON architecturedirection.standard_applications
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS architecturedirection.standard_application_history (
    tenant_id VARCHAR(50) NOT NULL,
    standard_application_id VARCHAR(255) NOT NULL,
    sequence INTEGER NOT NULL,
    application_id VARCHAR(255) NOT NULL,
    previous_application_id VARCHAR(255),
    narrative TEXT NOT NULL,
    set_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, standard_application_id, sequence)
);

CREATE INDEX IF NOT EXISTS idx_standard_application_history_aggregate
    ON architecturedirection.standard_application_history(tenant_id, standard_application_id, sequence DESC);

ALTER TABLE architecturedirection.standard_application_history ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.standard_application_history;
CREATE POLICY tenant_isolation_policy ON architecturedirection.standard_application_history
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.standard_applications TO easi_app';
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.standard_application_history TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.standard_applications TO easi_admin';
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.standard_application_history TO easi_admin';
    END IF;
END $$;
