# 213 — Retire the Enterprise Capability

> **Status:** done
> **Depends on:** [211_MaturityJourneys](211_MaturityJourneys_done.md), [212_UnifiedTimeVocabulary](212_UnifiedTimeVocabulary_done.md) — design: [docs/specs/enterprise-capability.md](../docs/specs/enterprise-capability.md)
> **Roadmap alignment:** SD1 / H1-1

---

## Problem Statement

The enterprise capability asks users to curate a taxonomy nobody browses. Since spec 172 the aggregate has been a shell whose only association with real capabilities runs through a Direction; the strategic-importance rating has never had a UI. With maturity now expressed as a journey (211) and the TIME suggestion sitting beside the assessment it advises (212), nothing of value depends on the concept any more.

Direction, Standard Application and Composition go with it. Each is a statement *about* an enterprise capability — a Direction's subject is one, a Standard Application is set per one, and Composition exists only to assemble one from source capabilities — and `DirectionPanel` and `StandardApplicationPanel` render nowhere except inside the enterprise capability detail panel. Removing the subject leaves them without a subject and without a surface.

Strategic fit analysis is the one thing on the enterprise architecture page worth keeping. It is already owned by Capability Mapping, scores domain capabilities against strategy pillars, and never referenced an enterprise capability. It needs a page of its own, nothing more.

Maturity analysis goes with the concept. It answered "which enterprise capabilities sit below their target maturity" — a question about a target that spec 211 moved onto the journey, where the gap is stated per capability at the moment the ambition is recorded. There is no enterprise capability left to roll the gap up to.

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
    And the Domain Board drawer still records grades and roles against them

  Scenario: Maturity gaps are read on the journey, not on a catalog entry
    Given a capability with a maturity journey
    When an architect looks for its maturity gap
    Then the journey states it
    And no enterprise-capability maturity analysis or maturity gap surface remains

  Scenario: The assistant loses the retired tools
    When an assistant user asks about enterprise capabilities
    Then no enterprise-capability, direction, standard-application, composition or maturity-analysis tool is available
```

---

## Business Rules & Invariants

1. **Four aggregates retire** — `EnterpriseCapability`, `EnterpriseStrategicImportance`, `Direction` and `StandardApplication`, with their commands, events, repositories, read models and tables. Composition is a service over Direction and retires with it, and maturity analysis is a read over composition and retires with that.
2. **Architecture Direction keeps only planning** — `CapabilityJourney`, `TimeAssessment` and `RealizationRole` survive, all keyed on a domain capability or a capability-application pair. The context speaks one language.
3. **Stored events are not rewritten** — retired event types stay in `infrastructure.events`, unread. No migration touches history; only read-model tables are dropped.
4. **OnePagers loses one subject type** — `enterprise-capability` retires from the subject-type value object, the relation catalog, the built-in field catalog, the subject index and the subject-deleted reactor.
5. **The archival sweep runs before the reactor mapping goes** — the subject-deleted reactor archives on `EnterpriseCapabilityDeleted`, an event retirement never emits, so a one-off archival dispatched inside OnePagers at deploy walks the subject index for the retired type and issues `ArchiveOnePagerFacts` for each. The reactor's mapping for the type is removed only after that sweep exists. No SQL touches the facts.
6. **Strategic fit analysis is untouched behind the API** — `GET /strategic-fit-analysis/{pillarId}` and its Capability Mapping ownership do not change; only the UI that calls it moves and the permission that guards it.
7. **Permissions retire everywhere, not just on routes** — the `enterprise-arch:*` group leaves the published language, the role definitions and every consumer: `/strategic-fit-analysis/{pillarId}` re-points at `capabilities:read`, and the assistant's write-link check, the session's one-pager-quality link check and the one-pager quality permission table drop the permission with the subject type it stood for. Roles already stored keep working: the permission is dropped from the role definitions, not migrated out of persisted rows.
8. **Agent tools retire with their routes** — the seven enterprise-capability tools plus the direction, standard-application, composition and maturity-analysis tools go: fourteen of Architecture Direction's twenty-two, leaving the eight that serve journeys, TIME assessments and realisation roles. The tool catalog guard must stay green.

---

## Acceptance Criteria

- [x] `EnterpriseCapability`, `EnterpriseStrategicImportance`, `Direction` and `StandardApplication` aggregates, the composition service and the maturity analysis read model no longer exist in the codebase
- [x] No `/enterprise-capabilities`, `/enterprise-capability-compositions` or direction route is registered
- [x] Strategic fit analysis is reachable from the main navigation, guarded by `capabilities:read`, and renders identically to before
- [x] The `enterprise-architecture` and `standard-application` frontend features are deleted; `architecture-direction` keeps its TIME assessment and realisation role surface and loses only its direction slice
- [x] OnePagers offers five subject types; the archival sweep has run and no unarchived facts remain for the retired type
- [x] The `enterprise-arch:*` permission group is gone and nothing — route, link check or quality table — references it
- [x] The agent tool catalog guard passes with fourteen tools removed and eight left in Architecture Direction
- [x] Read-model tables for the retired concepts are dropped in one migration; the event store is untouched
- [x] Journeys, TIME assessments and realisation roles pass their existing test suites unchanged, and the Domain Board drawer still assesses and roles a realisation end to end
- [x] The coverage assessment is re-scored: capability C3 drops from full to uncovered, C6 loses its enterprise-architecture rows, the context count stays at 14, and no boundary smell closes — this slice removes surface, not a smell

---

## Architecture

### Ownership

Architecture Direction (deletion) and OnePagers (subject type). Capability Mapping is unchanged behind its API.

### Domain Model

Delete the four aggregates and every value object exclusive to them (`EnterpriseCapabilityID`, `EnterpriseCapabilityName`, `Category`, `Importance`, `Rationale`, `SetAt`, `EnterpriseStrategicImportanceID`, direction status and horizon types), plus the composition service and the maturity analysis read model that reads over it. `TargetMaturity` does not die — spec 211 moves it onto the journey, and only the enterprise capability's copy goes. Retire the published event constants and, in OnePagers, the handling that consumed them.

### API Surface

Remove `/enterprise-capabilities/**` (including `maturity-analysis` and `{id}/maturity-gap`), `/enterprise-capability-compositions`, `/capabilities/source-candidates` and the direction and standard-application routes. `/strategic-fit-analysis/{pillarId}` keeps its handler and its Capability Mapping ownership; only its guard changes, from `enterprise-arch:read` to `capabilities:read`.

### Persistence

Migration 154 drops `enterprise_capabilities`, `enterprise_strategic_importance`, `directions`, `direction_source_capabilities`, `standard_applications`, `standard_application_history` and `capability_domain_cache`, the last of which only `DirectionReadModel` reads. Migration 155 deletes the retired type's row from `onepagers.one_pager_configurations`: each `OnePagerFactsArchived` event recomputes completeness in the write transaction, and with the row gone the retired type resolves to zero required fields instead of consulting a catalog that no longer exists. Two migrations, not one, because a non-backfill migration may touch only one schema (spec 209 guard).

The `one_pager_subject_index` rows for the retired type stay: they are the archival sweep's input, and they are unreachable behind the API — the quality list filters to readable subject types and the retired type no longer maps to any permission.

The caches that stay, because the surviving features read them: `capability_node_cache`, `realization_cache` and `reference_name_cache` back journeys, assessments and roles; `ea_importance_cache`, `ea_fit_score_cache` and `ea_strategy_pillar_cache` back the TIME suggestion that spec 212 composed into assessment reads. Dropping any of the six breaks a surviving feature. `infrastructure.events` is untouched.

### Frontend

Delete the `enterprise-architecture` and `standard-application` features. **Prune, do not delete, `architecture-direction`**: it is where TIME assessments and realisation roles live, and twenty files across `business-domains` and the test mocks import from it. Remove `api/directionApi.ts`, the direction components (`DirectionPanel`, `CaptureDirectionForm`, `EditDraftDirectionForm`, `DirectionStatusBadge`, `sourcePickerPrimitives`), `hooks/useDirection*.ts`, and the direction slice of `types.ts`, `queryKeys.ts` and `mutationEffects.ts`. Everything TIME and role keeps working untouched.

`StrategicFitTab` becomes a page of its own with a navigation entry, keeping its hook, query keys and API call. Remove the enterprise-architecture navigation entry and route path, and the enterprise-capability query keys, schemas, API types and mutation effects.

### Cross-Context Integration

OnePagers stops subscribing to the retired Architecture Direction events — the only cross-context consumption of them. Before those subscriptions go, OnePagers runs its own one-off archival over the retired subject type (rule 5): it walks its subject index for `enterprise-capability` and dispatches `ArchiveOnePagerFacts` per subject, exactly as the subject-deleted reactor would have. No new integration and no cross-schema SQL.

The sweep needs the tenant list, which OnePagers does not own: Auth publishes a `TenantDirectory` live read (`TenantIDs`) in its published language — auth owns tenancy — and the composition root injects it into `SetupOnePagersRoutes`, which runs the sweep synchronously at startup, per tenant, before the API serves. The sweep is idempotent: archiving deletes the facts rows from the read model, so later boots find nothing to archive.

---

## Design Decisions

1. **One slice, not several** — the concept's removal is a single behaviour, and splitting it front-end-first or back-end-first leaves a half-dead model in between: routes without surfaces, or surfaces without routes. The slice is wide but singular.
2. **Delete rather than deprecate** — the branch is pre-release for this concept, there is no external API contract to honour, and leaving the aggregates behind a feature flag preserves exactly the ambiguity this move removes.
3. **Leave stored events in place** — event stores are append-only, replaying retired streams is nobody's job now that the deserializers are gone, and a history-rewriting migration is risk without benefit.
4. **Archive one-pager facts through the context's own path** — OnePagers already archives facts when a subject is deleted; reusing it keeps the deletion inside the owning context instead of a cross-schema SQL sweep, which spec 209 forbids.
5. **Strategic fit gets a page, not a home inside capabilities** — it is a cross-capability, per-pillar analysis; folding it into a capability drawer would lose the comparison that makes it useful.
6. **Tenant enumeration is a published Auth read, not OnePagers SQL** — RLS hides other tenants' rows from the app user, so the sweep cannot scan its own tables cross-tenant. Auth already owns tenancy and live tenant reads; a one-method `TenantDirectory` in its published language keeps the boundary rule intact. Alternative — a BYPASSRLS scan — rejected: it would put tenancy knowledge and an RLS exemption inside OnePagers.
7. **The retired type's one-pager configuration row is deleted by migration, before the sweep** — projectors run in the write transaction, so a completeness recompute that errors on the retired type would roll the archival back. Deleting the configuration read-model row (the aggregate's events remain) makes the recompute resolve to zero required fields. Facts are still never touched by SQL.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Large deletion in one change set | Hard to review as a diff | Acceptance criteria are checkable independently; the architecture guard tests, tool catalog guard and existing journey suites pin what must still work |
| Recorded directions and standard applications are lost as a surface | History that architects may have relied on becomes unreadable | The events remain in the store; if a need appears, a read-side projection can resurrect them without the write model |
| Capability C3 — enterprise consolidation and standardisation analysis — drops from full to uncovered | Nothing will answer "the same capability is implemented in several domains; what do we consolidate, and which application is the standard?" | Deliberate: the assessment rated C3 full on machinery nobody used, and its own verdict left open whether the enterprise capability deserved a context at all. This slice answers that question. The question C3 poses stays worth answering; when it is asked again it gets a design informed by use, not the taxonomy being retired here |
| Maturity analysis disappears with composition | The catalog-wide "which capabilities are below target" list goes | Spec 211 put the target and the gap on the journey, where the ambition is recorded per capability; the roll-up had no subject left to roll up to |
| One-pager facts for the retired subject type are archived | Curated content is retired with the concept | Archival, not deletion — the facts remain recoverable through the same path used for any deleted subject |
| Spec 169 is blocked | Its resolution vocabulary names both retired concepts | Design doc decision 7 records the re-scope; 169 is pending and unscheduled |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [ ] User sign-off
