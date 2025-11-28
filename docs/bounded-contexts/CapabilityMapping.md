# Bounded Context Canvas: Capability Mapping

## Name
**Capability Mapping**

## Purpose
Enable enterprise architects to model business capabilities and map them to IT systems, creating a bridge between business strategy and technology implementation. Provides strategic insight into how organizational capabilities are realized, their maturity, dependencies, and alignment with business strategy.

**Key Stakeholders:**
- Enterprise Architects (capability modelers)
- Business Analysts (capability owners)
- Strategic Planners (capability strategists)
- Technology Leads (system-to-capability mappers)
- Portfolio Managers (investment decision makers)

**Value Proposition:**
- Understand which IT systems support which business capabilities
- Identify capability gaps (not realized by any system)
- Analyze capability dependencies and strategic importance
- Assess capability maturity and track improvement
- Align IT investment with strategic business capabilities
- Support portfolio rationalization decisions

## Strategic Classification

### Domain Importance
**Core Domain** - This is a key differentiator. Most EA tools don't provide sophisticated capability-to-system mapping with strategic analysis. This directly supports competitive advantage in IT portfolio management and strategic planning.

### Business Model
**Compliance Enforcer & Engagement Creator**
- Enforces strategic alignment between business capabilities and IT investments
- Creates engagement between business and IT through shared capability language

### Evolution Stage
**Custom-Built** - Highly tailored to this organization's strategic planning methodology and capability taxonomy.

## Domain Roles
- **Strategic Analyzer**: Analyzes capability portfolio for strategic planning
- **Integration Point**: Links business strategy to IT implementation
- **Maturity Assessor**: Tracks capability maturity evolution
- **Dependency Tracker**: Maps capability interdependencies
- **Gap Identifier**: Identifies capabilities not adequately supported by IT systems

## Inbound Communication

### Messages Received

**Commands** (from Frontend/API):
- `CreateCapability` - User creates new business capability
- `UpdateCapability` - User modifies capability properties
- `DeleteCapability` - User removes capability
- `UpdateCapabilityMetadata` - User changes maturity, ownership, strategy alignment
- `ChangeCapabilityParent` - User reorganizes capability hierarchy
- `AddCapabilityExpert` - User associates SME with capability
- `AddCapabilityTag` - User tags capability for categorization
- `CreateCapabilityDependency` - User defines capability dependency
- `DeleteCapabilityDependency` - User removes dependency
- `LinkSystemToCapability` - User maps system to capability
- `UpdateSystemRealization` - User changes realization level
- `DeleteSystemRealization` - User removes system-capability link

**Events** (from other contexts):
- From **Architecture Modeling**:
  - `ApplicationComponentDeleted` - Remove system from capability realizations

**Commands** (from other contexts - future):
- From **Enterprise Strategy** (future):
  - `AssignCapabilityToDomain` - Strategic process assigns capability to business domain
  - `UnassignCapabilityFromDomain` - Strategic process removes assignment

### Collaborators
- **Frontend UI**: Primary source of commands
- **Architecture Modeling Context**: Source of truth for system definitions
- **Enterprise Strategy Context** (future): Source of strategic domain assignments

### Relationship Types
- **Customer-Supplier** with Architecture Modeling: Depends on component definitions
- **Partnership** with Enterprise Strategy (future): Collaborative strategic modeling

## Outbound Communication

### Messages Sent

**Events** (published to event bus):
- `CapabilityCreated` - New capability defined
- `CapabilityUpdated` - Capability properties changed
- `CapabilityDeleted` - Capability removed
- `CapabilityMetadataUpdated` - Strategic metadata changed
- `CapabilityParentChanged` - Hierarchy reorganized
- `CapabilityExpertAdded` - SME associated
- `CapabilityTagAdded` - Tag added
- `CapabilityDependencyCreated` - Dependency established
- `CapabilityDependencyDeleted` - Dependency removed
- `SystemLinkedToCapability` - System-capability mapping created
- `SystemRealizationUpdated` - Realization level changed
- `SystemRealizationDeleted` - Mapping removed
- `CapabilityAssignedToDomain` (new) - L1 capability assigned to business domain
- `CapabilityUnassignedFromDomain` (new) - L1 capability removed from business domain

**Queries** (to other contexts):
- To **Architecture Modeling**: Read `ApplicationComponentReadModel` to get system details for capability realization

### Collaborators
- **Frontend UI**: Consumes events for real-time updates
- **Architecture Modeling Context**: Queries for system details
- **Enterprise Strategy Context** (future): Subscribes to capability events for impact analysis

### Integration Pattern
- **Event-driven integration** for capability changes
- **Query-based integration** for system details (read from Architecture Modeling)
- **Event subscription** for upstream system changes

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Business Capability** | A particular ability or capacity that a business possesses to achieve outcomes (e.g., "Customer Onboarding", "Financial Reporting") |
| **Capability Hierarchy** | Four-level taxonomy (L1→L2→L3→L4) organizing capabilities from strategic to operational |
| **L1 Capability** | Top-level strategic capability (e.g., "Customer Management") |
| **L2/L3/L4 Capability** | Sub-capabilities providing increasing operational detail |
| **Capability Realization** | The linkage between a capability and the IT systems that implement it |
| **Realization Level** | Degree to which a system realizes a capability (Partial, Full, Primary) |
| **Primary Realizer** | The main IT system responsible for a capability |
| **Capability Maturity** | Assessment of how well-developed a capability is (Initial, Managed, Defined, Quantified, Optimized) |
| **Capability Dependency** | A relationship where one capability requires or depends on another |
| **Dependency Type** | Nature of dependency (Requires, Enables, Supports) |
| **Capability Owner** | Business role/team responsible for the capability |
| **Ownership Model** | How capability ownership is distributed (Centralized, Federated, Hybrid) |
| **Strategy Pillar** | Strategic theme or initiative the capability supports |
| **Pillar Weight** | Importance of capability to each strategy pillar (1-5 scale) |
| **Capability Expert** | Subject matter expert associated with capability |
| **Capability Tag** | Label for categorization and filtering |
| **Capability Gap** | A capability with no IT system realization |
| **Business Domain** (new) | Strategic grouping of L1 capabilities (e.g., "Finance", "Customer Experience") |
| **Orphaned L1 Capability** (new) | L1 capability not assigned to any business domain |

## Business Decisions

### Core Business Rules
1. **Hierarchy constraints**:
   - L1 capabilities cannot have a parent
   - L2 capabilities must have L1 parent
   - L3 capabilities must have L2 parent
   - L4 capabilities must have L3 parent
   - Maximum hierarchy depth is L4

2. **Circular reference prevention**: Cannot create capability hierarchy cycles

3. **Dependency constraints**:
   - Cannot create self-dependencies
   - No duplicate dependencies between same two capabilities

4. **Realization constraints**:
   - Only one system can be marked as "Primary" realizer per capability
   - Systems must exist in Architecture Modeling to be linked

5. **Cascade deletion**: When capability deleted, all dependencies and realizations must be removed

6. **Business domain assignment** (new):
   - Only L1 capabilities can be assigned to business domains
   - L1 capabilities can belong to multiple business domains
   - L1 capabilities may have zero business domain assignments (orphaned)

### Policy Decisions
- Capability taxonomy is organization-wide (not tenant-specific for multi-tenancy)
- Maturity assessments are subjective (no automated calculation)
- Strategy pillar weights are relative (1-5 scale, not percentages)
- Tags are free-form (no controlled vocabulary)
- Experts are names/roles (not linked to user accounts)
- Capability codes are unique identifiers (immutable)

## Assumptions

1. **Capability count**: Fewer than 5,000 capabilities per tenant
2. **Hierarchy depth**: Most organizations use only L1-L3, rarely L4
3. **Realization ratio**: Most capabilities have 1-3 systems, rarely more than 10
4. **Dependency complexity**: Typical capability has 2-5 dependencies
5. **Update frequency**: Capability structure is relatively stable (quarterly changes)
6. **Maturity tracking**: Maturity levels change slowly (annually)
7. **Strategic alignment**: Strategy pillars remain consistent for 2-3 years
8. **Business domains**: 5-20 business domains per organization
9. **Domain assignments**: Most L1 capabilities assigned to 1-2 domains, rarely more than 3

## Verification Metrics

### Boundary Health Indicators
- **Hierarchy integrity**: Zero orphaned capabilities (except intentional L1s)
- **Reference integrity**: Zero capability dependencies referencing non-existent capabilities
- **Realization validity**: 100% of linked systems exist in Architecture Modeling
- **Event coupling**: Fewer than 30% of events trigger cross-context reactions

### Context Effectiveness Metrics
- **Capability coverage**: Percentage of L1 capabilities with at least one system realization
- **Realization completeness**: Percentage of capabilities with designated primary realizer
- **Dependency mapping**: Percentage of L1 capabilities with documented dependencies
- **Strategic alignment**: Percentage of L1 capabilities with strategy pillar weights
- **Domain coverage** (new): Percentage of L1 capabilities assigned to business domains

### Business Value Metrics
- **IT alignment score**: Correlation between strategy pillar weights and IT investment in realizing systems
- **Capability gap identification**: Number of high-priority capabilities with no/inadequate realization
- **Portfolio rationalization**: Number of redundant system realizations identified
- **Maturity improvement**: Average maturity level increase over time
- **Strategic focus** (new): Distribution of capabilities across business domains

## Open Questions

1. **Should capability maturity be auto-calculated?** Currently subjective. Could we derive from system metrics (uptime, performance, etc.)?

2. **Cross-tenant capability sharing?** Should there be reference capability models shared across tenants (industry standard taxonomies)?

3. **Capability lifecycle states?** Should capabilities have states (Planned, Active, Retiring, Retired)?

4. **System realization strength?** Is three levels (Partial, Full, Primary) sufficient, or do we need numeric scores?

5. **Dependency directionality semantics?** What's the difference between "A Requires B" and "B Enables A"?

6. **Should tags be controlled vocabulary?** Or remain free-form for flexibility?

7. **Expert verification?** Should experts confirm their association with capabilities?

8. **Capability ownership enforcement?** Should there be workflow to assign/approve owners?

9. **Business domain ownership?** Do business domains have separate owners from the capabilities within them?

10. **Domain assignment approval?** Should assigning capabilities to domains require approval workflow?

## Architecture Notes

### Implementation Location
`/backend/internal/capabilitymapping/`

### Key Packages
- `domain/` - Aggregates (Capability, CapabilityDependency, CapabilityRealization, BusinessDomain, BusinessDomainAssignment), Value Objects, Domain Events
- `application/` - Commands, Command Handlers, Projectors, Read Models
- `infrastructure/` - API routes, repository implementations

### Technical Patterns
- **CQRS with Event Sourcing**: Write model uses aggregates, read model uses projectors
- **Event Store**: PostgreSQL-backed event log
- **Read Models**: Denormalized views for hierarchical queries, dependency graphs, realization matrices
- **Multi-tenancy**: Tenant-aware database isolation

### API Style
- REST Level 3 with HATEOAS
- Hierarchical endpoint design (capabilities/{id}/children, capabilities/{id}/dependencies)
- Graph query support for dependency analysis

### Cross-Context Integration
- **Downstream of Architecture Modeling**: Listens to `ApplicationComponentDeleted` events
- **Upstream to Enterprise Strategy** (future): Publishes capability events for strategic analysis
- **Read-only access** to Architecture Modeling read models

## New Business Domain Features (Specs 053-058)

### Additional Aggregates
- **BusinessDomain**: Manages business domain lifecycle (CRUD)
- **BusinessDomainAssignment**: Manages many-to-many capability-to-domain assignments

### Integration with Enterprise Strategy Context (Future)
When Enterprise Strategy context is implemented:
- **Inbound**: Commands to assign/unassign capabilities during strategic consolidation/decomposition
- **Outbound**: Events about capability assignments for strategic impact analysis
- **Collaboration**: Partnership relationship for strategic capability planning

### Business Domain Queries
- Get all business domains
- Get capabilities for a specific business domain
- Get business domains for a specific capability
- Get orphaned L1 capabilities (not in any business domain)
- Get business domain composition (L1 capabilities + their full L2/L3/L4 hierarchy + realizing systems)
