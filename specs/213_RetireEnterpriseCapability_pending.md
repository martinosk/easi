# 213 — Retire the Enterprise Capability

> **Status:** pending
> **Depends on:** [211_MaturityJourneys](211_MaturityJourneys_done.md), [212_UnifiedTimeVocabulary](212_UnifiedTimeVocabulary_done.md) — design: [docs/specs/enterprise-capability.md](../docs/specs/enterprise-capability.md)
> **Roadmap alignment:** SD1 / H1-1

---

## Problem Statement

The enterprise capability asks users to curate a taxonomy nobody browses. Since spec 172 the aggregate has been a shell whose only association with real capabilities runs through a Direction; the strategic-importance rating has never had a UI. With maturity now expressed as a journey (211) and the TIME suggestion sitting beside the assessment it advises (212), nothing of value depends on the concept any more.

Direction, Standard Application and Composition go with it. Each is a statement *about* an enterprise capability — a Direction's subject is one, a Standard Application is set per one, and Composition exists only to assemble one from source capabilities — and `DirectionPanel` and `StandardApplicationPanel` render nowhere except inside the enterprise capability detail panel. Removing the subject leaves them without a subject and without a surface.

Strategic fit analysis is the one thing on the enterprise architecture page worth keeping. It is already owned by Capability Mapping, scores domain capabilities against strategy pillars, and never referenced an enterprise capability. It needs a page of its own, nothing more.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain / Enterprise Architect** | Reach strategic fit analysis without going through a page about a concept that no longer exists |
| **Enterprise Architect** | Stop maintaining a taxonomy that never informed a decision |
| **Assistant user** | Not be offered tools for concepts that no longer exist |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: The enterprise capability is gone

  Scenario: Strategic fit analysis has its own page
    Given an architect with capability read permission
    When they open strategic fit analysis from the main navigation
    Then they pick a strategy pillar and see the same analysis as before, unchanged

  Scenario: The enterprise architecture page is gone
    When a user navigates to the enterprise architecture page
    Then it no longer exists, and no navigation entry points at it

  Scenario: Enterprise capability endpoints are gone
    When any client calls an /enterprise-capabilities route
    Then it responds 404

  Scenario: One-pagers no longer offer the subject type
    Given a one-pager configuration is being created
    Then "enterprise-capability" is not among the subject types offered

  Scenario: Existing one-pager data for retired subjects is archived
    Given one-pagers held facts about enterprise capabilities
    When the retirement is deployed
    Then those facts are archived rather than left addressable

  Scenario: Journeys, TIME assessments and realisation roles are untouched
    Given a capability with a journey, a TIME assessment and a realisation role
    When the retirement is deployed
    Then all three read and write exactly as before

  Scenario: The assistant loses the retired tools
    When an assistant user asks about enterprise capabilities
    Then no enterprise-capability, direction, standard-application or composition tool is available
```

---

## Business Rules & Invariants

1. **Four aggregates retire** — `EnterpriseCapability`, `EnterpriseStrategicImportance`, `Direction` and `StandardApplication`, with their commands, events, repositories, read models and tables. Composition is a service over Direction and retires with it.
2. **Architecture Direction keeps only planning** — `CapabilityJourney`, `TimeAssessment` and `RealizationRole` survive, all keyed on a domain capability or a capability-application pair. The context speaks one language.
3. **Stored events are not rewritten** — retired event types stay in `infrastructure.events`, unread. No migration touches history; only read-model tables are dropped.
4. **OnePagers loses one subject type** — `enterprise-capability` retires from the subject-type value object, the relation catalog, the built-in field catalog, the subject index and the subject-deleted reactor. Existing facts for that subject type are archived through the context's own archival path, not deleted by SQL.
5. **Strategic fit analysis is untouched behind the API** — `GET /strategic-fit-analysis/{pillarId}` and its Capability Mapping ownership do not change; only the UI that calls it moves.
6. **Permissions retire** — the `enterprise-arch:*` group is removed, and every surviving route it guarded is re-pointed at an existing capability or architecture-direction permission.
7. **Agent tools retire with their routes** — the eight enterprise-capability tools plus the direction, standard-application and composition tools go; the tool catalog guard must stay green.

---

## Acceptance Criteria

- [ ] `EnterpriseCapability`, `EnterpriseStrategicImportance`, `Direction` and `StandardApplication` aggregates, and the composition service, no longer exist in the codebase
- [ ] No `/enterprise-capabilities`, `/enterprise-capability-compositions` or direction route is registered
- [ ] Strategic fit analysis is reachable from the main navigation and renders identically to before
- [ ] The `enterprise-architecture` frontend feature and the `architecture-direction` and `standard-application` features are deleted
- [ ] OnePagers offers five subject types; facts held for the retired type are archived
- [ ] The `enterprise-arch:*` permission group is gone and no route references it
- [ ] The agent tool catalog guard passes with the retired tools removed
- [ ] Read-model tables for the retired concepts are dropped in one migration; the event store is untouched
- [ ] Journeys, TIME assessments and realisation roles pass their existing test suites unchanged
- [ ] Context count and boundary smells re-measured against the coverage assessment

---

## Architecture

### Ownership

Architecture Direction (deletion) and OnePagers (subject type). Capability Mapping is unchanged behind its API.

### Domain Model

Delete the four aggregates and every value object exclusive to them (`EnterpriseCapabilityID`, `EnterpriseCapabilityName`, `Category`, `Importance`, `Rationale`, `SetAt`, `EnterpriseStrategicImportanceID`, direction status and horizon types). `TargetMaturity` does not die — spec 211 moves it onto the journey. Retire the published event constants and, in OnePagers, the handling that consumed them.

### API Surface

Remove `/enterprise-capabilities/**`, `/enterprise-capability-compositions`, `/capabilities/source-candidates` and the direction and standard-application routes. `/strategic-fit-analysis/{pillarId}` is unchanged.

### Persistence

One migration drops `enterprise_capabilities`, `enterprise_strategic_importance`, `directions`, `standard_applications`, `standard_application_history` and any cache exclusive to them. `infrastructure.events` is untouched.

### Frontend

Delete the `enterprise-architecture`, `architecture-direction` and `standard-application` features. `StrategicFitTab` becomes a page of its own with a navigation entry, keeping its hook, query keys and API call. Remove enterprise-capability query keys, schemas, API types and mutation effects.

### Cross-Context Integration

OnePagers stops subscribing to the retired Architecture Direction events — the only cross-context consumption of them. No new integration.

---

## Design Decisions

1. **One slice, not several** — the concept's removal is a single behaviour, and splitting it front-end-first or back-end-first leaves a half-dead model in between: routes without surfaces, or surfaces without routes. The slice is wide but singular.
2. **Delete rather than deprecate** — the branch is pre-release for this concept, there is no external API contract to honour, and leaving the aggregates behind a feature flag preserves exactly the ambiguity this move removes.
3. **Leave stored events in place** — event stores are append-only, replaying retired streams is nobody's job now that the deserializers are gone, and a history-rewriting migration is risk without benefit.
4. **Archive one-pager facts through the context's own path** — OnePagers already archives facts when a subject is deleted; reusing it keeps the deletion inside the owning context instead of a cross-schema SQL sweep, which spec 209 forbids.
5. **Strategic fit gets a page, not a home inside capabilities** — it is a cross-capability, per-pillar analysis; folding it into a capability drawer would lose the comparison that makes it useful.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Large deletion in one change set | Hard to review as a diff | Acceptance criteria are checkable independently; the architecture guard tests, tool catalog guard and existing journey suites pin what must still work |
| Recorded directions and standard applications are lost as a surface | History that architects may have relied on becomes unreadable | The events remain in the store; if a need appears, a read-side projection can resurrect them without the write model |
| One-pager facts for the retired subject type are archived | Curated content is retired with the concept | Archival, not deletion — the facts remain recoverable through the same path used for any deleted subject |
| Spec 169 is blocked | Its resolution vocabulary names both retired concepts | Design doc decision 7 records the re-scope; 169 is pending and unscheduled |

---

## Checklist

- [x] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
