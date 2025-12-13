-- Migration: Tenant Provisioning
-- Spec: 065_TenantProvisioning_pending.md
-- Description: Creates all tables required for multi-tenant provisioning

-- ============================================================================
-- Platform Tables (no RLS - managed by platform admin API)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    suspended_at TIMESTAMP,
    suspended_reason TEXT,

    CONSTRAINT chk_tenant_id CHECK (id ~ '^[a-z0-9-]{3,50}$'),
    CONSTRAINT chk_tenant_status CHECK (status IN ('active', 'suspended', 'archived'))
);

CREATE TABLE IF NOT EXISTS tenant_domains (
    domain VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_domain_format CHECK (domain ~ '^[a-z0-9][a-z0-9.-]*[a-z0-9]$')
);

CREATE INDEX IF NOT EXISTS idx_tenant_domains_tenant ON tenant_domains(tenant_id);

CREATE TABLE IF NOT EXISTS tenant_oidc_configs (
    tenant_id VARCHAR(50) PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    discovery_url TEXT NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    auth_method VARCHAR(20) NOT NULL DEFAULT 'client_secret',
    scopes VARCHAR(255) NOT NULL DEFAULT 'openid email profile',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_auth_method CHECK (auth_method IN ('client_secret', 'private_key_jwt'))
);

-- ============================================================================
-- Tenant-Scoped Tables (with RLS)
-- ============================================================================

CREATE TABLE IF NOT EXISTS invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id),
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    invited_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    revoked_at TIMESTAMP,

    CONSTRAINT chk_invitation_status CHECK (status IN ('pending', 'accepted', 'expired', 'revoked')),
    CONSTRAINT chk_invitation_role CHECK (role IN ('admin', 'architect', 'stakeholder'))
);

CREATE INDEX IF NOT EXISTS idx_invitations_tenant ON invitations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_invitations_status ON invitations(tenant_id, status);

ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON invitations;
CREATE POLICY tenant_isolation_policy ON invitations
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id),
    external_id VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    role VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    invitation_id UUID REFERENCES invitations(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,

    CONSTRAINT chk_user_role CHECK (role IN ('admin', 'architect', 'stakeholder')),
    CONSTRAINT chk_user_status CHECK (status IN ('active', 'disabled')),
    UNIQUE(tenant_id, email)
);

CREATE INDEX IF NOT EXISTS idx_users_tenant ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_external ON users(tenant_id, external_id);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON users;
CREATE POLICY tenant_isolation_policy ON users
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    user_agent TEXT,
    ip_address VARCHAR(45)
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON sessions;
CREATE POLICY tenant_isolation_policy ON sessions
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
