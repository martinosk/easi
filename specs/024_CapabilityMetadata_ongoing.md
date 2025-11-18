# Capability Metadata

## Description
Extend capabilities with strategic alignment, maturity, ownership, and status metadata.

## Purpose
Enable tracking of capability strategic importance, maturity progression, ownership model, and lifecycle status.

## Dependencies
- Spec 023: Capability Model (must be completed first)

## Command

### UpdateCapabilityMetadata
Updates capability metadata fields.

**Properties:**
- `Id` (Guid, required)
- `StrategyPillar` (enum, optional): AlwaysOn, Grow, Transform
- `PillarWeight` (int, optional): 0-100% allocation to pillar
- `MaturityLevel` (enum, required): Initial, Developing, Defined, Managed, Optimizing
- `OwnershipModel` (enum, required): TribeOwned, TeamOwned, Shared, EnterpriseService
- `PrimaryOwner` (string, required): Tribe/Team + individual name
- `EAOwner` (string, required): Enterprise Architect responsible
- `Status` (enum, required): Active, Planned, Deprecated

### AddCapabilityExpert
Adds subject matter expert to capability.

**Properties:**
- `CapabilityId` (Guid, required)
- `ExpertName` (string, required)
- `ExpertRole` (string, required)
- `ContactInfo` (string, required)

### AddCapabilityTag
Adds custom tag to capability.

**Properties:**
- `CapabilityId` (Guid, required)
- `Tag` (string, required): e.g., 'Legacy', 'API-first', 'Cloud-native'

## Events

### CapabilityMetadataUpdated
**Properties:** Id, StrategyPillar, PillarWeight, MaturityLevel, OwnershipModel, PrimaryOwner, EAOwner, Status

### CapabilityExpertAdded
**Properties:** CapabilityId, ExpertName, ExpertRole, ContactInfo, AddedAt

### CapabilityTagAdded
**Properties:** CapabilityId, Tag, AddedAt

## API Endpoints

### PUT /api/capabilities/{id}/metadata
Updates capability metadata.

### POST /api/capabilities/{id}/experts
Adds expert.

### GET /api/capabilities/{id}/experts
Lists experts.

### POST /api/capabilities/{id}/tags
Adds tag.

### GET /api/capabilities/{id}/tags
Lists tags.

## Domain Model

### Value Objects
- `StrategyPillar`: Enum (AlwaysOn, Grow, Transform)
- `PillarWeight`: Int (0-100) with validation
- `MaturityLevel`: Enum (Initial, Developing, Defined, Managed, Optimizing)
- `OwnershipModel`: Enum (TribeOwned, TeamOwned, Shared, EnterpriseService)
- `Owner`: String wrapper with validation
- `CapabilityStatus`: Enum (Active, Planned, Deprecated)
- `Expert`: Entity with Name, Role, Contact
- `Tag`: String wrapper

## Checklist
- [x] Specification ready
- [x] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [x] API Documentation updated in OpenAPI specification
- [ ] User sign-off
