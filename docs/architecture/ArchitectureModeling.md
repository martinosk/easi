# Bounded Context Canvas: Architecture Modeling

## Name
**Architecture Modeling**

## Purpose
Enable enterprise architects to model and manage the IT application landscape by defining application components and their relationships. Provides the foundational system-of-record for "what systems exist" and "how they interact."

**Key Stakeholders:**
- Enterprise Architects
- Solution Architects
- IT Portfolio Managers

**Value Proposition:**
- Single source of truth for application inventory
- Understand system dependencies and integration patterns
- Support impact analysis for changes
- Enable architecture governance

## Strategic Classification

### Domain Importance
**Supporting Domain** - Essential infrastructure for architecture management but not a competitive differentiator. Most EA tools provide similar capabilities.

### Business Model
**Engagement Creator** - Enables architects to engage with stakeholders through visual models and dependency analysis.

### Evolution Stage
**Custom-Built** - While similar to commercial EA tools, tailored specifically for this organization's architecture modeling methodology.

## Domain Roles
- **System of Record**: Authoritative source for application component definitions
- **Gateway**: Entry point for all architecture information that other contexts build upon
- **Information Holder**: Stores and provides application component and relationship data

## Inbound Communication

### Messages Received

**Commands** (from Frontend/API):
- `CreateApplicationComponent` - User creates new component
- `UpdateApplicationComponent` - User modifies component properties
- `DeleteApplicationComponent` - User removes component
- `CreateComponentRelation` - User creates dependency between components
- `UpdateComponentRelation` - User modifies relationship properties
- `DeleteComponentRelation` - User removes dependency

**Queries** (from other contexts):
- From **Architecture Views**: Read `ApplicationComponentReadModel` to get component details for views
- From **Capability Mapping**: Read `ApplicationComponentReadModel` to link systems to capabilities

### Collaborators
- **Frontend UI**: Primary source of commands (user-initiated actions)
- **Architecture Views Context**: Consumer of component data
- **Capability Mapping Context**: Consumer of component data for capability realization

### Relationship Types
- **Conformist** relationship with frontend (adapts to UI needs)
- **Published Language** relationship with other contexts (provides stable read models)

## Outbound Communication

### Messages Sent

**Events** (published to event bus):
- `ApplicationComponentCreated` - New component added to architecture
- `ApplicationComponentUpdated` - Component properties changed
- `ApplicationComponentDeleted` - Component removed from architecture
- `ComponentRelationCreated` - New dependency established
- `ComponentRelationUpdated` - Dependency properties changed
- `ComponentRelationDeleted` - Dependency removed

### Collaborators
- **Architecture Views Context**: Subscribes to component/relation events to maintain view consistency
- **Capability Mapping Context**: Subscribes to `ApplicationComponentDeleted` to cascade capability realization cleanup

### Integration Pattern
- **Event-driven integration** via shared `InMemoryEventBus`
- Events published after successful command execution
- No direct coupling to subscriber contexts

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Application Component** | A software system, application, service, or module in the IT landscape |
| **Component Relation** | A directed dependency or integration between two application components |
| **Relation Type** | The nature of the relationship (e.g., data flow, API call, sync/async) |
| **Component Properties** | Descriptive metadata about a component (name, description, type, owner, etc.) |
| **Component ID** | Globally unique identifier for an application component across all contexts |
| **Architecture Landscape** | The complete set of application components and their relationships |

## Business Decisions

### Core Business Rules
1. **Component uniqueness**: Component names should be unique within a tenant to avoid confusion
2. **Self-relations prohibited**: A component cannot have a relation to itself
3. **Duplicate relations prohibited**: Cannot create multiple identical relations between the same two components
4. **Deletion cascade**: When a component is deleted, all its relations must be deleted
5. **No orphaned relations**: Relations can only exist between components that exist

### Policy Decisions
- Components are tenant-scoped (multi-tenancy isolation)
- Component metadata is extensible (can add properties without schema changes)
- Relations are typed but types are not strictly enforced (flexible modeling)
- No versioning of components (single current state, history via event store)

## Assumptions

1. **Scale assumption**: A single tenant will have fewer than 10,000 components
2. **Update frequency**: Component definitions change infrequently (monthly, not daily)
3. **Read-heavy workload**: Components are read far more often than written (10:1 ratio)
4. **Simple relation model**: Relation properties are simple key-value pairs, not complex objects
5. **No external system integration**: Component data is manually entered, not synchronized from CMDB or service registry
6. **Event store reliability**: Events are never lost and are processed exactly once

## Verification Metrics

### Boundary Health Indicators
- **Event coupling ratio**: Less than 20% of events from this context cause changes in other contexts (indicates good autonomy)
- **Command success rate**: Greater than 95% of commands succeed without cross-context coordination
- **Read model query performance**: All queries under 100ms (indicates proper denormalization)

### Context Effectiveness Metrics
- **Component churn rate**: Track how often components are created/deleted (high churn may indicate boundary issues)
- **Relation consistency**: Zero orphaned relations in read models
- **Cross-context query count**: Fewer than 10% of queries need data from multiple contexts

### Business Value Metrics
- **Component catalog completeness**: Percentage of known IT systems modeled
- **Relationship coverage**: Percentage of known integrations documented
- **Architecture decision support**: Number of decisions informed by component models

## Open Questions

1. **Should components have versions?** Currently we only track current state. Should historical component definitions be queryable beyond event store?

2. **How should component decommissioning work?** Should deleted components be soft-deleted with "decommissioned" status, or hard-deleted from read models?

3. **What level of component granularity?** Should microservices be individual components, or grouped by domain/capability?

4. **Should relation properties be strongly typed?** Currently flexible key-value pairs. Should we enforce schemas for specific relation types?

5. **Integration with external CMDBs?** Should there be a synchronization mechanism with existing configuration management databases?

6. **Component ownership model?** Should components have formal ownership (team, product owner) enforced by this context?

7. **Relation cardinality constraints?** Should we enforce rules like "a database can have multiple clients but a client can only have one database of type X"?

## Architecture Notes

### Implementation Location
`/backend/internal/architecturemodeling/`

### Key Packages
- `domain/` - Aggregates (ApplicationComponent, ComponentRelation), Value Objects, Domain Events
- `application/` - Commands, Command Handlers, Projectors, Read Models
- `infrastructure/` - API routes, repository implementations, database adapters

### Technical Patterns
- **CQRS with Event Sourcing**: Write model uses aggregates, read model uses projectors
- **Event Store**: PostgreSQL-backed event log as source of truth
- **Read Models**: Denormalized views for query performance
- **Multi-tenancy**: Tenant-aware database isolation

### API Style
- REST Level 3 with HATEOAS
- OpenAPI specification
- Opaque cursor pagination
