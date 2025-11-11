# Multi-Tenancy Infrastructure

## Description
Implements multi-tenancy support across the entire application, enabling complete data isolation between different tenants (organizations/customers). Each tenant operates in a logically isolated environment while sharing the same physical infrastructure. This is a foundational architectural change that enables synthetic transaction testing, SaaS deployment models, and enterprise customer isolation.

Multi-tenancy is implemented as an **infrastructure concern**, not a domain concern. The domain layer remains pure and tenant-unaware. Tenant context is injected at the API boundary and automatically flows through all layers.

## Core Principles
- **Tenant ID is a value object** - Immutable, validated, type-safe
- **Domain purity** - Domain models don't contain tenant-aware business logic
- **Automatic scoping** - All queries/commands automatically scoped to current tenant
- **Infrastructure concern** - Tenant context managed in infrastructure layer
- **Event sourcing compatible** - Tenant ID included in all events for filtering
- **Zero cross-tenant leakage** - Impossible to accidentally access another tenant's data

## Tenant Context Flow

```
HTTP Request (tenant-id header)
  → API Middleware (extract & validate tenant)
    → Command/Query Handler (tenant context injected)
      → Aggregate Root (tenant ID as value object)
        → Event Store (tenant ID in events)
          → Read Model Projection (tenant-filtered)
```

## Domain Model Changes

### TenantId Value Object

**Properties:**
- `Value` (string, required): The unique tenant identifier (e.g., "acme-corp", "synthetic-monitoring")

**Validation Rules:**
- Must not be empty or whitespace
- Must match pattern: `^[a-z0-9-]{3,50}$` (lowercase alphanumeric with hyphens)
- Reserved tenant IDs: "system", "admin", "root"

**Special Tenants:**
- `synthetic-monitoring` - Used for production health checks
- `synthetic-load-test` - Used for load testing in production
- `default` - Default tenant for single-tenant deployments

### Aggregate Changes

All existing aggregates must include TenantId:

**ApplicationComponent Aggregate:**
```
- TenantId (TenantId, required)
- ComponentId (ComponentId, required)
- Name (ComponentName, required)
- Description (ComponentDescription, optional)
- CreatedAt (CreatedAt, required)
```

**ComponentRelation Aggregate:**
```
- TenantId (TenantId, required)
- RelationId (ComponentRelationId, required)
- SourceComponentId (ComponentId, required)
- TargetComponentId (ComponentId, required)
- RelationType (RelationType, required)
- Name (RelationName, optional)
- Description (RelationDescription, optional)
- CreatedAt (CreatedAt, required)
```

**View Aggregate (if exists):**
```
- TenantId (TenantId, required)
- ViewId (ViewId, required)
- Name (ViewName, required)
- [... other properties]
```

### Event Changes

All events must include TenantId for proper event store filtering:

**ApplicationComponentCreated:**
```json
{
  "tenantId": "string",
  "componentId": "guid",
  "name": "string",
  "description": "string",
  "createdAt": "datetime"
}
```

**ComponentRelationCreated:**
```json
{
  "tenantId": "string",
  "relationId": "guid",
  "sourceComponentId": "guid",
  "targetComponentId": "guid",
  "relationType": "Triggers | Serves",
  "name": "string",
  "description": "string",
  "createdAt": "datetime"
}
```

## API Changes

### Tenant Context Injection

All API requests must include tenant context via header:

```
X-Tenant-ID: acme-corp
```

**Authentication/Authorization (Future):**
- In production, tenant ID derived from authenticated user's organization
- For now, accept tenant ID from header (trusted environment)
- Middleware validates tenant ID format and existence

### Tenant Scoping Middleware

```go
func TenantScopingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantIDStr := r.Header.Get("X-Tenant-ID")

        // Default to "default" tenant if not specified
        if tenantIDStr == "" {
            tenantIDStr = "default"
        }

        tenantID, err := NewTenantID(tenantIDStr)
        if err != nil {
            http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
            return
        }

        // Inject tenant context into request context
        ctx := context.WithValue(r.Context(), TenantContextKey, tenantID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Updated API Endpoints

All existing endpoints automatically tenant-scoped:

**POST /api/application-component**
- Request header: `X-Tenant-ID: acme-corp`
- Creates component for specified tenant
- Returns tenant ID in response

**GET /api/application-component**
- Returns only components for current tenant
- Tenant ID from header

**GET /api/application-component/{id}**
- Returns 404 if component belongs to different tenant
- Automatic tenant boundary enforcement

### New Tenant Management Endpoints

**GET /api/tenants**
Gets all tenants (admin only - future).

**Response:** 200 OK
```json
[
  {
    "id": "acme-corp",
    "displayName": "Acme Corporation",
    "createdAt": "datetime",
    "_links": {
      "self": {
        "href": "/api/tenants/acme-corp"
      },
      "components": {
        "href": "/api/application-component",
        "title": "Tenant components"
      }
    }
  }
]
```

**POST /api/tenants**
Creates a new tenant (admin only - future).

**Request Body:**
```json
{
  "id": "new-tenant",
  "displayName": "New Tenant Organization"
}
```

**Response:** 201 Created

**DELETE /api/tenants/{tenantId}/data**
Deletes all data for a tenant (synthetic tenant cleanup).

**Validation:**
- Only allowed for tenants with prefix `synthetic-`
- Returns 403 Forbidden for regular tenants

**Response:** 204 No Content

## Infrastructure Changes

### Event Store

**Tenant Filtering in Event Queries:**
```go
func (s *EventStore) GetEventsForAggregate(
    tenantID TenantID,
    aggregateID string,
) ([]Event, error) {
    // Query events WHERE tenant_id = ? AND aggregate_id = ?
}

func (s *EventStore) GetEventsByType(
    tenantID TenantID,
    eventType string,
) ([]Event, error) {
    // Query events WHERE tenant_id = ? AND event_type = ?
}
```

**Event Store Schema:**
```sql
CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    version INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Composite index for tenant-scoped queries
    INDEX idx_events_tenant_aggregate (tenant_id, aggregate_id),
    INDEX idx_events_tenant_type (tenant_id, event_type)
);
```

### Read Models

**Automatic Tenant Filtering:**
```go
func (r *ComponentRepository) GetAll(ctx context.Context) ([]Component, error) {
    tenantID := GetTenantFromContext(ctx)
    // SELECT * FROM components WHERE tenant_id = ?
}

func (r *ComponentRepository) GetByID(
    ctx context.Context,
    id ComponentID,
) (*Component, error) {
    tenantID := GetTenantFromContext(ctx)
    // SELECT * FROM components WHERE tenant_id = ? AND id = ?
    // Returns nil if component belongs to different tenant
}
```

**Read Model Schema Updates:**
```sql
-- Add tenant_id to all read model tables

ALTER TABLE components
ADD COLUMN tenant_id VARCHAR(50) NOT NULL DEFAULT 'default';

ALTER TABLE component_relations
ADD COLUMN tenant_id VARCHAR(50) NOT NULL DEFAULT 'default';

ALTER TABLE views
ADD COLUMN tenant_id VARCHAR(50) NOT NULL DEFAULT 'default';

-- Create indexes for tenant-scoped queries
CREATE INDEX idx_components_tenant ON components(tenant_id);
CREATE INDEX idx_relations_tenant ON component_relations(tenant_id);
CREATE INDEX idx_views_tenant ON views(tenant_id);

-- Add composite unique constraints including tenant
ALTER TABLE components
ADD CONSTRAINT uq_components_tenant_name UNIQUE (tenant_id, name);
```

## Migration Strategy

### Phase 1: Add TenantId Infrastructure
- [ ] Create TenantId value object
- [ ] Add tenant_id column to all tables (default: "default")
- [ ] Implement tenant context middleware
- [ ] Update event store to include tenant_id

### Phase 2: Update Domain Models
- [ ] Add TenantId property to all aggregates
- [ ] Update all events to include tenant_id
- [ ] Update command handlers to use tenant context

### Phase 3: Update Repositories
- [ ] Add tenant filtering to all read model queries
- [ ] Update projections to include tenant_id
- [ ] Test cross-tenant isolation

### Phase 4: Update APIs
- [ ] Add X-Tenant-ID header handling
- [ ] Update all API responses to include tenant context
- [ ] Add tenant management endpoints

### Phase 5: Testing
- [ ] Unit tests for TenantId value object
- [ ] Integration tests with multiple tenants
- [ ] E2E tests verifying tenant isolation
- [ ] Security tests for cross-tenant access attempts

## Backward Compatibility

**Single-Tenant Mode:**
- If `X-Tenant-ID` header not provided, use `"default"` tenant
- Existing installations migrate to single "default" tenant
- No breaking changes for current users

**Migration Script:**
```sql
-- Backfill existing data to "default" tenant
UPDATE events SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE components SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE component_relations SET tenant_id = 'default' WHERE tenant_id IS NULL;
```

## Security Considerations

- **Tenant ID validation** - Prevent injection attacks via tenant ID
- **Access control** - Future: verify user has access to tenant
- **No cross-tenant queries** - Impossible by design (tenant in WHERE clause)
- **Audit logging** - Log tenant context with all operations
- **Tenant isolation testing** - Automated tests verify no data leakage

## Performance Considerations

- **Indexes** - All tenant-scoped queries use composite indexes
- **Query plans** - Verify query planner uses tenant_id indexes
- **Connection pooling** - Shared pool across tenants (not per-tenant)
- **Caching** - Cache keys must include tenant ID

## Checklist
- [ ] Specification ready
- [ ] TenantId value object created
- [ ] Database schema migration created
- [ ] Event store updated for tenant filtering
- [ ] All aggregates updated with TenantId
- [ ] All events updated with tenant_id
- [ ] Command handlers inject tenant context
- [ ] Read model repositories filter by tenant
- [ ] API middleware implements tenant scoping
- [ ] Tenant management endpoints created
- [ ] Migration script for existing data tested
- [ ] Unit tests implemented and passing
- [ ] Integration tests with multiple tenants passing
- [ ] E2E tests verify tenant isolation
- [ ] Performance testing with tenant indexes
- [ ] Security audit for cross-tenant access
- [ ] Documentation updated
- [ ] User sign-off
