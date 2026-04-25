---
name: easi-database-migrations
description: MUST load when creating, reviewing, or fixing database migrations in EASI. Load when adding new tables, altering columns, fixing migration errors, or working with the event store schema. Also load when configuring database users or reviewing RLS policies.
compatibility: opencode
---

# EASI Database Migrations

## Overview

EASI uses forward-only, immutable migrations. Once a migration file is committed it must never be edited — instead create a new migration to correct it. This ensures every deployed environment can be brought to the current schema by replaying the same migration sequence.

## Core Rules

| Rule | Rationale |
|------|-----------|
| **NEVER modify a committed migration file** | Other environments have already run it; modifying breaks schema divergence detection |
| **Forward-only — no down scripts** | Down scripts are rarely correct in practice and encourage unsafe rollbacks |
| **Sequential numbering** (001, 002, 003…) | Prevents ordering ambiguity; gaps are forbidden |
| **Each migration must be atomic and transactional** | Partial failures leave schema in a known rollback state |
| **No foreign key constraints** | Referential integrity is maintained by the domain model and event handlers, not the DB |
| **To fix a committed migration, write a new one** | Append a corrective migration; never edit history |

## Database Users

| User | Purpose | Permissions |
|------|---------|-------------|
| `easi_app` | Runtime application | SELECT, INSERT, UPDATE, DELETE (subject to RLS) |
| `easi_admin` | Migrations & admin | Full privileges — bypasses RLS |

Always run migrations as `easi_admin`. Application code runs as `easi_app` and is subject to Row-Level Security.

## Bounded Contexts and Schema Prefixes

Core domains use event sourcing. Events are stored in PostgreSQL. Each bounded context has:

- **Own event streams** — no cross-context event table sharing
- **Own read model tables** — denormalized projections per context
- **Tenant isolation** — enforced via Row-Level Security (RLS) at the PostgreSQL level

### Schema prefix rules

1. **Always qualify the table name**: `CREATE TABLE <schema>.<table>` — never bare `CREATE TABLE <table>`.
2. **Create the schema before using it**: `CREATE SCHEMA IF NOT EXISTS <schema>;` at the top of the migration (omit if the schema already exists).
3. **Schema name = bounded context directory name** under `backend/internal/`.
4. **All subsequent DDL on the same table must also be schema-qualified** — `ALTER TABLE <schema>.<table>`, `CREATE INDEX … ON <schema>.<table>`, `DROP POLICY … ON <schema>.<table>`, etc.

### Bounded Context → Schema Map

| Schema | Bounded Context | Type | `backend/internal/` directory |
|--------|----------------|------|-------------------------------|
| `infrastructure` | Event store | — | _(shared, not context-specific)_ |
| `shared` | Cross-cutting | — | _(shared, not context-specific)_ |
| `architecturemodeling` | Architecture Modeling | CQRS/ES | `architecturemodeling/` |
| `architectureviews` | Architecture Views | CQRS/ES | `architectureviews/` |
| `capabilitymapping` | Capability Mapping | CQRS/ES | `capabilitymapping/` |
| `metamodel` | MetaModel | CQRS/ES | `metamodel/` |
| `enterprisearchitecture` | Enterprise Architecture | CQRS/ES | `enterprisearchitecture/` |
| `valuestreams` | Value Streams | CQRS/ES | `valuestreams/` |
| `accessdelegation` | Access Delegation | CQRS/ES | `accessdelegation/` |
| `viewlayouts` | View Layouts | CRUD | `viewlayouts/` |
| `releases` | Releases | CRUD | `releases/` |
| `archassistant` | Arch Assistant | CRUD | `archassistant/` |
| `importing` | Importing | CRUD | `importing/` |
| `platform` | Platform | CRUD | `platform/` |
| `auth` | Auth | CRUD | `auth/` |

## Migration File Location

```
backend/deploy-scripts/migrations/
```

Full operational guide (execution, deployment, user configuration):
`/backend/deploy-scripts/migrations/README.md`

## Guidelines

1. **Never touch a committed migration** — write a new one instead
2. **Keep migrations atomic** — one DDL operation per migration when possible
3. **Always run as `easi_admin`** — not `easi_app`
4. **No FK constraints in schema** — domain model owns referential integrity
5. **New bounded contexts need their own tables** — never share event tables across contexts
6. **Tenant isolation via RLS** — do not bypass RLS in application code
