# PostgreSQL Schema-Based Bounded Context Ownership

## Problem

Table ownership is currently tracked via a hand-maintained `tableOwnership` map in `architecture_sql_test.go`. Every new table requires a manual entry, and cross-BC violations are detected by string-matching SQL literals against this map. This is fragile, easy to forget, and duplicates information that the database itself should encode structurally.

## Goal

Use PostgreSQL schemas to structurally enforce bounded context table ownership. Each bounded context gets its own schema. The architecture test becomes trivial: a read model in BC `X` should only reference tables in schema `X` (plus shared/infrastructure schemas). No manual map needed.

## Desired End State

- Each bounded context's tables live in a dedicated PostgreSQL schema (e.g., `capabilitymapping.capabilities`, `enterprisearchitecture.enterprise_capabilities`)
- Shared tables (e.g., `events`, `snapshots`, `sessions`) live in a `shared` or `infrastructure` schema
- The `tableOwnership` map and `allowedSQLCrossAccess` allowlist are deleted
- The architecture SQL test validates ownership structurally via schema names instead of a manual registry
- Cross-BC table access is immediately visible in any SQL string (you'd see the foreign schema prefix)
- All existing functionality, RLS policies, and migrations continue to work

## Why This Works Well Here

- No foreign keys exist between tables — cross-schema FK issues don't apply
- RLS works identically regardless of schema
- The ACL cache pattern (specs 136/137) is already eliminating legitimate cross-BC joins, so by the time this is implemented there should be near-zero cross-schema references

## Scope

- All bounded contexts and their tables as registered in the current `tableOwnership` map
- All SQL strings across readmodels, projectors, and migrations
- Database connection setup (schema search path configuration)
- RLS policies
- Architecture guardrail test rewrite

## Success Criteria

- `tableOwnership` map deleted from `architecture_sql_test.go`
- `allowedSQLCrossAccess` map deleted
- Architecture test validates BC isolation via PostgreSQL schemas
- All existing tests pass
- No behavioral changes to any API
