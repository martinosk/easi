# Architecture Guard Tests

## Description
Add compile-time architecture tests that prevent cross-bounded-context coupling at two levels: Go package imports and SQL table ownership. These tests form the safety net that catches violations before they reach production, and must be in place before any remediation work begins.

## Purpose
Developers inadvertently create coupling between bounded contexts by importing internal packages or writing SQL joins against foreign tables. These violations are invisible until runtime. Architecture guard tests make violations fail fast during `go test ./...`, catching them in development and CI.

## Dependencies
- None (this is the first spec to implement)

## Baseline Violations

The following violations exist today and must be **allowlisted** in the initial implementation. Each allowlisted entry will be removed as specs 135-138 resolve the underlying violation.

### Go Import Violations (initial allowlist)
| Importing BC | Imported Package | Reason |
|---|---|---|
| enterprisearchitecture | `capabilitymapping/infrastructure/metamodel` | StrategyPillarsGateway interface (fix: spec 135) |
| enterprisearchitecture | `auth/domain/valueobjects` | Permission constants (fix: spec 135) |
| enterprisearchitecture | `auth/infrastructure/session` | SessionManager (fix: spec 135) |
| capabilitymapping | `auth/domain/valueobjects` | Permission constants (fix: spec 135) |
| capabilitymapping | `auth/infrastructure/session` | SessionManager (fix: spec 135) |
| architecturemodeling | `auth/domain/valueobjects` | Permission constants (fix: spec 135) |
| architectureviews | `auth/domain/valueobjects` | Permission constants (fix: spec 135) |
| architectureviews | `auth/application/readmodels` | UserReadModel (fix: spec 138) |
| accessdelegation | `architecturemodeling/application/readmodels` | ArtifactNameResolver (fix: spec 138) |
| accessdelegation | `architectureviews/application/readmodels` | ArtifactNameResolver (fix: spec 138) |
| accessdelegation | `capabilitymapping/application/readmodels` | ArtifactNameResolver (fix: spec 138) |
| accessdelegation | `auth/application/readmodels` | UserReadModel (fix: spec 138) |
| importing | `architecturemodeling/application/commands` | Import orchestrator (fix: spec 138) |
| importing | `capabilitymapping/application/commands` | Import orchestrator (fix: spec 138) |
| viewlayouts | `architectureviews/publishedlanguage` | Acceptable (published language) |

### SQL Table Violations (initial allowlist)
| BC with Violation | Foreign Table | Owner BC | File |
|---|---|---|---|
| enterprisearchitecture | `capabilities` | capabilitymapping | `maturity_analysis_read_model.go` |
| enterprisearchitecture | `capabilities`, `capability_realizations`, `effective_capability_importance`, `application_fit_scores` | capabilitymapping | `time_suggestion_read_model.go` |
| enterprisearchitecture | `domain_capability_assignments` | capabilitymapping | `domain_capability_metadata_read_model.go` |
| capabilitymapping | `domain_capability_metadata` | enterprisearchitecture | `strategic_fit_analysis_read_model.go` |

## Part 1: Go Import Boundary Test

### Location
`backend/internal/architecture_test.go`

### Design

The test parses all `.go` source files (excluding `_test.go`) under `backend/internal/`, extracts import paths, and asserts that no bounded context imports another BC's internal packages unless explicitly allowed.

**Allowed cross-BC imports:**
- `<bc>/publishedlanguage` — always allowed (this is the defined integration point)
- `shared/` — always allowed (shared kernel)
- `infrastructure/` (top-level) — always allowed (shared infrastructure: database, eventstore, migrations)
- `platform/infrastructure/api` — always allowed (rate limiter, middleware shared utilities)
- Explicit allowlist entries for known violations (to be removed as fixes land)

### Rules

```
RULE 1: CLOSED BY DEFAULT. Every top-level directory under backend/internal/
        that is NOT in sharedPackages is a bounded context. New directories
        are automatically protected without any registration.

RULE 2: A bounded context may import from:
  - Its own packages (any depth)
  - Directories listed in sharedPackages (shared/, infrastructure/, testing/)
  - Specific packages listed in freeImportPackages (platform/infrastructure/api)
  - <other-bc>/publishedlanguage (published language only)
  - Allowlisted exceptions (temporary, tracked for removal)

RULE 3: Everything else is a violation.
```

### Implementation

The design is **closed by default**: every top-level directory under `backend/internal/` is assumed to be a bounded context unless explicitly listed as shared infrastructure. This means a new BC added by a developer is automatically protected — other BCs cannot import its internals without the test failing. No registration step required.

```go
//go:build !integration

package internal_test

import (
    "go/parser"
    "go/token"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

const modulePrefix = "easi/backend/internal/"

// sharedPackages lists directories under backend/internal/ that are NOT bounded
// contexts. These can be imported freely by any BC. Everything else is treated
// as a bounded context and locked down: only its publishedlanguage/ subpackage
// may be imported by other BCs.
//
// IMPORTANT: This list is intentionally restrictive. Adding a new directory
// under backend/internal/ without listing it here automatically treats it as
// a bounded context with full import protection. This is the desired behavior —
// new BCs are locked down by default.
var sharedPackages = map[string]bool{
    "shared":         true, // Shared kernel (cqrs, events, eventsourcing, valueobjects, etc.)
    "infrastructure": true, // Shared infrastructure (database, eventstore, migrations, api router)
    "testing":        true, // Shared test fixtures
}

// freeImportPackages lists specific packages from BCs that may be imported
// by any other BC because they serve as shared utilities (not domain logic).
// These should be migrated to shared/ or publishedlanguage/ over time.
var freeImportPackages = map[string]bool{
    "platform/infrastructure/api": true, // Rate limiter, shared middleware
}

// allowedCrossBCImports defines temporary exceptions for known violations.
// Format: "importing-bc -> imported-package-suffix"
// Each entry MUST reference the spec that will remove it.
var allowedCrossBCImports = map[string]string{
    // Spec 135: Published Language Expansion
    "enterprisearchitecture -> capabilitymapping/infrastructure/metamodel":  "spec-135",
    "enterprisearchitecture -> auth/domain/valueobjects":                    "spec-135",
    "enterprisearchitecture -> auth/infrastructure/session":                 "spec-135",
    "capabilitymapping -> auth/domain/valueobjects":                         "spec-135",
    "capabilitymapping -> auth/infrastructure/session":                      "spec-135",
    "architecturemodeling -> auth/domain/valueobjects":                      "spec-135",
    "architectureviews -> auth/domain/valueobjects":                         "spec-135",
    // Spec 138: Cross-BC Import Elimination
    "architectureviews -> auth/application/readmodels":                      "spec-138",
    "accessdelegation -> architecturemodeling/application/readmodels":       "spec-138",
    "accessdelegation -> architectureviews/application/readmodels":          "spec-138",
    "accessdelegation -> capabilitymapping/application/readmodels":          "spec-138",
    "accessdelegation -> auth/application/readmodels":                       "spec-138",
    "importing -> architecturemodeling/application/commands":                "spec-138",
    "importing -> capabilitymapping/application/commands":                   "spec-138",
}
```

**Detection logic:**

```
For each .go file under backend/internal/:
  1. Determine the owning top-level directory (e.g., "capabilitymapping")
  2. If the owner is in sharedPackages → skip (shared code can import anything)
  3. For each import starting with "easi/backend/internal/":
     a. Extract the imported top-level directory (e.g., "auth")
     b. If imported == owner → OK (same BC)
     c. If imported is in sharedPackages → OK (shared kernel/infra)
     d. If imported path is in freeImportPackages → OK (shared utility)
     e. If imported path contains "/publishedlanguage" → OK (published language)
     f. If "owner -> imported-suffix" is in allowedCrossBCImports → WARN (tracked)
     g. Otherwise → FAIL (cross-BC violation)
```

The test walks all `.go` files, identifies which directory each file belongs to, and applies these rules. Any directory not in `sharedPackages` is automatically a bounded context — **no registration needed**.

### Key Behavior
- Test runs with `go test ./internal/` (not behind integration tag)
- Each violation prints: `CROSS-BC VIOLATION: <file> imports <package> (from <foreign-bc>, only publishedlanguage allowed)`
- Allowlist entries print a reminder: `ALLOWLISTED: <file> imports <package> (tracked for removal by <spec>)`
- A separate subtest asserts that no allowlist entries are stale (i.e., if an allowlisted import no longer exists in the codebase, the test fails, reminding the developer to remove the allowlist entry)

## Part 2: SQL Table Ownership Test

### Location
`backend/internal/architecture_sql_test.go`

### Design

The test uses Go's `go/ast` to parse all read model source files, extracts SQL string literals, and identifies table names referenced in `FROM`, `JOIN`, `INSERT INTO`, `UPDATE`, and `DELETE FROM` clauses. It then asserts that each table belongs to the same BC as the file.

### Table Ownership Map

The map includes tables from specs 136-137 that don't exist yet. This is intentional: when those specs create the tables and read models reference them, the ownership map already has the correct entries. Tables that don't exist yet simply never appear in any source file scan, so they have no effect until the corresponding spec is implemented.

```go
var tableOwnership = map[string]string{
    // Event Store (shared infrastructure)
    "events":    "infrastructure",
    "snapshots": "infrastructure",

    // Shared (HTTP session storage)
    "sessions":  "shared",

    // Architecture Modeling
    "application_components":         "architecturemodeling",
    "component_relations":            "architecturemodeling",
    "application_component_experts":  "architecturemodeling",
    "acquired_entities":              "architecturemodeling",
    "vendors":                        "architecturemodeling",
    "internal_teams":                 "architecturemodeling",
    "acquired_via_relationships":     "architecturemodeling",
    "purchased_from_relationships":   "architecturemodeling",
    "built_by_relationships":         "architecturemodeling",

    // Architecture Views
    "architecture_views":        "architectureviews",
    "view_element_positions":    "architectureviews",
    "view_component_positions":  "architectureviews",
    "view_preferences":          "architectureviews",

    // Capability Mapping
    "capabilities":                    "capabilitymapping",
    "capability_dependencies":         "capabilitymapping",
    "capability_realizations":         "capabilitymapping",
    "capability_experts":              "capabilitymapping",
    "capability_tags":                 "capabilitymapping",
    "capability_component_cache":      "capabilitymapping",
    "domain_capability_assignments":   "capabilitymapping",
    "effective_capability_importance":  "capabilitymapping",
    "application_fit_scores":          "capabilitymapping",
    "cm_strategy_pillar_cache":        "capabilitymapping",
    "strategy_importance":             "capabilitymapping",
    "domain_composition_view":         "capabilitymapping",
    "business_domains":                "capabilitymapping",

    // Enterprise Architecture
    "enterprise_capabilities":          "enterprisearchitecture",
    "enterprise_capability_links":      "enterprisearchitecture",
    "enterprise_strategic_importance":   "enterprisearchitecture",
    "domain_capability_metadata":        "enterprisearchitecture",
    "capability_link_blocking":          "enterprisearchitecture",
    "ea_strategy_pillar_cache":          "enterprisearchitecture",

    // View Layouts
    "layout_containers":  "viewlayouts",
    "element_positions":  "viewlayouts",

    // Importing
    "import_sessions": "importing",

    // Auth / Platform
    "tenants":              "platform",
    "tenant_domains":       "platform",
    "tenant_oidc_configs":  "platform",
    "users":                "auth",
    "invitations":          "auth",

    // Access Delegation
    "edit_grants": "accessdelegation",

    // MetaModel
    "meta_model_configurations": "metamodel",

    // Releases
    "releases": "releases",

    // Value Streams
    "value_streams":                 "valuestreams",
    "value_stream_stages":           "valuestreams",
    "value_stream_stage_capabilities": "valuestreams",
    "value_stream_capability_cache": "valuestreams",
}
```

### Allowlist for SQL Violations

```go
var allowedSQLCrossAccess = map[string]string{
    // Spec 136: EA Read Model Decoupling
    "enterprisearchitecture/maturity_analysis_read_model.go -> capabilities":              "spec-136",
    "enterprisearchitecture/time_suggestion_read_model.go -> capability_realizations":      "spec-136",
    "enterprisearchitecture/time_suggestion_read_model.go -> capabilities":                 "spec-136",
    "enterprisearchitecture/time_suggestion_read_model.go -> effective_capability_importance": "spec-136",
    "enterprisearchitecture/time_suggestion_read_model.go -> application_fit_scores":       "spec-136",
    "enterprisearchitecture/domain_capability_metadata_read_model.go -> domain_capability_assignments": "spec-136",
    // Spec 137: CM Read Model Decoupling
    "capabilitymapping/strategic_fit_analysis_read_model.go -> domain_capability_metadata": "spec-137",
}
```

### SQL Parsing Strategy

Rather than a full SQL parser, use a regex-based table extractor that handles the patterns used in this codebase:
1. Extract all backtick-delimited or double-quoted string literals from Go source
2. Apply regex patterns for: `FROM\s+(\w+)`, `JOIN\s+(\w+)`, `INSERT\s+INTO\s+(\w+)`, `UPDATE\s+(\w+)`, `DELETE\s+FROM\s+(\w+)`
3. Ignore table aliases (single-letter identifiers after table names)
4. Match extracted table names against the ownership map

### Key Behavior
- Scans `backend/internal/*/application/readmodels/*.go` and `backend/internal/*/application/projectors/*.go`
- Each violation prints: `SQL CROSS-BC VIOLATION: <file> references table <table> (owned by <owner-bc>, file is in <current-bc>)`
- Infrastructure tables (`events`, `snapshots`) are allowed from any BC
- Shared tables (`sessions`) are allowed from any BC
- Tables in the `infrastructure` or `shared` ownership group are allowed from any BC

## Part 3: CI Integration

Add a step to the CI pipeline (if not already covered by `go test ./...`) that runs:
```bash
go test ./internal/ -run TestNoCrossBoundedContextImports
go test ./internal/ -run TestReadModelsOnlyReferenceOwnedTables
```

These tests are fast (no database needed, no integration tag) and run on every commit.

## Files to Create

```
backend/internal/architecture_test.go           # Go import boundary test
backend/internal/architecture_sql_test.go        # SQL table ownership test
```

## Files to Modify

None. This spec only adds new test files.

## Success Criteria

- `go test ./internal/ -run TestNoCrossBoundedContextImports` passes (all violations are in allowlist)
- `go test ./internal/ -run TestReadModelsOnlyReferenceOwnedTables` passes (all violations are in allowlist)
- Adding a new cross-BC import NOT in the allowlist causes a test failure
- Adding a new cross-BC SQL join NOT in the allowlist causes a test failure
- Each allowlist entry references the spec that will remove it
- Removing a violation from code and forgetting to remove the allowlist entry causes a test failure (stale entry detection)

## Checklist

- [x] Specification approved
- [x] Go import boundary test implemented
- [x] SQL table ownership test implemented
- [x] All existing violations allowlisted with spec references
- [x] Stale allowlist detection working
- [x] `go test ./...` passes (no integration tag)
- [x] Documentation updated in CLAUDE.md if needed
- [x] User sign-off
