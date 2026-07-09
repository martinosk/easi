-- Migration: Add One-Pager Configurations
-- Spec: 175_OnePagerConfiguration
-- Description: Add onepagers schema and one_pager_configurations read model table

CREATE SCHEMA IF NOT EXISTS onepagers;

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT USAGE ON SCHEMA onepagers TO easi_app';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA onepagers GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO easi_app';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA onepagers GRANT USAGE, SELECT ON SEQUENCES TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON SCHEMA onepagers TO easi_admin';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA onepagers GRANT ALL PRIVILEGES ON TABLES TO easi_admin';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA onepagers GRANT ALL PRIVILEGES ON SEQUENCES TO easi_admin';
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS onepagers.one_pager_configurations (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    subject_type VARCHAR(50) NOT NULL,
    configuration JSONB NOT NULL DEFAULT '{}'::jsonb,
    version INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL,
    modified_by VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_one_pager_configurations_subject_type
    ON onepagers.one_pager_configurations(tenant_id, subject_type);

ALTER TABLE onepagers.one_pager_configurations ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON onepagers.one_pager_configurations;
CREATE POLICY tenant_isolation_policy ON onepagers.one_pager_configurations
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA onepagers TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA onepagers TO easi_admin';
    END IF;
END $$;
