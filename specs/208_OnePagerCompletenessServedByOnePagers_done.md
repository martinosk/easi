# 208 — One-Pager Completeness Served by OnePagers

> **Status:** done
> **Depends on:** 178 (One-Pager Completeness), 189 (One-Pager Quality List)
> **Amends:** 178 design decision 4 (page-scoped decoration of subject lists)

---

## Problem Statement

Spec 178 put a per-row `onePagerComplete` indicator on the subject list responses of Architecture Modeling (applications, vendors, acquired entities, internal teams), Capability Mapping (capabilities) and Enterprise Architecture (enterprise capabilities). Each of those contexts obtains the value at request time from OnePagers' completeness query through a composition-root adapter. OnePagers, in turn, subscribes to those three contexts' events to maintain its subject index. Three runtime cycles result — AM ⇄ OP, CM ⇄ OP, EA ⇄ OP — and completeness, a OnePagers concept, is rendered by contexts that know nothing about one-pagers.

The dependency guard in spec 206 forbids these cycles. Completeness must be read from OnePagers.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | Still sees at a glance, in the navigation tree and the enterprise capabilities table, which subjects have an incomplete one-pager |
| **Tenant Administrator** | Indicator semantics from 178 rule 10 are unchanged: shown only for subject types with at least one required field |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: One-pager completeness read from OnePagers

  Scenario: Navigation tree shows incomplete indicators
    Given the application subject type has a required custom field
    And application A has recorded it and application B has not
    When the architect opens the navigation tree
    Then B shows the incomplete indicator and A does not

  Scenario: Enterprise capabilities table shows incomplete indicators
    Given the enterprise-capability subject type has a required field
    When the architect opens the enterprise capabilities table
    Then each enterprise capability without all required values shows the indicator

  Scenario: No indicator for a subject type without required fields
    Given the vendor subject type has no required field
    When the architect opens the navigation tree
    Then no vendor shows an indicator

  Scenario: Recording a required value clears the indicator
    Given application B shows the incomplete indicator
    When the architect records the missing required value on B's one-pager
    Then B's indicator disappears without a page reload

  Scenario: A newly created subject is indicated as incomplete
    Given the application subject type has a required field
    When the architect creates application C
    Then C shows the incomplete indicator

  Scenario: Subject list responses carry no completeness
    When any client lists applications, capabilities or enterprise capabilities
    Then the response rows contain no `onePagerComplete` field
```

---

## Business Rules & Invariants

1. **Completeness is a OnePagers resource** — `GET /one-pagers/{subjectType}/completeness` returns, for every subject of the type the caller may read, whether its one-pager is complete.
2. **Rule 10 of spec 178 is preserved** — when the subject type has no required field, the resource returns an empty collection and no indicator is rendered.
3. **Subject list DTOs of other contexts carry no completeness** — Architecture Modeling, Capability Mapping and Enterprise Architecture neither compute nor decorate it.
4. **Permission follows the subject type** — the same read permission per subject type that the quality list applies (178/189).
5. **Freshness follows OnePagers' own consistency** — recording or clearing a fact, changing the configuration, and subject creation or deletion update the indicator on the next read; the frontend invalidates the completeness query on those mutations.

---

## Acceptance Criteria

- [x] `GET /one-pagers/{subjectType}/completeness` returns `{ data: [{ subjectId, complete }], _links }` for all indexed subjects of the type; empty `data` when the type has no required field; 403 without the type's read permission; handler and query tests cover the three cases.
- [x] `onePagerComplete` and the `OnePagerCompleteness*` ports/sources are removed from Architecture Modeling, Capability Mapping and Enterprise Architecture handlers, routes, DTOs and Swagger; the composition-root completeness adapters are deleted.
- [x] Frontend: a `useOnePagerCompleteness(subjectType)` query; `OnePagerIncompleteIndicator` renders from it; navigation sections (applications, capabilities) and the enterprise capabilities table use it; `onePagerComplete` removed from API types and MSW stubs.
- [x] Mutation effects invalidate the completeness query on: record/clear fact, configuration changes, and create/delete of any subject type.
- [x] Spec 178's design decision 4 is annotated as amended by this spec.
- [x] `go test ./...`, spec 206 guard, `npm run build`, `npm test -- --run`, `npm run lint` green.

---

## Architecture

### Ownership

OnePagers owns completeness end to end. The three subject-owning contexts lose their completeness ports.

### Domain Model

No aggregate changes. The read uses OnePagers' subject index (which already stores required/filled counts per subject via `ApplyCompleteness`) and the existing completeness query.

### API Surface

New read resource on OnePagers, permission-gated per subject type. Removal of a field from three contexts' list DTOs.

### Persistence

None new.

### Frontend

One query hook in `features/one-pagers`, consumed by the indicator component; navigation sections and EC table pass the subject type; `mutationEffects.ts` of one-pagers, components, capabilities, enterprise-architecture and origin-entities return the completeness key where relevant.

### Cross-Context Integration

OnePagers ← AM/CM/EA events (unchanged). AM/CM/EA → OnePagers: **removed**.

---

## Design Decisions

1. **Per-subject-type collection instead of per-page decoration** — 178 D4 chose page-scoped decoration to avoid a second request; that choice is what created the cycles. One collection per subject type, cached client-side, serves every list surface. Alternative: an event-fed completeness cache in each subject-owning context (rejected: three caches with backfills for a purely presentational flag).
2. **Read from the subject index** — it already holds required/filled counts and is maintained by OnePagers' projectors; no new read model.
3. **Empty collection when nothing is required** — mirrors 178 rule 10 exactly; the frontend renders no indicator for ids absent from the collection.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| One more request per subject type on tree/table surfaces | Slight extra load on first render | One request per type, React Query cached, invalidated only on relevant mutations |
| Indicator no longer arrives with the row | Rows render before completeness resolves | Indicator is additive; rows are usable without it |

---

## Checklist

- [x] Specification ready — approved in session 2026-08-29 ("fix it in this change")
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off (2026-08-31)
