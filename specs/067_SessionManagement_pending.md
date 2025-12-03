# 067 - Session Management

**Depends on:** [066_SingleTenantLogin](066_SingleTenantLogin_pending.md)

## Description
Session introspection and logout. Users can check their current session status and log out.

## API Endpoints

### GET /auth/sessions/current
Returns current session with user identity.

**Response (200):**
```json
{
  "id": "session-uuid",
  "user": {
    "id": "user-uuid",
    "email": "john@acme.com",
    "name": "John Doe",
    "role": "architect",
    "permissions": ["components:read", "components:write", "views:read", "views:write"]
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

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 401 | Unauthorized | No valid session |

### DELETE /auth/sessions/current
Logs out current user by clearing session.

**Behavior:**
- Destroys server-side session
- Clears session cookie

**Response:** 204 No Content

## Session Expiry Behavior

### Token Lifetimes
| Token | Lifetime | Purpose |
|-------|----------|---------|
| Access token | 8 hours | Short-lived, triggers refresh |
| Refresh token | 7 days | Extends session without re-auth |
| Session cookie | 7 days | Matches refresh token lifetime |

### Automatic Token Refresh
- Session middleware checks access token expiry on each request
- If access token expired, refresh token is used to obtain new tokens
- Refresh happens transparently - user does not notice
- Session extended up to 7 days from last activity

### Session Expiry Flow
```
1. API request arrives with session cookie
2. Middleware loads session from database
3. Check access token expiry:
   a. Valid → continue to handler
   b. Expired → attempt refresh
4. Refresh attempt:
   a. Success → update session, continue to handler
   b. Failure → destroy session, return 401
5. Frontend receives 401 → redirect to /login
```

### Hard Expiry
- After 7 days without activity, refresh token expires
- User must re-authenticate via IdP
- This is a security boundary - no indefinite sessions

## UserIdentity Value Object

Injected into request context by session middleware:

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

## Roles & Permissions

### Roles
| Role | Description |
|------|-------------|
| admin | Full tenant access including user management |
| architect | Read/write architecture models |
| stakeholder | Read-only access |

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

## Frontend Components

### Session Check on App Load
1. On app initialization, call GET /auth/sessions/current
2. If 401, redirect to /login
3. If 200, store user in React context

### Handling 401 During Use
1. Any API call may return 401 (refresh token expired)
2. Global axios/fetch interceptor catches 401
3. Redirect to /login with return URL
4. After re-auth, redirect back to original page

### User Context Provider
```typescript
interface UserContext {
  user: {
    id: string;
    email: string;
    name: string;
    role: 'admin' | 'architect' | 'stakeholder';
    permissions: string[];
  };
  tenant: {
    id: string;
    name: string;
  };
  logout: () => Promise<void>;
}
```

### Logout Flow
1. Call DELETE /auth/sessions/current
2. Clear local user context
3. Redirect to /login

## Checklist
- [ ] GET /auth/sessions/current endpoint
- [ ] DELETE /auth/sessions/current endpoint
- [ ] Automatic token refresh in session middleware
- [ ] Session check on app load (frontend)
- [ ] Global 401 interceptor with redirect (frontend)
- [ ] User context provider (frontend)
- [ ] Unit tests for session validation
- [ ] Unit tests for role/permission mapping
- [ ] Integration test: transparent token refresh
- [ ] Integration test: session expiry after refresh token expires
- [ ] User sign-off
