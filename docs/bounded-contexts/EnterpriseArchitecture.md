# Bounded Context Canvas: Enterprise Architecture

## Name
**Enterprise Architecture**

## Purpose
Enable enterprise architects to discover capability overlap across business domains, track standardization requirements, and perform cross-domain maturity gap analysis for investment prioritization. Provides the analytical foundation for strategic consolidation decisions.

**Key Stakeholders:**
- Enterprise Architects (capability consolidation analysis)
- Portfolio Managers (investment prioritization)
- IT Strategy Team (standardization planning)
- Domain Architects (cross-domain visibility)

**Value Proposition:**
- Discover overlapping capabilities across domains ("We have 5 Payroll implementations")
- Track standardization requirements ("Payroll should be consolidated")
- Analyze cross-domain maturity gaps ("IT Support Payroll is Genesis, should be Product")
- Prioritize investments based on enterprise-wide view
- Provide analytical foundation for strategic consolidation decisions (feeds into Enterprise Strategy context)

## Relationship to Other Contexts

### Compared to Capability Mapping
- **Capability Mapping**: "What capabilities exist in each domain and how are they realized?"
- **Enterprise Architecture**: "Which capabilities across domains represent the same thing and where should we standardize?"

### Compared to Enterprise Strategy (Future)
- **Enterprise Architecture**: **Analyze** - "We have 5 payroll implementations with varying maturity, marked for standardization"
- **Enterprise Strategy**: **Act** - "Execute consolidation workflow: propose → approve → migrate"

These contexts are complementary:
1. Enterprise Architecture identifies consolidation opportunities
2. Enterprise Strategy executes consolidation workflows
3. Results flow back to Capability Mapping

## Strategic Classification

### Domain Importance
**Core Domain** - Cross-domain capability analysis and standardization tracking is a key differentiator. This enables evidence-based architecture governance that most EA tools lack.

### Business Model
**Engagement Creator & Decision Support**
- Creates engagement between domain architects and enterprise architects
- Supports evidence-based investment decisions
- Enables proactive standardization planning

### Evolution Stage
**Custom-Built** - Novel approach to bottom-up capability consolidation discovery. Most EA tools use top-down reference models; this supports organic discovery.

## Domain Roles
- **Overlap Discoverer**: Identifies capabilities that exist across multiple domains under different names
- **Standardization Tracker**: Tracks which enterprise capabilities should be consolidated
- **Gap Analyzer**: Compares maturity levels across domain instances of the same capability
- **Investment Advisor**: Provides data for cross-domain investment prioritization

## Inbound Communication

### Messages Received

**Commands** (from Frontend/API):
- `CreateEnterpriseCapability` - Create a new enterprise-wide capability grouping
- `UpdateEnterpriseCapability` - Modify name, description, category
- `DeleteEnterpriseCapability` - Soft delete grouping
- `LinkCapabilityToEnterpriseCapability` - Connect domain capability to enterprise grouping
- `UnlinkCapabilityFromEnterpriseCapability` - Remove connection
- `SetEnterpriseCapabilityImportance` - Rate importance for a strategy pillar (with optional rationale)
- `UpdateEnterpriseCapabilityImportance` - Change importance level or rationale
- `RemoveEnterpriseCapabilityImportance` - Remove importance rating

**Events** (from other contexts):
- From **MetaModel**:
  - `StrategyPillarAdded` - Update local pillar cache
  - `StrategyPillarUpdated` - Update local pillar cache
  - `StrategyPillarRemoved` - Mark pillar inactive in cache
- From **Capability Mapping**:
  - `CapabilityDeleted` - Remove links to deleted capability
  - `CapabilityMetadataUpdated` - Update maturity in analysis views

### Collaborators
- **Frontend UI**: Source of commands
- **MetaModel Context**: Source of strategy pillar definitions
- **Capability Mapping Context**: Source of domain capability data

### Relationship Types
- **Customer-Supplier** with MetaModel: Consumes pillar definitions
- **Customer-Supplier** with Capability Mapping: Consumes capability data for analysis

## Outbound Communication

### Messages Sent

**Events** (published to event bus):
- `EnterpriseCapabilityCreated` - New grouping created
- `EnterpriseCapabilityUpdated` - Grouping modified
- `EnterpriseCapabilityDeleted` - Grouping removed
- `CapabilityLinkedToEnterpriseCapability` - Domain capability connected
- `CapabilityUnlinkedFromEnterpriseCapability` - Connection removed
- `EnterpriseCapabilityImportanceSet` - Strategic importance rated for pillar
- `EnterpriseCapabilityImportanceUpdated` - Importance or rationale changed
- `EnterpriseCapabilityImportanceRemoved` - Importance rating removed

**Queries** (to other contexts):
- To **Capability Mapping**: Read capability details, maturity levels, business domains
- To **MetaModel**: Read maturity scale configuration for section names

### Collaborators
- **Enterprise Strategy Context** (future): Consumes enterprise capability data for consolidation proposals
- **Strategic Analytics Dashboard** (future): Consumes events for reporting

### Integration Pattern
- **Event-driven integration** for enterprise capability changes
- **Query-based integration** for reading domain capability data
- **Anti-Corruption Layer** maintains local cache of strategy pillars

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Enterprise Capability** | A logical capability that may exist across multiple business domains under different names |
| **Domain Capability** | A capability defined within a specific business domain (from Capability Mapping context) |
| **Capability Link** | Connection between a domain capability and an enterprise capability, indicating they represent the same logical capability |
| **Linked Capability** | A domain capability that has been connected to an enterprise capability |
| **Unlinked Capability** | A domain capability not connected to any enterprise capability (orphaned from enterprise view) |
| **Implementation Count** | Number of domain capabilities linked to an enterprise capability |
| **Domain Spread** | Number of different business domains with implementations of an enterprise capability |
| **Maturity Spread** | Difference between highest and lowest maturity among linked domain capabilities |
| **Maturity Gap** | Difference between a domain capability's current maturity and the target maturity |
| **Strategic Importance** | Rating (1-5) of how important an enterprise capability is for a strategy pillar, with optional rationale explaining why |
| **Importance Rationale** | Free-text explanation of why a capability has a particular strategic importance rating |
| **Standardization Candidate** | Enterprise capability with importance rating for standardization pillar and multiple implementations |
| **Investment Priority** | Calculated priority based on maturity gap and strategic importance |
| **Bottom-Up Discovery** | Process of identifying enterprise capabilities by linking existing domain capabilities |
| **Canonical Name** | The enterprise-wide standard name for a capability (defined in enterprise capability) |

## Business Decisions

### Core Business Rules

1. **Linking constraints**:
   - A domain capability can only be linked to ONE enterprise capability (prevents confusion)
   - Enterprise capability must exist and be active to receive links
   - Domain capability must exist to be linked

2. **Enterprise capability constraints**:
   - Names must be unique within tenant
   - Soft delete preserves links for historical analysis
   - Cannot physically delete enterprise capability with active links

3. **Strategic importance constraints**:
   - Importance rating 1-5, one rating per pillar per enterprise capability
   - Optional rationale (max 500 chars) explaining the rating
   - Enterprise-scoped importance is separate from domain-scoped importance

4. **Analysis constraints**:
   - Maturity gaps calculated from live capability mapping data
   - Standardization candidates must have 2+ implementations

### Policy Decisions
- Enterprise capabilities are optional (not all domain capabilities need linking)
- Linking is bottom-up (architects discover and link, not imposed top-down)
- Enterprise capability names are canonical but domain capabilities keep their local names
- Historical links preserved when enterprise capability soft-deleted
- No automatic linking suggestions (manual discovery process)

## Assumptions

1. **Enterprise capability count**: 50-200 enterprise capabilities per tenant
2. **Link count**: Average 2-5 domain capabilities per enterprise capability
3. **Discovery frequency**: Monthly discovery/linking sessions
4. **Standardization scope**: 20-40% of enterprise capabilities marked for standardization
5. **Analysis frequency**: Weekly gap analysis queries
6. **Stakeholder access**: Enterprise architects have cross-domain visibility

## Verification Metrics

### Boundary Health Indicators
- **Link integrity**: 100% of links reference existing capabilities
- **Pillar cache freshness**: Local cache updated within 5 seconds of pillar changes
- **Event coherence**: All enterprise capability changes emit corresponding events

### Context Effectiveness Metrics
- **Discovery coverage**: Percentage of domain capabilities linked to enterprise capabilities
- **Standardization tracking**: Percentage of multi-implementation capabilities with standardization importance rating
- **Gap visibility**: Number of identified maturity gaps above threshold

### Business Value Metrics
- **Overlap discovery**: Number of duplicate implementations identified
- **Consolidation readiness**: Number of standardization candidates with complete analysis
- **Investment clarity**: Reduction in time to identify investment priorities
- **Strategic execution**: Correlation between standardization importance ratings and consolidation execution

## Open Questions

1. **Auto-suggest linking?** Should system suggest potential links based on name similarity?

2. **Target maturity definition?** Should enterprise capability define target maturity, or derive from highest linked capability?

3. **Link inheritance?** If L1 domain capability is linked, should child capabilities inherit the link?

4. **Cross-tenant patterns?** Should reference enterprise capabilities be shareable across tenants?

5. **Consolidation workflow trigger?** Should high-importance standardization candidates auto-create Enterprise Strategy proposals?

6. **Historical analysis?** Should we track how enterprise capability composition changed over time?

7. **Dependency impact?** Should enterprise capability analysis include dependency relationships between linked capabilities?

8. **Cost analysis?** Should we integrate cost data for ROI analysis of consolidation?

## Architecture Notes

### Implementation Location
`/backend/internal/enterprisearchitecture/`

### Key Packages
- `domain/` - Aggregates (EnterpriseCapability, EnterpriseCapabilityLink, EnterpriseCapabilityStrategicImportance), Value Objects, Domain Events
- `application/` - Commands, Command Handlers, Projectors, Read Models
- `infrastructure/` - API routes, repository implementations, anti-corruption layer

### Technical Patterns
- **CQRS with Event Sourcing**: Full audit trail for enterprise capability evolution
- **Anti-Corruption Layer**: Local cache for strategy pillars from MetaModel
- **Read Models**: Denormalized views for gap analysis, standardization candidates
- **Multi-tenancy**: Tenant-aware database isolation

### API Style
- REST Level 3 with HATEOAS
- Base path: `/api/v1/enterprise-architecture/`
- Analysis-oriented endpoints (`/standardization-candidates`, `/maturity-gap-analysis`)

### Cross-Context Integration
- **Downstream of MetaModel**: Subscribes to strategy pillar events
- **Downstream of Capability Mapping**: Queries capability data, subscribes to deletion events
- **Upstream to Enterprise Strategy** (future): Provides analysis data for consolidation proposals

## Collaboration Patterns

### With MetaModel Context

**Pillar Synchronization**:
```
MetaModel → Enterprise Architecture (events)
- Event: StrategyPillarAdded
- Reaction: Insert into local available_strategy_pillars cache

Enterprise Architecture → MetaModel (query)
- Query: Get current maturity scale configuration
- Use: Translate maturity values to section names in gap analysis
```

### With Capability Mapping Context

**Capability Data Access**:
```
Enterprise Architecture → Capability Mapping (query)
- Query: Get capability details (name, maturity, business domain)
- Use: Display in enterprise capability analysis views

Capability Mapping → Enterprise Architecture (events)
- Event: CapabilityDeleted
- Reaction: Remove links referencing deleted capability
```

### With Enterprise Strategy Context (Future)

**Analysis to Action**:
```
Enterprise Strategy → Enterprise Architecture (query)
- Query: Get standardization candidates with high importance
- Use: Suggest consolidation proposals

Enterprise Architecture → Enterprise Strategy (data)
- Data: Enterprise capability composition, maturity gaps
- Use: Impact analysis during consolidation planning
```

## Implementation Priority

**Phase 1 (Spec 100: Create Enterprise Capability Groupings)**:
- EnterpriseCapability aggregate with CRUD
- EnterpriseCapabilityLink aggregate for linking domain capabilities
- EnterpriseCapabilityStrategicImportance aggregate with rationale
- Frontend: Enterprise Architecture page with capability list and detail views

**Phase 2 (Spec 101: Discover and Analyze Standardization Opportunities)**:
- Standardization candidates dashboard
- Maturity gap analysis views
- Investment priority indicators
- Unlinked capabilities discovery

**Phase 3 (Future)**:
- Integration with Enterprise Strategy for consolidation workflow
- Historical analysis and trend tracking
- Auto-suggest linking based on name similarity
