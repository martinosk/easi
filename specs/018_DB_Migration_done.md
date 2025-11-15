# DB Migration

## Requirements

Given a sql user with admin rights is available (easi_admin)
When migrations are run at deploy time
And connection to database is verified
Then all SQL scripts from /backend/migrations are executed in order

When a migration script has run successfully
Then a row is added to the schema_migrations table with the name of the script

Given a row with the name of a migration script already exists
When migration scripts are run
Then the script is skipped

## Design Decisions

### Deploy-Time Execution (Not App Startup)
Migrations run as a separate step during deployment, before the backend application starts.

**Rationale:**
- **Privilege separation**: Migrations use `easi_admin` (full privileges), app uses `easi_app` (restricted DML only)
- **Avoids race conditions**: Multiple backend instances starting simultaneously won't conflict
- **Controlled execution**: Explicit, observable migration step in deployment pipeline
- **Safer rollouts**: Migrations complete and are validated before new code runs
- **Zero-downtime deploys**: Enables coordination with blue-green deployment strategies

### Execution Contexts
1. **Local Development**: Docker Compose runs migrations before backend service
2. **CI/CD Pipeline**: Azure DevOps runs migrations as deployment step
3. **Manual**: Standalone migrate binary can be run independently

## Implementation Details

### Architecture

**Migration Runner Package** (`/backend/internal/infrastructure/migrations/`)
- Core migration logic (reusable library)
- Creates and manages `schema_migrations` tracking table
- Executes SQL scripts in transactions

**Migrate Binary** (`/backend/cmd/migrate/`)
- Standalone executable for running migrations
- Uses admin credentials (`easi_admin`)
- Invoked by deployment tooling (docker-compose, Azure DevOps)

**Backend API** (`/backend/cmd/api/`)
- Connects using restricted user (`easi_app`)
- No migration execution on startup
- Assumes schema is already up-to-date

### Migration Runner Features
- Creates `schema_migrations` table to track executed migrations
- Reads all `.sql` files from migrations directory in alphabetical order
- Skips migrations that have already been executed
- Executes each migration in a transaction for atomicity
- Comprehensive error handling and logging

### Configuration

**Connection Strings (Recommended):**
- `DB_ADMIN_CONN_STRING` - Admin connection string for migrations (format: `host=<host> port=<port> user=<user> password=<password> dbname=<dbname> sslmode=<mode>`)
- `DB_CONN_STRING` - Application connection string for backend runtime (format: `host=<host> port=<port> user=<user> password=<password> dbname=<dbname> sslmode=<mode>`)
- `MIGRATIONS_PATH` - Migration scripts directory (defaults to `./migrations`)

**Legacy Individual Parameters (Backward Compatible):**
- `DB_HOST`, `DB_PORT`, `DB_NAME` - Database connection (used if connection string not provided)
- `DB_ADMIN_USER`, `DB_ADMIN_PASSWORD` - Admin credentials for migrations (used if DB_ADMIN_CONN_STRING not provided)
- `DB_USER`, `DB_PASSWORD` - App credentials for backend (used if DB_CONN_STRING not provided)

**Note:** The implementation prefers connection strings but falls back to individual parameters for backward compatibility.

### Files Created/Modified

**Migration Infrastructure:**
- Created: `/backend/internal/infrastructure/migrations/runner.go` - Migration runner library
- Created: `/backend/internal/infrastructure/migrations/runner_test.go` - Comprehensive test suite
- Created: `/backend/cmd/migrate/main.go` - Standalone migration binary

**Docker Configuration:**
- Created: `/backend/Dockerfile.migrate` - Dockerfile for migration service
- Modified: `/docker-compose.yml` - Added migrate and backend services with dependency chain

**Backend Application:**
- Modified: `/backend/cmd/api/main.go` - Removed migration execution, app uses restricted credentials

**Deployment & CI/CD:**
- Modified: `/azure-pipelines.yml` - Added migration image build, push, and K8s job execution
- Created: `/k8s/migrate-job.yaml` - Kubernetes Job manifest for running migrations

**Documentation & Configuration:**
- Updated: `/backend/migrations/README.md` - Documented deploy-time approach with examples
- Updated: `/.env.example` - Added DB_ADMIN_CONN_STRING and DB_CONN_STRING configuration

### Changes in Reopened Version

**Connection String Approach:**
- Migrated from individual database connection parameters to connection strings
- Migration runner (`cmd/migrate/main.go`) now uses `DB_ADMIN_CONN_STRING` with fallback to individual params
- Backend API (`cmd/api/main.go`) now uses `DB_CONN_STRING` with fallback to individual params
- Kubernetes manifests (`k8s/migrate-job.yaml`) use connection strings from secrets
- Docker Compose (`docker-compose.yml`) uses connection strings with sensible defaults
- Updated `.env.example` to document connection string format

**Rationale:**
- Simplifies secret management (one value instead of 5)
- Aligns with PostgreSQL standard connection string format
- Easier to integrate with cloud secret managers (AWS Secrets Manager, Azure Key Vault)
- Reduces configuration complexity in Kubernetes/production environments
- Maintains backward compatibility with individual parameters for gradual migration

## Checklist
- [x] Specification ready
- [x] Initial implementation (individual parameters)
- [x] Refactored to connection strings with backward compatibility
- [ ] User sign-off