-- Migration: 134_add_cm_user_names_cache
-- Purpose: CM-owned cache of user display names so capability read models can
-- resolve EA owner user ids to names without crossing bounded context schemas.
-- Populated via subscription to auth's UserCreated published-language event;
-- backfilled here from auth.users (cross-schema access is allowed in migrations).

CREATE TABLE IF NOT EXISTS capabilitymapping.user_names (
    tenant_id VARCHAR(50) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, user_id)
);

ALTER TABLE capabilitymapping.user_names ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON capabilitymapping.user_names;
CREATE POLICY tenant_isolation_policy ON capabilitymapping.user_names
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON capabilitymapping.user_names TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON capabilitymapping.user_names TO easi_admin';
    END IF;
END $$;

INSERT INTO capabilitymapping.user_names (tenant_id, user_id, name, email)
SELECT tenant_id, id::text, COALESCE(name, ''), email
FROM auth.users
ON CONFLICT (tenant_id, user_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email;
