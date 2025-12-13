# 066 - Single-Tenant Login with Dev Mode

**Depends on:** [065_TenantProvisioning](065_TenantProvisioning_done.md) (database schema, secret management)

## Description
Complete end-to-end login flow for single tenant using Authorization Code flow with PKCE. User enters email, is redirected to IdP, authenticates, and receives a session. Backend handles all token operations - access tokens and refresh tokens never reach the browser.

## Libraries
| Library | Purpose | License |
|---------|---------|---------|
| [coreos/go-oidc](https://github.com/coreos/go-oidc) | OIDC discovery, ID token validation, JWKS caching | Apache 2.0 |
| [alexedwards/scs](https://github.com/alexedwards/scs) | Session management with PostgreSQL store | MIT |
| [golang.org/x/oauth2](https://github.com/golang/oauth2) | OAuth2 authorization code + PKCE exchange | BSD 3-Clause |

## OIDC Flow: Authorization Code with PKCE

### Why PKCE?
- Protects against authorization code interception attacks
- Required by OAuth 2.1 for all clients (including confidential)
- Defense-in-depth even with client secret

### Token Architecture
```
Browser                     Backend                      IdP
   │                           │                          │
   │ POST /auth/sessions       │                          │
   │   {email}                 │                          │
   ├──────────────────────────►│                          │
   │                           │ Generate:                │
   │                           │  - state (CSRF)          │
   │                           │  - nonce (replay)        │
   │                           │  - code_verifier (PKCE)  │
   │                           │  - code_challenge        │
   │                           │ Store in session         │
   │                           │                          │
   │ ◄─────────────────────────│                          │
   │   {authorizationUrl}      │                          │
   │                           │                          │
   │ ─────────────────────────────────────────────────────►
   │   Redirect to IdP with code_challenge                │
   │                           │                          │
   │ ◄─────────────────────────────────────────────────────
   │   Redirect with authorization code                   │
   │                           │                          │
   │ GET /auth/callback?code=  │                          │
   ├──────────────────────────►│                          │
   │                           │ Exchange code + verifier │
   │                           ├─────────────────────────►│
   │                           │                          │
   │                           │ ◄────────────────────────│
   │                           │  id_token                │
   │                           │  access_token            │
   │                           │  refresh_token           │
   │                           │                          │
   │                           │ Validate id_token        │
   │                           │ Store tokens in session  │
   │                           │ Create user session      │
   │                           │                          │
   │ ◄─────────────────────────│                          │
   │   Set-Cookie: easi_session (httpOnly, Secure)        │
   │   302 redirect to /       │                          │
```

### Security Properties
- **Tokens never reach browser**: Access and refresh tokens stored server-side only
- **Session cookie**: httpOnly, Secure, SameSite=Lax - cannot be read by JavaScript
- **PKCE**: Prevents code interception even if attacker sees authorization URL
- **Tenant isolation**: Tenant context derived from validated session, never from request

## Login Flow
```
1. User navigates to /login
2. User enters email (e.g., john@acme.com)
3. App extracts domain "acme.com"
4. App queries tenant_domains table → finds tenant "acme"
5. App loads OIDC config from database (tenant-specific)
6. App generates state, nonce, and PKCE code_verifier/code_challenge
7. App stores tenant_id, state, nonce, code_verifier in pre-auth session
8. App redirects to IdP authorization endpoint with code_challenge
9. User authenticates with their company IdP
10. IdP redirects to /auth/callback with authorization code
11. App validates state parameter matches session
12. App exchanges code + code_verifier for tokens (id, access, refresh)
13. App validates ID token (signature, issuer, audience, nonce, expiry)
14. App looks up user by (tenant_id, external_id OR email)
15. If no user exists and no pending invitation → block with 403
16. If invitation exists → accept invitation, create user, set status=active
17. If user exists but status != active → block with 403
18. App stores access_token and refresh_token in session
19. App creates authenticated session with user_id
20. App redirects to dashboard
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

OIDC configuration is stored per-tenant in the database (see spec 065). This enables multi-tenant support where each tenant uses their own Identity Provider.

**Loaded from database:**
- `discovery_url`: Tenant's IdP discovery endpoint
- `client_id`: OAuth client ID registered with tenant's IdP
- `auth_method`: Authentication method (`client_secret` or `private_key_jwt`)
- `scopes`: Requested scopes (default: `openid email profile offline_access`)

**Loaded from SecretProvider (see spec 065):**
- OIDC credentials are stored in AWS Secrets Manager, synced to K8s via External Secrets Operator
- The `SecretProvider` interface reads credentials from mounted K8s secrets at `/secrets/oidc/{tenant-id}/`
- Credentials include: `client_secret` (for client_secret auth) or `private_key`/`certificate` (for private_key_jwt auth)
- The `secretProvisioned` flag in tenant responses indicates whether credentials are available

**Environment variable:**
```bash
OIDC_REDIRECT_URL=https://easi.example.com/auth/callback
```

**Note:** The `offline_access` scope is required to receive refresh tokens. Most IdPs require this scope explicitly.

## Session Management

Using SCS with PostgreSQL store:
- Session lifetime: 8 hours (access token lifetime)
- Refresh extends session up to 7 days (refresh token lifetime)
- Cookie name: `easi_session`
- HTTP-only: true
- Secure: true (production)
- SameSite: Lax

### Session Data Structure
```go
type SessionData struct {
    // Pre-auth (before callback)
    TenantID      string
    State         string
    Nonce         string
    CodeVerifier  string

    // Post-auth (after successful login)
    UserID        uuid.UUID
    AccessToken   string
    RefreshToken  string
    TokenExpiry   time.Time
}
```

### Token Refresh Flow
```
1. API request arrives with session cookie
2. Middleware checks session validity
3. If access token expired but refresh token valid:
   a. Exchange refresh token for new tokens at IdP
   b. Update session with new access_token, refresh_token, expiry
   c. Continue processing request
4. If refresh token expired or invalid:
   a. Destroy session
   b. Return 401 Unauthorized
   c. Frontend redirects to login
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

```bash
curl -H "X-Tenant-ID: acme" \
     -H "X-Dev-Role: admin" \
     http://localhost:8080/api/v1/components
```

## Token Validation

### ID Token Validation (on login)
- Validate signature against JWKS (cache keys, refresh on unknown kid)
- Validate `iss` (issuer) matches OIDC discovery issuer
- Validate `aud` (audience) matches client ID
- Validate `exp` (expiration) with 5-minute clock skew tolerance
- Validate `nonce` matches value stored in pre-auth session
- Extract `sub` as external_id, `email`, and `name` claims

### PKCE Validation (by IdP)
- IdP validates `code_verifier` matches `code_challenge` from authorization request
- Uses S256 challenge method: `BASE64URL(SHA256(code_verifier))`

### Access Token Usage
- Access token stored in session, used for potential downstream API calls
- Not validated by EASI backend (session cookie is the auth mechanism)
- Token expiry tracked to trigger refresh

### Refresh Token Handling
- Stored encrypted in session
- Used to obtain new access/refresh tokens when access token expires
- Refresh token rotation: new refresh token issued on each refresh (if IdP supports)
- On refresh failure, session is invalidated (user must re-authenticate)

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
- [ ] OIDC config loading from database (per-tenant)
- [ ] PKCE code_verifier/code_challenge generation (S256)
- [ ] Session manager with SCS PostgreSQL store
- [ ] POST /auth/sessions endpoint (initiate login with PKCE)
- [ ] GET /auth/callback endpoint (complete login, exchange with code_verifier)
- [ ] ID token validation (signature, issuer, audience, nonce, expiry)
- [ ] Store access_token and refresh_token in session
- [ ] Token refresh middleware (refresh on access token expiry)
- [ ] Session middleware with LOCAL_DEV_MODE support
- [ ] Login page with email input (frontend)
- [ ] OIDC redirect handling (frontend)
- [ ] Integration test: complete login flow with mock OIDC
- [ ] Integration test: token refresh flow
- [ ] User sign-off
