CREATE TABLE IF NOT EXISTS architecturedirection.time_assessments (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    realization_id VARCHAR(255) NOT NULL,
    grade VARCHAR(20) NOT NULL,
    rationale TEXT NOT NULL DEFAULT '',
    assessed_by VARCHAR(255) NOT NULL,
    assessed_by_name VARCHAR(255) NOT NULL,
    assessed_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    capability_name VARCHAR(255),
    component_name VARCHAR(255),
    PRIMARY KEY (tenant_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_time_assessments_per_pair
    ON architecturedirection.time_assessments(tenant_id, capability_id, component_id);

CREATE INDEX IF NOT EXISTS idx_time_assessments_component
    ON architecturedirection.time_assessments(tenant_id, component_id);

CREATE INDEX IF NOT EXISTS idx_time_assessments_realization
    ON architecturedirection.time_assessments(tenant_id, realization_id);

ALTER TABLE architecturedirection.time_assessments ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.time_assessments;
CREATE POLICY tenant_isolation_policy ON architecturedirection.time_assessments
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.time_assessments TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.time_assessments TO easi_admin';
    END IF;
END $$;
