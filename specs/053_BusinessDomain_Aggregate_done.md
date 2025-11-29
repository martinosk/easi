# Business Domain Aggregate

## Description
Business domains are strategic groupings of L1 capabilities that represent business areas like "Finance", "Customer Experience", or "Operations".

## Purpose
Enable enterprise architects to organize L1 capabilities into meaningful business domains for strategic planning and visualization.

## Dependencies
- Capability aggregate (already exists)

## Aggregate Root: BusinessDomain

**Properties:**
- `ID` (BusinessDomainID value object): Unique identifier.
- `Name` (DomainName value object): Unique domain name
- `Description` (Description value object): Purpose and scope of the domain
- `CreatedAt` (timestamp)
- `UpdatedAt` (timestamp)

**Invariants:**
- Name must be unique across all business domains
- Name must not be empty and max 100 characters
- Description max 500 characters

## Commands

### CreateBusinessDomain
Creates a new business domain.

**Properties:**
- `Name` (string, required): Domain name
- `Description` (string, optional): Domain description

**Validation:**
- Name must not be empty
- Name must be unique (check via repository)
- Name max 100 characters
- Description max 500 characters if provided

### UpdateBusinessDomain
Updates an existing business domain.

**Properties:**
- `ID` (Guid, required): Domain identifier
- `Name` (string, required): Updated name
- `Description` (string, optional): Updated description

**Validation:**
- Domain must exist
- Name must not be empty
- Name must be unique (excluding self)
- Name max 100 characters
- Description max 500 characters if provided

### DeleteBusinessDomain
Deletes a business domain.

**Properties:**
- `ID` (Guid, required): Domain identifier

**Validation:**
- Domain must exist
- Domain must have no associated capabilities (check via read model)

## Events

### BusinessDomainCreated
Raised when a new business domain is created.

**Properties:**
- `ID` (Guid): Domain identifier
- `Name` (string): Domain name
- `Description` (string): Domain description
- `CreatedAt` (timestamp): Creation time

### BusinessDomainUpdated
Raised when a business domain is updated.

**Properties:**
- `ID` (Guid): Domain identifier
- `Name` (string): Updated name
- `Description` (string): Updated description
- `UpdatedAt` (timestamp): Update time

### BusinessDomainDeleted
Raised when a business domain is deleted.

**Properties:**
- `ID` (Guid): Domain identifier
- `DeletedAt` (timestamp): Deletion time

## Value Objects

### BusinessDomainID
Immutable GUID wrapper with "bd-" prefix for business domain identifiers.

**Validation:**
- Must be valid GUID format
- Must have "bd-" prefix

### DomainName
Immutable string wrapper for domain names.

**Validation:**
- Must not be empty or whitespace
- Must not exceed 100 characters
- Trims leading/trailing whitespace

## Repository Interface

```
GetByID(ctx, id) -> BusinessDomain
GetByName(ctx, name) -> BusinessDomain
NameExists(ctx, name, excludeID) -> bool
Save(ctx, aggregate) -> error
```

## Implementation Notes
- Business Domain is part of the CapabilityMapping bounded context
- No direct coupling to Capability aggregate - associations managed separately
- Uniqueness of name enforced at command handler level via repository check
- Domain deletion validation depends on read model query for associated capabilities
