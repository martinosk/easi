CREATE SCHEMA IF NOT EXISTS architecturedirection;

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT USAGE ON SCHEMA architecturedirection TO easi_app';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA architecturedirection GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO easi_app';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA architecturedirection GRANT USAGE, SELECT ON SEQUENCES TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON SCHEMA architecturedirection TO easi_admin';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA architecturedirection GRANT ALL PRIVILEGES ON TABLES TO easi_admin';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA architecturedirection GRANT ALL PRIVILEGES ON SEQUENCES TO easi_admin';
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS architecturedirection.directions (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    enterprise_capability_id VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    horizon VARCHAR(10) NOT NULL,
    narrative TEXT,
    placements JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id)
);

CREATE INDEX IF NOT EXISTS idx_directions_ec ON architecturedirection.directions(tenant_id, enterprise_capability_id);
CREATE INDEX IF NOT EXISTS idx_directions_status ON architecturedirection.directions(tenant_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS uq_directions_active_per_ec
    ON architecturedirection.directions(tenant_id, enterprise_capability_id)
    WHERE status != 'rejected';

ALTER TABLE architecturedirection.directions ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.directions;
CREATE POLICY tenant_isolation_policy ON architecturedirection.directions
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS architecturedirection.direction_source_capabilities (
    tenant_id VARCHAR(50) NOT NULL,
    direction_id VARCHAR(255) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    stale BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (tenant_id, direction_id, capability_id)
);

CREATE INDEX IF NOT EXISTS idx_direction_sources_capability
    ON architecturedirection.direction_source_capabilities(tenant_id, capability_id);

ALTER TABLE architecturedirection.direction_source_capabilities ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.direction_source_capabilities;
CREATE POLICY tenant_isolation_policy ON architecturedirection.direction_source_capabilities
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA architecturedirection TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA architecturedirection TO easi_admin';
    END IF;
END $$;
