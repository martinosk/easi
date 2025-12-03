# 069 - User Management

**Depends on:** [068_InvitationSystem](068_InvitationSystem_pending.md)

## Description
Admin can manage existing users: view user list, change roles, and disable/enable accounts.

## User States
| Status | Description |
|--------|-------------|
| active | User can login and access the system |
| disabled | User is blocked from login |

## API Endpoints

### GET /api/v1/users
Returns paginated list of users in tenant.

**Required Permission:** `users:read`

**Query Parameters:**
- `status`: Filter by status (active, disabled)
- `role`: Filter by role (admin, architect, stakeholder)
- `limit`: Page size (default 50, max 100)
- `after`: Cursor for pagination

**Response (200):**
```json
{
  "data": [
    {
      "id": "user-uuid",
      "email": "jane@acme.com",
      "name": "Jane Doe",
      "role": "architect",
      "status": "active",
      "createdAt": "2025-12-01T10:00:00Z",
      "lastLoginAt": "2025-12-02T08:30:00Z",
      "_links": {
        "self": "/api/v1/users/{id}",
        "changeRole": "/api/v1/users/{id}/change-role",
        "disable": "/api/v1/users/{id}/disable"
      }
    }
  ],
  "pagination": { "hasMore": false, "limit": 50 },
  "_links": { "self": "/api/v1/users" }
}
```

### GET /api/v1/users/{id}
Returns user details.

**Required Permission:** `users:read`

**Response (200):**
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

**Notes:**
- `_links.disable` only present if user is active
- `_links.enable` only present if user is disabled
- Actions not available for self (current user)

**Errors:** 404 if not found

### POST /api/v1/users/{id}/change-role
Changes user's role.

**Required Permission:** `users:manage`

**Request:**
```json
{
  "role": "admin"
}
```

**Behavior:**
- Updates user's role
- New role takes effect on next session check

**Response (200):** Updated user resource

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | Invalid role | Role not one of: admin, architect, stakeholder |
| 404 | User not found | User doesn't exist |
| 409 | Last admin | Cannot demote last admin in tenant |

### POST /api/v1/users/{id}/disable
Disables user account.

**Required Permission:** `users:manage`

**Behavior:**
- Sets user status = disabled
- User's existing sessions remain valid until expiry
- User cannot create new sessions (login blocked)

**Response (200):** Updated user resource

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 404 | User not found | User doesn't exist |
| 409 | Cannot disable self | Cannot disable your own account |
| 409 | Last admin | Cannot disable last admin in tenant |

### POST /api/v1/users/{id}/enable
Re-enables disabled user account.

**Required Permission:** `users:manage`

**Behavior:**
- Sets user status = active
- User can login again

**Response (200):** Updated user resource

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 404 | User not found | User doesn't exist |
| 409 | Already active | User is not disabled |

### GET /api/v1/tenants/current
Returns current tenant info. Available to all authenticated users.

**Response (200):**
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

**Notes:**
- `_links.users` and `_links.invitations` only present for admins

## Invariants
- At least one admin must exist per tenant
- Cannot demote or disable the last admin
- Cannot disable yourself

## Frontend Components

### User Management Page (Admin Only)
- Table of users with:
  - Email, name
  - Role badge
  - Status badge (active/disabled)
  - Last login timestamp
- Filter by status, role
- Actions:
  - Change role dropdown
  - Disable/Enable button

### Role-Based UI Visibility
- Admin: sees full menu (Users, Invitations, Settings)
- Architect: sees architecture tools only
- Stakeholder: sees read-only views

## Audit Events
| Event | When |
|-------|------|
| USER_ROLE_CHANGED | Admin changed user role |
| USER_DISABLED | Admin disabled user |
| USER_ENABLED | Admin enabled user |

## Checklist
- [ ] GET /api/v1/users endpoint
- [ ] GET /api/v1/users/{id} endpoint
- [ ] POST /api/v1/users/{id}/change-role endpoint
- [ ] POST /api/v1/users/{id}/disable endpoint
- [ ] POST /api/v1/users/{id}/enable endpoint
- [ ] GET /api/v1/tenants/current endpoint
- [ ] User management page (frontend, admin only)
- [ ] Role-based UI visibility (frontend)
- [ ] Integration tests: user management operations
- [ ] Integration test: cannot demote last admin
- [ ] Integration test: cannot disable self
- [ ] User sign-off
