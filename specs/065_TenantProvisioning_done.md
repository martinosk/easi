# 065 - Tenant Provisioning

**Depends on:** None (first slice)

## Description
Minimum viable tenant provisioning. Platform admin can create a new tenant with OIDC configuration, email domain, and first admin invitation.

OIDC credentials (client secrets or certificates) are stored in AWS Secrets Manager and synced to Kubernetes via External Secrets Operator. The database only stores a reference to the secret location.

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
    "authMethod": "private_key_jwt",
    "scopes": "openid email profile offline_access"
  },
  "firstAdminEmail": "john.doe@acme.com"
}
```

**OIDC Auth Methods:**
| Method | Description | Secret Required in Vault |
|--------|-------------|-------------------------|
| `client_secret` | Traditional client secret authentication | `client_secret` property |
| `private_key_jwt` | Certificate-based authentication (RFC 7523) | `private_key` and `certificate` properties |

**Behavior:**
- Validates tenant ID, name, domains, OIDC config
- Creates tenant, domains, OIDC config (no secrets stored in DB)
- Creates first admin invitation (status=pending, role=admin)
- Returns 201 Created with Location header
- Includes warning if OIDC secret is not yet provisioned in vault

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
    "authMethod": "private_key_jwt",
    "scopes": "openid email profile offline_access",
    "secretProvisioned": false
  },
  "createdAt": "2025-12-02T10:00:00Z",
  "_links": {
    "self": "/api/platform/v1/tenants/acme",
    "domains": "/api/platform/v1/tenants/acme/domains",
    "oidcConfig": "/api/platform/v1/tenants/acme/oidc-config",
    "suspend": "/api/platform/v1/tenants/acme/suspend",
    "users": "/api/v1/users?tenant=acme"
  },
  "_warnings": [
    "OIDC secret not provisioned. Users cannot authenticate until secret is created at: easi/tenants/acme/oidc"
  ]
}
```

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | Invalid tenant ID | ID doesn't match pattern `^[a-z0-9-]{3,50}$` |
| 400 | Invalid domain format | Domain format invalid |
| 400 | Invalid auth method | Must be `client_secret` or `private_key_jwt` |
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
Returns tenant details including OIDC configuration (no secrets).

**Response (200):** Same as POST response (includes `secretProvisioned` status)
**Errors:** 404 if not found

## Secret Provisioning (Manual)

OIDC credentials are stored in AWS Secrets Manager and synced to Kubernetes via External Secrets Operator. After creating a tenant, the platform admin must manually provision the OIDC secret.

### Secret Location Convention
```
easi/tenants/{tenant-id}/oidc
```

### Secret Structure (JSON)

**For `client_secret` auth method:**
```json
{
  "auth_method": "client_secret",
  "client_secret": "your-client-secret-here",
  "private_key": null,
  "certificate": null
}
```

**For `private_key_jwt` auth method:**
```json
{
  "auth_method": "private_key_jwt",
  "client_secret": null,
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
  "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
}
```

### AWS CLI Commands

**Create secret for client_secret auth:**
```bash
aws secretsmanager create-secret \
  --name "easi/tenants/acme-corp/oidc" \
  --secret-string '{
    "auth_method": "client_secret",
    "client_secret": "your-secret-here",
    "private_key": null,
    "certificate": null
  }'
```

**Create secret for private_key_jwt auth:**
```bash
# First, prepare the certificate and key (escape newlines for JSON)
PRIVATE_KEY=$(cat private-key.pem | awk '{printf "%s\\n", $0}')
CERTIFICATE=$(cat certificate.pem | awk '{printf "%s\\n", $0}')

aws secretsmanager create-secret \
  --name "easi/tenants/acme-corp/oidc" \
  --secret-string "{
    \"auth_method\": \"private_key_jwt\",
    \"client_secret\": null,
    \"private_key\": \"$PRIVATE_KEY\",
    \"certificate\": \"$CERTIFICATE\"
  }"
```

**Update existing secret:**
```bash
aws secretsmanager put-secret-value \
  --secret-id "easi/tenants/acme-corp/oidc" \
  --secret-string '{ ... }'
```

### Kubernetes ExternalSecret

Create one ExternalSecret per tenant, or use regex pattern for all tenants:

**Per-tenant ExternalSecret:**
```yaml
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: tenant-oidc-acme-corp
  namespace: easi
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: secrets-manager
    kind: SecretStore
  target:
    name: tenant-oidc-acme-corp
    creationPolicy: Owner
  data:
  - secretKey: auth-method
    remoteRef:
      key: "easi/tenants/acme-corp/oidc"
      property: "auth_method"
  - secretKey: client-secret
    remoteRef:
      key: "easi/tenants/acme-corp/oidc"
      property: "client_secret"
  - secretKey: private-key
    remoteRef:
      key: "easi/tenants/acme-corp/oidc"
      property: "private_key"
  - secretKey: certificate
    remoteRef:
      key: "easi/tenants/acme-corp/oidc"
      property: "certificate"
```

**Force sync after provisioning:**
```bash
kubectl annotate es tenant-oidc-acme-corp -n easi force-sync=$(date +%s) --overwrite
```

### Verifying Secret Provisioned

The application checks for secret availability by reading from the mounted K8s secret path:
- Path: `/secrets/oidc/{tenant-id}/`
- Files: `auth-method`, `client-secret`, `private-key`, `certificate`

The `secretProvisioned` field in API responses indicates whether the secret is available.

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

CREATE TABLE tenant_oidc_configs (
    tenant_id VARCHAR(50) PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    discovery_url TEXT NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    auth_method VARCHAR(20) NOT NULL DEFAULT 'client_secret',
    scopes TEXT NOT NULL DEFAULT 'openid email profile',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_auth_method CHECK (auth_method IN ('client_secret', 'private_key_jwt'))
);
-- Note: Actual secrets (client_secret, private_key, certificate) are stored in
-- AWS Secrets Manager at path: easi/tenants/{tenant_id}/oidc
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
- [x] Database migration for platform tables (tenants, tenant_domains)
- [x] Database migration for tenant-scoped tables (invitations, users, sessions)
- [x] TenantID value object with validation
- [x] TenantName value object
- [x] EmailDomain value object
- [x] Tenant aggregate
- [x] Platform admin authentication middleware (API key)
- [x] POST /api/platform/v1/tenants (create tenant with first admin invitation)
- [x] GET /api/platform/v1/tenants/{id}
- [x] GET /api/platform/v1/tenants
- [x] Integration test: create tenant and verify in database

### Security Hardening (from security review)
- [x] Update OIDCConfig to support auth_method (client_secret | private_key_jwt)
- [x] Remove client_secret from API request/database (secrets in vault only)
- [x] Database migration: replace client_secret_encrypted with auth_method
- [x] SecretProvider interface for reading OIDC credentials from mounted K8s secrets
- [x] FileSecretProvider implementation (reads from /secrets/oidc/{tenant-id}/)
- [x] Check secretProvisioned status in tenant responses
- [x] Fix SQL injection in RLS tenant context (defense in depth escaping)
- [x] Add rate limiting to platform admin API (100 requests/minute per IP)
- [x] User sign-off
