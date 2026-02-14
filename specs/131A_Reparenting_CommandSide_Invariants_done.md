# Reparenting Command-Side Invariants and Events

## Description
Enforce capability reparenting invariants in the command-side domain model and emit explicit events that describe reparenting and realization inheritance changes.

## Requirements
- Reparenting invariants are enforced in aggregates or a dedicated domain service
- Command handler emits a reparenting event with old and new parent identifiers
- Command handler emits explicit inheritance change events that describe additions and removals
- Event payloads include enough data for projectors to update read models without deriving changes
- Unit tests cover invariant enforcement and event emission
- Integration tests cover command handling if relevant

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] Documentation updated if needed
- [x] User sign-off
