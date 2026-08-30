# Bounded Context Canvas: Enterprise Architecture

## Name
**Enterprise Architecture**

## Purpose
Own the identity of enterprise capabilities — cross-domain strategic themes such as "Digital Customer Engagement" — together with the two ratings made about them: strategic importance per pillar and target maturity. Provide TIME suggestions computed from strategic importance, fit scores and realisations.

Which domain capabilities compose an enterprise capability is **not** decided here: since spec 172 the association is the source set of the enterprise capability's active direction, owned by Architecture Direction, and since spec 207 everything derived from it (composition, source eligibility, maturity-gap analysis) is served by Architecture Direction. This context depends on nothing in Architecture Direction; Architecture Direction depends on this context's published events.

**Key Stakeholders:**
- Enterprise Architects (define enterprise capabilities, rate importance, set target maturity)
- Portfolio Managers (read TIME suggestions)

**Value Proposition:**
- A stable, tenant-wide vocabulary of enterprise capabilities
- Strategic importance ratings with rationale per pillar
- Target maturity per enterprise capability as the yardstick for gap analysis
- TIME suggestions derived from fit and importance

## Relationship to Other Contexts

| Question | Context |
|----------|---------|
| Which enterprise capabilities exist, how important are they, what maturity should they reach? | **Enterprise Architecture** |
| Which domain capabilities compose them, what do we intend to do, where are the maturity gaps? | **Architecture Direction** |
| Which domain capabilities exist and how are they realised? | **Capability Mapping** |

## Strategic Classification

### Domain Importance
**Core Domain** — enterprise capabilities are the vocabulary strategic decisions are made in.

### Business Model
**Engagement Creator & Decision Support**

### Evolution Stage
**Custom-Built**

## Domain Roles
- **Identity Holder**: enterprise capability name, description, category, active flag
- **Rater**: strategic importance per pillar with rationale; target maturity
- **Suggester**: TIME classification suggestions from fit gaps

## Inbound Communication

### Messages Received

**Commands** (from Frontend/API):
- `CreateEnterpriseCapability`, `UpdateEnterpriseCapability`, `DeleteEnterpriseCapability`
- `SetTargetMaturity`
- `SetEnterpriseStrategicImportance`, `UpdateEnterpriseStrategicImportance`, `RemoveEnterpriseStrategicImportance`

**Events** (from other contexts):
- From **MetaModel**: `MetaModelConfigurationCreated`, `StrategyPillarAdded/Updated/Removed`, `PillarFitConfigurationUpdated` → strategy pillar cache
- From **Capability Mapping**: `Capability*`, `CapabilityAssignedToDomain/UnassignedFromDomain`, `BusinessDomainUpdated` → domain capability metadata cache (used by TIME suggestions); `SystemLinkedToCapability`, `SystemRealizationDeleted` → realisation cache; `EffectiveImportanceRecalculated` → importance cache; `ApplicationFitScoreSet/Removed` → fit score cache
- From **Architecture Modeling**: `ApplicationComponentUpdated` → component names in the realisation cache

### Relationship Types
- **Customer-Supplier** with MetaModel and Capability Mapping (events into local caches)
- **Declared composition-root bridge** (query-time): business-domain name from Capability Mapping at assignment time — see `backend/internal/architecture_bridges_test.go`

## Outbound Communication

### Messages Sent

**Events** (published to event bus):
- `EnterpriseCapabilityCreated`, `EnterpriseCapabilityUpdated`, `EnterpriseCapabilityDeleted`, `EnterpriseCapabilityTargetMaturitySet`
- `EnterpriseStrategicImportanceSet`, `EnterpriseStrategicImportanceUpdated`, `EnterpriseStrategicImportanceRemoved`

### Collaborators
- **Architecture Direction** caches enterprise capability identity and target maturity from these events and rejects the active direction when an enterprise capability is deleted
- **OnePagers** indexes enterprise capabilities as one-pager subjects and reads built-in fields through a declared bridge
- **Arch Assistant** calls the public API through loopback tools contributed by this context's published language

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Enterprise Capability** | A logical capability that exists across business domains under different local names |
| **Category** | Free-text grouping of enterprise capabilities |
| **Target Maturity** | The maturity (0–99 on the tenant's scale) an enterprise capability should reach |
| **Strategic Importance** | Rating 1–5 per strategy pillar with optional rationale |
| **TIME Suggestion** | Tolerate / Invest / Migrate / Eliminate classification computed from fit gaps, importance and realisations; the recorded assessment lives in Architecture Direction |
| **Composition** | Owned by Architecture Direction (see its canvas); this context exposes only the `x-composition` navigation link |

## Business Decisions

### Core Business Rules
1. Enterprise capability names are unique within a tenant.
2. Deletion is a soft delete (`active = false`); consumers react through `EnterpriseCapabilityDeleted`.
3. One strategic importance rating per pillar per enterprise capability; rationale at most 500 characters.
4. Target maturity is 0–99.

### Policy Decisions
- Enterprise capabilities are optional; domain capabilities need not be part of one.
- Enterprise-level importance is a separate judgement from domain-level importance in Capability Mapping (flagged as a watch item in the 2026-08-29 capability coverage analysis).

## Verification Metrics
- **Cache freshness**: pillar, realisation, importance and fit caches updated within one event round-trip
- **Guard**: `TestContextDependencyGraphIsAcyclic` — no edge from this context to Architecture Direction

## Architecture Notes

### Implementation Location
`/backend/internal/enterprisearchitecture/`

### Key Packages
- `domain/` — EnterpriseCapability and EnterpriseStrategicImportance aggregates, TIME suggestion calculator
- `application/` — commands, handlers, projectors (own events + ACL caches), read models (enterprise capability, importance, TIME suggestion, caches)
- `infrastructure/` — REST API, repositories, local strategy-pillar gateway

### API Style
- REST Level 3 with HATEOAS; enterprise capability DTOs carry `x-direction` and `x-composition` links to Architecture Direction resources
- Endpoints: `/enterprise-capabilities` (CRUD, target maturity, strategic importance), `/time-suggestions`

### Persistence
Schema `enterprisearchitecture`: `enterprise_capabilities`, `enterprise_strategic_importance`, `domain_capability_metadata`, `ea_realization_cache`, `ea_importance_cache`, `ea_fit_score_cache`, `ea_strategy_pillar_cache`. The legacy `link_count` / `domain_count` columns on `enterprise_capabilities` are unused since spec 207.
