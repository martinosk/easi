# 126 - Edit Grants (Access Delegation)

**Depends on:** [068_InvitationSystem](068_InvitationSystem_done.md), [112_HATEOAS_Driven_Frontend](112_HATEOAS_Driven_Frontend_done.md)

## Description
Architects/admins can grant stakeholders temporary edit access to specific artifacts (capabilities, components, views). This introduces resource-level permission delegation without modifying the core RBAC model. Modeled as a new `accessdelegation` bounded context.

## Edit Grant Lifecycle
```
Grantor creates edit grant
        |
    [active]
        |
        +-> TTL expires (30 days) -> [expired]
        |
        +-> Grantor/admin revokes -> [revoked]
        |
        +-> Artifact deleted -> [revoked] (cascade)
```

## Edit Grant States
| Status | Description |
|--------|-------------|
| active | Grant created, grantee can edit the artifact |
| expired | 30-day TTL elapsed |
| revoked | Grantor or admin revoked, or artifact was deleted |

## Design Decisions
- **Immediate activation**: No pending/accept step. Grant is active on creation.
- **30-day TTL**: Grants auto-expire. No perpetual delegation.
- **Revocable**: Grantor or any admin can revoke at any time.
- **Middleware-based**: `RequireWriteOrEditGrant` replaces `RequirePermission` on artifact PUT/PATCH routes. Checks native write permission first, falls back to edit grant lookup.
- **HATEOAS-gated UI**: `x-edit-grants` link on artifact responses controls frontend visibility of the "Invite to Edit" action.
- **Cross-context cleanup**: When an artifact is deleted, all active grants for that artifact are automatically revoked via event subscription.

## Requirements
- Grantor must have write permission on the artifact type, or `edit-grants:manage` permission
- Cannot grant edit access to yourself
- One active grant per (grantee, artifact) pair
- Grantee with an active grant can PUT/PATCH the specific artifact
- Grantee cannot POST (create new) or DELETE artifacts via edit grants

## API Endpoints

All endpoints require authentication. Create/revoke require write permission on the artifact type or `edit-grants:manage`.

### POST /api/v1/edit-grants
Creates a new edit grant (immediate activation).

**Request:**
```json
{
  "granteeEmail": "stakeholder@company.com",
  "artifactType": "capability",
  "artifactId": "uuid",
  "scope": "write",
  "reason": "Quarterly review input"
}
```

**Response (201):**
```json
{
  "id": "grant-uuid",
  "grantorId": "grantor-uuid",
  "grantorEmail": "architect@company.com",
  "granteeEmail": "stakeholder@company.com",
  "artifactType": "capability",
  "artifactId": "artifact-uuid",
  "scope": "write",
  "status": "active",
  "reason": "Quarterly review input",
  "createdAt": "2026-01-15T10:00:00Z",
  "expiresAt": "2026-02-14T10:00:00Z",
  "_links": {
    "self": { "href": "/api/v1/edit-grants/grant-uuid", "method": "GET" },
    "revoke": { "href": "/api/v1/edit-grants/grant-uuid", "method": "DELETE" }
  }
}
```

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | Cannot grant to self | Grantor and grantee are the same user |
| 400 | Invalid artifact type | Artifact type not one of: capability, component, view |
| 403 | Forbidden | Actor lacks write permission on artifact type |
| 409 | Duplicate grant | Active grant already exists for this grantee+artifact |

### GET /api/v1/edit-grants
Returns grants created by the current user (grantor's view).

### GET /api/v1/edit-grants/{id}
Returns a single edit grant by ID.

### DELETE /api/v1/edit-grants/{id}
Revokes an active edit grant.

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 404 | Not found | Grant doesn't exist |
| 409 | Already revoked | Grant was already revoked |
| 409 | Already expired | Grant has expired |

### GET /api/v1/edit-grants/artifact/{artifactType}/{artifactId}
Returns all grants for a specific artifact.

## HATEOAS Links

### On artifact responses (capabilities, components, views, business domains)
| Link | Condition | Points to |
|------|-----------|-----------|
| `x-edit-grants` | Actor can write artifact type OR has `edit-grants:manage` | `POST /api/v1/edit-grants` |

### On edit grant responses
| Link | Condition | Points to |
|------|-----------|-----------|
| `self` | Always | `GET /api/v1/edit-grants/{id}` |
| `revoke` | Status is active AND actor is grantor or admin | `DELETE /api/v1/edit-grants/{id}` |

## Authorization Middleware

```go
func RequireWriteOrEditGrant(checker EditGrantChecker, artifactType, idParam string) func(http.Handler) http.Handler
```

Applied to PUT/PATCH routes for capabilities, components, and views. Logic:
1. If `actor.CanWrite(artifactType)` -> pass through (normal RBAC)
2. Else check `HasActiveGrant(granteeEmail, artifactType, artifactID)`
3. If grant exists -> pass through
4. Else -> 403 Forbidden

POST (create) and DELETE routes are NOT affected by edit grants.

## Cross-Context Event Subscriptions

| Event | Source Context | Action |
|-------|---------------|--------|
| `CapabilityDeleted` | capabilitymapping | Revoke all active grants for that capability |
| `ApplicationComponentDeleted` | architecturemodeling | Revoke all active grants for that component |
| `ViewDeleted` | architectureviews | Revoke all active grants for that view |

## Domain Model

### Aggregate: EditGrant
- **Value Objects**: ArtifactRef (type + ID), GrantStatus (active/revoked/expired), GrantScope (write)
- **Events**: EditGrantActivated, EditGrantRevoked, EditGrantExpired
- **Invariants**: No self-grant, only active grants can be revoked/expired

### Published Language
- `EditGrantActivated`, `EditGrantRevoked`, `EditGrantExpired`

## Database

### Table: edit_grants (read model)
```sql
CREATE TABLE edit_grants (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    grantor_id VARCHAR(255) NOT NULL,
    grantor_email VARCHAR(255) NOT NULL,
    grantee_email VARCHAR(255) NOT NULL,
    artifact_type VARCHAR(20) NOT NULL,
    artifact_id UUID NOT NULL,
    scope VARCHAR(20) NOT NULL DEFAULT 'write',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    reason TEXT,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);
```
Key indexes: partial index on active grants by (tenant, grantee, artifact_type, artifact_id).

## Frontend Components

### InviteToEditButton
HATEOAS-gated button (visible when `x-edit-grants` link present). Opens InviteToEditDialog.

### InviteToEditDialog
Form with grantee email input and optional reason. Calls `POST /api/v1/edit-grants`.

### EditGrantsList
Admin table view with status filter (all/active/revoked/expired) and revoke action per grant.

### EditGrantBadge
Inline badge showing count of active grants for an artifact.

### MyEditGrants
Card-based view for stakeholders showing their active edit grants.

### Context Menu Integration
"Invite to Edit" item in capability context menu, gated on `x-edit-grants` HATEOAS link.

## Checklist
- [x] Domain model (aggregate, value objects, events)
- [x] Application layer (commands, handlers, read model, projectors)
- [x] Infrastructure (repository, migration, HTTP handlers, middleware)
- [x] Route wiring and HATEOAS links
- [x] Cross-context event subscriptions (artifact deletion cleanup)
- [x] Edit grant middleware wired into capability/component/view PUT routes
- [x] Permissions (`edit-grants:manage` for admin and architect)
- [x] Frontend feature module (components, hooks, API client)
- [x] Frontend integration (query keys, mutation effects, context menu)
- [x] Unit tests: aggregate invariants and value objects (37 tests)
- [ ] Unit tests: command handlers
- [x] Frontend tests: components and hooks (34 tests)
- [ ] Integration tests: full grant lifecycle
- [ ] User sign-off
