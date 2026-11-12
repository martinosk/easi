# 237 — Similar-Entity Suggestions at Creation

> **Status:** pending
> **Depends on:** none
> **Roadmap alignment:** `SD6` (application record) and `SD5` in spirit (guidance at the edge, as spec 218) — slice C of [`docs/specs/data-quality.md`](../docs/specs/data-quality.md) (decisions Q1, Q2). No bounded-context change.

---

## Problem Statement

Nothing tells a user that what they are about to create already exists. A reviewed production tenant held two applications named "Sevron", two named "MVS", "Legislation Update Server" beside "Legislation Update Service", "Hazcheck" beside "HazCheck IMDG" and "Hazcheck Professional", and three pairs of identically named sibling capabilities. Each duplicate splits realisations, relations and view placements between two records, and merging them afterwards is manual.

This slice shows existing entities with a similar name while the user types a new one, in every create surface for applications, capabilities, vendors, acquired entities and internal teams. The user opens the existing entity or continues; nothing is blocked.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect** | Find out that an entity exists before creating it again, without searching first |
| **Architect creating a genuinely new entity** | Not be slowed down or blocked |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Similar-entity suggestions at creation

  Scenario: An exact match is shown before creating
    Given the application "Sevron" exists
    When an architect types "sevron" as the name of a new application
    Then the dialog lists "Sevron" as an existing application with the same name
    And the create button reads "Create anyway"

  Scenario: Near matches are shown
    Given the application "Legislation Update Service" exists
    When an architect types "Legislation Update Server"
    Then the dialog lists "Legislation Update Service" as similar
    And the create button still reads "Create"

  Scenario: Opening the existing entity
    Given a suggestion for "Sevron" is shown
    When the architect selects it
    Then the dialog closes without creating anything
    And the existing application's details open

  Scenario: Continuing creates the entity
    Given a suggestion for "Sevron" is shown
    When the architect selects "Create anyway"
    Then a second application named "sevron" is created

  Scenario: A distinct name shows nothing
    Given no application has a name similar to "Dwell Time Prediction"
    When an architect types it
    Then no suggestion is shown

  Scenario: Capability suggestions prefer siblings
    Given "Load Execution" exists under "Load and Discharge Execution" and "Load Planning" exists elsewhere
    When an architect adds a child named "Load Execution" under "Load and Discharge Execution"
    Then the sibling "Load Execution" is listed first as having the same name under the same parent
    And "Load Planning" is listed after it with its parent's name

  Scenario: Suggestions show enough to recognise the entity
    Then each application suggestion shows its name, description excerpt, hosting and owner
    And each capability suggestion shows its name, level and parent path
    And each origin-entity suggestion shows its name and kind

  Scenario: Renaming shows suggestions too
    Given the applications "CRM" and "Portal" exist
    When an architect renames "Portal" to "CRM"
    Then "CRM" is listed as an existing application with the same name
    And "Portal" itself is never suggested

  Scenario: Origin entities are compared across kinds
    Given the vendor "Kalmar" exists
    When an architect types "Kalmar" as a new internal team
    Then the vendor "Kalmar" is listed as similar

  Scenario: Suggestions never block
    Given the suggestion request fails or is slow
    When the architect submits the form
    Then the entity is created

  Scenario: Creating from the canvas shows suggestions
    When an architect creates an application from a canvas handle
    Then the same suggestions are shown as in the create dialog
```

---

## Business Rules & Invariants

1. **Advisory only** — suggestions are a read; create and rename commands are unchanged and never depend on them. No uniqueness invariant is added.
2. **Similarity** — a name is normalised by trimming it, case-folding it and collapsing internal whitespace to one space; this one definition lives in the shared kernel. A candidate is similar to the typed name when their normalised names are equal (`exact`), or one contains the other as whole words, or their trigram similarity is at least 0.45 (`similar`). Names shorter than three characters after normalisation yield no suggestions.
3. **Ranking and size** — exact matches first, then by similarity descending, then name. At most five suggestions.
4. **Capabilities** — candidates are every capability of the tenant; a candidate with the same parent as the one being created (or both L1) ranks before others of the same match kind and is marked `sameParent`.
5. **Origin entities** — vendors, acquired entities and internal teams are compared as one set.
6. **Self-exclusion** — on rename the entity being renamed is excluded.
7. **Permissions** — the suggestion query requires the read permission of the collection and returns only what the caller could list.
8. **Tenant-scoped** — suggestions never cross a tenant.
9. **Every create and rename surface** — the create dialogs, the canvas handle-create, the capability tree's add-child, and inline rename in the details panels all show suggestions.

---

## Acceptance Criteria

- [ ] `GET /api/v1/components?similarTo={name}`, `GET /api/v1/capabilities?similarTo={name}&parentId={id}` and one origin-entity equivalent return at most five suggestions per rules 2–6, each with id, name, match kind, score and the recognition fields of the scenario, and `_links.self`
- [ ] `excludeId` removes the entity being renamed; a name under three characters returns an empty list with 200
- [ ] Similarity has tests for exact (case, whitespace), containment ("Hazcheck" in "Hazcheck Professional"), trigram ("Server" / "Service"), an unrelated name, and ordering
- [ ] Capability suggestions rank and mark the same-parent candidate first
- [ ] The query uses an index and answers within 100 ms for a tenant with 5,000 names
- [ ] Every surface in rule 9 shows suggestions, debounced 300 ms, with the latest response winning
- [ ] An exact match changes the submit label to "Create anyway" (or "Rename anyway"); submitting is never disabled by suggestions
- [ ] A failed or pending suggestion request does not delay or block submit and shows no error
- [ ] Selecting a suggestion closes the dialog and opens the entity through the existing deep links
- [ ] Without the collection's read permission the query answers 403 and the dialog shows no suggestions
- [ ] Tenant isolation test: a same-named entity in another tenant is never suggested
- [ ] Every BDD scenario has at least one corresponding test
- [ ] E2E: typing an existing application's name shows it and opening it creates nothing
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

Architecture Modeling serves suggestions for applications and origin entities; Capability Mapping for capabilities. Each queries its own read model. No cross-context reads.

### Domain Model

No aggregate changes. Matching is done in each read-model query; only the name normalisation is shared, in the shared kernel.

### API Surface

A `similarTo` filter on the existing collections, per `easi-api-standards`. Responses are the collection's usual envelope with a reduced item shape plus `match` (`exact` or `similar`), `score` and, for capabilities, `sameParent` and `parentPath`.

### Persistence

Trigram indexes on the name columns of the application, capability and origin-entity read models. The migration that enables the `pg_trgm` extension is a schema migration of the infrastructure, not a backfill.

### Frontend

One shared `SimilarEntities` component and hook, used by `CreateComponentDialog`, `CreateCapabilityDialog`, the origin-entity create dialogs, the canvas handle-create controller and the inline-rename fields. Mantine primitives per `easi-frontend-styling`.

### Cross-Context Integration

None.

---

## Design Decisions

1. **Advice, not a uniqueness rule** — the same name is legitimate for capabilities under different parents, existing tenants already hold duplicates that an invariant would make uneditable, and imports and agents would need an override path. The precedent is spec 218. Exact duplicates that are created anyway are reported by spec 236. Alternative: 409 on exact duplicates (rejected).
2. **Served by the owning context as a collection filter** — it is a search over a collection, so it is a query parameter on that collection, not a new resource or a search context.
3. **Trigram similarity in the database** — typo- and suffix-tolerant, indexable, no AI call per keystroke. Alternative: the assistant's AI channel (rejected: latency and cost on every keystroke, and names do not need semantics to match). Alternative: client-side matching over the loaded list (rejected: lists are paginated and not all surfaces load them).
4. **Origin entities compared across kinds** — the same organisation entered as a vendor and as an acquired entity is the duplicate that matters.
5. **"Create anyway" only on exact matches** — a changed label is the lightest signal that still makes continuing a deliberate act; near matches are too often legitimate to interrupt.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Advisory | A determined or scripted client can still create duplicates | Spec 236 reports exact duplicates; spec 241 merges applications |
| Fixed similarity threshold | Some noise, some misses | Five results, exact first; the threshold is one constant with tests on real cases from the review |
| A request per pause in typing | Extra load on create | Debounced, indexed, five rows |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
