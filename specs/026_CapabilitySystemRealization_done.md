# Capability System Realization

## Description
Link capabilities to application components that technically realize them, enabling gap analysis and alignment tracking.

## Purpose
Show which systems implement each capability, identify realization gaps, and track alignment between business capabilities and technical systems.

## Dependencies
- Spec 023: Capability Model (must be completed first)
- Spec 002: Application Component (already done)

## Command

### LinkSystemToCapability
Links application component to capability.

**Properties:**
- `CapabilityId` (Guid, required)
- `ComponentId` (Guid, required)
- `RealizationLevel` (enum, required): Full, Partial, Planned
- `Notes` (string, optional)

**Validation:**
- Capability must exist
- Component must exist
- No duplicate links

### UpdateSystemRealization
Updates realization details.

**Properties:**
- `Id` (Guid, required)
- `RealizationLevel` (enum, required)
- `Notes` (string, optional)

## Events

### SystemLinkedToCapability
**Properties:** Id, CapabilityId, ComponentId, RealizationLevel, Notes, LinkedAt

### SystemRealizationUpdated
**Properties:** Id, RealizationLevel, Notes

## API Endpoints

### POST /api/capabilities/{capabilityId}/systems
Links system to capability.

### GET /api/capabilities/{id}/systems
Lists systems realizing capability.

### GET /api/application-component/{id}/capabilities
Lists capabilities realized by system.

### PUT /api/capability-realizations/{id}
Updates realization.

### DELETE /api/capability-realizations/{id}
Removes link.

## Domain Model

### CapabilityRealization Aggregate
- Properties: Id, CapabilityId, ComponentId, RealizationLevel, Notes

### Value Objects
- `RealizationId`: Guid wrapper
- `RealizationLevel`: Enum (Full, Partial, Planned)

## Integration
Uses cross-context references by ComponentId only (no direct coupling to ArchitectureModeling context).

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [x] Documentation updated if needed
- [ ] User sign-off
