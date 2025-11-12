# Multi-Tenancy Infrastructure

## Description
Implement multi-tenancy across the application enabling complete data isolation between tenants. Tenant context is injected at the API boundary and flows through all layers as an infrastructure concern. Domain models remain tenant-unaware.

## Core Requirements

### Tenant ID Value Object
- String property matching pattern: `^[a-z0-9-]{3,50}$`
- Reserved IDs: "system", "admin", "root"
- Special tenants: "synthetic-monitoring", "synthetic-load-test", "default"

### Domain Layer - Tenant Unaware
**Domain models (aggregates, entities, events, value objects) must remain completely tenant-unaware:**
- Aggregates: No TenantID property
- Events: No tenantId field
- Domain logic focuses solely on business rules and invariants
- Tenant isolation is handled exclusively in infrastructure layer

This ensures domain models remain pure and focused on business concerns, following DDD principles.

### API Tenant Context

**Local Development Mode:**
- Accept X-Tenant-ID header (no auth required)
- Default to "default" tenant if header missing
- Enable via LOCAL_DEV_MODE=true environment variable

**Production Mode:**
- Tenant ID extracted from OAuth token claims (see Spec 015)
- X-Tenant-ID header ignored in production
- Middleware validates user has access to requested tenant

All existing endpoints are automatically tenant-scoped. Return 404 if resource belongs to different tenant.

### Infrastructure Layer - Tenant Context Flow

**Context Propagation:**
- API middleware extracts tenant ID from header (dev mode) or OAuth token (production)
- Tenant context stored in Go context.Context and flows through application layer
- Infrastructure services (event store, repositories) extract tenant from context
- Database connection wrapper sets PostgreSQL session variable from context

**Event Store Implementation:**
- SaveEvents: Extract tenant from context, add tenant_id when inserting to database
- GetEvents: Extract tenant from context, filter by tenant_id in WHERE clause
- Domain events remain unchanged - tenant is metadata managed by infrastructure

**Repository Implementation:**
- All queries include tenant_id filter extracted from context
- Projectors extract tenant from event metadata when updating read models
- Read model queries filtered by tenant from context

**Context Keys:**
```go
type contextKey string
const TenantContextKey contextKey = "tenant_id"

// Middleware sets: ctx = context.WithValue(ctx, TenantContextKey, tenantID)
// Infrastructure reads: tenantID := ctx.Value(TenantContextKey).(string)
```

### Infrastructure Changes

**Event Store:**
- Add tenant_id column (VARCHAR(50)) to events table
- Add tenant_id column to snapshots table
- Create composite indexes: (tenant_id, aggregate_id), (tenant_id, event_type)
- Enable RLS on events and snapshots tables

**Read Models:**
- Add tenant_id column (VARCHAR(50)) to all tables (default: 'default')
- Create indexes on tenant_id columns
- Add composite unique constraints including tenant_id
- Enable RLS on all read model tables

**Migration:**
- Backfill existing data to "default" tenant
- Existing single-tenant deployments continue using "default" tenant

### PostgreSQL Row-Level Security (RLS)

**Database-Level Tenant Isolation:**
RLS provides defense-in-depth by enforcing tenant isolation at the database layer, independent of application code.

**Session Variable Approach:**
- Set tenant context using PostgreSQL session variable: `SET app.current_tenant = 'tenant-id'`
- Application sets this immediately after acquiring database connection
- RLS policies use `current_setting('app.current_tenant')` to filter rows

**RLS Policy Pattern:**
```sql
-- Enable RLS on table
ALTER TABLE events ENABLE ROW LEVEL SECURITY;

-- Create policy for application user
CREATE POLICY tenant_isolation_policy ON events
  USING (tenant_id = current_setting('app.current_tenant', true));

-- For operations requiring both read and write
CREATE POLICY tenant_isolation_policy ON events
  FOR ALL
  USING (tenant_id = current_setting('app.current_tenant', true))
  WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
```

**Connection Management:**
- Execute `SET app.current_tenant = $1` immediately after connection acquisition from pool
- Use connection wrapper or middleware to ensure tenant context is set before any queries
- Session variable persists for connection lifetime in pool
- Re-establish on connection reset or error

**RLS for All Tenant Tables:**
Apply RLS policies to: events, snapshots, application_components, component_relations, views, and all future tenant-scoped tables.

**Bypass for System Operations:**
- Create dedicated database user for migrations and admin operations
- Grant BYPASSRLS privilege only to admin user
- Application user must NOT have BYPASSRLS privilege

### Security
- Defense in depth: ID validation, LOCAL_DEV_MODE flag, OAuth enforcement, RLS database isolation
- RLS policies prevent cross-tenant access even if application logic fails
- All queries automatically filtered by RLS (no explicit WHERE tenant_id needed, but recommended as second layer)
- Log all tenant context operations
- Test cross-tenant access prevention at both application and database levels
- Verify RLS policies cannot be circumvented by malicious queries
- Monitor for queries that fail due to missing tenant context

## Checklist
- [x] TenantId value object created (infrastructure layer)
- [x] Context key constants defined for tenant propagation
- [x] Database schema migration with tenant_id columns
- [x] RLS enabled on all tenant-scoped tables
- [x] RLS policies created for all tenant tables
- [x] Database connection wrapper sets tenant context from Go context (TenantAwareDB)
- [x] Dedicated admin database user with BYPASSRLS privilege
- [x] Event store extracts tenant from context and filters by tenant_id
- [x] Domain models verified to be tenant-unaware (no TenantID properties)
- [x] Domain events verified to be tenant-unaware (no tenantId fields)
- [x] Command handlers pass context through to infrastructure
- [x] Read repositories filter by tenant from context (ApplicationComponent, ComponentRelation, ArchitectureView)
- [x] API middleware extracts and injects tenant context (LOCAL_DEV_MODE support)
- [ ] Migration script tested
- [x] Unit tests for TenantId value object
- [ ] Integration tests with multiple tenants
- [ ] Backend integration tests verify tenant isolation
- [ ] Security tests prevent cross-tenant access at application layer
- [ ] Security tests verify RLS policies enforce isolation at database layer
- [ ] Test that missing tenant context fails safely
- [ ] User sign-off
