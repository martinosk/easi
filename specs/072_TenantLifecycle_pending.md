# 072 - Tenant Lifecycle

**Depends on:** [065_TenantProvisioning](065_TenantProvisioning_done.md)

## Description
Manage tenant lifecycle: update tenant name, suspend/activate tenants, and archive (soft delete) tenants.

## Tenant States
```
[active] ←→ [suspended] → [archived]
```

| Status | Description |
|--------|-------------|
| active | Tenant is operational, users can login |
| suspended | Tenant is temporarily disabled, users blocked |
| archived | Tenant is soft-deleted, data retained for compliance |

## API Endpoints

### PATCH /api/platform/v1/tenants/{id}
Updates tenant name.

**Request:**
```json
{
  "name": "Acme Corp International"
}
```

**Response (200):** Updated tenant resource

**Errors:** 404 if not found

### POST /api/platform/v1/tenants/{id}/suspend
Suspends tenant.

**Request:**
```json
{
  "reason": "Non-payment"
}
```

**Behavior:**
- All tenant users blocked from login (new logins rejected)
- Existing sessions invalidated (next API request returns 403 with message "Tenant suspended")
- Sets `suspended_at` timestamp and `suspended_reason`

**Response (200):** Updated tenant resource with status="suspended"

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 404 | Tenant not found | Tenant ID doesn't exist |
| 409 | Already suspended | Tenant already in suspended state |

### POST /api/platform/v1/tenants/{id}/activate
Reactivates suspended tenant.

**Behavior:**
- Clears `suspended_at` and `suspended_reason`
- Users can login again

**Response (200):** Updated tenant resource with status="active"

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 404 | Tenant not found | Tenant ID doesn't exist |
| 409 | Not suspended | Tenant not in suspended state |

### DELETE /api/platform/v1/tenants/{id}
Archives tenant (soft delete).

**Behavior:**
- Sets status to "archived"
- Requires all users to be disabled first
- Data retained for compliance
- Cannot be reversed via API

**Response:** 204 No Content

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 404 | Tenant not found | Tenant ID doesn't exist |
| 409 | Cannot archive with active users | Disable all users first |

### PATCH /api/platform/v1/tenants/{id}/oidc-config
Updates OIDC configuration (non-secret fields only).

**Request:**
```json
{
  "discoveryUrl": "https://...",
  "clientId": "new-client-id",
  "authMethod": "private_key_jwt",
  "scopes": "openid email profile"
}
```

**Note:** OIDC secrets (client_secret, private_key, certificate) are managed exclusively via AWS Secrets Manager. See spec 065 for secret provisioning instructions. This endpoint only updates non-secret configuration stored in the database.

**Behavior:**
- Validates discovery URL before accepting
- All fields optional (partial update)
- Does NOT handle secrets - those must be updated in AWS Secrets Manager

**Response (200):** Updated tenant resource

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | Invalid OIDC config | Discovery URL unreachable or invalid |
| 400 | Invalid auth method | Must be `client_secret` or `private_key_jwt` |
| 404 | Tenant not found | Tenant ID doesn't exist |

## Audit Events
| Event | Details |
|-------|---------|
| TENANT_UPDATED | id, changes |
| TENANT_SUSPENDED | id, reason |
| TENANT_ACTIVATED | id |
| TENANT_ARCHIVED | id |
| OIDC_CONFIG_UPDATED | tenant_id (no secrets logged) |

## Checklist
- [ ] PATCH /api/platform/v1/tenants/{id} (update name)
- [ ] POST /api/platform/v1/tenants/{id}/suspend
- [ ] POST /api/platform/v1/tenants/{id}/activate
- [ ] DELETE /api/platform/v1/tenants/{id} (archive)
- [ ] PATCH /api/platform/v1/tenants/{id}/oidc-config
- [ ] Session invalidation on tenant suspension
- [ ] Integration tests: lifecycle state transitions
- [ ] Integration test: cannot archive with active users
- [ ] User sign-off
