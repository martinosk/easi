# 074 - Audit & Security Hardening

**Depends on:** [069_UserManagement](069_UserManagement_pending.md)

## Description
Audit logging for authentication events and comprehensive security testing.

## Audit Events

### Authentication Events
| Event | When | Details |
|-------|------|---------|
| AUTH_SESSION_INITIATED | User submits email | email, tenant_id, ip_address |
| AUTH_SESSION_CREATED | User authenticated and session created | user_id, email, tenant_id |
| AUTH_SESSION_BLOCKED | User authenticated but no invitation/inactive | email, reason |
| AUTH_SESSION_FAILED | OIDC validation failed | email, error |
| AUTH_SESSION_ENDED | User logged out | user_id |

### Authorization Events
| Event | When | Details |
|-------|------|---------|
| AUTHZ_DENIED | User attempted unauthorized action | user_id, resource, action, permission_required |

### User Management Events
| Event | When | Details |
|-------|------|---------|
| INVITATION_CREATED | Admin created invitation | invitation_id, email, role, invited_by |
| INVITATION_ACCEPTED | User accepted invitation | invitation_id, user_id |
| INVITATION_REVOKED | Admin revoked invitation | invitation_id, revoked_by |
| INVITATION_EXPIRED | Invitation TTL elapsed | invitation_id |
| USER_ROLE_CHANGED | Admin changed user role | user_id, old_role, new_role, changed_by |
| USER_DISABLED | Admin disabled user | user_id, disabled_by |
| USER_ENABLED | Admin enabled user | user_id, enabled_by |

## Audit Record Fields
| Field | Type | Description |
|-------|------|-------------|
| timestamp | TIMESTAMP | When event occurred |
| event_type | VARCHAR(50) | Event type from tables above |
| tenant_id | VARCHAR(50) | Tenant context |
| user_id | UUID | User who performed action (if known) |
| user_email | VARCHAR(255) | Email of user |
| ip_address | VARCHAR(45) | Client IP address |
| user_agent | TEXT | Client user agent |
| details | JSONB | Event-specific details |

## Audit Middleware

Logs all authenticated API requests:
- Request path and method
- User identity (from session)
- Response status code
- Duration

For security-sensitive operations, detailed audit records are created within handlers.

## Security Test Cases

### Token Validation Tests
- [ ] Valid token accepted
- [ ] Expired token rejected with 401
- [ ] Token with invalid signature rejected
- [ ] Token with wrong issuer rejected
- [ ] Token with wrong audience rejected
- [ ] Token replay attack prevented (nonce validation)

### Session Security Tests
- [ ] Session cookie is HTTP-only
- [ ] Session cookie has Secure flag (HTTPS)
- [ ] Session cookie has SameSite=Lax
- [ ] Session ID regenerated on login (fixation prevention)
- [ ] Access token expires after 8 hours (triggers refresh)
- [ ] Session expires after 7 days without activity (refresh token expiry)
- [ ] Logged-out session cannot be reused

### Tenant Isolation Tests
- [ ] User A cannot access User B's tenant resources
- [ ] Cross-tenant access returns 404 (not 403)
- [ ] RLS prevents data leakage at database level
- [ ] Tenant ID in session matches authenticated user

### Authorization Tests
- [ ] Unauthenticated requests to protected routes return 401
- [ ] Authenticated user without permission returns 403
- [ ] Admin can access user management
- [ ] Architect cannot access user management
- [ ] Stakeholder has read-only access
- [ ] Disabled user cannot access any routes

### Invitation Security Tests
- [ ] Uninvited user blocked with 403
- [ ] Expired invitation blocked
- [ ] Revoked invitation blocked
- [ ] Invitation for wrong tenant blocked
- [ ] Cannot create invitation for unregistered domain

### OIDC Security Tests
- [ ] State parameter validated (CSRF protection)
- [ ] Authorization code cannot be reused
- [ ] ID token nonce validated
- [ ] Discovery URL validated before use
- [ ] PKCE code_verifier required for token exchange
- [ ] PKCE code_challenge uses S256 method

### Token Refresh Security Tests
- [ ] Refresh token obtains new access token
- [ ] Expired refresh token returns 401
- [ ] Refresh token rotation works (if IdP supports)
- [ ] Invalid refresh token destroys session

## E2E Test Scenarios

### Happy Path: New User Onboarding
1. Admin creates invitation for jane@acme.com
2. Jane navigates to login, enters email
3. Jane authenticates with IdP
4. Invitation accepted, user created
5. Jane lands on dashboard with correct role

### Happy Path: Returning User Login
1. Existing user navigates to login
2. Enters email, redirected to IdP
3. Authenticates successfully
4. Session created, lands on dashboard

### Error Path: Uninvited User
1. Unknown user navigates to login
2. Enters email, redirected to IdP
3. Authenticates successfully
4. Blocked with 403, no user created

### Error Path: Disabled User
1. Disabled user navigates to login
2. Enters email, redirected to IdP
3. Authenticates successfully
4. Blocked with 403

### Error Path: Expired Session
1. User with expired session makes API request
2. Returns 401
3. Frontend redirects to login

## Implementation Notes

### Audit Logging Strategy
- Use database transactions to ensure audit records are created atomically with operations
- Never log sensitive data (passwords, tokens, secrets)
- Include request ID for correlation across logs

### Security Headers
Ensure these headers are set:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `Content-Security-Policy: default-src 'self'`

## Checklist
- [ ] Audit logging middleware
- [ ] Audit events for authentication operations
- [ ] Audit events for user management operations
- [ ] Security tests for token validation
- [ ] Security tests for session handling
- [ ] Cross-tenant isolation tests
- [ ] Authorization permission tests
- [ ] E2E tests for login flows
- [ ] E2E test for uninvited user blocked
- [ ] E2E test for disabled user blocked
- [ ] Security headers configured
- [ ] User sign-off
