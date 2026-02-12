# Realization Inheritance Read Model Refactor and Migration

## Description
Refactor read-model projectors to apply explicit inheritance events without business logic and clean up existing invalid inherited realizations.

## Requirements
- Projectors apply explicit inheritance change events without deriving additional realizations
- Remove inheritance inference logic from projector implementations
- Read models remain consistent with command-side event payloads
- Data migration removes invalid inherited realizations and repairs inconsistent lineage
- Integration tests cover projector behavior for inheritance change events

## Checklist
- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] User sign-off
