# Maturity Scale Configuration Aggregate

## Description
Implement the MetaModelConfiguration aggregate with MaturityScaleConfig as its first configurable element. This aggregate manages tenant-specific configuration for how the maturity scale is structured.

## Purpose
Allow tenants to customize the maturity scale section names and boundaries while maintaining data integrity through immutable value objects and event sourcing.

## Dependencies
- Spec 090: MetaModel Bounded Context (parent context definition)
- Spec 089: GenericEventSourcedRepository

## Aggregate: MetaModelConfiguration

### Identity
- Aggregate ID: `MetaModelConfigurationID` (UUID-based value object)
- Reference: `TenantID` (1:1 relationship, used for lookup but not as identity)

### Architectural Decision: Focused Aggregate Scope
This aggregate focuses solely on maturity scale configuration. Future meta-model extensions (strategy pillars, element types, relationship rules) will be implemented as **separate aggregates** to prevent a "God Aggregate" anti-pattern and reduce contention.

### Creation
Created automatically when a tenant is provisioned (via TenantCreated event handler).

**Command**: (internal - triggered by event handler)
```
CreateMetaModelConfiguration
- id: MetaModelConfigurationID (required, generated UUID)
- tenantID: TenantID (required, from Platform context)
- maturityScaleConfig: MaturityScaleConfig (defaults to standard Wardley mapping)
```

**Event**: MetaModelConfigurationCreated

### Update Maturity Scale

**Command**: UpdateMaturityScale
```
- id: MetaModelConfigurationID (aggregate ID)
- sections: []MaturitySectionInput (exactly 4)
- version: int (for optimistic locking)
```

**MaturitySectionInput**:
```
- order: int (1-4)
- name: string (non-empty, max 50 chars)
- minValue: int (0-99)
- maxValue: int (0-99)
```

**Validation Rules** (enforced in value objects):
- Exactly 4 sections required
- Each section has unique order (1, 2, 3, 4)
- Section names are non-empty, max 50 characters
- minValue and maxValue must be 0-99
- maxValue >= minValue within each section
- Sections must be contiguous: section[n].maxValue + 1 = section[n+1].minValue
- First section must start at 0
- Last section must end at 99

**Event**: MaturityScaleConfigUpdated

### Reset to Defaults

**Command**: ResetMaturityScale
```
- id: MetaModelConfigurationID (aggregate ID)
```

**Behavior**: Restores the maturity scale to default configuration (Genesis, Custom Built, Product, Commodity with 25-point ranges).

**Event**: MaturityScaleConfigReset

## Value Objects

### MaturityScaleConfig
Immutable configuration containing exactly 4 sections.

**Properties:** Fixed array of 4 MaturitySection

**Behaviors:**
- Constructor validates section contiguity and coverage
- Factory method for creating default configuration
- Lookup section by numeric value (0-99)
- Equality comparison

### MaturitySection
Single section of the maturity scale.

**Properties:**
- `order`: 1-4
- `name`: SectionName
- `minValue`: MaturityValue
- `maxValue`: MaturityValue

**Behaviors:**
- Constructor validates order range and min <= max
- Check if a value falls within section range

### SectionName
**Validation:** Non-empty, max 50 characters, trimmed whitespace.

### MaturityValue
**Validation:** Must be 0-99 inclusive.

## Read Model: MaturityScaleReadModel

### Table: maturity_scale_configs
```sql
CREATE TABLE maturity_scale_configs (
    id VARCHAR(50) PRIMARY KEY,           -- MetaModelConfigurationID
    tenant_id VARCHAR(50) NOT NULL UNIQUE, -- 1:1 relationship with tenant
    section_1_name VARCHAR(50) NOT NULL,
    section_1_min INT NOT NULL,
    section_1_max INT NOT NULL,
    section_2_name VARCHAR(50) NOT NULL,
    section_2_min INT NOT NULL,
    section_2_max INT NOT NULL,
    section_3_name VARCHAR(50) NOT NULL,
    section_3_min INT NOT NULL,
    section_3_max INT NOT NULL,
    section_4_name VARCHAR(50) NOT NULL,
    section_4_min INT NOT NULL,
    section_4_max INT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    is_default BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL,
    modified_at TIMESTAMPTZ NOT NULL,
    modified_by VARCHAR(255)
);

CREATE INDEX idx_maturity_scale_configs_tenant ON maturity_scale_configs(tenant_id);
```

### Projector: MaturityScaleProjector

Handles events:
- `MetaModelConfigurationCreated` → INSERT row
- `MaturityScaleConfigUpdated` → UPDATE row, set is_default = false
- `MaturityScaleConfigReset` → UPDATE row, set is_default = true

## Event Handlers

### TenantCreatedHandler (in MetaModel context)

Subscribes to `TenantCreated` event from Platform context.

**Behavior**:
1. Check if configuration already exists for tenant (idempotency)
2. Create default MaturityScaleConfig
3. Create MetaModelConfiguration aggregate with tenant ID
4. Save aggregate (triggers MetaModelConfigurationCreated event)

## Repository

Uses generic event-sourced repository pattern from Spec 089.

**Required Operations:**
- `GetByID`: Load aggregate by MetaModelConfigurationID (for command handling)
- `GetByTenantID`: Load aggregate by TenantID (for idempotency checks during provisioning)
- `Save`: Persist aggregate with new events

## Checklist
- [x] Specification ready
- [x] Value objects implemented with validation
- [x] Aggregate implemented with event sourcing
- [x] Events defined and serializable
- [x] Command handlers implemented
- [x] TenantCreated event handler implemented
- [x] Projector implemented
- [x] Read model table migration created
- [x] Repository implemented
- [x] Unit tests for value objects
- [x] Unit tests for aggregate
- [x] Integration tests for repository
- [x] User sign-off
