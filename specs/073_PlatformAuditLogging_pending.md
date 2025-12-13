# 073 - Platform Audit Logging

**Depends on:** [065_TenantProvisioning](065_TenantProvisioning_done.md)

## Description
Audit trail for all platform administration operations. Provides visibility into tenant lifecycle events for compliance and debugging.

## Audit Events

| Event | When | Details |
|-------|------|---------|
| TENANT_CREATED | POST /tenants | id, name, domains, first_admin_email |
| TENANT_UPDATED | PATCH /tenants/{id} | id, changes |
| TENANT_SUSPENDED | POST /tenants/{id}/suspend | id, reason |
| TENANT_ACTIVATED | POST /tenants/{id}/activate | id |
| TENANT_ARCHIVED | DELETE /tenants/{id} | id |
| DOMAIN_ADDED | POST /tenants/{id}/domains | tenant_id, domain |
| DOMAIN_REMOVED | DELETE /tenants/{id}/domains/{domain} | tenant_id, domain |
| OIDC_CONFIG_UPDATED | PATCH /tenants/{id}/oidc-config | tenant_id (no secrets logged) |

## Audit Record Fields
| Field | Type | Description |
|-------|------|-------------|
| id | SERIAL | Auto-incrementing ID |
| event_type | VARCHAR(50) | Event type from table above |
| tenant_id | VARCHAR(50) | Tenant affected (nullable for platform-wide events) |
| user_id | UUID | User who performed action (nullable for platform admin) |
| user_email | VARCHAR(255) | Email of user who performed action |
| performed_by | VARCHAR(255) | Identifier of actor (platform-admin, user email) |
| ip_address | VARCHAR(45) | Client IP address |
| user_agent | TEXT | Client user agent |
| details | JSONB | Event-specific details |
| created_at | TIMESTAMP | When event occurred |

## Database Schema

```sql
CREATE TABLE platform_audit_log (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    tenant_id VARCHAR(50),
    user_id UUID,
    user_email VARCHAR(255),
    performed_by VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_platform_audit_tenant ON platform_audit_log(tenant_id);
CREATE INDEX idx_platform_audit_event ON platform_audit_log(event_type);
CREATE INDEX idx_platform_audit_created ON platform_audit_log(created_at);
```

## API Endpoints

### GET /api/platform/v1/tenants/{id}/audit-log
Returns paginated audit log for a specific tenant.

**Query Parameters:**
- `eventType`: Filter by event type
- `from`: Start date (ISO 8601)
- `to`: End date (ISO 8601)
- `limit`: Page size (default 50, max 100)
- `after`: Cursor for pagination

**Response (200):**
```json
{
  "data": [
    {
      "id": 123,
      "eventType": "TENANT_SUSPENDED",
      "tenantId": "acme",
      "performedBy": "platform-admin",
      "ipAddress": "192.168.1.100",
      "details": { "reason": "Non-payment" },
      "createdAt": "2025-12-02T10:00:00Z"
    }
  ],
  "pagination": {
    "hasMore": true,
    "limit": 50,
    "cursor": "eyJpZCI6MTIzfQ=="
  },
  "_links": {
    "self": "/api/platform/v1/tenants/acme/audit-log",
    "next": "/api/platform/v1/tenants/acme/audit-log?after=eyJpZCI6MTIzfQ=="
  }
}
```

### GET /api/platform/v1/audit-log
Returns paginated platform-wide audit log.

**Query Parameters:**
- `tenantId`: Filter by tenant
- `eventType`: Filter by event type
- `from`: Start date (ISO 8601)
- `to`: End date (ISO 8601)
- `limit`: Page size (default 50, max 100)
- `after`: Cursor for pagination

**Response (200):** Same structure as tenant-specific endpoint

## Implementation Notes
- Audit logging happens in the same transaction as the operation
- Failed operations are not logged (only successful operations)
- Secrets (client_secret) must never be logged in details

## Checklist
- [ ] Database migration for platform_audit_log table
- [ ] Audit logging for all platform admin operations
- [ ] GET /api/platform/v1/tenants/{id}/audit-log
- [ ] GET /api/platform/v1/audit-log
- [ ] Integration tests: audit log queries
- [ ] Integration test: verify secrets not logged
- [ ] User sign-off
