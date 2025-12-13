-- Migration: Seed test tenant for local development
-- Spec: 066_SingleTenantLogin, 067_SessionManagement
-- Description: Creates a test tenant with Dex OIDC configuration and test users for local development

INSERT INTO tenants (id, name, status)
VALUES ('acme', 'ACME Corporation', 'active')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    status = EXCLUDED.status;

INSERT INTO tenant_domains (domain, tenant_id)
VALUES ('acme.com', 'acme')
ON CONFLICT (domain) DO NOTHING;

INSERT INTO tenant_oidc_configs (tenant_id, discovery_url, issuer_url, client_id, auth_method, scopes)
VALUES (
    'acme',
    'http://dex:5556/dex',
    'http://localhost:5556/dex',
    'easi-test',
    'client_secret',
    'openid email profile offline_access'
)
ON CONFLICT (tenant_id) DO UPDATE SET
    discovery_url = EXCLUDED.discovery_url,
    issuer_url = EXCLUDED.issuer_url,
    client_id = EXCLUDED.client_id,
    auth_method = EXCLUDED.auth_method,
    scopes = EXCLUDED.scopes;

-- Seed test users (must bypass RLS for seeding)
INSERT INTO users (id, tenant_id, email, name, role, status)
VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'acme', 'testuser@acme.com', 'Test User', 'architect', 'active'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'acme', 'admin@acme.com', 'Admin User', 'admin', 'active')
ON CONFLICT (tenant_id, email) DO UPDATE SET
    name = EXCLUDED.name,
    role = EXCLUDED.role,
    status = EXCLUDED.status;
