# 071 - Domain Management

**Depends on:** [065_TenantProvisioning](065_TenantProvisioning_pending.md)

## Description
Add and remove email domains from tenants. Email domains are used for tenant discovery during login.

## Invariants
- At least one email domain must be registered per tenant
- Email domains must be globally unique (no two tenants share a domain)
- Domain format must match `^[a-z0-9][a-z0-9.-]*[a-z0-9]$`

## API Endpoints

### POST /api/platform/v1/tenants/{id}/domains
Adds email domain to tenant.

**Request:**
```json
{
  "domain": "acme.de"
}
```

**Response (201):**
```json
{
  "domain": "acme.de",
  "tenantId": "acme",
  "createdAt": "2025-12-02T10:00:00Z",
  "_links": {
    "self": "/api/platform/v1/tenants/acme/domains/acme.de",
    "delete": "/api/platform/v1/tenants/acme/domains/acme.de",
    "tenant": "/api/platform/v1/tenants/acme"
  }
}
```

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | Invalid domain format | Domain format invalid |
| 404 | Tenant not found | Tenant ID doesn't exist |
| 409 | Domain already registered | Domain belongs to another tenant |

### DELETE /api/platform/v1/tenants/{id}/domains/{domain}
Removes email domain from tenant.

**Response:** 204 No Content

**Errors:**
| Status | Error | Description |
|--------|-------|-------------|
| 404 | Tenant not found | Tenant ID doesn't exist |
| 404 | Domain not found | Domain not registered to this tenant |
| 409 | Cannot remove last domain | At least one domain required |

### GET /api/platform/v1/domains
Returns all domain → tenant mappings. Useful for debugging domain conflicts.

**Response (200):**
```json
{
  "data": [
    {
      "domain": "acme.com",
      "tenantId": "acme",
      "createdAt": "2025-12-02T10:00:00Z",
      "_links": {
        "tenant": "/api/platform/v1/tenants/acme"
      }
    },
    {
      "domain": "contoso.com",
      "tenantId": "contoso",
      "createdAt": "2025-12-01T10:00:00Z",
      "_links": {
        "tenant": "/api/platform/v1/tenants/contoso"
      }
    }
  ],
  "_links": {
    "self": "/api/platform/v1/domains"
  }
}
```

## Audit Events
| Event | Details |
|-------|---------|
| DOMAIN_ADDED | tenant_id, domain |
| DOMAIN_REMOVED | tenant_id, domain |

## Checklist
- [ ] POST /api/platform/v1/tenants/{id}/domains
- [ ] DELETE /api/platform/v1/tenants/{id}/domains/{domain}
- [ ] GET /api/platform/v1/domains (all domain mappings)
- [ ] Integration tests: domain add/remove operations
- [ ] Integration test: cannot remove last domain
- [ ] Integration test: domain uniqueness across tenants
- [ ] User sign-off
