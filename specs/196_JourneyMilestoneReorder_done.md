# 196 — Journey Milestone Reorder

> **Status:** done
> **Depends on:** 182 (journeys)
> **Design doc:** [`docs/specs/capability-journeys.md`](../docs/specs/capability-journeys.md)
> **Mockup:** [`196_JourneyMilestoneReorder_mockup.html`](196_JourneyMilestoneReorder_mockup.html)

---

## Problem Statement

Spec 182 defines milestones as an ordered list (rule 8) but fixed the order to insertion order, explicitly deferring a reorder operation until architects asked for it (182 decision 10). They have asked (field feedback 2026-08): milestones should read as a sequenced plan, and plans are not written in the order their steps will execute. Today the only way to fix an order is to remove and re-add milestones, which destroys their status and event history.

This slice adds the deferred reorder operation: one discrete, evented way to set the milestone sequence of an active journey.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain / Enterprise Architect** | Arrange a journey's milestones into the sequence the plan will execute, without recreating them. |
| **Engineer / Product Manager / Stakeholder** | Read a milestone list top-to-bottom as the intended order of events. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Reordering journey milestones

  Scenario: An architect reorders milestones
    Given an active journey with milestones "Contract signed", "Rollout", "Pilot" in that order
    When the architect reorders them to "Contract signed", "Pilot", "Rollout"
    Then every view of the journey lists the milestones in the new order
    And each milestone keeps its label, target period, and status

  Scenario: Reordering is one discrete recorded event
    When the architect reorders a journey's milestones
    Then exactly one past-tense event records the resulting order with actor and timestamp
    And replaying the journey's events reconstructs the order

  Scenario: Terminal journeys are frozen
    Given a done or abandoned journey
    When the architect attempts to reorder its milestones
    Then the operation is rejected

  Scenario: The new order must account for every milestone
    Given an active journey with three milestones
    When a reorder omits one of them, repeats one, or names an unknown milestone
    Then the operation is rejected with a clear error

  Scenario: A no-change reorder is rejected
    When the architect submits the order the journey already has
    Then the operation is rejected and no event is recorded

  Scenario: Order is stable under add and remove
    Given a reordered journey
    When the architect adds a milestone
    Then it appears at the end of the list
    And when the architect removes a milestone, the remaining order is unchanged

  Scenario: Read-only users get no reorder affordance
    Given a user without the architect permission
    When they fetch a journey
    Then the milestone list is readable but carries no reorder affordance
```

---

## Business Rules & Invariants

1. **Reorder is a full permutation** — one operation submits the complete list of the journey's milestone ids in the intended order. Omissions, duplicates, and unknown ids are rejected.
2. **Active journeys only** — reorder follows 182 rules 6 and 8: terminal journeys are frozen entirely.
3. **One discrete event** — a reorder produces exactly one past-tense event carrying the resulting order, the acting user, and the occurred-at timestamp; replay reconstructs the order.
4. **No-op rejected** — an order identical to the current order is rejected; no event is recorded.
5. **Order is stable** — a newly added milestone appends to the end; removal compacts the remaining order without other changes (182 decision 10 posture, unchanged).
6. **Authorisation** — reorder requires the architect permission and is HATEOAS-gated, like every journey write (182 rule 14).

---

## Acceptance Criteria

- [x] An architect can reorder an active journey's milestones; the new order is returned by every read surface (single journey, history, bulk query)
- [x] Rules 1–4 violations are rejected with clear errors; a valid reorder produces exactly one event and replay reconstructs the order
- [x] Add-after-reorder appends; remove-after-reorder compacts without disturbing the rest
- [x] The reorder affordance is HATEOAS-gated; read-only users receive none; the drawer milestone list exposes reordering only when the link is present
- [x] Every BDD scenario has at least one corresponding test; every business rule has a unit test
- [x] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

`architecturedirection` — a new behaviour on the existing `CapabilityJourney` aggregate. No other context is touched.

### Domain Model

`CapabilityJourney` gains a reorder operation validating rules 1–2 and 4, producing one new past-tense event that carries the full resulting milestone order. Milestone identity, labels, periods, and statuses are untouched by reordering.

### API Surface

One write operation on the journey's milestone collection accepting the ordered id list, exposed as a HATEOAS link on the journey alongside the existing milestone affordances. Shape per `easi-api-standards`.

### Persistence

Event-sourced like all journey changes; read models persist and return the explicit order so every query surface lists milestones identically.

### Frontend

The drawer's milestone list gains a reorder affordance (shown only when the link is present) wired through the existing journey mutation hooks and `mutationEffects.ts` invalidation.

### Cross-Context Integration

None.

---

## Design Decisions

1. **Full-permutation event, not per-move events** — one reorder gesture is one intent; recording it as N move events was rejected as replay noise that still requires permutation validation at the end.
2. **Reorder is a distinct operation, not part of milestone update** — updating one milestone (182) and sequencing the list are different intents with different validation; overloading the update operation was rejected.
3. **No-op rejection over silent acceptance** — accepting an unchanged order would either record a meaningless event (violating the discrete-event posture) or silently do nothing while reporting success; an explicit rejection keeps the contract honest.

4. **Event is `JourneyMilestonesReordered`** (decided during implementation) — the mockup hint names it in the singular; the plural names what happened (the journey's milestones were reordered) and matches the full-permutation payload (`milestoneIds`). Replay validates the permutation against the current list and fails as a corrupted event otherwise.

5. **Operation is `PUT /capability-journeys/{journeyId}/milestone-order`, link rel `x-reorder-milestones`** (decided during implementation) — a sibling resource of the milestone collection, so it cannot collide with `PUT /milestones/{milestoneId}`. Permutation violations map to 400 (omission, duplicate) or 404 (unknown id); the no-op rejection maps to 409 as a business-rule conflict.

6. **Link only when there is something to reorder** (decided during implementation) — `x-reorder-milestones` is emitted for architects on active journeys with two or more milestones; a single-milestone list renders inert.

7. **Reorder affordance is native drag-and-drop plus keyboard** (decided during implementation) — a grip handle per row is draggable and focusable; ArrowUp/ArrowDown on the handle moves the milestone one step. A move that would leave the order unchanged is not submitted, so the client never triggers the rule-4 rejection on its own. No drag-and-drop library was added.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Full-list permutation | Concurrent milestone add/remove makes an in-flight reorder stale | Rule 1 validation rejects the stale list; the client refetches and reapplies |
| Sequence is authoritative | The arranged order can contradict the milestones' target periods | Spec 205 marks such rows on read; the timeline (197) still places by period |
| Explicit order in read models | Order must be maintained on add/remove projections | Append/compact semantics are unchanged from 182; only reorder rewrites positions |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
