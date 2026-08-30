# Bounded Context Canvas: Architecture Direction

## Name
**Architecture Direction**

## Purpose
Govern what the architecture group intends to do with each enterprise capability and record the judgements that follow from it: the direction (consolidate, decompose, stay) with its source capabilities and horizon, the standard application, TIME assessments and standard/legacy roles of realisations, and the capability journeys that execute the direction. Since spec 172 the direction's source set *is* the association between an enterprise capability and its domain capabilities, so this context also owns everything derived from that association: composition, source eligibility, composition summaries and maturity-gap analysis (spec 207).

**Key Stakeholders:**
- Enterprise Architects (capture and agree directions, designate standard applications)
- Domain Architects (plan and run capability journeys)
- Portfolio Managers (read maturity analysis and TIME rollups)

**Value Proposition:**
- One place where intent about an enterprise capability is recorded, proposed and agreed
- Composition of an enterprise capability follows from its agreed sources, never from a separate link
- Judgements about realisations (TIME, standard/legacy) sit next to the direction they support
- Journeys turn an agreed direction into milestones and progress

## Strategic Classification

### Domain Importance
**Core Domain** — the decision and execution layer that other EA tools rarely model explicitly.

### Business Model
**Decision Support & Compliance Enforcer**

### Evolution Stage
**Custom-Built**

## Domain Roles
- **Decision Record**: holds directions and their lifecycle (draft → proposed → agreed / rejected)
- **Association Owner**: a direction's sources define which domain capabilities compose the enterprise capability
- **Composition Resolver**: derives composition, carve-outs, counts and eligibility from active directions
- **Assessor**: records TIME grades and standard/legacy roles per realisation
- **Journey Planner**: milestones, target period and progress per capability journey

## Inbound Communication

### Messages Received

**Commands** (from Frontend/API):
- Direction: `CaptureDirection`, `UpdateDirection`, `ProposeDirection`, `AgreeDirection`, `RejectDirection`, `AddDirectionSource`, `RemoveDirectionSource`
- Standard application: `SetStandardApplication`
- TIME: `AssessRealization`, `RemoveTimeAssessment`
- Realisation roles: `AssignRealizationRole`, `ClearRealizationRole`
- Journeys: `PlanJourney`, `StartJourney`, `CompleteJourney`, `AbandonJourney`, `UpdateJourneyProgress`, `UpdateJourneyDetails`, `ChangeJourneySourceApplications`, milestone add / update / remove / reorder

**Queries** (from Frontend/API): direction per enterprise capability, composition, composition summaries, source candidates, composition preview, maturity analysis and gap detail, standard application (+history), TIME assessments and rollups, realisation roles, journeys (+history)

**Events** (from other contexts):
- From **Enterprise Architecture**: `EnterpriseCapabilityCreated/Updated/Deleted/TargetMaturitySet` → enterprise capability cache; `EnterpriseCapabilityDeleted` → reject the active direction
- From **Capability Mapping**: `Capability*` and `BusinessDomain*` → capability node cache (hierarchy, effective domain, maturity), reference-name cache, stale-source detection; `SystemRealizationDeleted` → remove TIME assessment and realisation role
- From **Architecture Modeling**: `ApplicationComponent*` → reference-name cache, stale standard applications
- From **Auth**: `UserCreated` → assessor / planner display names

### Collaborators
- **Enterprise Architecture**: supplier of enterprise capability identity and target maturity
- **Capability Mapping**: supplier of the capability tree, domains and realisations
- **Architecture Modeling**: supplier of application identity

### Relationship Types
- **Customer-Supplier** with Enterprise Architecture, Capability Mapping and Architecture Modeling (events into local caches)
- **Declared composition-root bridges** (query-time, see `backend/internal/architecture_bridges_test.go`): capability / business-domain existence, direct realisation lookup and effective-domain check from Capability Mapping; component existence from Architecture Modeling

## Outbound Communication

### Messages Sent

**Events** (published to event bus): `DirectionDrafted/Proposed/Agreed/Rejected/NarrativeUpdated/HorizonChanged/PlacementsChanged/SourceCapabilitiesChanged`, `StandardApplicationSet`, `TimeAssessmentRecorded/Removed`, `RealizationRoleAssigned/Cleared`, `Journey*` and `JourneyMilestone*`

### Collaborators
- **OnePagers** reads composition through a declared bridge for the enterprise-capability relation field
- **Arch Assistant** calls the public API through loopback tools contributed by this context's published language

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Direction** | The architecture group's intent for an enterprise capability: consolidate, decompose or stay, with narrative, horizon (now / next / later) and placements |
| **Source Capability** | A domain capability explicitly named by a direction; a capability is the explicit source of at most one active direction |
| **Composition** | The source capabilities of the active direction plus their subtrees, minus subtrees carved out by a more specific source on another enterprise capability |
| **Carve-out** | A subtree claimed by another enterprise capability's direction and therefore excluded |
| **Composition Summary** | Per enterprise capability: source, included, carved-out and domain counts plus direction status |
| **Source Eligibility** | Whether a capability may be added as a source without conflicting with another active direction |
| **Standard Application** | The application the group recorded as the one that should realise the enterprise capability |
| **TIME Assessment** | Tolerate / Invest / Migrate / Eliminate grade recorded per capability–application realisation; stale after 12 months |
| **Realisation Role** | Standard or legacy, per capability–application realisation |
| **Capability Journey** | Execution plan for a direction: kind, target period, milestones, progress, status |
| **Capability Node Cache** | Local copy of the capability tree (level, parent, L1 ancestor, effective domain, maturity) fed by Capability Mapping events |
| **Enterprise Capability Cache** | Local copy of enterprise capability identity, activity and target maturity fed by Enterprise Architecture events |

## Business Decisions

### Core Business Rules
1. A direction can only be captured on an existing, active enterprise capability; one active direction per enterprise capability.
2. A domain capability may be the explicit source of at most one active direction (R1); conflicts are rejected with the conflicting enterprise capability named.
3. Composition is resolved with most-specific-wins carve-outs (R2, spec 172) and is computed per request from local caches.
4. Propose-time cardinality: consolidate ≥ 2 sources, decompose exactly 1, stay exactly 1; a narrative is required to propose.
5. Deleting the enterprise capability rejects its active direction and releases its sources.
6. TIME assessments and realisation roles are removed when the underlying realisation is deleted.

### Policy Decisions
- Composition, source eligibility and maturity analysis are served under the enterprise-architecture read permission they had before spec 207.
- Caches are eventually consistent with their upstream contexts and are backfilled by migration on deployment.

## Verification Metrics
- **Cache freshness**: capability and enterprise capability changes visible in composition within the event round-trip
- **Eligibility integrity**: zero capabilities that are explicit sources of two active directions
- **Guard**: `TestContextDependencyGraphIsAcyclic` — this context depends on Enterprise Architecture, never the reverse

## Architecture Notes

### Implementation Location
`/backend/internal/architecturedirection/`

### Key Packages
- `domain/aggregates/` — Direction, StandardApplication, TimeAssessment, RealizationRoles, CapabilityJourney
- `domain/services/` — composition resolver (R2), reference and eligibility services
- `application/services/` — composition service (composition, counts, preview, source candidates, direction status)
- `application/readmodels/` — direction, standard application, TIME, realisation role, journey read models; capability node cache; enterprise capability cache; maturity analysis
- `application/projectors/` — projectors for own events, cache projectors for Capability Mapping and Enterprise Architecture events, stale-reference projectors
- `infrastructure/api/` — routes, handlers, composition wiring (`NewCompositionService` is the constructor OnePagers' bridge uses)

### Persistence
Schema `architecturedirection`; tables include `directions`, `direction_source_capabilities`, `standard_applications`, `time_assessments`, `realization_roles`, `capability_journeys`, `reference_name_cache`, `capability_domain_cache`, `capability_node_cache`, `enterprise_capability_cache`. Migrations 137/138 add and backfill the two spec-207 caches.
