# Strategy Pillars Meta Model

## Description
Implement user-definable strategy pillars as a meta model configuration. Strategy pillars allow organizations to categorize business capabilities by strategic alignment. Each capability can be assigned to multiple strategy pillars **within the context of a specific business domain**, with individual strategic importance ratings.

## Purpose
Replace the hardcoded strategy pillar enum (AlwaysOn, Grow, Transform) with user-definable pillars. Enable capabilities to align with multiple strategic initiatives, where strategic importance is scoped per business domain.

## Dependencies
- Spec 090: MetaModel Bounded Context
- Spec 089: GenericEventSourcedRepository
- Spec 091: MaturityScale Aggregate (similar pattern)
- Spec 053: Business Domain Aggregate
- Spec 054: Business Domain Assignment Aggregate

## Related Specs
- Spec 099: Strategy Pillars REST API
- Spec 100: Enterprise Capability (enterprise-scoped strategy alignment)

## Strategic Design Decisions

### Bounded Context Placement
- **Pillar Definitions**: Metamodel bounded context (configuration/taxonomy)
- **Pillar Assignments**: Capability Mapping bounded context (operational data)

### Domain-Scoped Strategic Importance

**Key Insight**: Strategic importance is a **domain-specific evaluation**, not a global property. The same capability can have different strategic importance for different business domains.

**Example**:
- Capability: "Customer Data Management"
- Digital Banking Domain: Critical (5) for "Transform" pillar (heavy consumer, strategic differentiator)
- Traditional Lending Domain: Average (3) for "Transform" pillar (light consumer, hygiene factor)

**Rationale**:
1. Strategic value is fundamentally contextual per domain
2. Different organizational units have different strategic priorities
3. Enables domain-specific portfolio management
4. Supports federated governance (domain teams manage their own priorities)
5. Avoids false precision of forced global consensus

### Relationship to BusinessDomainAssignment

These are **orthogonal concerns**:
- `BusinessDomainAssignment`: **Ownership/Responsibility** - "Which domain owns this capability?"
- `DomainCapabilityStrategyAlignment`: **Strategic Value** - "How important is this capability for achieving domain-specific strategic objectives?"

A domain might:
- Own a capability but rate it as low importance (legacy maintenance)
- Not own a capability but rate its strategic importance if they consume it
- Own a capability and rate it as critically important (core competitive advantage)

### Domain-Scoped vs Enterprise-Scoped Strategy Alignment

This specification covers **domain-scoped** strategy alignment. For **enterprise-scoped** alignment, see Spec 100.

| Scope | Aggregate | Question Answered | Example |
|-------|-----------|-------------------|---------|
| **Domain-scoped** (this spec) | DomainCapabilityStrategyAlignment | "How important is THIS capability for THIS domain's strategy?" | "Customer Data Management is Critical (5) for Transform in Digital Banking" |
| **Enterprise-scoped** (Spec 100) | EnterpriseCapabilityStrategyAlignment | "Should this capability be standardized across the enterprise?" | "Payroll is Critical (5) for Standardization enterprise-wide" |

**Key Distinction**:
- Domain-scoped alignment is about domain-specific investment priorities
- Enterprise-scoped alignment (especially Standardization pillar) is about cross-domain consolidation

### Migration of Existing Data
The current `StrategyPillar` and `PillarWeight` fields in Capability aggregate will be deprecated. Existing data will be migrated to the new structure.

---

## Part 1: Metamodel - Strategy Pillar Configuration

### Aggregate: StrategyPillarConfiguration

Manages the CRUD of user-defined strategy pillar definitions for a tenant.

#### Identity
- Aggregate ID: `StrategyPillarConfigurationID` (UUID-based value object)
- Reference: `TenantID` (1:1 relationship, used for lookup)

#### Creation
Created automatically when a tenant is provisioned (via TenantCreated event handler), similar to MetaModelConfiguration.

**Command**: (internal - triggered by event handler)
```
CreateStrategyPillarConfiguration
- id: StrategyPillarConfigurationID (required, generated UUID)
- tenantID: TenantID (required, from Platform context)
- pillars: []PillarDefinition (defaults to AlwaysOn, Grow, Transform for migration compatibility)
```

**Event**: StrategyPillarConfigurationCreated

#### Add Pillar

**Command**: AddStrategyPillar
```
- configurationID: StrategyPillarConfigurationID
- name: string (non-empty, max 100 chars)
- description: string (optional, max 500 chars)
- displayOrder: int (positive)
- version: int (for optimistic locking)
```

**Validation Rules**:
- Name must be unique within tenant (case-insensitive)
- Maximum 20 pillars per tenant
- Display order must be unique

**Event**: StrategyPillarAdded

#### Update Pillar

**Command**: UpdateStrategyPillar
```
- configurationID: StrategyPillarConfigurationID
- pillarID: PillarID
- name: string (non-empty, max 100 chars)
- description: string (optional, max 500 chars)
- displayOrder: int (positive)
- version: int (for optimistic locking)
```

**Validation Rules**:
- Pillar must exist
- New name must be unique within tenant (excluding self)

**Event**: StrategyPillarUpdated

#### Remove Pillar

**Command**: RemoveStrategyPillar
```
- configurationID: StrategyPillarConfigurationID
- pillarID: PillarID
- version: int (for optimistic locking)
```

**Validation Rules**:
- Pillar must exist
- Minimum 1 pillar must remain
- Soft delete: marks pillar as inactive rather than physical delete

**Event**: StrategyPillarRemoved

### Value Objects

#### PillarID
UUID-based identifier for individual pillars.

**Validation**: Valid UUID format.

#### PillarName
**Validation**: Non-empty, max 100 characters, trimmed whitespace.

#### PillarDescription
**Validation**: Max 500 characters, allows empty.

#### DisplayOrder
**Validation**: Positive integer (>= 1).

#### PillarDefinition
Immutable value object representing a single strategy pillar.

**Properties:**
- `id`: PillarID
- `name`: PillarName
- `description`: PillarDescription
- `displayOrder`: DisplayOrder
- `isActive`: bool (for soft delete)

### Read Model: strategy_pillar_configs

```sql
CREATE TABLE strategy_pillar_configs (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    modified_at TIMESTAMPTZ NOT NULL,
    modified_by VARCHAR(255),
    CONSTRAINT fk_strategy_pillar_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE UNIQUE INDEX idx_strategy_pillar_configs_tenant ON strategy_pillar_configs(tenant_id);

CREATE TABLE strategy_pillars (
    id VARCHAR(50) PRIMARY KEY,
    configuration_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    display_order INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_strategy_pillar_config FOREIGN KEY (configuration_id)
        REFERENCES strategy_pillar_configs(id) ON DELETE CASCADE
);

CREATE INDEX idx_strategy_pillars_config ON strategy_pillars(configuration_id);
CREATE UNIQUE INDEX idx_strategy_pillars_name ON strategy_pillars(configuration_id, name) WHERE is_active = true;
```

### Projector: StrategyPillarConfigurationProjector

Handles events:
- `StrategyPillarConfigurationCreated` → INSERT config row + default pillars
- `StrategyPillarAdded` → INSERT pillar row
- `StrategyPillarUpdated` → UPDATE pillar row
- `StrategyPillarRemoved` → UPDATE pillar row (set is_active = false)

### Event Handler: TenantCreatedHandler

Subscribes to `TenantCreated` event from Platform context.

**Behavior**:
1. Check if configuration already exists for tenant (idempotency)
2. Create default pillars (AlwaysOn, Grow, Transform) for migration compatibility
3. Create StrategyPillarConfiguration aggregate
4. Save aggregate (triggers StrategyPillarConfigurationCreated event)

---

## Part 2: Capability Mapping - Domain-Scoped Strategy Alignment

### Aggregate: DomainCapabilityStrategyAlignment

Manages the alignment of capabilities to strategy pillars **within the context of a specific business domain**.

#### Identity
- Aggregate ID: `AlignmentID` (UUID-based value object with "dom-strat-" prefix)
- References: `BusinessDomainID`, `CapabilityID`, `PillarID` (by ID only)

#### Align Capability to Pillar (Domain-Scoped)

**Command**: AlignDomainCapabilityToStrategyPillar
```
- businessDomainID: BusinessDomainID (required)
- capabilityID: CapabilityID (required)
- pillarID: PillarID (required)
- strategicImportance: int (1-5, required)
```

**Validation Rules**:
- Business domain must exist
- Capability must exist
- Pillar must exist and be active
- Combination of businessDomainID + capabilityID + pillarID must be unique

**Event**: DomainCapabilityAlignedToStrategyPillar

#### Update Strategic Importance

**Command**: UpdateDomainStrategicImportance
```
- alignmentID: AlignmentID (required)
- strategicImportance: int (1-5, required)
```

**Validation Rules**:
- Alignment must exist
- New importance must differ from current

**Event**: DomainCapabilityStrategyAlignmentUpdated

#### Unalign Capability from Pillar

**Command**: UnalignDomainCapabilityFromStrategyPillar
```
- alignmentID: AlignmentID (required)
```

**Validation Rules**:
- Alignment must exist

**Event**: DomainCapabilityUnalignedFromStrategyPillar

### Value Objects

#### AlignmentID
UUID-based identifier with "dom-strat-" prefix.

**Validation**: Valid UUID format with correct prefix.

#### StrategicImportance
Numeric rating of how important a capability is for a particular strategy pillar within a domain context.

**Validation**: Integer in range 1-5 inclusive.

**Semantic Scale**:
- 1: Low importance
- 2: Below average importance
- 3: Average importance
- 4: Above average importance
- 5: Critical importance

### Read Model: domain_capability_strategy_alignments

```sql
CREATE TABLE domain_capability_strategy_alignments (
    id VARCHAR(50) PRIMARY KEY,
    business_domain_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(50) NOT NULL,
    pillar_id VARCHAR(50) NOT NULL,
    strategic_importance INT NOT NULL CHECK (strategic_importance BETWEEN 1 AND 5),
    aligned_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_alignment_domain FOREIGN KEY (business_domain_id) REFERENCES business_domains(id),
    CONSTRAINT fk_alignment_capability FOREIGN KEY (capability_id) REFERENCES capabilities(id),
    CONSTRAINT fk_alignment_pillar FOREIGN KEY (pillar_id) REFERENCES strategy_pillars(id)
);

CREATE UNIQUE INDEX idx_domain_cap_pillar_unique
    ON domain_capability_strategy_alignments(business_domain_id, capability_id, pillar_id);

CREATE INDEX idx_domain_alignments
    ON domain_capability_strategy_alignments(business_domain_id, strategic_importance DESC);

CREATE INDEX idx_capability_alignments
    ON domain_capability_strategy_alignments(capability_id);

CREATE INDEX idx_pillar_alignments
    ON domain_capability_strategy_alignments(pillar_id, strategic_importance DESC);
```

### Projector: DomainCapabilityStrategyAlignmentProjector

Handles events:
- `DomainCapabilityAlignedToStrategyPillar` → INSERT alignment row
- `DomainCapabilityStrategyAlignmentUpdated` → UPDATE strategic_importance
- `DomainCapabilityUnalignedFromStrategyPillar` → DELETE alignment row

### Event Handlers

#### OnCapabilityDeleted
When a capability is deleted, automatically unalign from all strategy pillars in all domains.

**Action**:
- Query all alignments for the deleted capability
- Issue UnalignDomainCapabilityFromStrategyPillar command for each

#### OnBusinessDomainDeleted
When a business domain is deleted, automatically unalign all capability-pillar relationships for that domain.

**Action**:
- Query all alignments for the deleted domain
- Issue UnalignDomainCapabilityFromStrategyPillar command for each

#### OnStrategyPillarRemoved
When a strategy pillar is removed (soft deleted), keep existing alignments for historical data but prevent new alignments.

---

## Part 3: Anti-Corruption Layer

### Pillar Read Model in Capability Mapping Context

The capability mapping context maintains a local read model of available strategy pillars:

```sql
CREATE TABLE available_strategy_pillars (
    pillar_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INT NOT NULL
);
```

### Event Handler: StrategyPillarEventHandler

Subscribes to metamodel events:
- `StrategyPillarAdded` → INSERT into local cache
- `StrategyPillarUpdated` → UPDATE local cache
- `StrategyPillarRemoved` → UPDATE is_active = false

---

## Part 4: Query Views for Portfolio Management

### Domain Strategic Portfolio View

```sql
CREATE VIEW domain_strategic_portfolio AS
SELECT
    bd.id as domain_id,
    bd.name as domain_name,
    c.id as capability_id,
    c.name as capability_name,
    c.level,
    sp.name as pillar_name,
    dcsa.strategic_importance,
    dcsa.aligned_at
FROM domain_capability_strategy_alignments dcsa
JOIN business_domains bd ON dcsa.business_domain_id = bd.id
JOIN capabilities c ON dcsa.capability_id = c.id
JOIN strategy_pillars sp ON dcsa.pillar_id = sp.id;
```

### Cross-Domain Capability Comparison View

```sql
CREATE VIEW capability_strategic_variance AS
SELECT
    c.id as capability_id,
    c.name as capability_name,
    sp.id as pillar_id,
    sp.name as pillar_name,
    AVG(dcsa.strategic_importance) as avg_importance,
    MIN(dcsa.strategic_importance) as min_importance,
    MAX(dcsa.strategic_importance) as max_importance,
    COUNT(DISTINCT dcsa.business_domain_id) as domain_count
FROM domain_capability_strategy_alignments dcsa
JOIN capabilities c ON dcsa.capability_id = c.id
JOIN strategy_pillars sp ON dcsa.pillar_id = sp.id
GROUP BY c.id, c.name, sp.id, sp.name
HAVING COUNT(DISTINCT dcsa.business_domain_id) > 1;
```

---

## Part 5: Migration Strategy

### Phase 1: Deploy New Infrastructure
1. Create new database tables for strategy pillar configs and alignments
2. Deploy metamodel aggregate and API
3. Deploy capability mapping alignment aggregate and API

### Phase 2: Data Migration
1. For each tenant, create StrategyPillarConfiguration with default pillars matching legacy values
2. For each capability with non-empty strategyPillar:
   - For each BusinessDomainAssignment for that capability:
     - Create DomainCapabilityStrategyAlignment
     - Map pillarWeight (0-100) to strategicImportance (1-5):
       - 0-20 → 1, 21-40 → 2, 41-60 → 3, 61-80 → 4, 81-100 → 5

### Phase 3: Deprecation
1. Mark `strategyPillar` and `pillarWeight` fields in Capability as deprecated
2. Remove from UpdateMetadata command
3. Update frontend to use new alignment APIs

### Phase 4: Cleanup
1. Remove deprecated fields from Capability aggregate
2. Create new event version: CapabilityMetadataUpdatedV2 (without pillar fields)

---

## Governance Model

- **Domain Owners**: Manage alignments for their domain
- **Portfolio Managers**: Get aggregated cross-domain views
- **Capability Owners**: Can see how their capability is valued across domains (insight, not control)

---

## Checklist
- [ ] Specification ready
- [ ] Part 1: Strategy Pillar Configuration
  - [ ] Value objects implemented
  - [ ] Aggregate implemented with event sourcing
  - [ ] Events defined and serializable
  - [ ] Command handlers implemented
  - [ ] TenantCreated event handler
  - [ ] Projector implemented
  - [ ] Read model migration created
  - [ ] Repository implemented
  - [ ] Unit tests
- [ ] Part 2: Domain-Scoped Capability Strategy Alignment
  - [ ] Value objects implemented
  - [ ] Aggregate implemented with event sourcing
  - [ ] Events defined and serializable
  - [ ] Command handlers implemented
  - [ ] Event handlers (OnCapabilityDeleted, OnBusinessDomainDeleted)
  - [ ] Projector implemented
  - [ ] Read model migration created
  - [ ] Repository implemented
  - [ ] Unit tests
- [ ] Part 3: Anti-Corruption Layer
  - [ ] Local pillar cache table
  - [ ] Event handlers for metamodel events
- [ ] Part 4: Query Views
  - [ ] Domain strategic portfolio view
  - [ ] Cross-domain comparison view
- [ ] Part 5: Migration
  - [ ] Migration script for existing data
  - [ ] Deprecation of old fields
- [ ] User sign-off
