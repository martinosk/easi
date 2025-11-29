# Business Domain Assignment Aggregate

## Description
Manages the many-to-many association between business domains and L1 capabilities.

## Purpose
Enable architects to assign L1 capabilities to one or more business domains and track these associations with full audit trail through event sourcing.

## Dependencies
- Spec 053: Business Domain Aggregate
- Capability aggregate (already exists)

## Aggregate Root: BusinessDomainAssignment

**Properties:**
- `ID` (AssignmentID value object): Unique identifier with "assign-" prefix
- `BusinessDomainID` (BusinessDomainID value object): Reference to business domain
- `CapabilityID` (CapabilityID value object): Reference to L1 capability
- `AssignedAt` (timestamp): When the assignment was created

**Invariants:**
- Only L1 capabilities can be assigned to business domains
- No duplicate assignments (same domain + capability pair)
- Both domain and capability must exist at assignment time

## Commands

### AssignCapabilityToDomain
Assigns an L1 capability to a business domain.

**Properties:**
- `BusinessDomainID` (Guid, required): Target business domain
- `CapabilityID` (Guid, required): Capability to assign
- `CapabilityLevel` (string, required): Capability level for validation

**Validation:**
- Business domain must exist (repository check)
- Capability must exist (repository check)
- Capability must be L1 level
- Assignment must not already exist (repository check)

### UnassignCapabilityFromDomain
Removes a capability from a business domain.

**Properties:**
- `BusinessDomainID` (Guid, required): Business domain
- `CapabilityID` (Guid, required): Capability to remove

**Validation:**
- Assignment must exist

## Events

### CapabilityAssignedToDomain
Raised when an L1 capability is assigned to a business domain.

**Properties:**
- `AssignmentID` (Guid): Unique assignment identifier
- `BusinessDomainID` (Guid): Business domain identifier
- `CapabilityID` (Guid): Capability identifier
- `AssignedAt` (timestamp): Assignment time

### CapabilityUnassignedFromDomain
Raised when a capability is removed from a business domain.

**Properties:**
- `AssignmentID` (Guid): Assignment identifier
- `BusinessDomainID` (Guid): Business domain identifier
- `CapabilityID` (Guid): Capability identifier
- `UnassignedAt` (timestamp): Unassignment time

## Value Objects

### AssignmentID
Immutable GUID wrapper with "assign-" prefix for assignment identifiers.

**Validation:**
- Must be valid GUID format
- Must have "assign-" prefix

## Repository Interface

```
GetByDomainAndCapability(ctx, domainID, capabilityID) -> BusinessDomainAssignment
AssignmentExists(ctx, domainID, capabilityID) -> bool
Save(ctx, aggregate) -> error
```

## Event Handlers

### OnCapabilityDeleted
When a capability is deleted, automatically unassign it from all business domains.

**Action:**
- Query all assignments for the deleted capability
- Issue UnassignCapabilityFromDomain command for each assignment

### OnBusinessDomainDeleted
When a business domain is deleted, automatically unassign all capabilities.

**Action:**
- Query all assignments for the deleted domain
- Issue UnassignCapabilityFromDomain command for each assignment

## Implementation Notes
- Assignment is a first-class aggregate with its own event stream
- L1-only restriction enforced at command validation level
- Event handlers maintain eventual consistency across aggregates
- No cascading deletes - explicit unassignment commands maintain audit trail
