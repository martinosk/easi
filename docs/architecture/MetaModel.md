# Bounded Context Canvas: MetaModel

## Name
**MetaModel**

## Purpose
Manage configurable meta-model elements that control how the architecture modeling tool behaves within each tenant. The meta-model defines the vocabulary and rules for modeling, allowing organizations to customize the tool to match their specific methodologies.

**Key Stakeholders:**
- Platform Administrators (configure tool behavior)
- Enterprise Architects (consumers of configured options)
- Tenant Administrators (manage tenant-specific settings)

**Value Proposition:**
- Customize maturity scales to match organizational terminology
- Adapt the tool to different architectural methodologies
- Maintain consistency within a tenant while allowing cross-tenant diversity
- Enable evolution of modeling concepts without code changes

## Strategic Classification

### Domain Importance
**Supporting Domain** - Enables customization of core domain concepts but is not a competitive differentiator itself.

### Business Model
**Compliance Enabler** - Ensures the tool can comply with various organizational standards and methodologies.

### Evolution Stage
**Product** - Standard configuration patterns, well-understood requirements.

## Domain Roles
- **Configuration Manager**: Stores and retrieves tenant-specific configuration
- **Default Provider**: Provisions default configuration for new tenants
- **Validation Enforcer**: Ensures configuration changes maintain invariants

## Inbound Communication

### Messages Received

**Commands** (from Frontend/API):
- `UpdateMaturityScale` - Modify maturity scale section names and boundaries
- `ResetMaturityScale` - Restore maturity scale to default configuration

**Events** (from other contexts):
- From **Platform**:
  - `TenantCreated` - Provision default meta-model configuration for new tenant

### Collaborators
- **Frontend UI**: Source of configuration commands
- **Platform Context**: Publishes tenant lifecycle events

### Relationship Types
- **Customer-Supplier** with Platform: MetaModel is downstream, conforms to Platform's tenant model

## Outbound Communication

### Messages Sent

**Events** (published to event bus):
- `MetaModelConfigurationCreated` - Default configuration provisioned
- `MaturityScaleConfigUpdated` - Maturity scale modified
- `MaturityScaleConfigReset` - Maturity scale reset to defaults

**Queries** (from other contexts):
- Other contexts query MetaModel's read models for current configuration

### Collaborators
- **CapabilityMapping Context**: Queries maturity scale configuration
- **ArchitectureModeling Context** (future): May query element type configuration

### Integration Pattern
- **Event-driven integration** for configuration changes
- **REST API** as Published Language for read operations
- **Direct read model queries** acceptable for same-database deployment

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Meta-Model** | The set of configurable elements that define how the architecture tool behaves |
| **Meta-Model Configuration** | A tenant's complete set of meta-model settings |
| **Maturity Scale** | A 0-99 numeric scale divided into named sections representing component/capability evolution |
| **Section** | A named range within the maturity scale (e.g., "Genesis" covering 0-24) |
| **Section Boundary** | The min/max numeric values that define a section's range |
| **Default Configuration** | The standard configuration provisioned for new tenants |

## Business Decisions

### Core Business Rules

1. **Maturity Scale Invariants**:
   - Always exactly 4 sections
   - Sections must cover 0-99 with no gaps or overlaps
   - First section starts at 0, last section ends at 99
   - Sections must be contiguous

2. **Section Validation**:
   - Section names must be non-empty, max 50 characters
   - Section boundaries must be 0-99 inclusive
   - maxValue must be >= minValue within each section

3. **Configuration Identity**:
   - One configuration per tenant (1:1 relationship)
   - TenantID serves as aggregate identity

4. **Default Provisioning**:
   - New tenants automatically receive default configuration
   - Default: Genesis (0-24), Custom Built (25-49), Product (50-74), Commodity (75-99)

### Policy Decisions
- Configuration changes are event-sourced for full audit trail
- Maturity scale changes do not require migration of existing capabilities (capabilities store numeric values)
- Configuration is tenant-isolated (row-level security)

## Assumptions

1. **Configuration frequency**: Changes to meta-model are infrequent (monthly/quarterly)
2. **Section count**: Fixed at 4 sections per maturity scale
3. **Tenant count**: Configuration scales linearly with tenants
4. **Query frequency**: Configuration is cached; reads are frequent but writes are rare

## Verification Metrics

### Boundary Health Indicators
- **Configuration completeness**: 100% of tenants have meta-model configuration
- **Event consistency**: All configuration changes captured as events

### Context Effectiveness Metrics
- **Default usage**: Percentage of tenants using default configuration (baseline)
- **Customization rate**: Percentage of tenants with modified configuration

## Open Questions

1. **Strategy Pillar Configuration**: Should strategy pillars be configurable similar to maturity scale?
2. **Element Type Configuration**: Should custom element types be definable per tenant?
3. **Import/Export**: Should configuration be exportable for sharing between tenants?
4. **Version History**: Should users be able to view/restore previous configurations?

## Architecture Notes

### Implementation Location
`/backend/internal/metamodel/`

### Key Packages
- `domain/` - MetaModelConfiguration aggregate, value objects, events
- `application/` - Commands, handlers, projectors
- `infrastructure/` - API routes, repositories

### Technical Patterns
- **CQRS with Event Sourcing**: Full audit trail for configuration changes
- **Aggregate per Tenant**: TenantID = Aggregate ID
- **Value Object Immutability**: All configuration elements are immutable value objects
- **Event-Driven Provisioning**: Subscribe to TenantCreated for automatic setup

### API Style
- REST Level 3 with HATEOAS
- Single resource endpoints (`/api/v1/meta-model/maturity-scale`)
- Optimistic locking via version field

### Cross-Context Integration
- **Downstream of Platform**: Subscribes to `TenantCreated`
- **Upstream to CapabilityMapping**: Provides configuration via read models/API
- **Published Language**: Well-defined DTOs for configuration data
