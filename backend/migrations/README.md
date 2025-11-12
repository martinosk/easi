# Database Migrations

This directory contains all database schema migrations for the EASI backend. All DDL (Data Definition Language) operations are managed through sequential migration scripts.

## Migration Philosophy

- **No Down Scripts**: Migrations are forward-only. We do not support rollbacks via down scripts.
- **Sequential Execution**: Migrations must be run in numeric order (001, 002, 003, etc.)
- **Idempotent**: Migrations use `IF NOT EXISTS` and `IF EXISTS` clauses to be safely re-runnable
- **Supports Both**:
  - **New Databases**: Can spin up a complete schema from empty database
  - **Existing Databases**: Can migrate existing databases by only applying changes

## Migration Files

You can use:
- [golang-migrate](https://github.com/golang-migrate/migrate)

Example with golang-migrate:
```bash
migrate -path ./migrations -database "postgres://easi:easi@localhost:5432/easi?sslmode=disable" up
```

## Database Users

After running migrations, you'll have two database users:

### easi_app
- **Purpose**: Runtime application user
- **Permissions**: SELECT, INSERT, UPDATE, DELETE on all tables
- **RLS**: Subject to Row-Level Security policies
- **Password**: Change `change_me_in_production` in migration 003

### easi_admin
- **Purpose**: Administrative tasks and migrations
- **Permissions**: Full privileges on database
- **RLS**: BYPASSRLS privilege (not subject to RLS policies)
- **Password**: Change `change_me_in_production` in migration 003

## Application Configuration

After running migrations, configure your application to:

1. **Use easi_app user** for runtime operations:
   ```env
   DB_USER=easi_app
   DB_PASSWORD=your_secure_password
   ```

2. **Set tenant context** after acquiring each database connection:
   ```go
   // Via SQL
   _, err := conn.ExecContext(ctx, "SET app.current_tenant = $1", tenantID)

   // Via helper function
   _, err := conn.ExecContext(ctx, "SELECT set_tenant_context($1)", tenantID)
   ```

## Development vs Production

### Development
- Use `easi_admin` user for convenience
- Can bypass RLS if needed for debugging
- Can use migration tools with full privileges

### Production
- Application MUST use `easi_app` user
- Enable connection pooling with tenant context
- Use `easi_admin` only for migrations and maintenance
- Change default passwords in migration 003
- Use secure password management (e.g., AWS Secrets Manager, Vault)