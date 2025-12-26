# MetaModel Bounded Context

## Description
Introduce a new MetaModel bounded context that manages configurable meta-model elements for the architecture modeling tool. This context controls how the tool behaves and what options are available to users within each tenant.

## Purpose
Enable tenant-specific customization of the tool's modeling concepts (e.g., maturity scales, element types, relationship rules) while maintaining domain integrity and audit trails through event sourcing.

## Strategic Classification
**Supporting Domain** - Enables customization of core domain concepts but is not a differentiator itself.

## Context Relationships

### Upstream Contexts (MetaModel receives from)
- **Platform Context**: Subscribes to `TenantCreated` event to provision default meta-model configuration

### Downstream Contexts (MetaModel publishes to)
- **CapabilityMapping Context**: Publishes `MaturityScaleConfigUpdated` event when maturity scale changes
- **ArchitectureModeling Context** (future): May subscribe to meta-model changes

### Integration Pattern
- **Event-driven integration** for change notifications
- **REST API** for read operations (Published Language pattern)
- **No direct database access** between contexts

## Aggregate: MetaModelConfiguration

### Identity
- Uses `MetaModelConfigurationID` as aggregate ID (proper value object, UUID-based)
- Maintains 1:1 relationship with tenant via `TenantID` reference (not identity)
- Event stream keyed by MetaModelConfigurationID

### Properties
- `id`: MetaModelConfigurationID (value object)
- `tenantID`: TenantID (reference to Platform context - by ID only, not identity)
- `maturityScaleConfig`: MaturityScaleConfig (value object)
- `createdAt`: Timestamp (value object)
- `modifiedAt`: Timestamp (value object)
- `modifiedBy`: UserEmail (value object)
- `version`: Version (value object)

### Architectural Decision: Focused Aggregates
Future meta-model extensions (strategy pillars, element types, relationship configurations) will be implemented as **separate aggregates** rather than expanding this aggregate. This prevents a "God Aggregate" anti-pattern and reduces contention when multiple users configure different aspects.

Future aggregates (each with own ID and event stream):
- `StrategyPillarConfiguration` - manages custom strategy pillars
- `ElementTypeConfiguration` - manages element type definitions
- `RelationshipRuleConfiguration` - manages allowed relationships

All future aggregates will reference the tenant by TenantID but have their own aggregate identity.

## Value Objects

### MetaModelConfigurationID
Unique identifier for the aggregate.

**Properties:**
- `value`: string (UUID format)

**Validation:**
- Must be valid UUID format

### Timestamp
Immutable wrapper for time values.

**Properties:**
- `value`: time.Time

**Validation:**
- Must not be zero value

### UserEmail
Immutable wrapper for user email addresses.

**Properties:**
- `value`: string

**Validation:**
- Must be valid email format
- Max 255 characters

### Version
Optimistic locking version counter.

**Properties:**
- `value`: int

**Validation:**
- Must be positive integer (>= 1)

### MaturityScaleConfig
Immutable configuration for the maturity scale with exactly 4 sections.

**Properties:**
- `sections`: [4]MaturitySection

**Invariants:**
- Exactly 4 sections
- Sections cover range 0-99 with no gaps or overlaps
- First section starts at 0, last section ends at 99
- Sections are contiguous (section N's maxValue + 1 = section N+1's minValue)

### MaturitySection
Single section of the maturity scale.

**Properties:**
- `order`: int (1-4)
- `name`: SectionName (non-empty, max 50 chars)
- `minValue`: MaturityValue (0-99)
- `maxValue`: MaturityValue (0-99, >= minValue)

## Default Configuration

When a tenant is created, the following default maturity scale is provisioned:

| Order | Name | Min | Max |
|-------|------|-----|-----|
| 1 | Genesis | 0 | 24 |
| 2 | Custom Built | 25 | 49 |
| 3 | Product | 50 | 74 |
| 4 | Commodity | 75 | 99 |

## Events

### MetaModelConfigurationCreated
Published when a new tenant's meta-model configuration is initialized.

**Properties:**
- `ID`: string (MetaModelConfigurationID)
- `TenantID`: string
- `Sections`: []MaturitySectionData
- `CreatedAt`: timestamp
- `CreatedBy`: string (user email)

### MaturityScaleConfigUpdated
Published when the maturity scale configuration is modified.

**Properties:**
- `ID`: string (MetaModelConfigurationID)
- `TenantID`: string
- `Version`: int
- `NewSections`: []MaturitySectionData
- `ModifiedAt`: timestamp
- `ModifiedBy`: string (user email)

### MaturityScaleConfigReset
Published when the maturity scale is reset to defaults.

**Properties:**
- `ID`: string (MetaModelConfigurationID)
- `TenantID`: string
- `Version`: int
- `Sections`: []MaturitySectionData
- `ModifiedAt`: timestamp
- `ModifiedBy`: string (user email)

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Meta-Model** | The set of configurable elements that define how the architecture tool behaves |
| **Maturity Scale** | A 0-99 numeric scale divided into named sections representing component/capability maturity |
| **Section** | A named range within the maturity scale (e.g., "Genesis" covering 0-24) |
| **Section Boundary** | The min/max numeric values that define a section's range |

## Directory Structure

```
backend/internal/metamodel/
├── domain/
│   ├── aggregates/      # Aggregate roots
│   ├── valueobjects/    # Immutable value objects
│   └── events/          # Domain events
├── application/
│   ├── commands/        # Command definitions
│   ├── handlers/        # Command and event handlers
│   ├── projectors/      # Event projectors for read models
│   └── readmodels/      # Read model queries
└── infrastructure/
    ├── api/             # HTTP handlers and routes
    └── repositories/    # Persistence implementations
```

## Dependencies
- Spec 065: TenantProvisioning (subscribes to TenantCreated event)
- Spec 089: GenericEventSourcedRepository (uses generic repository pattern)

## Checklist
- [x] Specification ready
- [x] Bounded context documentation created in /docs/bounded-contexts
- [x] Directory structure created
- [x] Implementation complete
- [x] Unit tests implemented and passing
- [x] Integration tests implemented
- [x] API documentation updated (Swagger annotations in handlers)
- [x] User sign-off
