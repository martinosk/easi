# Database Conventions

## Migration Rules

- **NEVER modify a committed migration file** - migrations are immutable once committed
- Migrations are forward-only (no down scripts)
- Sequential numbering (001, 002, 003...)
- Each migration must be atomic and transactional
- No foreign key constraints - referential integrity maintained by domain model and event handlers
- To fix issues in committed migrations, create a new migration

## Operational Guide

For detailed migration execution, deployment, and user configuration, see:
[`/backend/deploy-scripts/migrations/README.md`](/backend/deploy-scripts/migrations/README.md)

## Database Users

| User | Purpose | Permissions |
|------|---------|-------------|
| `easi_app` | Runtime application | SELECT, INSERT, UPDATE, DELETE (subject to RLS) |
| `easi_admin` | Migrations & admin | Full privileges (bypasses RLS) |

## Event Store Schema

Core domains use event sourcing with events stored in PostgreSQL. Each bounded context has:
- Own event streams
- Own read model tables
- Tenant isolation via Row-Level Security (RLS)
