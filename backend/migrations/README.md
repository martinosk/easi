# Database Migrations

This directory contains all database schema migrations for the EASI backend. All DDL (Data Definition Language) operations are managed through sequential migration scripts.

## Migration Philosophy

- **No Down Scripts**: Migrations are forward-only. We do not support rollbacks via down scripts.
- **Sequential Execution**: Migrations must be run in numeric order (001, 002, 003, etc.)
- **Idempotent**: Migrations use `IF NOT EXISTS` and `IF EXISTS` clauses to be safely re-runnable
- **Supports Both**:
  - **New Databases**: Can spin up a complete schema from empty database
  - **Existing Databases**: Can migrate existing databases by only applying changes

## Running Migrations

### Deploy-Time Migrations (Recommended)

Migrations run as a **separate step before the backend application starts**. This provides better separation of concerns, privilege isolation, and avoids race conditions with multiple backend instances.

#### Local Development (Docker Compose)

```bash
# Starts postgres → runs migrations → starts backend
docker-compose up

# Or step by step:
docker-compose up postgres          # Start database
docker-compose up migrate           # Run migrations
docker-compose up backend           # Start application
```

The docker-compose configuration:
- **migrate service**: Runs migration binary with `easi_admin` credentials
- **backend service**: Starts only after migrations complete successfully
- Backend uses `easi_app` (restricted) credentials

#### Manual Execution

Build and run the migration binary directly:

```bash
# From backend directory
cd backend

# Build migrate binary
go build -o migrate cmd/migrate/main.go

# Run migrations with admin credentials
DB_HOST=localhost \
DB_PORT=5432 \
DB_ADMIN_USER=easi_admin \
DB_ADMIN_PASSWORD=change_me_in_production \
DB_NAME=easi \
MIGRATIONS_PATH=./migrations \
./migrate
```

#### CI/CD Pipeline (Azure DevOps)

Add a migration step before deploying the backend:

```yaml
- task: Bash@3
  displayName: 'Run Database Migrations'
  inputs:
    targetType: 'inline'
    script: |
      cd backend
      go build -o migrate cmd/migrate/main.go
      ./migrate
  env:
    DB_HOST: $(DB_HOST)
    DB_PORT: $(DB_PORT)
    DB_ADMIN_USER: $(DB_ADMIN_USER)
    DB_ADMIN_PASSWORD: $(DB_ADMIN_PASSWORD)
    DB_NAME: $(DB_NAME)
    MIGRATIONS_PATH: ./migrations

- task: Deploy Backend
  dependsOn: RunMigrations
  ...
```

### Migration Tracking

The system creates a `schema_migrations` table to track executed migrations:

```sql
CREATE TABLE schema_migrations (
    id SERIAL PRIMARY KEY,
    migration_name VARCHAR(255) NOT NULL UNIQUE,
    executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)
```

**How it works:**
1. Database connection is verified
2. `schema_migrations` table is created (if not exists)
3. All `.sql` files in migrations directory are read in alphabetical order
4. Already-executed migrations (found in tracking table) are skipped
5. Pending migrations are executed in transactions
6. Successfully executed migrations are recorded

### Alternative: golang-migrate

You can also use [golang-migrate](https://github.com/golang-migrate/migrate):

```bash
migrate -path ./migrations -database "postgres://easi_admin:password@localhost:5432/easi?sslmode=disable" up
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