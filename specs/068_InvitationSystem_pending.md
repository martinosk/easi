# 068 - Invitation System

**Depends on:** [067_SessionManagement](067_SessionManagement_pending.md)

## Description
Admin can invite users, users can accept invitations via login. User onboarding is modeled as a long-running process via the Invitation resource.

## Invitation Lifecycle
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

## Invitation States
| Status | Description |
|--------|-------------|
| pending | Invitation created, awaiting user login |
| accepted | User authenticated, invitation fulfilled |
| expired | TTL elapsed without user login |
| revoked | Admin cancelled the invitation |

## Invitation Requirements
- **TTL**: 7 days from creation
- **Expiration handling**: Lazy evaluation (check expiry on login attempt, mark expired if TTL elapsed)
- **Cleanup**: Expired/revoked invitations retained for audit purposes (no automatic deletion)

## Uninvited Users
Users who authenticate via IdP without an invitation are blocked:
- 403 response: "Access denied. Contact your administrator for access."
- No user record created (invitation-only model)

## API Endpoints

All invitation endpoints require `invitations:manage` permission (admin only).

### POST /api/v1/invitations
Creates a new invitation.

**Request:**
```json
{
  "email": "jane@acme.com",
  "role": "architect"
}
```

**Behavior:**
- Creates invitation with status=pending
- Sets expires_at = now + 7 days
- Email must match tenant's registered domains

**Response (201):**
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

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | Invalid email | Email format invalid |
| 400 | Invalid role | Role not one of: admin, architect, stakeholder |
| 400 | Domain mismatch | Email domain not registered to tenant |
| 409 | User exists | Active user with this email already exists |
| 409 | Invitation pending | Pending invitation for this email already exists |

### GET /api/v1/invitations
Returns paginated list of invitations.

**Query Parameters:**
- `status`: Filter by status (pending, accepted, expired, revoked)
- `limit`: Page size (default 50, max 100)
- `after`: Cursor for pagination

**Response (200):**
```json
{
  "data": [
    {
      "id": "invitation-uuid",
      "email": "jane@acme.com",
      "role": "architect",
      "status": "pending",
      "invitedBy": { "id": "admin-uuid", "email": "admin@acme.com" },
      "createdAt": "2025-12-02T10:00:00Z",
      "expiresAt": "2025-12-09T10:00:00Z",
      "_links": {
        "self": "/api/v1/invitations/{id}",
        "revoke": "/api/v1/invitations/{id}/revoke"
      }
    }
  ],
  "pagination": { "hasMore": false, "limit": 50 },
  "_links": { "self": "/api/v1/invitations" }
}
```

### GET /api/v1/invitations/{id}
Returns invitation details.

**Response (200):** Same as single invitation in list

**Errors:** 404 if not found

### POST /api/v1/invitations/{id}/revoke
Revokes a pending invitation.

**Behavior:**
- Sets status = revoked
- Sets revoked_at = now

**Response (200):** Updated invitation resource

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 404 | Not found | Invitation doesn't exist |
| 409 | Cannot revoke | Invitation not in pending status |

## Login Flow with Invitation Checking

Modified callback flow (from spec 066):
```
...
10. App validates ID token against JWKS
11. App looks up user by (tenant_id, external_id OR email)
12. If user exists and status=active → create session, redirect
13. If user exists and status=disabled → block with 403
14. If no user exists:
    a. Look up pending invitation by (tenant_id, email)
    b. If no invitation → block with 403
    c. If invitation expired (expires_at < now) → mark expired, block with 403
    d. Accept invitation:
       - Set invitation.status = accepted
       - Set invitation.accepted_at = now
       - Create user with role from invitation
       - Create session, redirect
```

## Authorization Middleware

Protect routes by required permission:

```go
r.Route("/api/v1/invitations", func(r chi.Router) {
    r.Use(middleware.RequirePermission(PermInvitationsManage))
    r.Post("/", invitationHandlers.Create)
    r.Get("/", invitationHandlers.List)
    r.Get("/{id}", invitationHandlers.Get)
    r.Post("/{id}/revoke", invitationHandlers.Revoke)
})
```

Middleware behavior:
- Check user.status == "active"
- Check required permission exists in user.permissions
- Return 403 Forbidden if denied

## Frontend Components

### Invitation Management Page (Admin Only)
- Table of invitations with status badges
- Filter by status
- "Invite User" button opens modal
- "Revoke" action for pending invitations

### Invite User Modal
- Email input (validated against tenant domains)
- Role dropdown (admin, architect, stakeholder)
- Submit creates invitation

## Audit Events
| Event | When |
|-------|------|
| INVITATION_CREATED | Admin created invitation |
| INVITATION_ACCEPTED | User accepted invitation via login |
| INVITATION_REVOKED | Admin revoked invitation |
| INVITATION_EXPIRED | Invitation TTL elapsed (on login attempt) |

## Checklist
- [ ] POST /api/v1/invitations endpoint
- [ ] GET /api/v1/invitations endpoint
- [ ] GET /api/v1/invitations/{id} endpoint
- [ ] POST /api/v1/invitations/{id}/revoke endpoint
- [ ] Login flow with invitation checking (accept on first login)
- [ ] Invitation expiration handling (lazy evaluation)
- [ ] Authorization middleware (RequirePermission)
- [ ] Invitation management page (frontend, admin only)
- [ ] Integration tests: invitation lifecycle (create → accept)
- [ ] Integration test: invitation expiration
- [ ] Integration test: uninvited user blocked
- [ ] User sign-off
