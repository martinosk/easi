# 241 — Merge Applications

> **Status:** pending
> **Depends on:** 214 (ownership), 215 (hosting), 216 (composition), 234 (view membership integrity)
> **Roadmap alignment:** `SD6` (application record) — slice G of [`docs/specs/data-quality.md`](../docs/specs/data-quality.md) (decisions Q7, Q8). No new bounded context; one new published event in Architecture Modeling's language, consumed under standing invariant 1.

---

## Problem Statement

When the same application has been entered twice, the only remedy is to re-link by hand whatever the unwanted record carries and then delete it. In a reviewed production tenant two such pairs were merged that way: realisations were re-created on the surviving record, and the deleted record's view placements, relations and one-pager facts were simply lost. Anything the person merging forgets is lost silently, and the steps are not atomic.

This slice adds a merge: one application absorbs another. Everything the absorbed application had moves to the survivor in every context, the absorbed application is deleted, and the whole operation succeeds or fails as one.

---

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Survivor** | The application that remains. |
| **Absorbed** | The application that is merged into the survivor and then deleted. |
| **Merge** | Moving the absorbed application's facts to the survivor, then deleting the absorbed application. |
| **Survivor wins** | Where both applications hold a value for the same thing, the survivor's is kept and the absorbed one's is dropped. |

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect** | Resolve a duplicate without losing what either record knew |
| **View owner** | Their view shows the survivor where it showed the absorbed application |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Merge applications

  Scenario: Realisations move to the survivor
    Given "Sevron" realises "Terminals Health & Safety"
    And "Sevron (old)" realises "Chemical Safety"
    When an architect merges "Sevron (old)" into "Sevron"
    Then "Sevron" realises both capabilities
    And "Sevron (old)" no longer exists

  Scenario: Survivor wins on a shared capability
    Given both applications realise "Chemical Safety"
    And the survivor's realisation is graded Invest and the absorbed one's Tolerate
    When they are merged
    Then "Sevron" realises "Chemical Safety" once, graded Invest

  Scenario: An assessment moves with its realisation
    Given only the absorbed application realises "Chemical Safety", graded Tolerate with the role "Standard"
    When they are merged
    Then the survivor realises "Chemical Safety" graded Tolerate with the role "Standard"

  Scenario: Relations are re-pointed
    Given "Sevron (old)" serves "GTMS" and is triggered by "Portal"
    When it is merged into "Sevron"
    Then "Sevron" serves "GTMS" and is triggered by "Portal"

  Scenario: Relations that would duplicate or loop are dropped
    Given both applications serve "GTMS"
    And "Sevron (old)" triggers "Sevron"
    When they are merged
    Then "Sevron" serves "GTMS" once
    And "Sevron" has no relation to itself

  Scenario: Empty fields are filled from the absorbed application
    Given the survivor has hosting unknown, no owner and no description
    And the absorbed application is "saas", owned by "Ole Berg" and described
    When they are merged
    Then the survivor is "saas", owned by "Ole Berg" and carries that description

  Scenario: Set fields are kept
    Given the survivor is "cloud" and the absorbed application is "saas"
    When they are merged
    Then the survivor is "cloud"

  Scenario: Experts and origins are combined
    Given the absorbed application has an expert and a purchased-from vendor the survivor lacks
    When they are merged
    Then the survivor has both

  Scenario: One-pager facts move where the survivor has none
    Given the absorbed application's one-pager records Criticality "High" and the survivor's records nothing
    When they are merged
    Then the survivor's one-pager records Criticality "High"

  Scenario: Views show the survivor in place of the absorbed application
    Given a view contains "Sevron (old)" at a position and not "Sevron"
    When they are merged
    Then the view contains "Sevron" at that position
    And the view's history records the replacement

  Scenario: A view that held both keeps the survivor only
    Given a view contains both applications
    When they are merged
    Then the view contains "Sevron" at its own position and not "Sevron (old)"

  Scenario: Another user's private view is updated
    Given another user's private view contains "Sevron (old)"
    When an architect who cannot edit that view merges the applications
    Then that view contains "Sevron"

  Scenario: Parts move to the survivor
    Given "Sevron (old)" has the part "Sevron API"
    When it is merged into "Sevron"
    Then "Sevron API" is a part of "Sevron"

  Scenario: A merge that would break composition is refused
    Given "Sevron (old)" has parts and "Sevron" is itself a part of "Safety Suite"
    When an architect merges "Sevron (old)" into "Sevron"
    Then the merge is refused because parts cannot have parts
    And both applications are unchanged

  Scenario: Journeys follow the survivor
    Given a consolidation journey names "Sevron (old)" as a source
    When it is merged into "Sevron"
    Then the journey names "Sevron" as a source

  Scenario: Edit grants on the absorbed application end
    Given "Eva Dahl" holds an edit grant on "Sevron (old)"
    When it is merged into "Sevron"
    Then "Eva Dahl" holds no grant on "Sevron"

  Scenario: The dialog shows what will happen before merging
    When an architect opens "Merge" for "Sevron (old)" and picks "Sevron" as survivor
    Then the dialog lists what moves, what is dropped because the survivor wins, and what the survivor's fields will be
    And offers to swap which application survives

  Scenario: Merging is offered only to those who may do it
    Given the caller may edit applications but not delete them
    Then no application offers "Merge"

  Scenario: An application cannot absorb itself or a deleted application
    When a merge names the same application twice, or an application that does not exist
    Then it is refused and nothing changes

  Scenario: All or nothing
    Given moving the absorbed application's facts fails in any context
    Then neither application is changed and the absorbed application still exists

  Scenario: The merge is in the audit trail
    When "Sevron (old)" is merged into "Sevron"
    Then the history of "Sevron" records that it absorbed "Sevron (old)", by whom and when
```

---

## Business Rules & Invariants

1. **One command, one outcome** — a merge moves every fact in rules 3–9 and deletes the absorbed application, in one transaction. Any failure rolls all of it back.
2. **Preconditions** — both applications exist in the tenant and differ; the caller may edit and delete applications; the result respects the two-level composition cap (rule 5). Otherwise the merge is refused with nothing changed.
3. **Application record** — the survivor keeps its name. Description, hosting (`unknown` counts as empty) and ownership (state `unknown` counts as empty) are taken from the absorbed application only where the survivor's is empty. Experts are combined without duplicates.
4. **Relations and origins** — every relation of the absorbed application is re-pointed to the survivor; a relation that would equal an existing one (same type, direction and other end) or connect the survivor to itself is dropped. Each origin link kind (acquired-via, purchased-from, built-by) moves where the survivor has none of that kind.
5. **Composition** — the absorbed application's parts become parts of the survivor with their kind. If the absorbed application is a part and the survivor is not, the survivor takes its place under that parent. If both are parts, the survivor keeps its parent. If the outcome would give a part its own parts, or make an application part of itself, the merge is refused.
6. **Realisations** — each direct realisation of the absorbed application moves to the survivor with its level, notes, TIME assessment and realisation role. Where the survivor already realises that capability, the survivor's realisation, assessment and role are kept and the absorbed one's are dropped. Inherited realisations are recomputed, not moved. Strategic fit scores move per pillar where the survivor has none.
7. **Journeys** — every journey reference to the absorbed application (source, target, milestone subject) is re-pointed to the survivor; a reference that would duplicate one the journey already holds is dropped.
8. **One-pager facts** — each recorded fact of the absorbed application moves to the survivor where the survivor has no fact for that field.
9. **Views** — in every non-deleted view, private or public, whoever owns it: where only the absorbed application is contained, the survivor is added at the absorbed application's position and colour and the absorbed application removed, both through the view's history; where both are contained, the absorbed application is removed. View ownership and edit permission are not consulted (spec 234 rule 4).
10. **Edit grants do not move** — grants on the absorbed application end with it, as on deletion.
11. **The absorbed application is then deleted** by the ordinary deletion, after every context has moved its facts, so deletion cleanup finds nothing left to remove.
12. **Recorded** — the survivor's history holds an `ApplicationMerged` event naming the absorbed application's id and name and the actor.
13. **Not reversible** — there is no unmerge.

---

## Acceptance Criteria

- [ ] `POST /api/v1/components/{id}/merges` with `{ "absorbedId" }` performs the merge on the survivor `{id}` and answers 201 with the survivor; 404 when either application is missing, 409 when the composition rule refuses it, 400 when both ids are equal, 403 without both `components:write` and `components:delete`
- [ ] Every application the caller may edit and delete carries `x-merge` (POST) in its links; others do not
- [ ] Rules 3–9 each have tests for moved, survivor-wins and dropped-duplicate cases, including a self-relation, a shared capability with different grades, a fact present on both, a view holding both, and each composition case of rule 5
- [ ] `ApplicationMerged` (survivor id, absorbed id, absorbed name) is published in Architecture Modeling's language and every consuming context moves its facts in the publishing transaction before `ApplicationComponentDeleted` for the absorbed application is published
- [ ] Each consuming context applies its move through its own aggregates or primary records, so its own history shows the change (a realisation linked to the survivor, `ComponentAddedToView` and `ComponentRemovedFromView` on each view, a fact recorded on the survivor)
- [ ] A forced failure in any one consuming context leaves both applications and all their facts unchanged
- [ ] After a merge, no context holds a reference to the absorbed application's id other than in history
- [ ] Views of other users, including private ones, are updated; a soft-deleted view is skipped
- [ ] Stewardship's caches and Audit reflect the merge through the events already subscribed
- [ ] The merge dialog shows both applications side by side, the planned outcome per rule (moved, kept, dropped), a swap control, and requires typing nothing but an explicit confirm
- [ ] The dialog is reachable from the application details menu and from a duplicate-name attention item (spec 240), preselecting the other application
- [ ] Mutation effects invalidate components, realisations, relations, origins, views, one-pagers, journeys, home and attention queries
- [ ] Tenant isolation: an absorbed id belonging to another tenant answers 404
- [ ] Every BDD scenario has at least one corresponding test
- [ ] E2E: merge two applications with different realisations and a view placement and find all of it on the survivor
- [ ] `docs/architecture/ArchitectureModeling.md` and `docs/backend/cross-context-events.md` describe `ApplicationMerged` and its consumers
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

Architecture Modeling owns the command, the application-record rules (3–5) and the published event. Capability Mapping, Architecture Direction, OnePagers and Architecture Views each own moving their own facts in reaction. Access Delegation, Stewardship and Audit need no merge-specific behaviour: the existing deletion and creation events cover them.

### Domain Model

- **Architecture Modeling** — `MergeApplication` command handled over the two `ApplicationComponent` aggregates, the relations, origin links and the `ComponentContainments` aggregate. New event `ApplicationMerged` on the survivor, published. The absorbed application is deleted by the existing delete behaviour as the last step.
- **Capability Mapping** — on `ApplicationMerged`: re-link direct realisations and fit scores per rule 6.
- **Architecture Direction** — on `ApplicationMerged`: move TIME assessments and realisation roles with their realisations, re-point journey references per rule 7.
- **OnePagers** — on `ApplicationMerged`: move facts per rule 8.
- **Architecture Views** — on `ApplicationMerged`: per view, add the survivor and remove the absorbed application through the `ArchitectureView` aggregate, carrying position and colour per rule 9.

Ordering between consumers matters in one place: Architecture Direction's assessments follow Capability Mapping's realisations. Both react to the same event; Architecture Direction moves an assessment for a capability exactly when the survivor did not already hold an assessment for it, which is the same condition Capability Mapping applies to the realisation, so neither depends on the other having run.

### API Surface

One new operation, `POST /api/v1/components/{id}/merges`, and the `x-merge` affordance. Shapes, status codes and Swagger per `easi-api-standards`. No preview endpoint: the dialog composes its preview from existing reads of both applications.

### Persistence

No new tables. The event store gains `ApplicationMerged` events. No migration.

### Frontend

`MergeApplicationDialog` in the `components` feature: survivor picker (searching applications, with similar names first via spec 237 when available), side-by-side comparison, outcome list, swap, confirm. Entry points in the application details menu and the attention list.

### Cross-Context Integration

| Direction | Event | Purpose |
|-----------|-------|---------|
| Architecture Modeling → Capability Mapping | `ApplicationMerged` | Move realisations and fit scores |
| Architecture Modeling → Architecture Direction | `ApplicationMerged` | Move assessments, roles, journey references |
| Architecture Modeling → OnePagers | `ApplicationMerged` | Move facts |
| Architecture Modeling → Architecture Views | `ApplicationMerged` | Replace placements |

All in the publishing transaction (in-transaction projection); consumers stay database-only.

---

## Design Decisions

1. **A published event, not a client-side sequence** — the facts live in six contexts and only each context knows all of its own. A client that re-links and deletes loses whatever it does not know about and cannot be atomic. Alternative: Architecture Modeling calling each context's commands (rejected: it would have to know every other context's model, which is the coupling spec 209 removed).
2. **Survivor wins, everywhere** — one rule the user can hold in their head; the dialog's swap control is how they choose the other outcome. Alternative: per-field choice (rejected: a merge form with dozens of decisions, for a case where the two records are meant to be the same thing).
3. **Empty fields are filled from the absorbed application** — the point of merging is to lose nothing; `unknown` hosting and ownership are absences, not values.
4. **Edit grants do not move** — a grant is permission given for one record; extending it to another record by side effect is an access-control change nobody approved.
5. **Views are updated regardless of owner** — the same reasoning as deletion cleanup in spec 234: the alternative is views pointing at an application that no longer exists.
6. **The absorbed application is really deleted** — no redirect or alias from the old id. Links to it answer 404 as for any deleted application. Alternative: keep a tombstone that redirects (rejected: every read path in every context would have to resolve aliases).
7. **No capability merge here** — a capability carries a subtree, domain assignments, dependencies, importance ratings and journeys; it is a different operation with different rules, and the duplicates found so far were empty copies resolved by deletion (design doc Q8).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| All consumers in one transaction | A large merge holds a long transaction, and a bug in any consumer blocks merging | Merges are rare and per pair; the failure is loud, not silent |
| Irreversible | A wrong merge cannot be undone | The dialog shows the full outcome first; every moved fact is in the survivor's and the contexts' histories |
| Survivor wins without per-field choice | The absorbed record's differing values are dropped | They are listed in the dialog before confirming; swapping survivor keeps the other set |
| Frontend-composed preview | The preview can miss facts the client cannot read | The merge itself is server-side and complete; the preview states which sections the caller may not see |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
