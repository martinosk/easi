# 066 - Single-Tenant Login with Dev Mode

**Depends on:** [065_TenantProvisioning](065_TenantProvisioning_pending.md) (database schema)

## Description
Complete end-to-end login flow for single tenant. User enters email, is redirected to IdP, authenticates, and receives a session. Includes LOCAL_DEV_MODE for development without IdP.

## Libraries
| Library | Purpose | License |
|---------|---------|---------|
| [coreos/go-oidc](https://github.com/coreos/go-oidc) | OIDC discovery, ID token validation, JWKS caching | Apache 2.0 |
| [alexedwards/scs](https://github.com/alexedwards/scs) | Session management with PostgreSQL store | MIT |
| [golang.org/x/oauth2](https://github.com/golang/oauth2) | OAuth2 authorization code exchange | BSD 3-Clause |

## Login Flow
```
1. User navigates to /login
2. User enters email (e.g., john@acme.com)
3. App extracts domain "acme.com"
4. App queries tenant_domains table → finds tenant "acme"
5. App loads OIDC config from environment variables
6. App stores tenant_id + state in session, redirects to IdP
7. User authenticates with their company IdP
8. IdP redirects to /auth/callback with authorization code
9. App exchanges code for tokens using client credentials
10. App validates ID token against JWKS
11. App looks up user by (tenant_id, external_id OR email)
12. If no user exists and no pending invitation → block with 403
13. If invitation exists → accept invitation, create user, set status=active
14. If user exists but status != active → block with 403
15. App creates session cookie with user_id
16. App redirects to dashboard
```

## API Endpoints

### POST /auth/sessions
Initiates login flow by looking up tenant and generating authorization URL.

**Request:**
```json
{
  "email": "john@acme.com"
}
```

**Behavior:**
- Extracts domain from email
- Looks up tenant by domain in `tenant_domains` table
- Generates OIDC authorization URL with state parameter
- Stores tenant_id and state in session

**Response (200):**
```json
{
  "authorizationUrl": "https://login.microsoftonline.com/...",
  "_links": {
    "authorize": "https://login.microsoftonline.com/..."
  }
}
```

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | Invalid email | Email format invalid |
| 404 | Domain not registered | No tenant owns this email domain |
| 503 | IdP unavailable | Identity provider temporarily unavailable |

### GET /auth/callback
Handles OIDC callback after user authenticates with IdP.

**Query Parameters:**
- `code`: Authorization code from IdP
- `state`: State parameter for CSRF protection

**Behavior:**
1. Validates state matches session
2. Exchanges code for tokens at IdP token endpoint
3. Validates ID token signature against JWKS
4. Extracts user info (email, name, external_id)
5. Checks for existing user or pending invitation
6. Creates user if invitation exists
7. Creates session with user_id
8. Redirects to frontend dashboard

**Success:** 302 redirect to `/`

**Errors:**
- Invalid state: 400 Bad Request
- Token exchange failed: 502 Bad Gateway
- ID token invalid: 401 Unauthorized
- No invitation/user blocked: 403 Forbidden (redirects to error page)

## OIDC Configuration

Environment variables (stored in K8s secrets):
```bash
OIDC_DISCOVERY_URL=https://login.microsoftonline.com/{tenant}/v2.0/.well-known/openid-configuration
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_REDIRECT_URL=https://easi.example.com/auth/callback
```

## Session Management

Using SCS with PostgreSQL store:
- Session lifetime: 24 hours
- Cookie name: `easi_session`
- HTTP-only: true
- Secure: true (production)
- SameSite: Lax

## Local Development Mode

When `LOCAL_DEV_MODE=true`:
- No OIDC authentication required
- Accept headers for mock identity:
  - `X-Tenant-ID`: Tenant ID (default: "default")
  - `X-Dev-Role`: Role (default: "architect")
  - `X-Dev-Email`: Email (default: "developer@localhost")
  - `X-Dev-User-ID`: User ID (default: generated UUID)
- Session middleware creates mock UserIdentity from headers

```bash
curl -H "X-Tenant-ID: acme" \
     -H "X-Dev-Role: admin" \
     http://localhost:8080/api/v1/components
```

## Token Validation
- Validate signature against JWKS (cache keys, refresh periodically)
- Validate issuer matches OIDC discovery
- Validate audience matches client ID
- Validate expiration with 5-minute clock skew tolerance
- Validate nonce matches session state

## Frontend Components

### Login Page
1. Email input form
2. Submit calls POST /auth/sessions
3. Receive authorization URL
4. Redirect browser to IdP

### Callback Handling
1. IdP redirects to /auth/callback
2. Backend processes, sets session cookie
3. Redirects to dashboard

## Checklist
- [ ] Add dependencies: coreos/go-oidc, alexedwards/scs, golang.org/x/oauth2
- [ ] OIDC config from environment variables (K8s secret)
- [ ] Session manager with SCS PostgreSQL store
- [ ] POST /auth/sessions endpoint (initiate login)
- [ ] GET /auth/callback endpoint (complete login)
- [ ] Session middleware with LOCAL_DEV_MODE support
- [ ] Login page with email input (frontend)
- [ ] OIDC redirect handling (frontend)
- [ ] Integration test: complete login flow with mock OIDC
- [ ] User sign-off
