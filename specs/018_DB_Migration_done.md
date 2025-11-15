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
- `DB_HOST`, `DB_PORT`, `DB_NAME` - Database connection
- `DB_ADMIN_USER`, `DB_ADMIN_PASSWORD` - Admin credentials for migrations (easi_admin)
- `DB_USER`, `DB_PASSWORD` - App credentials for backend (easi_app)
- `MIGRATIONS_PATH` - Migration scripts directory (defaults to `./migrations`)

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
- Updated: `/.env.example` - Added DB_ADMIN_USER and DB_USER separation

## Checklist
- [x] Specification ready
- [x] Spec implemented
- [ ] User sign-off