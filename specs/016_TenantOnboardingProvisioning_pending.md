# Platform Administration

## Description
Implement platform administration for tenant lifecycle management. This is a **Supporting Domain** bounded context that handles tenant provisioning, OIDC configuration, and platform-level operations. Tenants are provisioned by the platform admin (SaaS owner) via API.

**Dependencies:** Spec 013 (Multi-Tenancy Infrastructure)

## Architecture Decisions

### Supporting Domain (Not Event Sourced)
Platform Administration is a supporting domain with simple CRUD-like operations. No event sourcing needed:
- Simple state transitions (active → suspended → archived)
- No temporal query requirements
- Audit logging is sufficient for compliance
- Simple tables with foreign keys

### Platform Admin API
REST API for platform administration:
- Enables programmatic access, integrations, automation
- Future: Self-service tenant registration portal

### Tenant as Aggregate Root
Tenant is the aggregate root containing:
- OIDC configuration (value object)
- Email domains (entities)
- Invariants enforced at aggregate level

### Separation from Tenant Users
Platform admins (SaaS operators) are separate from tenant users:
- Platform admins access the Platform Admin API
- Tenant users access the main application API
- Different authentication mechanisms

## Domain Model

### Tenant Aggregate

```go
type Tenant struct {
    ID             TenantID
    Name           TenantName
    Status         TenantStatus
    Domains        []EmailDomain
    OIDCConfig     OIDCConfiguration
    CreatedAt      time.Time
    UpdatedAt      time.Time
    SuspendedAt    *time.Time
    SuspendedReason *string
}

type TenantStatus string
const (
    TenantStatusActive    TenantStatus = "active"
    TenantStatusSuspended TenantStatus = "suspended"
    TenantStatusArchived  TenantStatus = "archived"
)
```

### Value Objects

```go
type TenantID struct {
    value string  // Pattern: ^[a-z0-9-]{3,50}$
}

type TenantName struct {
    value string  // 1-255 characters
}

type EmailDomain struct {
    value string  // Valid domain format
}

type OIDCConfiguration struct {
    DiscoveryURL    URL
    ClientID        ClientID
    ClientSecret    EncryptedSecret
    Scopes          Scopes
}
```

### Invariants
- Tenant ID is immutable after creation
- Tenant ID must match pattern `^[a-z0-9-]{3,50}$`
- Tenant ID cannot be reserved words (system, admin, root, default)
- At least one email domain must be registered
- Email domains must be globally unique (no two tenants share a domain)
- OIDC discovery URL must be valid and reachable
- Cannot archive tenant with active users (must disable all first)

## Platform Admin API

All endpoints require Platform Admin authentication via API key.

### Authentication
```
Header: X-Platform-Admin-Key: <api-key>
```

API keys are configured via environment variable and validated by middleware.

### Tenant Endpoints

**POST /api/platform/v1/tenants**
- Creates new tenant with OIDC configuration
- Body:
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
- Validates OIDC discovery URL is reachable
- Creates tenant, domains, OIDC config
- Creates first admin invitation (status=invited, role=admin)
- Returns 201 Created with Location header
- Response:
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

**GET /api/platform/v1/tenants**
- Returns paginated list of all tenants
- Filter by status: ?status=active
- Filter by domain: ?domain=acme.com
- Response uses standard pagination

**GET /api/platform/v1/tenants/{id}**
- Returns tenant details
- Client secret is masked in response
- Returns 404 if not found

**PATCH /api/platform/v1/tenants/{id}**
- Updates tenant name
- Body: `{ "name": "Acme Corp International" }`
- Returns 200 OK with updated resource

**POST /api/platform/v1/tenants/{id}/suspend**
- Suspends tenant
- Body: `{ "reason": "Non-payment" }`
- All tenant users blocked from login
- Returns 200 OK with updated resource

**POST /api/platform/v1/tenants/{id}/activate**
- Reactivates suspended tenant
- Returns 200 OK with updated resource

**DELETE /api/platform/v1/tenants/{id}**
- Archives tenant (soft delete)
- Requires all users to be disabled first (returns 409 otherwise)
- Returns 204 No Content

### Domain Endpoints

**POST /api/platform/v1/tenants/{id}/domains**
- Adds email domain to tenant
- Body: `{ "domain": "acme.de" }`
- Returns 409 if domain already registered to another tenant
- Returns 201 Created

**DELETE /api/platform/v1/tenants/{id}/domains/{domain}**
- Removes email domain from tenant
- Cannot remove last domain (returns 409)
- Returns 204 No Content

**GET /api/platform/v1/domains**
- Returns all domain → tenant mappings
- Useful for debugging domain conflicts

### OIDC Configuration Endpoints

**PATCH /api/platform/v1/tenants/{id}/oidc-config**
- Updates OIDC configuration
- Body:
```json
{
  "discoveryUrl": "https://...",
  "clientId": "new-client-id",
  "clientSecret": "new-secret",
  "scopes": "openid email profile"
}
```
- Validates discovery URL before accepting
- Returns 200 OK with updated tenant resource

### Audit Log Endpoints

**GET /api/platform/v1/tenants/{id}/audit-log**
- Returns paginated audit log for tenant
- Filter by event type: ?eventType=TENANT_SUSPENDED
- Filter by date range: ?from=2025-01-01&to=2025-12-31

**GET /api/platform/v1/audit-log**
- Returns paginated platform-wide audit log
- Filter by tenant: ?tenantId=acme

## Database Schema

### Tables (No RLS - Platform Admin Only)

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

-- Future: tenant_oidc_config table for multi-tenant OIDC
-- For now, OIDC config comes from environment variables (K8s secrets)
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

### Audit Log Table

```sql
CREATE TABLE platform_audit_log (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    tenant_id VARCHAR(50),
    user_id UUID,
    user_email VARCHAR(255),
    performed_by VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_platform_audit_tenant ON platform_audit_log(tenant_id);
CREATE INDEX idx_platform_audit_event ON platform_audit_log(event_type);
CREATE INDEX idx_platform_audit_created ON platform_audit_log(created_at);
```

## OIDC Configuration

### Current Approach: Environment Variables
For now (single tenant), OIDC configuration is stored in Kubernetes secrets and loaded via environment variables:

```bash
# K8s Secret / Environment Variables
OIDC_DISCOVERY_URL=https://login.microsoftonline.com/{tenant}/v2.0/.well-known/openid-configuration
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_REDIRECT_URL=https://easi.example.com/auth/callback
```

This is simple and secure (K8s secrets are encrypted at rest).

### Future: Database-Based Configuration
When multi-tenant onboarding is needed, migrate to database-based OIDC config:
- Store per-tenant OIDC settings in `tenant_oidc_config` table
- Use external secrets manager (Vault, AWS Secrets Manager) for client secrets
- Create separate spec for tenant self-service onboarding

## OIDC Discovery Validation

When creating/updating OIDC config, validate using coreos/go-oidc:

```go
import "github.com/coreos/go-oidc/v3/oidc"

func validateOIDCDiscovery(ctx context.Context, discoveryURL string) error {
    // This performs full OIDC discovery and validates the response
    _, err := oidc.NewProvider(ctx, discoveryURL)
    if err != nil {
        return fmt.Errorf("OIDC discovery failed: %w", err)
    }
    return nil
}
```

The `oidc.NewProvider` function:
- Fetches `/.well-known/openid-configuration`
- Validates required fields (issuer, authorization_endpoint, token_endpoint, jwks_uri)
- Fetches and caches JWKS for token validation

## Audit Events

| Event | Details |
|-------|---------|
| TENANT_CREATED | id, name, domains, first_admin_email |
| TENANT_UPDATED | id, changes |
| TENANT_SUSPENDED | id, reason |
| TENANT_ACTIVATED | id |
| TENANT_ARCHIVED | id |
| DOMAIN_ADDED | tenant_id, domain |
| DOMAIN_REMOVED | tenant_id, domain |
| OIDC_CONFIG_UPDATED | tenant_id (no secrets logged) |

## Customer Onboarding Documentation

Provide customers with setup instructions for their IdP:

### Azure Entra (Azure AD)
1. Register new application in Azure Portal → App registrations
2. Set redirect URI: `https://easi.example.com/auth/callback`
3. Go to "Certificates & secrets" → Create client secret
4. Note: Application (client) ID from Overview page
5. Note: Directory (tenant) ID from Overview page
6. Discovery URL: `https://login.microsoftonline.com/{tenant-id}/v2.0/.well-known/openid-configuration`

### Okta
1. Admin Console → Applications → Create App Integration
2. Select "OIDC - OpenID Connect" and "Web Application"
3. Set redirect URI: `https://easi.example.com/auth/callback`
4. Note: Client ID and Client Secret from General tab
5. Discovery URL: `https://{your-domain}.okta.com/.well-known/openid-configuration`

### Google Workspace
1. Google Cloud Console → APIs & Services → Credentials
2. Create OAuth 2.0 Client ID (Web application)
3. Set redirect URI: `https://easi.example.com/auth/callback`
4. Configure OAuth consent screen with internal user type
5. Discovery URL: `https://accounts.google.com/.well-known/openid-configuration`

### Generic OIDC Provider
1. Create OIDC client/application in your IdP
2. Set redirect URI: `https://easi.example.com/auth/callback`
3. Required scopes: `openid email profile`
4. Provide to EASI admin: Discovery URL, Client ID, Client Secret

## Error Handling

### API Errors
| Status | Error | Description |
|--------|-------|-------------|
| 400 | Invalid tenant ID | ID doesn't match pattern |
| 400 | Invalid domain format | Domain format invalid |
| 400 | Invalid OIDC config | Discovery URL unreachable or invalid |
| 404 | Tenant not found | Tenant ID doesn't exist |
| 409 | Tenant already exists | Duplicate tenant ID |
| 409 | Domain already registered | Domain belongs to another tenant |
| 409 | Cannot remove last domain | At least one domain required |
| 409 | Cannot archive with active users | Disable users first |

## Directory Structure

```
backend/internal/
├── platformadmin/           # Platform Administration bounded context
│   ├── domain/
│   │   ├── tenant.go        # Tenant aggregate
│   │   └── valueobjects/
│   │       ├── tenant_id.go
│   │       ├── tenant_name.go
│   │       └── email_domain.go
│   ├── application/
│   │   ├── tenant_service.go
│   │   └── audit_service.go
│   └── infrastructure/
│       ├── api/
│       │   ├── handlers.go
│       │   └── routes.go
│       └── repositories/
│           ├── tenant_repository.go
│           └── audit_repository.go
```

## Checklist

### Phase 1: Domain & Database
- [ ] TenantID value object with validation
- [ ] TenantName value object
- [ ] EmailDomain value object
- [ ] Tenant aggregate
- [ ] Database migration for platform tables (tenants, tenant_domains)
- [ ] Database migration for tenant-scoped tables (invitations, users, sessions)
- [ ] OIDC config from environment variables

### Phase 2: Platform Admin API
- [ ] Platform admin authentication middleware (API key)
- [ ] POST /api/platform/v1/tenants
- [ ] GET /api/platform/v1/tenants
- [ ] GET /api/platform/v1/tenants/{id}
- [ ] PATCH /api/platform/v1/tenants/{id}
- [ ] POST /api/platform/v1/tenants/{id}/suspend
- [ ] POST /api/platform/v1/tenants/{id}/activate
- [ ] DELETE /api/platform/v1/tenants/{id}
- [ ] POST /api/platform/v1/tenants/{id}/domains
- [ ] DELETE /api/platform/v1/tenants/{id}/domains/{domain}
- [ ] GET /api/platform/v1/domains
- [ ] Audit logging for all operations

### Phase 3: Documentation & Testing
- [ ] Customer onboarding guide (Azure Entra)
- [ ] Customer onboarding guide (Okta)
- [ ] Customer onboarding guide (Google Workspace)
- [ ] Generic OIDC setup guide
- [ ] Unit tests for value objects
- [ ] Integration tests for API endpoints

### Final
- [ ] OpenAPI specification for Platform Admin API
- [ ] Security review
- [ ] User sign-off
