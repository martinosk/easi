# Spec 088: Generic Value Object Base Types

## Overview
Eliminate duplication of value object implementations across bounded contexts by creating generic base types in the shared domain layer.

## Problem Statement
The codebase contains significant value object duplication:
- `Description` value object is 100% identical across 3 bounded contexts (architecturemodeling, architectureviews, capabilitymapping)
- UUID-based ID value objects (`ComponentID`, `ViewID`, `CapabilityID`, `RelationID`, `BusinessDomainID`, etc.) follow identical patterns with ~50 lines each
- Each ID type implements: `New*()`, `New*FromString()`, `Value()`, `Equals()`, `String()`

**Estimated duplicate code**: ~500+ lines across ID types alone

## Requirements

### R1: Generic UUID Identifier Base
Create a type-safe generic base for UUID-based identifiers:
- Must provide `New()` for generating new UUIDs
- Must provide `FromString()` for parsing existing UUIDs with validation
- Must implement `Value() string`, `Equals()`, and `String()` methods
- Must return appropriate domain errors (`ErrEmptyValue`, `ErrInvalidValue`)
- Must maintain type safety (a `ComponentID` cannot be assigned to `ViewID`)

### R2: Shared Description Value Object
Move `Description` to `shared/domain/valueobjects/`:
- Must support empty descriptions (optional field)
- Must trim whitespace on construction
- Must implement `Value()`, `IsEmpty()`, `Equals()`, `String()`
- All bounded contexts must import from shared location

### R3: Bounded Context Wrapper Structs
Each bounded context must define wrapper structs (NOT bare type aliases):
- Wrapper structs preserve each context's ability to evolve independently
- Maintains semantic meaning within the context
- Allows future context-specific behavior without breaking encapsulation
- Example:
```go
// Correct - wrapper struct allows future extension
type ComponentID struct {
    shared.UUIDIdentifier
}

// Incorrect - bare type alias prevents future customization
type ComponentID = shared.UUIDIdentifier[componentIDMarker]
```

### R4: Migration Strategy
- New code must use the shared types
- Existing code can be migrated incrementally per bounded context
- No functional changes to API contracts

## Affected Files

### Files to Create
- `backend/internal/shared/domain/valueobjects/uuid_identifier.go`
- `backend/internal/shared/domain/valueobjects/description.go` (move from context)

### Files to Modify (examples per context)
- `backend/internal/architecturemodeling/domain/valueobjects/component_id.go` (simplify to alias/wrapper)
- `backend/internal/architectureviews/domain/valueobjects/view_id.go` (simplify to alias/wrapper)
- `backend/internal/capabilitymapping/domain/valueobjects/capability_id.go` (simplify to alias/wrapper)
- All other ID value objects across contexts

### Files to Delete (after migration)
- `backend/internal/architecturemodeling/domain/valueobjects/description.go`
- `backend/internal/architectureviews/domain/valueobjects/description.go`
- `backend/internal/capabilitymapping/domain/valueobjects/description.go`

## Checklist
- [x] Specification approved
- [x] Generic UUID identifier base created with tests
- [x] Shared Description value object created with tests
- [x] architecturemodeling context migrated
- [x] architectureviews context migrated
- [x] capabilitymapping context migrated
- [x] viewlayouts context migrated
- [x] auth context migrated
- [x] All existing tests pass
- [x] User sign-off
