# Architecture Direction — Strategic DDD Memo

Strategic design for the Direction / Logical-vs-Physical capability model defined in `architecture-direction-model.md`. This memo defines bounded contexts, aggregates, integration patterns, events, and language. It does **not** define endpoints, schemas, or feature scope — that is the next agent's job.

The five-second-alignment-decision test from the spec is the load-bearing constraint here. Everything below is filtered through it: if a model element does not help a non-architect answer "is this with the target or against it?" in five seconds, it is rejected.

## 1. Bounded Context Shape

**Recommendation: a new bounded context `architecturedirection`, sibling to the existing two.**

```
                +--------------------------+
                |   capabilitymapping      |   AS-IS, operationally real
                |   (Physical layer)       |   PhysicalCapability (=Capability),
                |                          |   BusinessDomain, CapabilityRealization
                +-----------+--------------+
                            |
                            | Published Language
                            | (CapabilityCreated/Deleted/AssignedToDomain,
                            |  ApplicationFitScoreSet, etc.)
                            v
                +--------------------------+      +----------------------------+
                |  enterprisearchitecture  |<---->|   architecturedirection    |
                |  (Logical layer)         |      |   (Change-proposal layer)  |
                |  LogicalCapability,      |      |   Direction,               |
                |  StrategicImportance,    |      |   StandardAppDesignation,  |
                |  TIME, TargetMaturity    |      |   DiscoveryCandidate       |
                +--------------------------+      +----------------------------+
```

**Why split rather than absorb into `enterprisearchitecture`:**

- *Cohesion test.* `enterprisearchitecture` today answers "how do we describe and classify the steady-state landscape?" (TIME framework, target maturity, strategic importance, the logical grouping itself). The new concepts answer a different question: "what change are we proposing, and where is it on the agenda?" Steady-state classification and change proposals are distinct activities with distinct lifecycles, distinct audiences, and — critically — distinct ubiquitous languages. Stuffing both into one context will erode the language inside a year.
- *Coupling test.* A `Direction` references `PhysicalCapability` IDs (in `capabilitymapping`) far more often than it references `LogicalCapability`. If `Direction` lived inside `enterprisearchitecture`, that context would carry two integration relationships of equal weight to `capabilitymapping`, with no hierarchy between them. A separate context lets each integration line carry one purpose.
- *Evolution test.* Direction status workflow, horizons, group decision semantics, and Discover candidates will evolve quickly while the spec settles. The Logical layer (renamed EnterpriseCapability + StrategicImportance + TIME) is the stable, mature half. Don't drag the stable half through the volatility of the new half.
- *Five-second test.* "Is there a Direction on this capability?" is the daily-decision query. It deserves a context whose name announces the answer.

**What stays in `enterprisearchitecture`:** the rename `EnterpriseCapability → LogicalCapability` (and `EnterpriseCapabilityLink → LogicalCapabilityMapping`), `EnterpriseStrategicImportance`, target maturity, TIME classification. These are properties of the logical layer itself, not of any change proposal.

## 2. Aggregate Boundaries

| Aggregate | Context | Invariants enforced inside the boundary |
|---|---|---|
| `LogicalCapability` (renamed from `EnterpriseCapability`) | `enterprisearchitecture` | Name uniqueness within tenant; description/category constraints; target-maturity validity. Lifecycle: created → updated → deleted. Unchanged from today. |
| `LogicalCapabilityMapping` (renamed `EnterpriseCapabilityLink`) | `enterprisearchitecture` | A physical capability maps to **at most one** logical capability at a time (the spec says "0..1"). The mapping cannot be created against an inactive logical capability — already enforced. |
| `Direction` | `architecturedirection` | Status transition rules (`draft → proposed → agreed`/`rejected`, terminal states); type/source-cardinality consistency (`consolidate` requires N ≥ 2 sources, `decompose` requires exactly 1, `stay` requires exactly 1); horizon validity; immutability of the source-capability set after `agreed`. Narrative + decision audit trail attached at every status transition. |
| `StandardAppDesignation` | `architecturedirection` | One active `agreed` designation per `LogicalCapability` at a time (replacement supersedes). Status transition rules identical to `Direction`. References exactly one `Application` and one `LogicalCapability`. |
| `DiscoveryCandidate` | `architecturedirection` | Lifecycle: `open → promoted-to-direction` / `promoted-to-grouping` / `rejected`. Carries the candidate signal (the multi-domain theme) and provenance, not domain truth. Once promoted, stores the resulting `Direction` or `LogicalCapability` ID for traceability. |

**Consistency boundaries — must be transactionally consistent:**

- A `Direction`'s status, narrative, source-capabilities, placements, and horizon. (One aggregate, one txn.)
- A `StandardAppDesignation`'s status, application reference, narrative.
- A `LogicalCapabilityMapping`'s existence (mapping a physical to a logical).

**Consistency boundaries — eventually consistent:**

- The set of physical capabilities that a `Direction` references. The `Direction` stores IDs; if a `PhysicalCapability` is deleted in `capabilitymapping`, the `Direction` learns asynchronously and must surface this as a "stale reference" condition rather than block the deletion. The strategic-communication tool does not own physical reality.
- The link from a `DiscoveryCandidate` to its produced `Direction` / `LogicalCapability` after promotion.
- All read models combining physical + logical + direction (Direction Map, Logical Capability Map, Target Architecture Synthesis).

**What is explicitly *not* an aggregate:**

- "Horizon" is a value object (enum: `now | next | later`) on `Direction` and `StandardAppDesignation`. Not an aggregate. Has no ID, no lifecycle.
- "Placement" is a value object: a target `BusinessDomainID` plus an optional resulting-name hint. A `Direction` carries 1..N placements as part of its own state. It is not a separate aggregate.
- A "phased path" is *not* a concept. It is a `LogicalCapability` (with optional `StandardAppDesignation` on horizon `now`) plus a `Direction` (on horizon `later`) sharing source capabilities. Resist any urge to model `Phase` as an entity — that path leads to project-management semantics. The five-second test forbids it.
- "Group decision" / "agenda item" is *not* an aggregate. Status transitions on `Direction` and `StandardAppDesignation` carry the decision narrative directly. There is no separate `Decision` aggregate.

## 3. Context Relationships

| From | To | Pattern | Mechanism |
|---|---|---|---|
| `architecturedirection` | `capabilitymapping` | **Customer / Conformist** (downstream) | Subscribes to `capabilitymapping/publishedlanguage` events: `CapabilityCreated`, `CapabilityDeleted`, `CapabilityAssignedToDomain`, `CapabilityUnassignedFromDomain`. Stores `PhysicalCapabilityID` (= `CapabilityID`) as an opaque ID value object. Never imports the `Capability` aggregate. |
| `architecturedirection` | `enterprisearchitecture` | **Customer / Conformist** (downstream) | Subscribes to `LogicalCapabilityCreated/Deleted` (post-rename). Stores `LogicalCapabilityID` as an opaque value object. |
| `architecturedirection` | external Application catalog | **Customer / Anti-Corruption Layer** | `StandardAppDesignation` references an `ApplicationID`. The application model lives outside both contexts (the existing component/system catalog). Wrap the lookup in an ACL inside `architecturedirection/infrastructure` so a future move of the catalog doesn't ripple. |
| `enterprisearchitecture` | `capabilitymapping` | **Customer / Conformist** (already in place) | Existing pattern via projectors importing `cmPL` — keep as-is post-rename. |
| `enterprisearchitecture` | `architecturedirection` | **None — no inbound dependency** | The Logical layer must not know about Directions. Read models that combine the two are projected from events of *both* contexts inside a query/read-model module (a separate context, or under `architectureviews`). |

This gives `architecturedirection` exactly **three** outbound integration lines (physical, logical, application) and **zero** inbound from the existing two. That is the asymmetry the split is designed to produce.

**Anti-corruption surface.** Each inbound relationship is one-directional and event-shaped. The `architecturedirection` context never calls into a sibling's repository or aggregate. It maintains its own read-side projection of "physical capabilities I care about" and "logical capabilities I care about" — populated by event subscription, not by query. This is identical to the pattern already used by `enterprisearchitecture` projectors today.

## 4. Domain Events

```
architecturedirection
├── DiscoveryCandidateOpened          (theme detected from cross-domain signal)
├── DiscoveryCandidatePromotedToDirection
├── DiscoveryCandidatePromotedToGrouping
├── DiscoveryCandidateRejected
├── DirectionDrafted
├── DirectionProposed
├── DirectionAgreed                   <-- the high-value event
├── DirectionRejected
├── DirectionPlacementsChanged
├── DirectionHorizonChanged
├── DirectionNarrativeUpdated
├── DirectionSourceCapabilitiesChanged   (only legal pre-`agreed`)
├── StandardAppDesignationDrafted
├── StandardAppDesignationProposed
├── StandardAppDesignationAgreed
├── StandardAppDesignationRejected
└── StandardAppDesignationSuperseded   (when a new `agreed` designation replaces an old one)
```

**Cross-context flow when `DirectionAgreed` is published:**

- The Discover read model marks the source candidate (if any) as resolved.
- The Logical Capability Map read model marks affected logical capabilities with a "has agreed direction" badge — this is the five-second-decision payload.
- The Target Architecture Synthesis view recomputes its horizon timeline.
- A notification consumer (existing infra) can push to subscribed architects.
- Critically: nothing in `capabilitymapping` reacts. Agreement is a strategic statement; physical reality changes only when humans act on it. Realization tracking is explicitly out of scope per the spec.

**Status-as-events vs status-as-field.** Each transition is its own past-tense event (`DirectionProposed`, `DirectionAgreed`, …) rather than a generic `DirectionStatusChanged`. This matches the EASI codebase pattern (one event per business fact) and lets projections subscribe to "agreed" without filtering. The aggregate's current status is reconstructed from event history.

## 5. Ubiquitous Language

| Spec term | Codebase name | Notes |
|---|---|---|
| Physical capability | `Capability` (existing aggregate, unchanged) | Keep the type name `Capability`. **Do not rename to `PhysicalCapability` in code.** The spec uses "Physical capability" as a *clarifying adjective* against "Logical"; in code, `Capability` is unambiguous because `LogicalCapability` is its own type. Renaming the existing aggregate would generate a massive churn migration for zero semantic gain. Document the equivalence in package-level Godoc and in the `capabilitymapping` README. |
| Logical capability | `LogicalCapability` | Rename from `EnterpriseCapability`. Aggregate, value objects, events, repository, read models, API routes all rename. Data migrates 1:1. This is a self-contained spec (item 1 in the spec's feature list). |
| Logical mapping | `LogicalCapabilityMapping` | Rename from `EnterpriseCapabilityLink`. "Mapping" reads cleaner than "Link" given the 0..1 cardinality and the strategic-classification semantics. |
| Direction | `Direction` (aggregate) | **Naming conflict warning.** "Direction" is overloaded in software (graph edge direction, sort direction, navigation). Inside `architecturedirection` the package qualifier disambiguates. Cross-context APIs and DTOs must spell it `ArchitectureDirection` or use the package-qualified form. Read models exposed via HTTP should use `architecture-direction` in URLs, never bare `direction`. |
| StandardAppDesignation | `StandardAppDesignation` | The full name is worth its weight; abbreviating to `Standard` or `Designation` collides with too many neighbours. Alternative `StandardApplicationDesignation` was considered — `App` is the term DFDS architects actually use, so keep it. |
| Horizon | `Horizon` (value object, enum) | Avoid `Timeframe`, `TimeHorizon`, `Phase` — all carry project-planning baggage. `Horizon` is the spec's word and matches the EA literature. |
| Discovery candidate | `DiscoveryCandidate` | Keep. "Candidate" cleanly communicates "not yet a decision." |
| `now / next / later` | preserved as-is | Resist any temptation to add dates to these. The whole point of the horizon enum is that it does *not* have a date — adding one breaks the five-second test. |
| Status values | `draft`, `proposed`, `agreed`, `rejected` | Spelled exactly as the spec writes them. These are the words the architects say in the room. |

**Term to retire:** "Enterprise capability." It conflated two concepts and the rename is the whole point. After the migration, no API, package, table, or doc may use the term except in migration notes.

## 6. Migration Posture

Existing data must travel. Concretely:

- Event-store events of type `EnterpriseCapability*` are **kept unchanged on disk** (event-sourcing rule: never mutate history). Add **upcasters** in the new `LogicalCapability` aggregate that map the old event type names into the new ones at deserialization time. Pattern is already established in `easi-go-backend-patterns`.
- Read-model tables rename via migration. The `LogicalCapabilityMapping` table maps 1:1 from `enterprise_capability_link`.
- Public API routes get new paths but the old paths can stay for a release as redirects or duplicate handlers. A separate spec scopes this.
- `architecturedirection` is greenfield. No migration. Tables, schemas, repositories, projectors, read models all new.

## 7. What This Memo Deliberately Does Not Say

- No HTTP route shapes, DTO field lists, or HATEOAS structure. (Next-spec territory.)
- No UI flow, canvas layout, or read-model SQL.
- No decision on how Discover candidates are *generated* (manual entry vs analyser vs cross-domain signal). The aggregate exists; how it is opened is downstream.
- No commitment on whether `DiscoveryCandidate` is its own aggregate forever or eventually folds into `Direction` as a pre-state. Start it as a separate aggregate; revisit if it stays thin.
- No decision on group-decision permissions / RBAC. That is the access-delegation context's concern.

## 8. Hand-off to the product-spec writer

The work decomposes into roughly seven self-contained vertical slices, each one shippable:

1. Rename `EnterpriseCapability → LogicalCapability` (event upcasters, table renames, API renames; behaviour unchanged).
2. Create the `architecturedirection` bounded context skeleton (package layout, published-language stubs, no aggregates yet).
3. Add `Direction` aggregate, status workflow, source-capability binding, narrative.
4. Add `StandardAppDesignation` aggregate, including the supersession rule.
5. Subscribe `architecturedirection` to `capabilitymapping` and `enterprisearchitecture` events; build the local "physical-of-interest" and "logical-of-interest" read projections.
6. Add `DiscoveryCandidate` aggregate and the two-path promotion handlers.
7. Compose the cross-context read views (Direction Map, Logical Capability Map, Target Architecture Synthesis) — placement TBD: either inside `architectureviews` or a new query-side module. The spec writer should make that call once the projections in (5) are real.

Slices 1 and 2 can run in parallel. Slice 3 depends on 2. Slice 4 depends on 2 and the `ApplicationID` ACL. Slice 5 depends on 1 and 2. Slice 6 depends on 3, 4, 5. Slice 7 depends on everything.
