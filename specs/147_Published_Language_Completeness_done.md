# Published Language Completeness and Boundary Purity

## Description
Ensure every bounded context that exposes cross-BC contracts has an explicit `publishedlanguage` package with a clear two-tier structure, remove framework leakage from published contracts, and add architecture test guardrails that enforce ACL boundaries.

## Dependencies
- Spec 146 for payload DTO definitions and ACL principles.

## Published Language Package Structure

Each `publishedlanguage` package has two tiers with distinct import rules:

```
<bc>/publishedlanguage/
├── events.go          # Event name constants (cross-BC importable)
├── ids.go             # ID type aliases and parse helpers (cross-BC importable)
└── contracts/
    └── events.go      # Payload DTOs documenting wire format (documentation-only)
```

**Root package** (`publishedlanguage/`): Constants, ID types, gateway interfaces. Importable by any BC in production code. These are stable identifiers — a string constant or type alias is not domain model leakage.

**Contracts sub-package** (`publishedlanguage/contracts/`): Payload DTOs documenting the serialized event format. NOT imported in downstream production code. Used only as:
- The authoritative specification for writing local ACL structs
- Test imports for contract verification (ensuring ACL structs match the documented contract)

## Required Changes

### 1) Create missing published language packages where integration exists
Create `publishedlanguage` only for BCs with actual cross-BC contracts. Do not create empty placeholder packages.

Minimum expected in this cycle:
- `enterprisearchitecture/publishedlanguage/events.go` if EA events are consumed outside EA.

### 2) Add contracts sub-packages for payload DTOs
For each `publishedlanguage` package that defines externally consumed events, add a `contracts/` sub-package with the payload DTOs defined per spec 146.

### 3) Remove framework coupling from published contracts
`publishedlanguage` packages must not implement or depend on shared event-sourcing domain interfaces. Published contracts are plain DTO/value types plus parsing helpers. Specifically:
- `auth/publishedlanguage` types must not implement `domain.ValueObject` or similar framework interfaces.
- `contracts/` sub-packages must only depend on stdlib (no domain imports).

### 4) Normalize contract shape
Published language packages should expose:
- Constants for event names (root package)
- ID type aliases and parse helpers (root package)
- Payload DTOs for externally consumed events (contracts sub-package)
- Command/query DTOs only where cross-BC integration requires them

### 5) Add architecture test guardrails
Extend `architecture_test.go` with new rules:

**Rule: Block cross-BC imports of `publishedlanguage/contracts` in production code.**
The existing `isAllowedCrossBCImport` function currently allows any path containing `/publishedlanguage`. Update it to exclude `/publishedlanguage/contracts`:

```go
func isAllowedCrossBCImport(ownerBC, importSuffix string) bool {
    // ... existing checks ...

    // Allow publishedlanguage imports, but NOT contracts sub-package
    if strings.Contains(importSuffix, "/publishedlanguage/contracts") {
        return false
    }
    return strings.Contains(importSuffix, "/publishedlanguage")
}
```

Since the architecture test already skips `_test.go` files, this naturally allows contract imports in tests for verification while blocking them in production code.

**Rule: Published language framework purity.**
Add a new test `TestPublishedLanguagePurity` that scans all files in `*/publishedlanguage/contracts/` packages and verifies they only import stdlib packages (no domain, infrastructure, or framework imports).

## Out of Scope
- Typing every event constant as a custom `EventType` (follow-up if event bus APIs are updated).

## Acceptance Criteria
- No BC with active cross-BC integration lacks a `publishedlanguage` package.
- `auth/publishedlanguage` has zero dependency on internal domain framework interfaces.
- `publishedlanguage/contracts/` sub-packages only depend on stdlib.
- Architecture test blocks cross-BC imports of `publishedlanguage/contracts` in production code.
- Architecture test verifies `publishedlanguage/contracts/` framework purity.
- All existing architecture and import tests pass.

## Verification
- `go test ./internal -run Architecture`
- `go test ./internal/...`
- Manually verify that adding a cross-BC `publishedlanguage/contracts` import in a projector causes a test failure.
