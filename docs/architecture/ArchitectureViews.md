# Bounded Context Canvas: Architecture Views

## Name
**Architecture Views**

## Purpose
Enable enterprise architects to create, manage, and customize visual representations of the architecture. Provides multiple perspectives on the same underlying architecture model through different views, each with custom layouts, styling, and visualization preferences.

**Key Stakeholders:**
- Enterprise Architects (view creators)
- Solution Architects (view consumers)
- Business Stakeholders (presentation audiences)
- Governance Teams (review participants)

**Value Proposition:**
- Multiple perspectives on the same architecture (integration view, security view, etc.)
- Customizable visual presentation for different audiences
- Support architecture review sessions with stakeholder-specific views
- Maintain visual context separate from underlying component definitions

## Strategic Classification

### Domain Importance
**Supporting Domain** - Visualization and presentation layer for architecture models. Important for communication but not core to architecture modeling itself.

### Business Model
**Engagement Creator** - Primary purpose is enabling effective communication and stakeholder engagement through tailored visual representations.

### Evolution Stage
**Custom-Built** - Visualization needs are specific to this organization's architecture review process and presentation standards.

## Domain Roles
- **Presentation Layer**: Transforms component data into visual representations
- **Context Manager**: Maintains multiple perspectives/lenses on the same architecture
- **Customization Engine**: Enables view-specific styling, layout, and configuration

## Inbound Communication

### Messages Received

**Commands** (from Frontend/API):
- `CreateView` - User creates new architecture view
- `DeleteView` - User removes view
- `RenameView` - User changes view name
- `AddComponentToView` - User includes component in view
- `RemoveComponentFromView` - User excludes component from view
- `UpdateComponentPosition` - User repositions component in view
- `UpdateMultiplePositions` - User repositions multiple components
- `UpdateViewEdgeType` - User changes relationship visualization style
- `UpdateViewLayoutDirection` - User changes layout orientation
- `UpdateViewColorScheme` - User applies color scheme
- `UpdateElementColor` - User customizes individual element color
- `ClearElementColor` - User resets element to default color

**Events** (from other contexts):
- From **Architecture Modeling**:
  - `ApplicationComponentDeleted` - Remove component from all views
  - `ComponentRelationDeleted` - Update views displaying this relation

### Collaborators
- **Frontend UI**: Primary source of commands (user-initiated view management)
- **Architecture Modeling Context**: Source of truth for component existence
- **Multi-tenant Infrastructure**: Provides tenant isolation for views

### Relationship Types
- **Customer-Supplier** with Architecture Modeling: Consumes component definitions, must adapt to upstream changes
- **Conformist** with Frontend: Adapts to UI visualization needs

## Outbound Communication

### Messages Sent

**Events** (published to event bus):
- `ViewCreated` - New view created
- `ViewDeleted` - View removed
- `ViewRenamed` - View name changed
- `ComponentAddedToView` - Component included in view
- `ComponentRemovedFromView` - Component excluded from view
- `DefaultViewChanged` - User's default view changed
- `ViewVisibilityChanged` - View visibility updated

**Queries** (to other contexts):
- To **Architecture Modeling**: Read `ApplicationComponentReadModel` to get component details for display in views

### Collaborators
- **Frontend UI**: Consumes events for real-time view updates
- **Architecture Modeling Context**: Queries for component details to enrich view data

### Integration Pattern
- **Event-driven integration** for view lifecycle and membership changes (publish to event bus)
- **Query-based integration** for component details (read from Architecture Modeling read models)
- **Event subscription** for upstream component changes (maintain consistency)

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Architecture View** | A specific visual representation of a subset of the architecture with custom layout and styling |
| **View Component** | An instance of an application component within a specific view, with view-specific properties (position, color) |
| **Component Position** | X/Y coordinates for component placement in view canvas |
| **Edge Type** | Visual style for displaying relationships (straight, curved, orthogonal) |
| **Layout Direction** | Orientation for automatic layout (top-down, left-right, radial) |
| **Color Scheme** | Predefined color palette applied to views (default, high-contrast, categorical) |
| **Element Color** | Custom color override for individual component or relation in a view |
| **Default View** | User's preferred view that loads automatically |
| **View Canvas** | The visual workspace where components and relations are displayed |

## Business Decisions

### Core Business Rules
1. **View membership**: A component can appear in multiple views with different positions/styling
2. **View isolation**: Changes to component position/color in one view do not affect other views
3. **Consistency enforcement**: Cannot add a component to a view if the component doesn't exist in Architecture Modeling
4. **Cascade cleanup**: When a component is deleted from Architecture Modeling, it's removed from all views
5. **View ownership**: Each view belongs to a single tenant (no cross-tenant views)
6. **Default view per user**: Each user can have one default view

### Policy Decisions
- Views are not versioned (single current state)
- No explicit view sharing permissions (all views visible to all users in tenant)
- Position coordinates are absolute (not relative or grid-based)
- Color customization overrides scheme settings
- Layout direction is a hint for auto-layout, not enforced
- Views can be empty (no components required)

## Assumptions

1. **View count per tenant**: Fewer than 100 views per tenant
2. **Components per view**: Fewer than 500 components per view (performance constraint)
3. **Update frequency**: View layouts change more frequently than component definitions (daily adjustments)
4. **Real-time updates**: Users expect immediate visual feedback when repositioning components
5. **Client-side rendering**: Frontend handles actual graph visualization, backend only stores coordinates
6. **No collaborative editing**: Only one user edits a view at a time (no conflict resolution needed)
7. **Eventual consistency acceptable**: Brief delay between component deletion and view update is acceptable

## Verification Metrics

### Boundary Health Indicators
- **View consistency rate**: 100% of components in views must exist in Architecture Modeling
- **Event handler success**: Greater than 99% success rate for `ApplicationComponentDeleted` handler
- **Cross-context coupling**: Fewer than 5% of view operations require synchronous calls to Architecture Modeling

### Context Effectiveness Metrics
- **View orphan rate**: Zero orphaned components in views (components that don't exist in Architecture Modeling)
- **Position update latency**: Less than 200ms for component position updates
- **Color scheme application**: Greater than 90% of views use standard schemes (not fully custom colors)

### Business Value Metrics
- **Views per architect**: Average number of views created per active user
- **View reuse**: Percentage of views accessed by multiple users
- **Stakeholder engagement**: Number of views created for specific review sessions
- **Visualization diversity**: Number of different edge types and color schemes in active use

## Open Questions

1. **Should views be shareable across tenants?** Currently tenant-isolated. Use case: Shared reference architectures?

2. **View templates or libraries?** Should there be pre-defined view templates (e.g., "Integration View Template", "Security View Template")?

3. **Auto-layout algorithms?** Should the backend calculate component positions based on layout direction, or is this purely frontend concern?

4. **View versioning or snapshots?** Should views be snapshotted at specific times for historical comparison or audit?

5. **Collaborative editing?** What happens when two users edit the same view simultaneously?

6. **View access control?** Should there be granular permissions (private views, team views, public views)?

7. **Component filtering rules?** Should views support rule-based filtering (e.g., "show all components owned by Team X")?

8. **Relation visibility control?** Should views allow hiding specific relations while keeping components visible?

9. **Export formats?** Should views be exportable to image/PDF/SVG directly from backend, or is this frontend responsibility?

## Architecture Notes

### Implementation Location
`/backend/internal/architectureviews/`

### Key Packages
- `domain/` - Aggregates (ArchitectureView), Value Objects, Domain Events
- `application/` - Commands, Command Handlers, Projectors, Read Models
- `infrastructure/` - API routes, repository implementations

### Technical Patterns
- **CQRS with Event Sourcing**: Write model uses aggregates, read model uses projectors
- **Event Store**: PostgreSQL-backed event log
- **Read Models**: Denormalized views for fast query
- **Event Subscriptions**: Listens to Architecture Modeling events
- **Multi-tenancy**: Tenant-aware database isolation

### API Style
- REST Level 3 with HATEOAS
- PATCH operations for incremental updates
- Bulk update support for multi-select operations

### Cross-Context Integration
- **Downstream of Architecture Modeling**: Must handle component/relation deletion events
- **Read-only access** to Architecture Modeling read models
- **No circular dependencies**: Does not publish events consumed by Architecture Modeling

## Collaboration Patterns

### With Architecture Modeling Context
```
Architecture Modeling → Architecture Views
- Event: ApplicationComponentDeleted
- Action: Remove component from all views

Architecture Modeling (read models) ← Architecture Views
- Query: Get component details for view rendering
- Pattern: Eventually consistent read
```

### Consistency Strategy
- **Eventual consistency** for component deletions (acceptable brief lag)
- **Query-based validation** when adding components to views (check existence)
- **Event-driven cleanup** for maintaining view consistency
