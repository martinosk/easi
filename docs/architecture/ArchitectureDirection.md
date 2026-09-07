# Bounded Context Canvas: Architecture Direction

## Name
**Architecture Direction**

## Purpose
Plan and track change on domain capabilities. The context records the judgements architects make about a realisation — the TIME grade and the standard/legacy role — and the capability journeys that carry a planned change (migration, consolidation, carve-out, move, maturity) through milestones, a target period and progress. TIME suggestions computed from fit gaps are composed into assessment reads as advice beside the recorded judgement (spec 212). The enterprise capability, direction, standard application and composition concepts were retired by spec 213; everything here is keyed on a domain capability or a capability–application pair.

**Key Stakeholders:**
- Domain Architects (assess realisations, assign roles, plan and run capability journeys)
- Portfolio Managers (read TIME rollups and suggestions, journey timelines)

## Strategic Classification

### Domain Importance
**Core Domain** — the decision and execution layer that other EA tools rarely model explicitly, in the vocabulary strategic decisions are made in.

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **TIME Assessment** | Tolerate / Invest / Migrate / Eliminate grade recorded per capability–application realisation; stale after 12 months |
| **TIME Suggestion** | Tolerate / Invest / Migrate / Eliminate grade computed from fit gaps, importance and realisations; advice, never a recorded judgement |
| **Realisation Role** | Standard or legacy, per capability–application realisation |
| **Capability Journey** | A planned change to one capability over time: kind, target period, milestones, progress, status |
| **Journey Kind** | migration, consolidation, carve-out, move, or maturity; each kind defines its own required fields |
| **Maturity Journey** | A journey whose declared outcome is a maturity level above the capability's current maturity |
| **Capability Node Cache** | Local copy of the capability tree (level, parent, L1 ancestor, effective domain, maturity) fed by Capability Mapping events |
| **Reference-Name Cache** | Local copy of component and business-domain identity and names fed by Architecture Modeling and Capability Mapping events; answers existence checks |
| **Realisation Cache** | Local copy of direct capability–application realisations fed by Capability Mapping events; answers the direct-realisation lookup |

## Inbound Communication

**Commands** (from Frontend/API):
- TIME: `AssessRealization`, `RemoveTimeAssessment`
- Realisation roles: `AssignRealizationRole`, `ClearRealizationRole`
- Journeys: `PlanJourney`, `StartJourney`, `CompleteJourney`, `AbandonJourney`, `UpdateJourneyProgress`, `UpdateJourneyDetails`, `ChangeJourneySourceApplications`, milestone add / update / remove / reorder

**Queries** (from Frontend/API): TIME assessments (with composed suggestions) and rollups, realisation roles, journeys (+history)

**Events** (from other contexts):
- From **MetaModel**: `MetaModelConfigurationCreated`, `StrategyPillarAdded/Updated/Removed`, `PillarFitConfigurationUpdated` → strategy pillar cache
- From **Capability Mapping**: `Capability*` and `BusinessDomain*` → capability node cache and reference-name cache; `SystemLinkedToCapability`, `SystemRealizationDeleted` → realisation cache; `SystemRealizationDeleted` → remove TIME assessment and realisation role; `EffectiveImportanceRecalculated` → importance cache; `ApplicationFitScoreSet/Removed` → fit score cache
- From **Architecture Modeling**: `ApplicationComponent*` → reference-name cache and component names in the realisation cache
- From **Auth**: `UserCreated` → assessor / planner display names

### Relationship Types
- **Customer-Supplier** with MetaModel, Capability Mapping and Architecture Modeling (events into local, backfilled caches; every existence and realisation check reads a local cache — spec 209)

## Outbound Communication

**Events** (published to event bus):
- `TimeAssessmentRecorded/Removed`, `RealizationRoleAssigned/Cleared`, `Journey*` and `JourneyMilestone*`

No other context consumes these events. **Arch Assistant** calls the public API through loopback tools contributed by this context's published language.

## Business Rules

1. TIME assessments and realisation roles may only be recorded on an existing direct realisation.
2. TIME assessments and realisation roles are removed when the underlying realisation is deleted.
3. One active journey per capability; a journey's kind fixes its source-application cardinality and target fields.
4. A maturity journey carries no applications, requires a target maturity, and the target must exceed the capability's current maturity at planning time.
5. A move journey requires a target business domain; its target parent must effectively belong to that domain.
6. A journey's milestone order lists every milestone exactly once.

## Design Constraints

1. Every existence and realisation check reads a local, event-fed cache; the context issues no query to another context at request time (spec 209).
2. TIME suggestions are computed at query time, never stored — advice goes stale whenever fit scores move (spec 212).
3. Caches are eventually consistent with their upstream contexts and are backfilled by migration on deployment.

## Boundary Health
- **Cache freshness**: capability, pillar, realisation, importance and fit changes visible in suggestions within the event round-trip
- **Guard**: `TestNoCrossBoundedContextImports` — only published language crosses the boundary

## Architecture Notes

### Implementation Location
`/backend/internal/architecturedirection/`

### Key Packages
- `domain/aggregates/` — TimeAssessment, RealizationRoles, CapabilityJourney
- `domain/services/` — journey reference checks, direct-realisation lookup, TIME suggestion calculator
- `application/readmodels/` — TIME assessment, realisation role, journey read models; the TIME assessment view that composes suggestions into assessment reads; TIME suggestion read model; capability node / reference-name / realisation / pillar / importance / fit caches
- `application/projectors/` — projectors for own events, cache projectors for MetaModel, Capability Mapping and Architecture Modeling events, deletion reactors
- `infrastructure/api/` — routes and handlers; `SetupRoutes` needs only shared infrastructure (router, buses, database, HATEOAS, auth middleware, session provider)
- `infrastructure/metamodel/` — local strategy-pillar gateway

### API Style
- REST Level 3 with HATEOAS
- Endpoints: `/capabilities/{id}/components/{componentId}/time-assessment`, `/time-assessments` (+rollups), `/capabilities/{id}/components/{componentId}/realization-role`, `/realization-roles`, `/capabilities/{id}/journey` (+history), `/capability-journeys`

### Persistence
Schema `architecturedirection`; tables: `time_assessments`, `realization_roles`, `realization_role_aggregates`, `capability_journeys`, `capability_journey_sources`, `capability_journey_milestones`, `capability_node_cache`, `reference_name_cache`, `realization_cache`, `ea_importance_cache`, `ea_fit_score_cache`, `ea_strategy_pillar_cache`. Migration 154 dropped the read-model tables of the retired EnterpriseCapability, EnterpriseStrategicImportance, Direction and StandardApplication aggregates (spec 213); their stored events remain in `infrastructure.events`, unread.
