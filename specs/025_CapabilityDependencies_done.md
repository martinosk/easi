# Capability Dependencies

## Description
Model dependencies between capabilities to understand cross-domain relationships.

## Purpose
Enable visualization and analysis of how capabilities depend on each other across business domains.

## Dependencies
- Spec 023: Capability Model (must be completed first)

## Command

### CreateCapabilityDependency
Creates dependency between two capabilities.

**Properties:**
- `SourceCapabilityId` (Guid, required): Dependent capability
- `TargetCapabilityId` (Guid, required): Capability depended upon
- `DependencyType` (enum, required): Requires, Enables, Supports
- `Description` (string, optional): Nature of dependency

**Validation:**
- Both capabilities must exist
- Cannot create self-dependency
- No duplicate dependencies

## Events

### CapabilityDependencyCreated
**Properties:** Id, SourceCapabilityId, TargetCapabilityId, DependencyType, Description, CreatedAt

## API Endpoints

### POST /api/capability-dependencies
Creates dependency with HATEOAS links.

### GET /api/capability-dependencies
Lists all dependencies.

### GET /api/capabilities/{id}/dependencies/outgoing
Lists capabilities this one depends on.

### GET /api/capabilities/{id}/dependencies/incoming
Lists capabilities that depend on this one.

### DELETE /api/capability-dependencies/{id}
Removes dependency.

## Domain Model

### CapabilityDependency Aggregate
- Properties: Id, SourceCapabilityId, TargetCapabilityId, DependencyType, Description

### Value Objects
- `DependencyId`: Guid wrapper
- `DependencyType`: Enum (Requires, Enables, Supports)

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [x] Documentation updated if needed
- [ ] User sign-off
