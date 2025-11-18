# Capability Model

## Description
Implement capability modeling with hierarchical structure (L1-L4 levels) representing business capabilities independent of technical implementation.

## Purpose
Enable enterprise architects to model what the business can do, creating a foundation for capability-based enterprise architecture.

## Command

### CreateCapability
Creates a new capability in the hierarchy.

**Properties:**
- `Name` (string, required): Capability name
- `Description` (string, optional): Purpose and scope
- `ParentId` (Guid, optional): Parent capability for hierarchy
- `Level` (enum, required): L1 (Domain), L2 (Area), L3 (Capability), L4 (Sub-capability)

**Validation:**
- Name must not be empty
- Level must be valid (L1-L4)
- If ParentId provided, parent must exist and be one level above
- L1 capabilities cannot have parents

### UpdateCapability
Updates existing capability.

**Properties:**
- `Id` (Guid, required)
- `Name` (string, required)
- `Description` (string, optional)

## Events

### CapabilityCreated
**Properties:** Id, Name, Description, ParentId, Level, CreatedAt

### CapabilityUpdated
**Properties:** Id, Name, Description

## API Endpoints

### POST /api/capabilities
Creates capability.

**Request:**
```json
{
  "name": "string",
  "description": "string",
  "parentId": "guid (optional)",
  "level": "L1" | "L2" | "L3" | "L4"
}
```

**Response:** 201 Created with HATEOAS links (self, parent, children)

### GET /api/capabilities
Lists all capabilities with hierarchy.

### GET /api/capabilities/{id}
Retrieves specific capability with HATEOAS links.

### GET /api/capabilities/{id}/children
Lists child capabilities.

### PUT /api/capabilities/{id}
Updates capability.

## Domain Model

### Capability Aggregate
- Properties: Id, Name, Description, ParentId, Level, CreatedAt
- All properties as value objects

### Value Objects
- `CapabilityId`: Guid wrapper
- `CapabilityName`: String with validation
- `CapabilityLevel`: Enum (L1, L2, L3, L4)
- Reuse: `Description` from ArchitectureModeling

## Bounded Context
New bounded context: **CapabilityMapping** (core domain)

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [x] API Documentation updated in OpenAPI specification
- [ ] User sign-off
