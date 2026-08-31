# Bounded Context Canvas: Architecture Direction

## Name
**Architecture Direction**

## Purpose
Own the enterprise capability — the cross-domain strategic theme such as "Digital Customer Engagement" — and govern what the architecture group intends to do with it: the direction (consolidate, decompose, stay) with its source capabilities and horizon, the standard application, TIME assessments and standard/legacy roles of realisations, and the capability journeys that execute the direction. Since spec 172 the direction's source set *is* the association between an enterprise capability and its domain capabilities, so this context also owns everything derived from that association: composition, source eligibility, composition summaries and maturity-gap analysis (spec 207). Since spec 210 it also owns enterprise capability identity itself, together with the two ratings made about it — strategic importance per pillar and target maturity — and the TIME suggestions computed from fit gaps.

**Key Stakeholders:**
- Enterprise Architects (define enterprise capabilities, rate importance, set target maturity, capture and agree directions, designate standard applications)
- Domain Architects (plan and run capability journeys)
- Portfolio Managers (read maturity analysis, TIME rollups and TIME suggestions)

**Value Proposition:**
- One place where intent about an enterprise capability is recorded, proposed and agreed
- A stable, tenant-wide vocabulary of enterprise capabilities with strategic importance and target maturity
- Composition of an enterprise capability follows from its agreed sources, never from a separate link
- Judgements about realisations (TIME, standard/legacy) sit next to the direction they support
- Journeys turn an agreed direction into milestones and progress

## Strategic Classification

### Domain Importance
**Core Domain** — the decision and execution layer that other EA tools rarely model explicitly, in the vocabulary strategic decisions are made in.

### Business Model
**Decision Support & Compliance Enforcer**

### Evolution Stage
**Custom-Built**

## Domain Roles
- **Identity Holder**: enterprise capability name, description, category, active flag
- **Rater**: strategic importance per pillar with rationale; target maturity
- **Decision Record**: holds directions and their lifecycle (draft → proposed → agreed / rejected)
- **Association Owner**: a direction's sources define which domain capabilities compose the enterprise capability
- **Composition Resolver**: derives composition, carve-outs, counts and eligibility from active directions
- **Assessor**: records TIME grades and standard/legacy roles per realisation
- **Suggester**: TIME classification suggestions from fit gaps
- **Journey Planner**: milestones, target period and progress per capability journey

## Inbound Communication

### Messages Received

**Commands** (from Frontend/API):
- Enterprise capability: `CreateEnterpriseCapability`, `UpdateEnterpriseCapability`, `DeleteEnterpriseCapability`, `SetTargetMaturity`
- Strategic importance: `SetEnterpriseStrategicImportance`, `UpdateEnterpriseStrategicImportance`, `RemoveEnterpriseStrategicImportance`
- Direction: `CaptureDirection`, `UpdateDirection`, `ProposeDirection`, `AgreeDirection`, `RejectDirection`, `AddDirectionSource`, `RemoveDirectionSource`
- Standard application: `SetStandardApplication`
- TIME: `AssessRealization`, `RemoveTimeAssessment`
- Realisation roles: `AssignRealizationRole`, `ClearRealizationRole`
- Journeys: `PlanJourney`, `StartJourney`, `CompleteJourney`, `AbandonJourney`, `UpdateJourneyProgress`, `UpdateJourneyDetails`, `ChangeJourneySourceApplications`, milestone add / update / remove / reorder

**Queries** (from Frontend/API): enterprise capabilities and their strategic importance, direction per enterprise capability, composition, composition summaries, source candidates, composition preview, maturity analysis and gap detail, TIME suggestions, standard application (+history), TIME assessments and rollups, realisation roles, journeys (+history)

**Events** (from other contexts):
- From **MetaModel**: `MetaModelConfigurationCreated`, `StrategyPillarAdded/Updated/Removed`, `PillarFitConfigurationUpdated` → strategy pillar cache
- From **Capability Mapping**: `Capability*` and `BusinessDomain*` → capability node cache (hierarchy, effective domain, maturity, existence), domain capability metadata cache (used by TIME suggestions), business-domain name cache, reference-name cache (domain existence and names; rows removed on `BusinessDomainDeleted`), stale-source detection; `SystemLinkedToCapability`, `SystemRealizationDeleted` → realisation caches of direct realisations (the lookup behind TIME assessments and realisation roles); `SystemRealizationDeleted` → remove TIME assessment and realisation role; `EffectiveImportanceRecalculated` → importance cache; `ApplicationFitScoreSet/Removed` → fit score cache
- From **Architecture Modeling**: `ApplicationComponent*` → reference-name cache (component existence and names; rows removed on `ApplicationComponentDeleted`), component names in the realisation cache, stale standard applications
- From **Auth**: `UserCreated` → assessor / planner display names

### Collaborators
- **MetaModel**: supplier of strategy pillars and fit configuration
- **Capability Mapping**: supplier of the capability tree, domains, realisations, importance and fit scores
- **Architecture Modeling**: supplier of application identity

### Relationship Types
- **Customer-Supplier** with MetaModel, Capability Mapping and Architecture Modeling (events into local, backfilled caches; every existence and realisation check reads a local cache — spec 209)

## Outbound Communication

### Messages Sent

**Events** (published to event bus):
- `EnterpriseCapabilityCreated`, `EnterpriseCapabilityUpdated`, `EnterpriseCapabilityDeleted`, `EnterpriseCapabilityTargetMaturitySet`
- `EnterpriseStrategicImportanceSet`, `EnterpriseStrategicImportanceUpdated`, `EnterpriseStrategicImportanceRemoved`
- `DirectionDrafted/Proposed/Agreed/Rejected/NarrativeUpdated/HorizonChanged/PlacementsChanged/SourceCapabilitiesChanged`, `StandardApplicationSet`, `TimeAssessmentRecorded/Removed`, `RealizationRoleAssigned/Cleared`, `Journey*` and `JourneyMilestone*`

### Collaborators
- **OnePagers** indexes enterprise capabilities as one-pager subjects and caches their built-in fields from the `EnterpriseCapability*` events
- **Arch Assistant** calls the public API through loopback tools contributed by this context's published language
- No other context consumes composition; the enterprise-capability one-pager no longer shows it (spec 209)

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Enterprise Capability** | A logical capability that exists across business domains under different local names |
| **Category** | Free-text grouping of enterprise capabilities |
| **Target Maturity** | The maturity (0–99 on the tenant's scale) an enterprise capability should reach |
| **Strategic Importance** | Rating 1–5 per strategy pillar with optional rationale |
| **Direction** | The architecture group's intent for an enterprise capability: consolidate, decompose or stay, with narrative, horizon (now / next / later) and placements |
| **Source Capability** | A domain capability explicitly named by a direction; a capability is the explicit source of at most one active direction |
| **Composition** | The source capabilities of the active direction plus their subtrees, minus subtrees carved out by a more specific source on another enterprise capability |
| **Carve-out** | A subtree claimed by another enterprise capability's direction and therefore excluded |
| **Composition Summary** | Per enterprise capability: source, included, carved-out and domain counts plus direction status |
| **Source Eligibility** | Whether a capability may be added as a source without conflicting with another active direction |
| **Standard Application** | The application the group recorded as the one that should realise the enterprise capability |
| **TIME Assessment** | Tolerate / Invest / Migrate / Eliminate grade recorded per capability–application realisation; stale after 12 months |
| **TIME Suggestion** | Tolerate / Invest / Migrate / Eliminate classification computed from fit gaps, importance and realisations; advice, never a recorded judgement |
| **Realisation Role** | Standard or legacy, per capability–application realisation |
| **Capability Journey** | Execution plan for a direction: kind, target period, milestones, progress, status |
| **Capability Node Cache** | Local copy of the capability tree (level, parent, L1 ancestor, effective domain, maturity) fed by Capability Mapping events |
| **Reference-Name Cache** | Local copy of component and business-domain identity and names fed by Architecture Modeling and Capability Mapping events; answers existence checks |
| **Realisation Cache** | Local copy of direct capability–application realisations fed by Capability Mapping events; answers the direct-realisation lookup |

## Business Decisions

### Core Business Rules
1. Enterprise capability names are unique within a tenant.
2. Deletion of an enterprise capability is a soft delete (`active = false`); consumers react through `EnterpriseCapabilityDeleted`.
3. One strategic importance rating per pillar per enterprise capability; rationale at most 2000 characters. Target maturity is 0–99.
4. A direction can only be captured on an existing, active enterprise capability; one active direction per enterprise capability.
5. A domain capability may be the explicit source of at most one active direction (R1); conflicts are rejected with the conflicting enterprise capability named.
6. Composition is resolved with most-specific-wins carve-outs (R2, spec 172) and is computed per request from local caches.
7. Propose-time cardinality: consolidate ≥ 2 sources, decompose exactly 1, stay exactly 1; a narrative is required to propose.
8. Deleting the enterprise capability rejects its active direction and releases its sources.
9. TIME assessments and realisation roles are removed when the underlying realisation is deleted.

### Policy Decisions
- Enterprise capabilities are optional; domain capabilities need not be part of one.
- Enterprise-level importance is a separate judgement from domain-level importance in Capability Mapping (flagged as a watch item in the 2026-08-29 capability coverage analysis).
- Enterprise capability, composition, source eligibility, maturity analysis and TIME suggestions are served under the `enterprise-arch:*` permissions they had before specs 207 and 210; direction, standard application and journey writes use `architecture-direction:*`.
- Caches are eventually consistent with their upstream contexts and are backfilled by migration on deployment.

## Verification Metrics
- **Cache freshness**: capability, pillar, realisation, importance and fit changes visible in composition and suggestions within the event round-trip
- **Eligibility integrity**: zero capabilities that are explicit sources of two active directions
- **Guard**: `TestNoCrossBoundedContextImports` — only published language crosses the boundary

## Architecture Notes

### Implementation Location
`/backend/internal/architecturedirection/`

### Key Packages
- `domain/aggregates/` — EnterpriseCapability, EnterpriseStrategicImportance, Direction, StandardApplication, TimeAssessment, RealizationRoles, CapabilityJourney
- `domain/services/` — composition resolver (R2), reference and eligibility services, TIME suggestion calculator
- `application/services/` — composition service (composition, counts, preview, source candidates, direction status)
- `application/readmodels/` — enterprise capability, strategic importance, TIME suggestion, direction, standard application, TIME, realisation role, journey read models; capability node cache; domain capability metadata; pillar / realisation / importance / fit / business-domain-name caches; maturity analysis
- `application/projectors/` — projectors for own events, cache projectors for MetaModel, Capability Mapping and Architecture Modeling events, stale-reference projectors
- `infrastructure/api/` — routes, handlers, composition wiring; `SetupRoutes` needs only shared infrastructure (router, buses, database, HATEOAS, auth middleware, session provider)
- `infrastructure/metamodel/` — local strategy-pillar gateway

### API Style
- REST Level 3 with HATEOAS; enterprise capability DTOs carry `x-direction` and `x-composition` links
- Endpoints: `/enterprise-capabilities` (CRUD, target maturity, strategic importance, direction, standard application, composition, maturity gap), `/enterprise-capability-compositions`, `/time-suggestions`, `/time-assessments`, `/realization-roles`, `/capability-journeys`

### Persistence
Schema `architecturedirection`; tables include `enterprise_capabilities`, `enterprise_strategic_importance`, `directions`, `direction_source_capabilities`, `standard_applications`, `time_assessments`, `realization_roles`, `capability_journeys`, `reference_name_cache`, `capability_domain_cache`, `capability_node_cache`, `realization_cache`, `domain_capability_metadata`, `business_domain_name_cache`, `ea_realization_cache`, `ea_importance_cache`, `ea_fit_score_cache`, `ea_strategy_pillar_cache`. Migrations 137/138 add and backfill the spec-207 capability node cache; 145/146 add and backfill the realisation cache, reconcile the reference-name cache and make `maturity_value` non-null (spec 209); 150/151 re-parent the enterprisearchitecture tables into this schema, drop the now-redundant `enterprise_capability_cache` and re-seed the relocated backfills (spec 210). The legacy `link_count` / `domain_count` columns on `enterprise_capabilities` are unused since spec 207.
