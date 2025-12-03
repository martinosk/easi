# 065 - Tenant Provisioning

**Depends on:** None (first slice)

## Description
Minimum viable tenant provisioning. Platform admin can create a new tenant with OIDC configuration, email domain, and first admin invitation.

## API Endpoints

### Authentication
```
Header: X-Platform-Admin-Key: <api-key>
```

**API Key Management:**
- API key configured via `PLATFORM_ADMIN_API_KEY` environment variable
- Stored in K8s secret alongside OIDC credentials
- Validated by middleware on every request
- Returns 401 Unauthorized if missing or invalid

### POST /api/platform/v1/tenants
Creates new tenant with OIDC configuration.

**Request:**
```json
{
  "id": "acme",
  "name": "Acme Corporation",
  "domains": ["acme.com", "acme.co.uk"],
  "oidcConfig": {
    "discoveryUrl": "https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
    "clientId": "client-id-here",
    "clientSecret": "client-secret-here",
    "scopes": "openid email profile"
  },
  "firstAdminEmail": "john.doe@acme.com"
}
```

**Behavior:**
- Validates OIDC discovery URL is reachable
- Creates tenant, domains, OIDC config
- Creates first admin invitation (status=pending, role=admin)
- Returns 201 Created with Location header

**Response (201):**
```json
{
  "id": "acme",
  "name": "Acme Corporation",
  "status": "active",
  "domains": ["acme.com", "acme.co.uk"],
  "oidcConfig": {
    "discoveryUrl": "...",
    "clientId": "...",
    "scopes": "openid email profile"
  },
  "createdAt": "2025-12-02T10:00:00Z",
  "_links": {
    "self": "/api/platform/v1/tenants/acme",
    "domains": "/api/platform/v1/tenants/acme/domains",
    "oidcConfig": "/api/platform/v1/tenants/acme/oidc-config",
    "suspend": "/api/platform/v1/tenants/acme/suspend",
    "users": "/api/v1/users?tenant=acme"
  }
}
```

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | Invalid tenant ID | ID doesn't match pattern `^[a-z0-9-]{3,50}$` |
| 400 | Invalid domain format | Domain format invalid |
| 400 | Invalid OIDC config | Discovery URL unreachable or invalid |
| 409 | Tenant already exists | Duplicate tenant ID |
| 409 | Domain already registered | Domain belongs to another tenant |

### GET /api/platform/v1/tenants
Returns paginated list of all tenants.

**Query Parameters:**
- `status`: Filter by status (active, suspended, archived)
- `domain`: Filter by domain

**Response (200):**
```json
{
  "data": [
    {
      "id": "acme",
      "name": "Acme Corporation",
      "status": "active",
      "domains": ["acme.com"],
      "createdAt": "2025-12-02T10:00:00Z",
      "_links": { "self": "/api/platform/v1/tenants/acme" }
    }
  ],
  "pagination": { "hasMore": false, "limit": 50 },
  "_links": { "self": "/api/platform/v1/tenants" }
}
```

### GET /api/platform/v1/tenants/{id}
Returns tenant details. Client secret is masked in response.

**Response (200):** Same as POST response
**Errors:** 404 if not found

## Database Schema

```sql
CREATE TABLE tenants (
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

CREATE TABLE tenant_domains (
    domain VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_domain_format CHECK (domain ~ '^[a-z0-9][a-z0-9.-]*[a-z0-9]$')
);

CREATE INDEX idx_tenant_domains_tenant ON tenant_domains(tenant_id);
```

### Tenant-Scoped Tables (With RLS)

```sql
CREATE TABLE invitations (
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

CREATE INDEX idx_invitations_tenant ON invitations(tenant_id);
CREATE INDEX idx_invitations_email ON invitations(tenant_id, email);
CREATE INDEX idx_invitations_status ON invitations(tenant_id, status);

ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON invitations
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE users (
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

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_external ON users(tenant_id, external_id);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON users
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    user_agent TEXT,
    ip_address VARCHAR(45)
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON sessions
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
```

## Checklist
- [ ] Database migration for platform tables (tenants, tenant_domains)
- [ ] Database migration for tenant-scoped tables (invitations, users, sessions)
- [ ] TenantID value object with validation
- [ ] TenantName value object
- [ ] EmailDomain value object
- [ ] Tenant aggregate
- [ ] Platform admin authentication middleware (API key)
- [ ] POST /api/platform/v1/tenants (create tenant with first admin invitation)
- [ ] GET /api/platform/v1/tenants/{id}
- [ ] GET /api/platform/v1/tenants
- [ ] Integration test: create tenant and verify in database
- [ ] User sign-off
