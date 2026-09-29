-- Migration: 166_add_stewardship_context
-- Purpose: spec 226 — the Stewardship context. stewardships holds one row per
-- live (domain, concern) stewardship, removed on release; the unique index is the
-- backstop behind the handler's one-steward-per-pair check. domain_cache and
-- user_cache are fed by Capability Mapping and Auth events and seeded by
-- migration 167; user_cache.active backs the rule that only active users are
-- assignable.

CREATE SCHEMA IF NOT EXISTS stewardship;

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT USAGE ON SCHEMA stewardship TO easi_app';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA stewardship GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON SCHEMA stewardship TO easi_admin';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA stewardship GRANT ALL PRIVILEGES ON TABLES TO easi_admin';
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS stewardship.stewardships (
    tenant_id VARCHAR(50) NOT NULL,
    id VARCHAR(255) NOT NULL,
    domain_id VARCHAR(255) NOT NULL,
    concern VARCHAR(50) NOT NULL,
    steward_id VARCHAR(255) NOT NULL,
    assigned_by VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_stewardships_domain_concern
    ON stewardship.stewardships(tenant_id, domain_id, concern);

CREATE TABLE IF NOT EXISTS stewardship.domain_cache (
    tenant_id VARCHAR(50) NOT NULL,
    domain_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    domain_architect_id VARCHAR(255),
    PRIMARY KEY (tenant_id, domain_id)
);

CREATE TABLE IF NOT EXISTS stewardship.user_cache (
    tenant_id VARCHAR(50) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL DEFAULT '',
    email VARCHAR(255) NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (tenant_id, user_id)
);

ALTER TABLE stewardship.stewardships ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON stewardship.stewardships;
CREATE POLICY tenant_isolation_policy ON stewardship.stewardships
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

ALTER TABLE stewardship.domain_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON stewardship.domain_cache;
CREATE POLICY tenant_isolation_policy ON stewardship.domain_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

ALTER TABLE stewardship.user_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON stewardship.user_cache;
CREATE POLICY tenant_isolation_policy ON stewardship.user_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON stewardship.stewardships, stewardship.domain_cache, stewardship.user_cache TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON stewardship.stewardships, stewardship.domain_cache, stewardship.user_cache TO easi_admin';
    END IF;
END $$;
