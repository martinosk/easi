-- Migration: 163_add_am_component_containment
-- Purpose: spec 216 — two-level application composition. A part's parent
-- reference and containment kind (composition/aggregation) are projected onto
-- the component read model from the ComponentAttached/ComponentDetached
-- events; NULL means standalone, so every existing row reads as standalone
-- without a backfill. The registry table maps a tenant to its single
-- ComponentContainments aggregate (primary key = one aggregate per tenant).

ALTER TABLE architecturemodeling.application_components
    ADD COLUMN IF NOT EXISTS parent_component_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS containment_kind VARCHAR(20);

CREATE INDEX IF NOT EXISTS idx_application_components_parent
    ON architecturemodeling.application_components(tenant_id, parent_component_id);

CREATE TABLE IF NOT EXISTS architecturemodeling.component_containment_aggregates (
    tenant_id VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id)
);

ALTER TABLE architecturemodeling.component_containment_aggregates ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON architecturemodeling.component_containment_aggregates;
CREATE POLICY tenant_isolation_policy ON architecturemodeling.component_containment_aggregates
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturemodeling.component_containment_aggregates TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturemodeling.component_containment_aggregates TO easi_admin';
    END IF;
END $$;
