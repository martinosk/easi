# Testing Guide

## Test Tiers

| Tier | Files | Build tag | Needs |
|------|-------|-----------|-------|
| Unit | `*_test.go` | none | nothing |
| Integration | `*_integration_test.go` | `//go:build integration` | a migrated PostgreSQL; the auth package also needs Dex |

`go test ./...` runs unit tests only. Integration tests run only with `-tags=integration`.

## Where the Database Comes From

| Environment | Database | Migrations | Dex |
|-------------|----------|------------|-----|
| Dev container (`.devcontainer/`) | `postgres:5432`, started with the container; `INTEGRATION_TEST_DB_HOST=postgres` is preset | `migrate` service runs on container start; migrations added afterwards: `cd backend && make migrate` | started with the container, reachable as `localhost:5556` |
| Host with Compose | `localhost:5432` after `podman compose up -d` (or `docker compose up -d`) | `migrate` service; rebuild it after adding migrations: `podman compose build migrate && podman compose up migrate` | started by the same command |

Both are the same Compose stack from `docker-compose.yml`. A host `podman compose up -d` while the dev container is open recreates the postgres, migrate and dex containers (the data volume is kept) and the container reconnects to them by name.

The integration tests read `INTEGRATION_TEST_DB_HOST`, `INTEGRATION_TEST_DB_PORT`, `INTEGRATION_TEST_DB_USER`, `INTEGRATION_TEST_DB_PASSWORD`, `INTEGRATION_TEST_DB_NAME` and `INTEGRATION_TEST_DB_SSLMODE` (defaults `localhost`, `5432`, `easi_app`, `localdev`, `easi`, `disable`).

## Running Tests

```bash
cd backend

# Unit
go test ./...

# Integration, every package (skips the auth package when Dex is unreachable)
make test-integration

# Integration, one package
go test -v -tags=integration ./internal/onepagers/... -count=1

# Coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
go test -tags=integration -coverprofile=coverage_integration.out ./internal/onepagers/... -count=1
```

Use `-count=1` on integration runs so a cached result never hides a database change.

## Applying Migrations From the Dev Container

The container has no Docker CLI. `make migrate` runs the migrate binary directly against `postgres:5432` with the Compose defaults (`easi`/`easi` superuser, `localdev` for `easi_app` and `easi_admin`); it is idempotent and runs the `deploy-scripts/pre` and `post` scripts like the `migrate` service does.
