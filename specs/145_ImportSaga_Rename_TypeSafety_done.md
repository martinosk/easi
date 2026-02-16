# Import Process Manager — Clarify Guarantees and Type Safety

## Status: done

## Problem

The import saga (process manager) has two issues:

1. **Implicit failure semantics**: The saga uses best-effort continuation (skip failed items, keep going), but this is not documented or systematically tested across all phases.
2. **Raw string IDs flow between contexts**: Gateway interfaces use `string` for all IDs. While the current layering is correct (adapters translate to commands, command handlers validate), the saga's internal state (`sagaState`) treats all mapped IDs as untyped strings, making it easy to mix up a component ID with a capability ID.

## Decision

Fix these in two separate, focused steps. **Do NOT change the gateway port interfaces** — those are the Anti-Corruption Layer boundary and correctly use `string`. Type safety belongs inside the saga's internal state, not at the cross-context port boundary.

## What NOT To Do

- **Do NOT import other contexts' domain value objects or published language types into the importing context's ports.** The gateway interfaces (`ComponentGateway`, `CapabilityGateway`, `ValueStreamGateway`) must remain context-neutral. They return `string` IDs because that is what the ACL adapters produce after dispatching commands.
- **Do NOT create `publishedlanguage/ids.go` files** that re-export domain value objects as type aliases. The existing published language pattern in this codebase exposes event constants and behavioral contracts, not domain types.
- **Do NOT add `isZero*` helper functions** to check for empty value objects. If zero-value detection is needed, the type system should prevent zero values from being stored in the first place.

## Required Changes

### Step 1: Typed wrapper IDs inside the saga (internal only)

Define thin ID wrapper types **inside the importing context** (e.g., `importing/application/saga/ids.go`) that distinguish between different kinds of mapped IDs at compile time:

```go
package saga

type mappedComponentID string
type mappedCapabilityID string
type mappedValueStreamID string
type mappedStageID string
type mappedRelationID string
```

These are the importing context's own vocabulary for "an ID we received back from a gateway call." They are **not** domain value objects from other contexts. They are simple string wrappers that prevent accidental mixing.

Update `sagaState` to use these types:

```go
type sagaState struct {
    sourceToComponentID   map[string]mappedComponentID
    sourceToCapabilityID  map[string]mappedCapabilityID
    sourceToValueStreamID map[string]mappedValueStreamID
    sourceToStageID       map[string]mappedStageID
    createdCapabilityIDs  []mappedCapabilityID
}
```

Gateway calls return `string` — wrap at the point of storage:

```go
id, err := s.components.CreateComponent(ctx, comp.Name, comp.Description)
// ...
state.sourceToComponentID[comp.SourceID] = mappedComponentID(id)
```

Gateway calls that accept IDs — unwrap at the point of use:

```go
s.capabilities.LinkSystem(ctx, string(capabilityID), string(componentID), ...)
```

Empty-check uses the natural string zero value: `mappedComponentID("") == ""` is `false` for any successfully mapped ID.

### Step 2: Failure-path test coverage

Add tests that verify best-effort continuation across phases. The `TestImportSaga_FailureIsBestEffortAndKeepsSuccessfulProgress` test already covers component failures (added separately). Extend coverage to:

- A capability creation failure does not prevent subsequent value stream creation.
- A value stream failure does not prevent subsequent realization creation.
- A realization failure is recorded but does not prevent domain assignment.

These tests document the process manager's guarantee: **best-effort continuation with error accumulation, no compensating rollback.**

## Scope

- `backend/internal/importing/application/saga/` — saga implementation and new ID types
- `backend/internal/importing/application/saga/*_test.go` — failure-path tests

**Nothing outside the importing context is touched.**

## Out of Scope

- Renaming `saga` to `processmanager` (low value, high churn)
- Changing gateway port interfaces
- Published language changes in other contexts

## Acceptance Criteria

- `sagaState` maps use distinct wrapper types — compiler rejects mixing a `mappedComponentID` where a `mappedCapabilityID` is expected.
- Gateway interfaces remain unchanged (`string` IDs).
- Failure-path tests cover at least 3 distinct phase-failure scenarios.
- All existing tests pass: `go test ./internal/importing/...`

## Verification

```bash
go test ./internal/importing/...
go vet ./internal/importing/...
```
