-- Migration: 139_add_auth_tenant_caches
-- Purpose: Auth-owned caches of the tenant, e-mail-domain and OIDC configuration published by
-- Platform's TenantCreated event (spec 209). They replace Auth's query-time reads of
-- platform.tenants, platform.tenant_domains and platform.tenant_oidc_configs.
-- RLS: deliberately not enabled. These rows are resolved during login, by e-mail domain, before any
-- tenant context exists, so an app.current_tenant policy would make login impossible. This mirrors
-- how the source tables in the platform schema are protected today (migrations 039 and 103):
-- tenant-level lookup data with no RLS. Every query filters by tenant_id or domain explicitly.

CREATE TABLE IF NOT EXISTS auth.tenant_cache (
    tenant_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS auth.tenant_domain_cache (
    domain VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_tenant_domain_cache_tenant
    ON auth.tenant_domain_cache(tenant_id);

CREATE TABLE IF NOT EXISTS auth.tenant_oidc_cache (
    tenant_id VARCHAR(50) PRIMARY KEY,
    discovery_url TEXT NOT NULL,
    issuer_url TEXT,
    client_id VARCHAR(255) NOT NULL,
    auth_method VARCHAR(20) NOT NULL,
    scopes VARCHAR(255) NOT NULL
);

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON auth.tenant_cache TO easi_app';
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON auth.tenant_domain_cache TO easi_app';
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON auth.tenant_oidc_cache TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON auth.tenant_cache TO easi_admin';
        EXECUTE 'GRANT ALL PRIVILEGES ON auth.tenant_domain_cache TO easi_admin';
        EXECUTE 'GRANT ALL PRIVILEGES ON auth.tenant_oidc_cache TO easi_admin';
    END IF;
END $$;
