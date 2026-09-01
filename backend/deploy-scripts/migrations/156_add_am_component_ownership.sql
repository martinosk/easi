-- Migration: 156_add_am_component_ownership
-- Purpose: spec 214 — ownership state machine on application components
-- (unknown/nominated/owned/managed with a typed owner reference) plus an
-- AM-owned cache of user display names so owner references can be validated
-- and rendered without crossing bounded context schemas. The cache is
-- populated via subscription to auth's UserCreated published-language event
-- and seeded by migration 157 (backfill).

ALTER TABLE architecturemodeling.application_components
    ADD COLUMN IF NOT EXISTS ownership_state VARCHAR(20) NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS owner_kind VARCHAR(10),
    ADD COLUMN IF NOT EXISTS owner_id VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_application_components_ownership_state
    ON architecturemodeling.application_components(tenant_id, ownership_state);

CREATE INDEX IF NOT EXISTS idx_application_components_owner
    ON architecturemodeling.application_components(tenant_id, owner_kind, owner_id);

CREATE TABLE IF NOT EXISTS architecturemodeling.user_names (
    tenant_id VARCHAR(50) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, user_id)
);

ALTER TABLE architecturemodeling.user_names ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON architecturemodeling.user_names;
CREATE POLICY tenant_isolation_policy ON architecturemodeling.user_names
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON architecturemodeling.user_names TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON architecturemodeling.user_names TO easi_admin';
    END IF;
END $$;
