# Published Language Event Payload Contracts

## Description
Define stable event payload DTOs in published language packages as the documented serialization contract for cross-BC events. Downstream projectors maintain their own local structs (ACL boundary) but use the published DTOs as the authoritative specification for field names, types, and JSON tags.

## Architectural Intent
Published Language DTOs and ACLs serve complementary but distinct roles:

- **Published Language** documents what the upstream BC serializes. It is the specification.
- **ACL (downstream local structs)** controls what the downstream BC deserializes into. It is the isolation boundary.

Downstream projectors do NOT import upstream PL types at runtime. The PL DTOs exist so that local ACL structs can be written and verified against a documented contract, rather than reverse-engineered from raw JSON. This preserves BC autonomy: if the upstream renames a field, only the downstream ACL adapter changes — not every file that touches the struct.

## Scope (Phase 1)
- Producer BCs with active cross-BC event consumers:
    - `capabilitymapping/publishedlanguage/events.go`
    - `architecturemodeling/publishedlanguage/events.go`
    - `valuestreams/publishedlanguage/events.go` (if consumed cross-BC)

## Out of Scope (Phase 2+)
- Creating DTOs for events with no external consumers yet
- Event versioning framework changes

## Required Changes

### 1) Add payload DTOs next to event constants
For each externally consumed event constant, add an exported payload DTO with explicit JSON tags representing the serialized contract. These DTOs document the wire format and are the single source of truth for what the producer serializes.

### 2) Audit downstream ACL structs against published contracts
Review each downstream projector's local event struct and verify it matches the documented PL DTO (field names, JSON tags, types). Fix any drift. Local structs remain locally owned — they are not replaced by PL imports.

### 3) Keep contracts minimal and stable
Only include fields used by consumers today. Additive evolution only; breaking rename/removal requires a new event version.

## Acceptance Criteria
- Every externally consumed event has a payload DTO in the producer's `publishedlanguage` package.
- Payload DTOs are exported, JSON-tagged, and document the serialization contract.
- Downstream projectors do NOT import upstream `publishedlanguage` payload types — they maintain locally owned ACL structs.
- Local ACL structs are verified to match the documented PL contract (no silent drift).
- All affected cross-BC projector tests pass.

## Verification
- `go test ./internal/...`
- Confirm no downstream projector imports upstream `publishedlanguage` payload DTOs.
- Confirm local ACL structs match the documented PL contracts in field names, JSON tags, and types.
