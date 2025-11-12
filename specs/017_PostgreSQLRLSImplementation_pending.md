# PostgreSQL Row-Level Security Implementation

## Description
Implement PostgreSQL Row-Level Security (RLS) for database-level tenant isolation. RLS enforces tenant boundaries at the database layer as defense-in-depth, preventing cross-tenant data access even if application logic fails.

**Dependencies:** Spec 013 (Multi-Tenancy Infrastructure)

## Core Requirements

### RLS Session Variable Pattern
Use session variables to establish tenant context rather than per-tenant database users, enabling efficient connection pooling.

**Set tenant context:**
```sql
SET app.current_tenant = 'tenant-id';
```

**Reference in policies:**
```sql
current_setting('app.current_tenant', true)
```

The second parameter `true` makes current_setting return NULL instead of error if variable not set.

### Database Users

**Application User:**
- Used by application for all normal operations
- Does NOT have BYPASSRLS privilege
- Subject to all RLS policies
- Owns connection pool

**Admin User:**
- Used only for migrations, schema changes, and administrative tasks
- HAS BYPASSRLS privilege
- Never used by application runtime
- Access tightly controlled

### RLS Policy Implementation

**Enable RLS on tables:**
```sql
ALTER TABLE events ENABLE ROW LEVEL SECURITY;
ALTER TABLE snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE application_components ENABLE ROW LEVEL SECURITY;
ALTER TABLE component_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE views ENABLE ROW LEVEL SECURITY;
```

**Create tenant isolation policies:**
```sql
-- Policy for events table
CREATE POLICY tenant_isolation_policy ON events
  FOR ALL
  TO application_user
  USING (tenant_id = current_setting('app.current_tenant', true))
  WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Repeat for all tenant-scoped tables
```

**USING clause:** Filters rows for SELECT, UPDATE, DELETE operations
**WITH CHECK clause:** Validates rows for INSERT, UPDATE operations

### Connection Wrapper

Implement database connection wrapper that:
1. Acquires connection from pool
2. Immediately executes `SET app.current_tenant = $1` with tenant ID
3. Returns wrapped connection to application
4. Re-establishes tenant context on connection reset

**Go implementation approach:**
```go
type TenantAwareDB struct {
    db *sql.DB
}

func (t *TenantAwareDB) SetTenantContext(ctx context.Context, conn *sql.Conn, tenantID string) error {
    _, err := conn.ExecContext(ctx, "SET app.current_tenant = $1", tenantID)
    return err
}

func (t *TenantAwareDB) WithTenantContext(ctx context.Context, tenantID string, fn func(*sql.Conn) error) error {
    conn, err := t.db.Conn(ctx)
    if err != nil {
        return err
    }
    defer conn.Close()

    if err := t.SetTenantContext(ctx, conn, tenantID); err != nil {
        return err
    }

    return fn(conn)
}
```

### Integration Points

**API Middleware:**
- Extract tenant ID from OAuth token or X-Tenant-ID header
- Store in request context
- Connection wrapper reads from context

**Event Store:**
- All SaveEvents and GetEvents operations use tenant-aware connection
- RLS automatically filters events by tenant
- Double defense: application WHERE clause + RLS policy

**Read Model Repositories:**
- All queries use tenant-aware connection
- RLS enforces filtering even if WHERE clause omitted (but keep WHERE clause)

**Projectors:**
- Process events with tenant context set
- Updates to read models automatically tenant-scoped by RLS

### Migration Strategy

**Phase 1: Add tenant_id columns**
- Add tenant_id VARCHAR(50) to all tables
- Backfill with 'default' tenant
- Add NOT NULL constraint

**Phase 2: Enable RLS**
- Create application_user role if not exists
- Enable RLS on all tenant tables
- Create policies for each table
- Test with sample tenants

**Phase 3: Verify isolation**
- Attempt cross-tenant queries (should return empty)
- Verify queries without tenant context fail safely
- Load test with concurrent tenant operations

### Security Considerations

**Defense in Depth:**
RLS is the final layer. Previous layers:
1. TenantId value object validation
2. API middleware tenant scoping
3. Application-level WHERE clauses
4. RLS policies (database enforcement)

**Fail-Safe Behavior:**
- If tenant context not set, policies return zero rows (due to NULL comparison)
- Prevents accidental data exposure
- Log warnings for queries without tenant context

**Policy Testing:**
```sql
-- Test as application user
SET ROLE application_user;
SET app.current_tenant = 'tenant-a';
SELECT * FROM events; -- Should only see tenant-a

SET app.current_tenant = 'tenant-b';
SELECT * FROM events WHERE tenant_id = 'tenant-a'; -- Should return empty due to RLS

-- Attempt to bypass
RESET app.current_tenant;
SELECT * FROM events; -- Should return empty (NULL context)
```

**Attack Vectors to Test:**
- SQL injection attempting to bypass RLS
- Connection reuse with stale tenant context
- Concurrent requests with different tenants
- Race conditions in connection pool
- Malicious WHERE clauses attempting cross-tenant access

### Performance Considerations

**Index Strategy:**
- Create indexes on (tenant_id, other_columns) for efficient filtering
- RLS policies leverage existing indexes
- Monitor query plans to ensure index usage

**Connection Pooling:**
- Session variable persists for connection lifetime
- Reset tenant context when returning connection to pool
- Use pgBouncer in transaction mode carefully (loses session state)

**Query Performance:**
- RLS adds WHERE clause to every query
- Combined with application WHERE clause (both should use same index)
- Minimal overhead with proper indexing

### Monitoring and Observability

**Metrics to Track:**
- Queries executed without tenant context
- RLS policy violations (should be zero in normal operation)
- Query performance impact of RLS policies

**Logging:**
- Log tenant context establishment
- Log any RLS policy failures
- Alert on queries without tenant context in production

## Checklist
- [ ] Create application_user database role without BYPASSRLS
- [ ] Create admin_user database role with BYPASSRLS
- [ ] Enable RLS on all tenant-scoped tables
- [ ] Create RLS policies for each table
- [ ] Implement TenantAwareDB connection wrapper
- [ ] Update EventStore to use tenant-aware connections
- [ ] Update all repositories to use tenant-aware connections
- [ ] Integrate with API middleware for context propagation
- [ ] Migration scripts for schema changes
- [ ] Test RLS policies with multiple tenants
- [ ] Test cross-tenant access prevention
- [ ] Test missing tenant context behavior
- [ ] Verify index usage in query plans
- [ ] Performance testing with RLS enabled
- [ ] Security audit of RLS implementation
- [ ] Documentation for developers
- [ ] User sign-off
