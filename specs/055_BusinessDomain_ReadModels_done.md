# Business Domain Read Models

## Description
Read models for querying business domains and their capability associations efficiently.

## Purpose
Enable fast queries for domain composition, capability assignments, and orphaned capabilities without reconstructing aggregates from event streams.

## Dependencies
- Spec 053: Business Domain Aggregate
- Spec 054: Business Domain Assignment Aggregate

## Read Model: BusinessDomainDTO

**Purpose:** List and display business domains

**Properties:**
- `ID` (string): Domain identifier
- `Name` (string): Domain name
- `Description` (string): Domain description
- `CapabilityCount` (int): Number of L1 capabilities in this domain
- `CreatedAt` (timestamp): Creation time
- `UpdatedAt` (timestamp): Last update time

**Projected from events:**
- `BusinessDomainCreated` - Creates new row
- `BusinessDomainUpdated` - Updates name, description, updatedAt
- `BusinessDomainDeleted` - Removes row
- `CapabilityAssignedToDomain` - Increments CapabilityCount
- `CapabilityUnassignedFromDomain` - Decrements CapabilityCount

**Queries:**
- `GetAll()` - List all domains with pagination
- `GetByID(id)` - Get single domain
- `GetByName(name)` - Find domain by name

## Read Model: DomainCapabilityAssignmentDTO

**Purpose:** Track which capabilities belong to which domains

**Properties:**
- `AssignmentID` (string): Assignment identifier
- `BusinessDomainID` (string): Domain identifier
- `BusinessDomainName` (string): Domain name (denormalized)
- `CapabilityID` (string): Capability identifier
- `CapabilityCode` (string): Capability code (denormalized)
- `CapabilityName` (string): Capability name (denormalized)
- `CapabilityLevel` (string): Should always be "L1"
- `AssignedAt` (timestamp): Assignment time

**Projected from events:**
- `CapabilityAssignedToDomain` - Creates new row with denormalized data
- `CapabilityUnassignedFromDomain` - Removes row
- `BusinessDomainUpdated` - Updates denormalized domain name
- `CapabilityUpdated` - Updates denormalized capability data

**Queries:**
- `GetCapabilitiesForDomain(domainID)` - List capabilities in domain with pagination
- `GetDomainsForCapability(capabilityID)` - List domains containing capability
- `GetAssignment(domainID, capabilityID)` - Check if assignment exists
- `GetAssignmentsForDomain(domainID)` - Get all assignments for deletion validation

## Read Model: DomainCompositionDTO

**Purpose:** Display full capability hierarchy within a business domain

**Properties:**
- `BusinessDomainID` (string): Domain identifier
- `L1CapabilityID` (string): Root L1 capability
- `L1CapabilityCode` (string): L1 capability code
- `L1CapabilityName` (string): L1 capability name
- `ChildCapabilities` (JSON): Nested L2/L3/L4 hierarchy
- `RealizingSystems` (JSON): Systems implementing these capabilities
- `AssignedAt` (timestamp): When added to domain

**Projected from events:**
- `CapabilityAssignedToDomain` - Adds L1 with full child hierarchy
- `CapabilityUnassignedFromDomain` - Removes L1 and children
- `CapabilityCreated` (child) - Updates ChildCapabilities JSON
- `CapabilityUpdated` - Updates denormalized data
- `ComponentToCapabilityAssociated` - Updates RealizingSystems JSON

**Queries:**
- `GetDomainComposition(domainID)` - Full hierarchy for visualization

## Read Model: UnassignedCapabilityDTO

**Purpose:** Identify L1 capabilities without business domain assignment

**Properties:**
- `CapabilityID` (string): Capability identifier
- `CapabilityCode` (string): Capability code
- `CapabilityName` (string): Capability name
- `CapabilityLevel` (string): Should always be "L1"
- `CreatedAt` (timestamp): Capability creation time

**Projected from events:**
- `CapabilityCreated` (level L1) - Adds to list
- `CapabilityDeleted` - Removes from list
- `CapabilityAssignedToDomain` - Removes from list (now assigned)
- `CapabilityUnassignedFromDomain` - Re-adds if no other domains (check count)

**Queries:**
- `GetUnassignedL1Capabilities()` - List orphaned L1 capabilities with pagination

## Database Schema

### business_domains table
```
id              UUID PRIMARY KEY
name            VARCHAR(100) UNIQUE NOT NULL
description     VARCHAR(500)
capability_count INTEGER DEFAULT 0
created_at      TIMESTAMPTZ NOT NULL
updated_at      TIMESTAMPTZ
```

### domain_capability_assignments table
```
assignment_id          UUID PRIMARY KEY
business_domain_id     UUID NOT NULL
business_domain_name   VARCHAR(100) NOT NULL
capability_id          UUID NOT NULL
capability_code        VARCHAR(50) NOT NULL
capability_name        VARCHAR(200) NOT NULL
capability_level       VARCHAR(2) NOT NULL
assigned_at            TIMESTAMPTZ NOT NULL

UNIQUE(business_domain_id, capability_id)
INDEX idx_dca_domain (business_domain_id)
INDEX idx_dca_capability (capability_id)
```

### domain_composition_view table
```
business_domain_id    UUID NOT NULL
l1_capability_id      UUID NOT NULL
l1_capability_code    VARCHAR(50) NOT NULL
l1_capability_name    VARCHAR(200) NOT NULL
child_capabilities    JSONB
realizing_systems     JSONB
assigned_at           TIMESTAMPTZ NOT NULL

PRIMARY KEY (business_domain_id, l1_capability_id)
INDEX idx_dc_domain (business_domain_id)
```

## Implementation Notes
- All read models are eventually consistent with event streams
- Denormalization used strategically to avoid joins in API queries
- JSONB columns for hierarchical data to simplify queries
- No foreign key constraints as per event sourcing principles
- Queries support cursor-based pagination using opaque tokens
