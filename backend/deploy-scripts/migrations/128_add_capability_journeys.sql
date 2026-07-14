CREATE TABLE IF NOT EXISTS architecturedirection.capability_journeys (
    tenant_id VARCHAR(50) NOT NULL,
    id VARCHAR(255) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    kind VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    progress INT,
    target_year INT,
    target_quarter SMALLINT,
    note TEXT NOT NULL DEFAULT '',
    planned_by VARCHAR(255) NOT NULL,
    planned_by_name VARCHAR(255) NOT NULL DEFAULT '',
    planned_at TIMESTAMP NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    abandoned_at TIMESTAMP,
    updated_at TIMESTAMP,
    capability_name VARCHAR(255) NOT NULL DEFAULT '',
    capability_stale BOOLEAN NOT NULL DEFAULT FALSE,
    to_component_id VARCHAR(255) NOT NULL,
    to_component_name VARCHAR(255) NOT NULL DEFAULT '',
    to_component_stale BOOLEAN NOT NULL DEFAULT FALSE,
    target_domain_id VARCHAR(255),
    target_domain_name VARCHAR(255) NOT NULL DEFAULT '',
    target_domain_stale BOOLEAN NOT NULL DEFAULT FALSE,
    target_parent_id VARCHAR(255),
    target_parent_name VARCHAR(255) NOT NULL DEFAULT '',
    target_parent_stale BOOLEAN NOT NULL DEFAULT FALSE,
    resulting_name VARCHAR(255) NOT NULL DEFAULT '',
    PRIMARY KEY (tenant_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_capability_journeys_single_active
    ON architecturedirection.capability_journeys(tenant_id, capability_id)
    WHERE status IN ('planned', 'in-flight');

CREATE INDEX IF NOT EXISTS idx_capability_journeys_capability
    ON architecturedirection.capability_journeys(tenant_id, capability_id);

CREATE INDEX IF NOT EXISTS idx_capability_journeys_to_component
    ON architecturedirection.capability_journeys(tenant_id, to_component_id);

ALTER TABLE architecturedirection.capability_journeys ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.capability_journeys;
CREATE POLICY tenant_isolation_policy ON architecturedirection.capability_journeys
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS architecturedirection.capability_journey_sources (
    tenant_id VARCHAR(50) NOT NULL,
    journey_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    position INT NOT NULL,
    component_name VARCHAR(255) NOT NULL DEFAULT '',
    component_stale BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (tenant_id, journey_id, component_id)
);

CREATE INDEX IF NOT EXISTS idx_capability_journey_sources_component
    ON architecturedirection.capability_journey_sources(tenant_id, component_id);

ALTER TABLE architecturedirection.capability_journey_sources ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.capability_journey_sources;
CREATE POLICY tenant_isolation_policy ON architecturedirection.capability_journey_sources
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS architecturedirection.capability_journey_milestones (
    tenant_id VARCHAR(50) NOT NULL,
    journey_id VARCHAR(255) NOT NULL,
    milestone_id VARCHAR(255) NOT NULL,
    position INT NOT NULL,
    label VARCHAR(255) NOT NULL,
    target_year INT,
    target_quarter SMALLINT,
    status VARCHAR(20) NOT NULL,
    updated_at TIMESTAMP,
    PRIMARY KEY (tenant_id, journey_id, milestone_id)
);

ALTER TABLE architecturedirection.capability_journey_milestones ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON architecturedirection.capability_journey_milestones;
CREATE POLICY tenant_isolation_policy ON architecturedirection.capability_journey_milestones
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.capability_journeys TO easi_app';
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.capability_journey_sources TO easi_app';
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturedirection.capability_journey_milestones TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.capability_journeys TO easi_admin';
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.capability_journey_sources TO easi_admin';
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturedirection.capability_journey_milestones TO easi_admin';
    END IF;
END $$;
