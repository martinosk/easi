# Authentication & Authorization

## Description
Implement multi-tenant authentication supporting any OIDC-compliant identity provider. Each tenant configures their own IdP (Azure Entra, Okta, Google Workspace, etc.). Users are discovered via email domain, authenticated against their tenant's IdP, and authorized via locally-stored roles (RBAC).

**Dependencies:** Spec 016 (Platform Administration)

## Architecture Decisions

### Multi-IdP Support
Each tenant configures their own OIDC provider. The app is IdP-agnostic and validates tokens against the tenant's configured JWKS endpoint.

### Email Domain Discovery
Users enter their email address. The app extracts the domain, looks up which tenant owns that domain, and redirects to that tenant's IdP.

### Local Role Storage (RBAC)
Roles are stored in EASI database, not in IdP claims. This means:
- Customers don't need to configure custom claims in their IdP
- Tenant admins manage roles via EASI web UI
- Simple OIDC setup for customers (standard scopes only)

### Authorization Model: RBAC with ABAC Extensibility
Start with Role-Based Access Control. Design allows future Attribute-Based policies (domain-scoped access, resource ownership) without breaking changes.

### Infrastructure vs Domain
Authentication (OIDC, sessions, JWT validation) is **infrastructure**.
User/role management is **domain** (Platform Administration bounded context - see Spec 016).

### Libraries
Use established libraries for the heavy lifting:

| Library | Purpose | License |
|---------|---------|---------|
| [coreos/go-oidc](https://github.com/coreos/go-oidc) | OIDC discovery, ID token validation, JWKS caching | Apache 2.0 |
| [alexedwards/scs](https://github.com/alexedwards/scs) | Session management with PostgreSQL store | MIT |
| [golang.org/x/oauth2](https://github.com/golang/oauth2) | OAuth2 authorization code exchange | BSD 3-Clause |

Custom code required (~200-300 lines):
- Multi-tenant OIDC provider manager (caches provider per tenant)
- Email domain → tenant lookup
- Middleware wiring

## Roles & Permissions

### Roles
| Role | Description |
|------|-------------|
| admin | Full tenant access including user management |
| architect | Read/write architecture models (future: domain-scoped) |
| stakeholder | Read-only access (future: proposals) |

### Permissions
| Permission | Admin | Architect | Stakeholder |
|------------|-------|-----------|-------------|
| components:read | ✓ | ✓ | ✓ |
| components:write | ✓ | ✓ | |
| components:delete | ✓ | | |
| views:read | ✓ | ✓ | ✓ |
| views:write | ✓ | ✓ | |
| views:delete | ✓ | | |
| capabilities:read | ✓ | ✓ | ✓ |
| capabilities:write | ✓ | ✓ | |
| capabilities:delete | ✓ | | |
| domains:read | ✓ | ✓ | ✓ |
| domains:write | ✓ | ✓ | |
| domains:delete | ✓ | | |
| users:read | ✓ | | |
| users:manage | ✓ | | |
| invitations:manage | ✓ | | |

## Login Flow

```
1. User navigates to /login
2. User enters email (e.g., john@acme.com)
3. App extracts domain "acme.com"
4. App queries tenant_domains table → finds tenant "acme"
5. App loads tenant's OIDC config from tenant_oidc_config
6. App stores tenant_id + state in session, redirects to IdP
7. User authenticates with their company IdP
8. IdP redirects to /auth/callback with authorization code
9. App exchanges code for tokens using tenant's client credentials
10. App validates ID token against tenant's JWKS
11. App looks up user by (tenant_id, external_id OR email)
12. If no user exists and no pending invitation → block with 403
13. If invitation exists → accept invitation, create user, set status=active
14. If user exists but status != active → block with 403
15. App creates session cookie with user_id
16. App redirects to dashboard
```

## User Onboarding via Invitations

User onboarding is modeled as a long-running process via the Invitation resource.

### Invitation Lifecycle
```
Admin creates invitation
        ↓
    [pending]
        │
        ├─→ User logs in via IdP → [accepted] → User created with role
        │
        ├─→ TTL expires → [expired]
        │
        └─→ Admin cancels → [revoked]
```

### Invitation States
| Status | Description |
|--------|-------------|
| pending | Invitation created, awaiting user login |
| accepted | User authenticated, invitation fulfilled |
| expired | TTL elapsed without user login |
| revoked | Admin cancelled the invitation |

### Uninvited Users
Users who authenticate via IdP without an invitation are blocked:
- 403 response: "Access denied. Contact your administrator for access."
- No user record created (invitation-only model)

## API Endpoints

### Authentication Endpoints (Infrastructure)

**POST /auth/sessions**
- Body: `{ "email": "john@acme.com" }`
- Looks up tenant by email domain
- Returns 404 if domain not registered
- Returns 200 with authorization URL for client redirect
- Stores state in server-side session (CSRF protection)
- Response:
```json
{
  "authorizationUrl": "https://login.microsoftonline.com/...",
  "_links": {
    "authorize": "https://login.microsoftonline.com/..."
  }
}
```

**GET /auth/callback**
- Query params: code, state
- Validates state matches session
- Exchanges code for tokens at tenant's token endpoint
- Validates ID token signature against tenant's JWKS
- Processes user login (check invitation or existing user)
- Sets HTTP-only session cookie
- Redirects to frontend (or returns error page)

**DELETE /auth/sessions/current**
- Clears session cookie
- Returns 204 No Content

**GET /auth/sessions/current**
- Returns current session with user identity
- Returns 401 if not authenticated
- Response:
```json
{
  "id": "session-uuid",
  "user": {
    "id": "user-uuid",
    "email": "john@acme.com",
    "name": "John Doe",
    "role": "architect",
    "permissions": ["components:read", "components:write", ...]
  },
  "tenant": {
    "id": "acme",
    "name": "Acme Corporation"
  },
  "expiresAt": "2025-12-02T12:00:00Z",
  "_links": {
    "self": "/auth/sessions/current",
    "logout": "/auth/sessions/current",
    "user": "/api/v1/users/{user-id}",
    "tenant": "/api/v1/tenants/current"
  }
}
```

### Invitation Endpoints (Admin Only)

**POST /api/v1/invitations**
- Body: `{ "email": "jane@acme.com", "role": "architect" }`
- Creates invitation with status=pending
- Email must match tenant's registered domains
- Returns 409 if active user or pending invitation already exists
- Returns 201 Created with invitation resource
- Response:
```json
{
  "id": "invitation-uuid",
  "email": "jane@acme.com",
  "role": "architect",
  "status": "pending",
  "invitedBy": {
    "id": "admin-uuid",
    "email": "admin@acme.com"
  },
  "createdAt": "2025-12-02T10:00:00Z",
  "expiresAt": "2025-12-09T10:00:00Z",
  "_links": {
    "self": "/api/v1/invitations/{id}",
    "revoke": "/api/v1/invitations/{id}/revoke"
  }
}
```

**GET /api/v1/invitations**
- Returns paginated list of invitations
- Filter by status: ?status=pending
- Response uses RespondPaginated with _links

**GET /api/v1/invitations/{id}**
- Returns invitation details
- Returns 404 if not found

**POST /api/v1/invitations/{id}/revoke**
- Revokes a pending invitation
- Returns 409 if invitation not in pending status
- Returns 200 OK with updated invitation resource

### User Endpoints (Admin Only)

**GET /api/v1/users**
- Returns paginated list of users in tenant
- Filter by status: ?status=active, ?role=architect
- Response uses RespondPaginated with _links
- Each user includes _links based on current state

**GET /api/v1/users/{id}**
- Returns user details
- Returns 404 if not found
- Response:
```json
{
  "id": "user-uuid",
  "email": "jane@acme.com",
  "name": "Jane Doe",
  "role": "architect",
  "status": "active",
  "invitedBy": {
    "id": "admin-uuid",
    "email": "admin@acme.com"
  },
  "createdAt": "2025-12-01T10:00:00Z",
  "lastLoginAt": "2025-12-02T08:30:00Z",
  "_links": {
    "self": "/api/v1/users/{id}",
    "changeRole": "/api/v1/users/{id}/change-role",
    "disable": "/api/v1/users/{id}/disable"
  }
}
```

**POST /api/v1/users/{id}/change-role**
- Body: `{ "role": "admin" }`
- Changes user's role
- Cannot demote last admin (returns 409)
- Returns 200 OK with updated user resource

**POST /api/v1/users/{id}/disable**
- Disables user account
- Cannot disable self (returns 409)
- Cannot disable last admin (returns 409)
- Returns 200 OK with updated user resource

**POST /api/v1/users/{id}/enable**
- Re-enables disabled user account
- Returns 200 OK with updated user resource

### Tenant Endpoints (All Authenticated Users)

**GET /api/v1/tenants/current**
- Returns current tenant info (read-only for non-admins)
- Response:
```json
{
  "id": "acme",
  "name": "Acme Corporation",
  "domains": ["acme.com", "acme.co.uk"],
  "_links": {
    "self": "/api/v1/tenants/current",
    "users": "/api/v1/users",
    "invitations": "/api/v1/invitations"
  }
}
```

## Middleware Architecture

### Request Flow
```
Request
   ↓
CORS Middleware (existing)
   ↓
Session Middleware (NEW)
   - Load session from cookie
   - If valid: load user from DB, inject UserIdentity into context
   - If LOCAL_DEV_MODE: create mock identity from headers
   - Public endpoints (/auth/*, /health): skip
   ↓
Tenant Middleware (MODIFIED)
   - Extract tenant from UserIdentity
   - Set RLS session variable
   ↓
Authorization Middleware (NEW)
   - Check user.status == "active"
   - Check required permission for route
   - Return 403 if denied
   ↓
Audit Middleware (NEW)
   - Log request with user context
   ↓
Handler
```

### UserIdentity Value Object (Infrastructure)

```go
type UserIdentity struct {
    UserID      uuid.UUID
    Email       string
    Name        string
    TenantID    valueobjects.TenantID
    Role        Role
    Permissions []Permission
    Status      UserStatus
}
```

### Route Protection Pattern

```go
r.Route("/api/v1/components", func(r chi.Router) {
    r.Use(middleware.RequirePermission(PermComponentsRead))
    r.Get("/", handlers.GetAll)
    r.Get("/{id}", handlers.GetByID)

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequirePermission(PermComponentsWrite))
        r.Post("/", handlers.Create)
        r.Put("/{id}", handlers.Update)
    })

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequirePermission(PermComponentsDelete))
        r.Delete("/{id}", handlers.Delete)
    })
})

r.Route("/api/v1/invitations", func(r chi.Router) {
    r.Use(middleware.RequirePermission(PermInvitationsManage))
    r.Post("/", invitationHandlers.Create)
    r.Get("/", invitationHandlers.List)
    r.Get("/{id}", invitationHandlers.Get)
    r.Post("/{id}/revoke", invitationHandlers.Revoke)
})

r.Route("/api/v1/users", func(r chi.Router) {
    r.Use(middleware.RequirePermission(PermUsersRead))
    r.Get("/", userHandlers.List)
    r.Get("/{id}", userHandlers.Get)

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequirePermission(PermUsersManage))
        r.Post("/{id}/change-role", userHandlers.ChangeRole)
        r.Post("/{id}/disable", userHandlers.Disable)
        r.Post("/{id}/enable", userHandlers.Enable)
    })
})
```

## Local Development Mode

When `LOCAL_DEV_MODE=true`:
- No OIDC authentication required
- Accept headers for mock identity:
  - `X-Tenant-ID`: Tenant ID (default: "default")
  - `X-Dev-Role`: Role (default: "architect")
  - `X-Dev-Email`: Email (default: "developer@localhost")
  - `X-Dev-User-ID`: User ID (default: generated UUID)
- Session middleware creates mock UserIdentity from headers
- All OIDC validation skipped

```bash
curl -H "X-Tenant-ID: acme" \
     -H "X-Dev-Role: admin" \
     http://localhost:8080/api/v1/users
```

## Session Management

### Session Storage (SCS)
```go
import "github.com/alexedwards/scs/v2"
import "github.com/alexedwards/scs/postgresstore"

sessionManager := scs.New()
sessionManager.Store = postgresstore.New(db)
sessionManager.Lifetime = 24 * time.Hour
sessionManager.Cookie.Name = "easi_session"
sessionManager.Cookie.HttpOnly = true
sessionManager.Cookie.Secure = true  // HTTPS only in production
sessionManager.Cookie.SameSite = http.SameSiteLaxMode
```

### OIDC Provider (Single Tenant - Environment Variables)

For now (single tenant), OIDC config comes from environment variables:

```go
import (
    "github.com/coreos/go-oidc/v3/oidc"
    "golang.org/x/oauth2"
    "os"
)

type OIDCConfig struct {
    Provider     *oidc.Provider
    OAuth2Config *oauth2.Config
    Verifier     *oidc.IDTokenVerifier
}

func NewOIDCConfigFromEnv(ctx context.Context) (*OIDCConfig, error) {
    discoveryURL := os.Getenv("OIDC_DISCOVERY_URL")
    clientID := os.Getenv("OIDC_CLIENT_ID")
    clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
    redirectURL := os.Getenv("OIDC_REDIRECT_URL")

    // Create OIDC provider (handles discovery + JWKS caching)
    provider, err := oidc.NewProvider(ctx, discoveryURL)
    if err != nil {
        return nil, fmt.Errorf("OIDC discovery failed: %w", err)
    }

    return &OIDCConfig{
        Provider: provider,
        OAuth2Config: &oauth2.Config{
            ClientID:     clientID,
            ClientSecret: clientSecret,
            Endpoint:     provider.Endpoint(),
            RedirectURL:  redirectURL,
            Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
        },
        Verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
    }, nil
}
```

### Future: Multi-Tenant OIDC Provider Manager

When multi-tenant onboarding is needed, wrap this with per-tenant caching:

```go
type OIDCProviderManager struct {
    tenantRepo  TenantRepository
    providers   sync.Map  // tenantID → *OIDCConfig
    callbackURL string
}

func (m *OIDCProviderManager) GetProvider(ctx context.Context, tenantID string) (*OIDCConfig, error) {
    // Check cache, load from DB if not found, create provider
}
```

### Token Handling
- ID token validated once at login using `provider.Verifier().Verify(ctx, rawIDToken)`
- Access token not needed (app doesn't call IdP APIs)
- Refresh token not used (re-authenticate on session expiry)

## Audit Logging

### Events to Log
| Event | When |
|-------|------|
| AUTH_SESSION_INITIATED | User submits email |
| AUTH_SESSION_CREATED | User authenticated and session created |
| AUTH_SESSION_BLOCKED | User authenticated but no invitation/inactive |
| AUTH_SESSION_FAILED | OIDC validation failed |
| AUTH_SESSION_ENDED | User logged out |
| INVITATION_CREATED | Admin created invitation |
| INVITATION_ACCEPTED | User accepted invitation |
| INVITATION_REVOKED | Admin revoked invitation |
| INVITATION_EXPIRED | Invitation TTL elapsed |
| USER_ROLE_CHANGED | Admin changed user role |
| USER_DISABLED | Admin disabled user |
| USER_ENABLED | Admin enabled user |
| AUTHZ_DENIED | User attempted unauthorized action |

### Audit Record Fields
- timestamp
- event_type
- tenant_id
- user_id (if known)
- user_email
- ip_address
- user_agent
- details (JSON)

## Security Considerations

### Token Validation
- Validate signature against tenant's JWKS (cache keys, refresh periodically)
- Validate issuer matches tenant's OIDC discovery
- Validate audience if configured
- Validate expiration with 5-minute clock skew tolerance
- Validate nonce matches session state

### Session Security
- HTTP-only cookie (no JavaScript access)
- Secure flag (HTTPS only in production)
- SameSite=Lax (CSRF protection)
- Regenerate session ID on login

### Tenant Isolation
- Users can only authenticate to their email domain's tenant
- RLS enforces database-level isolation
- Cross-tenant access returns 404 (not 403)

## Frontend Integration

### Login Component
1. Show email input form
2. POST /auth/sessions with email
3. Receive authorization URL
4. Redirect to IdP
5. After callback, check GET /auth/sessions/current
6. If authenticated, load app
7. If 401/403, show appropriate message

### Session Handling
- Check /auth/sessions/current on app load
- If 401, redirect to login
- Store user info in React context
- Show role-appropriate UI elements

### Invitation Management Component (Admin)
- List pending/accepted/expired invitations
- Create new invitation modal
- Revoke pending invitations

### User Management Component (Admin)
- List users with status badges
- Change role actions
- Disable/enable user actions

## Testing Strategy

### Unit Tests
- Role/permission mapping
- UserIdentity creation
- JWT validation logic (mock JWKS)
- Email domain extraction
- Invitation state transitions

### Integration Tests
- Login flow with mock OIDC provider
- Session creation and validation
- Invitation lifecycle (create → accept)
- Invitation expiration
- Permission checks on protected routes
- Tenant isolation (user A can't see tenant B)

### Security Tests
- Invalid token signatures rejected
- Expired tokens rejected
- Cross-tenant access prevented
- Session fixation prevented
- Uninvited users blocked

## Checklist

### Phase 1: Infrastructure
- [ ] Add dependencies: coreos/go-oidc, alexedwards/scs, golang.org/x/oauth2
- [ ] OIDC config from environment variables (K8s secret)
- [ ] Session manager with SCS PostgreSQL store
- [ ] Session middleware with LOCAL_DEV_MODE support
- [ ] Authorization middleware (RequirePermission)
- [ ] Audit logging middleware

### Phase 2: Authentication Flow
- [ ] POST /auth/sessions endpoint
- [ ] GET /auth/callback endpoint
- [ ] DELETE /auth/sessions/current endpoint
- [ ] GET /auth/sessions/current endpoint
- [ ] Login flow with invitation checking

### Phase 3: Invitation Management
- [ ] POST /api/v1/invitations endpoint
- [ ] GET /api/v1/invitations endpoint
- [ ] GET /api/v1/invitations/{id} endpoint
- [ ] POST /api/v1/invitations/{id}/revoke endpoint
- [ ] Invitation expiration handling (background job or lazy)

### Phase 4: User Management
- [ ] GET /api/v1/users endpoint
- [ ] GET /api/v1/users/{id} endpoint
- [ ] POST /api/v1/users/{id}/change-role endpoint
- [ ] POST /api/v1/users/{id}/disable endpoint
- [ ] POST /api/v1/users/{id}/enable endpoint
- [ ] GET /api/v1/tenants/current endpoint

### Phase 5: Frontend
- [ ] Login page with email input
- [ ] OIDC redirect handling
- [ ] Session check on app load
- [ ] User context provider
- [ ] Invitation management page (admin)
- [ ] User management page (admin)
- [ ] Role-based UI visibility

### Phase 6: Testing & Security
- [ ] Unit tests for authorization logic
- [ ] Integration tests with mock OIDC
- [ ] Security tests for token validation
- [ ] Cross-tenant isolation tests
- [ ] E2E tests for login flow

### Final
- [ ] Security review
- [ ] OpenAPI specification
- [ ] User sign-off
