-- Migration: 140_backfill_auth_tenant_caches
-- Purpose: Populates the Auth tenant caches from Platform's tables so that login by e-mail domain
-- and invitation domain checks are complete for tenants provisioned before spec 209. The
-- TenantCreated projector keeps them current from this point on. Idempotent.

INSERT INTO auth.tenant_cache (tenant_id, name, status)
SELECT id, name, status
FROM platform.tenants
ON CONFLICT (tenant_id) DO UPDATE SET
    name = EXCLUDED.name,
    status = EXCLUDED.status;

INSERT INTO auth.tenant_domain_cache (domain, tenant_id)
SELECT domain, tenant_id
FROM platform.tenant_domains
ON CONFLICT (domain) DO UPDATE SET
    tenant_id = EXCLUDED.tenant_id;

INSERT INTO auth.tenant_oidc_cache (tenant_id, discovery_url, issuer_url, client_id, auth_method, scopes)
SELECT tenant_id, discovery_url, issuer_url, client_id, auth_method, scopes
FROM platform.tenant_oidc_configs
ON CONFLICT (tenant_id) DO UPDATE SET
    discovery_url = EXCLUDED.discovery_url,
    issuer_url = EXCLUDED.issuer_url,
    client_id = EXCLUDED.client_id,
    auth_method = EXCLUDED.auth_method,
    scopes = EXCLUDED.scopes;
