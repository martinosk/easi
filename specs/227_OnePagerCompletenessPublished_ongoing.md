# 227 — One-Pager Completeness Published

> **Status:** pending
> **Depends on:** 208 (completeness served by OnePagers from the subject index), 209 (events-only integration)
> **Roadmap alignment:** `SD8 / H2-6` — slice B of [`docs/specs/personal-home.md`](../docs/specs/personal-home.md) (decision D6)

---

## Problem Statement

A one-pager's completeness (every required field filled) is known only inside OnePagers. It is computed in the `one_pager_subject_index` projector and served as the per-row One-Pager Quality list and the per-type completeness map, and nothing leaves the context: OnePagers has no published language, and a guard test (`TestOnePagersExposesNoPublishedLanguage`) and the context canvas both say so.

The stewardship read side needs completeness to raise *documentation* attention items (slice D), and spec 209 forbids it reading OnePagers' tables at runtime. Recomputing completeness from field-value events in the consumer was rejected in the design doc (D6): the rule — active required custom fields plus required built-ins, including relation fields — is OnePagers' own, and duplicating it recreates the drift spec 208 removed.

This slice gives OnePagers a published language with one event, raised by the subject-index projector whenever a subject's completeness changes. It has no consumer in this slice; slice D subscribes to it.

---

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Completeness** | A subject's standing in the subject index: `complete` (filled ≥ required, required > 0), `incomplete` (filled < required), `not-applicable` (nothing is required for the subject type). The existing signal of spec 208. |
| **Missing count** | `max(required − filled, 0)`, as the index stores it. |
| **Completeness change** | A write to a subject's index row after which its completeness or missing count differs from before, including the first computation of a newly created subject. |
| **Recalculated event** | An event published by a projection rather than appended by an aggregate; never stored, never replayed (spec 135 naming convention). |

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Consuming context** (stewardship, slice D) | Know, per subject, whether its one-pager is complete and how many required fields are missing, without reading OnePagers' tables. |
| **Architect** | Nothing visible changes; the One-Pager Quality list and completeness map behave as before. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: One-pager completeness is published

  Scenario: Filling the last missing field publishes the change
    Given the application "CRM" requires 3 fields and has 2 filled
    When the third required field is recorded
    Then SubjectCompletenessRecalculated is published for application "CRM" with completeness "complete" and missing count 0

  Scenario: Clearing a field publishes the change
    Given the capability "Billing" is complete with 4 required fields
    When one required field is cleared
    Then SubjectCompletenessRecalculated is published with completeness "incomplete" and missing count 1

  Scenario: A change in missing count alone is published
    Given the application "CRM" is incomplete with missing count 3
    When one required field is recorded
    Then SubjectCompletenessRecalculated is published with completeness "incomplete" and missing count 2

  Scenario: An edit that changes nothing publishes nothing
    Given the application "CRM" is incomplete with missing count 2
    When an optional field is recorded
    Then no SubjectCompletenessRecalculated is published

  Scenario: A new subject publishes its first completeness
    Given applications require 3 fields
    When the application "Portal" is created
    Then SubjectCompletenessRecalculated is published for "Portal" with completeness "incomplete" and missing count 3

  Scenario: A new subject of a type with no requirements
    Given nothing is required for vendors
    When the vendor "Acme" is created
    Then SubjectCompletenessRecalculated is published for "Acme" with completeness "not-applicable" and missing count 0

  Scenario: Making a field required publishes for every affected subject
    Given 10 applications, 4 of which have "Business criticality" filled
    When "Business criticality" becomes required for applications
    Then SubjectCompletenessRecalculated is published once for each of the 6 applications without it
    And nothing is published for the 4 that have it filled, whose completeness and missing count are unchanged

  Scenario: A relation change that fills a required built-in publishes
    Given capabilities require the built-in "Business domains" and "Billing" has none
    When "Billing" is assigned to a business domain
    Then SubjectCompletenessRecalculated is published for "Billing"

  Scenario: Deleting a subject publishes nothing
    When the application "CRM" is deleted
    Then no SubjectCompletenessRecalculated is published for "CRM"

  Scenario: Existing surfaces are unchanged
    Given a user who may read one-pagers
    When they open One-Pager Quality or a subject type's completeness map
    Then the rows, counts and ordering are as before this change
```

---

## Business Rules & Invariants

1. **One published event** — OnePagers' published language holds exactly one event constant, `SubjectCompletenessRecalculated`. Its payload: `subjectType` (one of the five subject types of `SubjectType`), `subjectId`, `completeness` (`complete` | `incomplete` | `not-applicable`), `requiredCount`, `missingCount`, `recalculatedAt`.
2. **Published on change only** — the event is published for a subject exactly when a subject-index write leaves its completeness or missing count different from the row's previous values. A write that changes neither publishes nothing. The first computation of a new subject is a change (from no row).
3. **One event per changed subject** — a bulk recompute (configuration change, relation change touching several subjects) publishes one event per subject whose standing changed, in one publish call.
4. **Deletion publishes nothing** — consumers learn of a subject's end from the supplier's own deletion event, which they already subscribe to.
5. **Completeness is not redefined** — the event carries the values the index stores; the bucket and missing-count definitions of spec 208 are unchanged.
6. **Not stored, not replayable** — the event is published to the bus by the projector and is not appended to the event store (spec 135 convention, `EffectiveImportanceRecalculated` precedent). A consumer seeds its cache with a backfill migration from `onepagers.one_pager_subject_index`.
7. **Suppliers of OnePagers never consume it** — Architecture Modeling, Capability Mapping, MetaModel and Auth, whose published languages OnePagers imports, must not subscribe; the dependency graph would cycle (enforced by `TestContextDependencyGraphIsAcyclic`).
8. **Same transaction** — the event is published inside the transaction of the write that caused it, like every bus delivery; a rolled-back write publishes nothing a consumer keeps.

---

## Acceptance Criteria

- [ ] `onepagers/publishedlanguage` exists with the single constant `SubjectCompletenessRecalculated`; `TestOnePagersExposesNoPublishedLanguage` is replaced by a guard asserting the package declares only that constant
- [ ] Each subject-index write path (create, subject update, fact recorded/cleared/archived, configuration change, relation recompute) publishes per rules 2–3, with a projector test per path covering the change and the no-change case
- [ ] A new subject publishes its first completeness, including `not-applicable` for a type with no requirements
- [ ] Deleting a subject publishes nothing
- [ ] The event payload carries every field of rule 1; a payload test pins the field names
- [ ] One-Pager Quality and the completeness map return identical results before and after (existing tests unchanged and passing)
- [ ] `docs/backend/cross-context-events.md` lists OnePagers' published event with a note that it is projection-published and needs a backfill; `docs/architecture/OnePagers.md` no longer says "Events published: none"
- [ ] Every BDD scenario has at least one corresponding test
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

OnePagers owns the event. No other context changes in this slice.

### Domain Model

- **`SubjectCompletenessRecalculated`** — an event struct in `onepagers/domain/events` in the `EffectiveImportanceRecalculated` shape (`BaseEvent` plus `EventData()`), aggregate id = subject id.
- **Change detection** — the subject-index store returns the previous (completeness, missing count) for every row it writes, from the same statement (`UPDATE … RETURNING` against the pre-update values, or a read-before-write in the transaction). The projector compares previous and new values and collects events for changed subjects. Today no store method returns previous state; this is the one structural change.
- The projector receives `events.EventBus` from `SetupRoutes` (already available as `deps.EventBus`) and publishes the collected events once per handled event.

### API Surface

None.

### Persistence

No schema change. The subject index is read as it is; the returned previous values come from the existing columns.

### Cross-Context Integration

| Direction | Event | Purpose |
|-----------|-------|---------|
| OnePagers → (any, except its suppliers) | `SubjectCompletenessRecalculated` | Completeness for the stewardship read side (slice D); no consumer in this slice |

---

## Design Decisions

1. **Named `…Recalculated`, not `…Changed`** — spec 135 reserves the suffix for projection-published events so a reader knows the event is derived and not replayable. The design doc's working name `SubjectCompletenessChanged` is replaced. Alternative: keep `…Changed` (rejected — breaks the one naming convention that marks non-replayable events).
2. **Projection-published, not an aggregate event** — completeness is a derived value over configuration, facts and relations; no aggregate owns it. Alternative: a `SubjectCompleteness` aggregate appending events (rejected — an aggregate over a read-model value, with its own consistency problems, to gain replay the backfill already gives).
3. **Carry completeness and counts, not the missing field names** — consumers rank and count; the names are computed only for a single one-pager view and a consumer that needs them links to the one-pager. Alternative: include missing field ids (rejected — larger events on every bulk recompute, and a consumer would learn OnePagers' field vocabulary).
4. **Publish only on change** — a configuration change recomputes every subject of a type; publishing unchanged rows would flood consumers with no-ops. Alternative: publish on every write (rejected — consumers would have to diff anyway).
5. **Replace the no-published-language guard rather than delete it** — the guard's intent (OnePagers leaks nothing of its internals) survives as "exactly this one event". Alternative: delete it (rejected — the next event added would pass without review).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Not replayable | A consumer that misses an event (deployed later, cache wiped) drifts | Consumers backfill from the subject index; the bus is synchronous and in-transaction, so a running consumer does not miss events |
| Read of previous values on every index write | One extra read per write, including bulk recomputes | `RETURNING` keeps it in one statement; bulk writes are already one statement over `unnest` |
| Bulk recompute publishes many events | A required-field change on a large type publishes one event per changed subject | One publish call; consumers apply each as a single-row upsert |

---

## Checklist

- [x] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
