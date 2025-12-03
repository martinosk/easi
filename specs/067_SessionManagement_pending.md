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
- Sessions expire after 24 hours
- Expired session on API request returns 401 Unauthorized
- Frontend receives 401, redirects user to login page
- No automatic session refresh (user must re-authenticate)

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
- [ ] Session check on app load (frontend)
- [ ] User context provider (frontend)
- [ ] Unit tests for session validation
- [ ] Unit tests for role/permission mapping
- [ ] User sign-off
