# Authentication & Authorization

## Description
Implement OAuth 2.0 with OpenID Connect (OIDC) to securely map users to tenants. Tenant ID is extracted from OAuth token claims and injected into request context for tenant scoping.

**Dependencies:** Spec 013 (Multi-Tenancy Infrastructure)

## Core Requirements

### OAuth Provider Integration
- Support Auth0, Okta, Google OAuth, GitHub OAuth, or custom OIDC provider
- Configure via environment variables: OAUTH_PROVIDER, OAUTH_CLIENT_ID, OAUTH_CLIENT_SECRET, OAUTH_DISCOVERY_URL
- Standard OIDC authorization code flow

### Token Structure
**Required Claims:**
- iss: Token issuer
- sub: User ID
- aud: API audience
- exp: Token expiration
- org: Organization/Tenant ID (maps to TenantId)
- email: User email

**Permissions/Scopes:**
- components:read, components:write, components:delete
- tenants:admin (future)
- users:manage (future)

### API Endpoints

**GET /auth/login**
- Initiates OAuth login flow
- Redirects to OAuth provider authorization endpoint
- Optional redirect_uri parameter (must be whitelisted)

**GET /auth/callback**
- OAuth provider callback endpoint
- Validates CSRF state token
- Exchanges authorization code for access token
- Extracts tenant ID from token claims
- Creates secure HTTP-only session cookie
- Logs authentication event
- Redirects to dashboard

**GET /auth/logout**
- Clears session cookie
- Optionally revokes refresh token at provider
- Logs logout event
- Redirects to homepage

**GET /auth/me**
- Returns current user information from token
- Includes: user ID, email, display name, tenant ID, permissions
- Returns 401 if not authenticated

### Token Validation Middleware
- All API endpoints (except /auth/* and /health) require valid Bearer token
- Extract token from Authorization header
- Validate signature, issuer, audience, expiration
- Extract tenant ID from "org" claim
- Check permissions for requested operation
- Inject user ID and tenant context into request
- Set PostgreSQL session variable for RLS: `SET app.current_tenant = $1` (see Spec 013)
- Return 401 for invalid/expired tokens
- Return 403 for insufficient permissions
- Audit log all validation failures

### Development Mode
**Configuration:**
- LOCAL_DEV_MODE=true and SKIP_OAUTH=true
- No OAuth authentication required
- Accept X-Tenant-ID header directly
- Set PostgreSQL session variable from X-Tenant-ID header
- /auth/me returns mock user
- MUST be disabled in production

### Tenant Mapping
**Single Tenant per User (implement now):**
- User belongs to one organization
- Extract "org" claim as tenant ID
- Simple authorization

**Multi-Tenant User (future):**
- Support "orgs" array and "current_org" claim
- Allow tenant switching in UI

### Permission Model (RBAC)
**Viewer:** components:read
**Editor:** components:read, components:write
**Admin:** components:read, components:write, components:delete, tenants:admin, users:manage

### Audit Logging
Log all authentication events with: timestamp, event type, user ID, tenant ID, email, IP address, user agent, error reason, duration

**Events:** LOGIN_SUCCESS, LOGOUT, TOKEN_VALIDATION_FAILED, PERMISSION_DENIED, TOKEN_REFRESH_FAILED, SUSPICIOUS_ACTIVITY

### Security
- Validate token signature with provider's public key
- Enforce token expiration with grace period for clock skew
- CSRF protection on callback endpoint
- Secure, HTTP-only session cookies
- Verify user has access to requested tenant
- Return 403 if user attempts to access different tenant's resources
- Prevent LOCAL_DEV_MODE in production (check both flags don't conflict)

## Checklist
- [ ] OAuth provider account configured
- [ ] OIDC discovery endpoints integrated
- [ ] Bearer token middleware with JWT validation
- [ ] Login/callback/logout endpoints
- [ ] Token claims extraction and tenant mapping
- [ ] Permission checking middleware
- [ ] Audit logging for all auth events
- [ ] Session management with secure cookies
- [ ] LOCAL_DEV_MODE for development
- [ ] Frontend OAuth integration
- [ ] Token refresh logic
- [ ] Unit tests for token validation
- [ ] Integration tests with mock OAuth provider
- [ ] E2E tests for login/logout flow
- [ ] Security tests for token attacks
- [ ] User sign-off
