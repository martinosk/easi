# 195 — Sub-Capability Journeys in the Plan View

> **Status:** pending
> **Depends on:** 182 (journeys), drawer and deep-link machinery from 179/183
> **Design doc:** [`docs/specs/capability-journeys.md`](../docs/specs/capability-journeys.md)
> **Mockup:** [`195_SubCapabilityJourneys_mockup.html`](195_SubCapabilityJourneys_mockup.html)

---

## Problem Statement

A journey attaches to exactly one domain capability (182 decision 1), and capability-hierarchy relationships are derived, never stored (design doc D6). But the plan view never composes the two: an L1 journey and the journeys of its sub-capabilities are invisible to each other. Field feedback (2026-08): an architect planning "Ferry Booking → Control Tower" (L1, Q4 2028) alongside "Hazardous → Control Tower Hazardous" (L2, Q3 2026) had to hand-copy the Hazardous milestone onto the Ferry Booking journey to make the L1 plan tell the whole story — and the copies will now drift apart silently.

The sub-capability journey *is* the milestone the architect duplicated. This slice surfaces the relationship the hierarchy already implies: a capability's journey section shows the journeys running beneath it, and a sub-capability's journey section shows the ancestor journey it is part of. Pure read-side derivation — nothing new is stored.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain / Enterprise Architect** | See the whole change story of an L1 — its own journey plus every journey beneath it — without duplicating milestones by hand. |
| **Engineer / Product Manager / Stakeholder** | Open a sub-capability and understand which larger plan its journey belongs to. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Journeys across the capability hierarchy

  Scenario: A parent's plan shows its sub-capability journeys
    Given L1 capability "Ferry Booking" has an in-flight migration journey to "Control Tower" with target period Q4 2028
    And its L2 sub-capability "Hazardous" has a planned migration journey to "Control Tower Hazardous" with target period Q3 2026
    When a user opens the "Ferry Booking" drawer
    Then the journey section lists "Hazardous" under sub-capability journeys with its kind, status, target period, and progress
    And no milestone had to be added to the "Ferry Booking" journey to show it

  Scenario: Sub-capability journeys show even without a parent journey
    Given "Ferry Booking" has no journey of its own
    And "Hazardous" has an active journey
    When a user opens the "Ferry Booking" drawer
    Then the sub-capability journeys list shows the "Hazardous" journey

  Scenario: A completed sub-capability journey reads as progress
    Given the "Hazardous" journey was completed
    When a user opens the "Ferry Booking" drawer
    Then the "Hazardous" row shows status done

  Scenario: Abandoned sub-capability journeys are not shown
    Given the "Hazardous" journey was abandoned and no newer journey exists on "Hazardous"
    When a user opens the "Ferry Booking" drawer
    Then "Hazardous" does not appear in the sub-capability journeys list

  Scenario: A sub-capability's drawer names the ancestor journey it is part of
    Given the journeys from the first scenario
    When a user opens the "Hazardous" drawer
    Then the journey section shows that "Hazardous" sits under the "Ferry Booking" journey, with its status and target period

  Scenario: Rows navigate to the referenced capability
    When the user activates the "Hazardous" row in the "Ferry Booking" drawer
    Then the "Hazardous" drawer opens with its board card scrolled into view
    And activating the ancestor line in the "Hazardous" drawer opens the "Ferry Booking" drawer

  Scenario: No hierarchy journeys, no section
    Given a capability whose sub-capabilities have no journeys and whose ancestors have no active journey
    When a user opens its drawer
    Then no sub-capability journeys list and no ancestor line render

  Scenario: Read-only users see the composition
    Given a user without the architect permission
    When they open the "Ferry Booking" drawer
    Then the sub-capability journeys list renders identically, with navigation only
```

---

## Business Rules & Invariants

1. **Descendant scope** — the sub-capability journeys list covers all capabilities effectively below the drawer's capability, at any depth.
2. **Journey selection per descendant** — the descendant's current journey (active, or most recent terminal); shown when `planned`, `in-flight`, or `done`; an `abandoned` current journey is never shown.
3. **List ordering** — by target period ascending, undated journeys last, ties by capability name. Deterministic.
4. **Ancestor context** — every ancestor of the drawer's capability with an *active* journey is shown, nearest ancestor first.
5. **Pure derivation** — no stored association, no new events, no new aggregate state. Capturing, transitioning, or abandoning any journey is reflected on the next fetch with no further bookkeeping.
6. **Navigation only** — rows and ancestor lines carry no write affordances; all journey writes stay on the owning capability's own journey section.
7. **Permission** — visibility follows board read permission; the composition exposes nothing the caller could not already read.

---

## Acceptance Criteria

- [ ] The drawer journey section lists descendant journeys per rules 1–3, including when the capability has no journey of its own
- [ ] Done descendant journeys render with done status; abandoned ones are absent
- [ ] The drawer of a capability with an active ancestor journey shows the ancestor line per rule 4
- [ ] Activating a row or ancestor line opens that capability's drawer with its card scrolled into view
- [ ] The section is absent when rules 1–4 produce nothing
- [ ] No new backend state, events, or endpoints exist for this feature; removing a child journey removes its row on the next fetch
- [ ] Read-only users see the identical composition with no write affordances
- [ ] Every BDD scenario has at least one corresponding test
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

No backend change. The frontend composes two things it already loads for the board: the bulk journey query (182) and the capability hierarchy in the board view model. Consistent with design doc D6 — hierarchy relationships are derived at read time, never stored.

### Domain Model

None.

### API Surface

None new. The existing bulk journey query by capability set serves the composition.

### Persistence

None.

### Frontend

The drawer's journey section gains a sub-capability journeys list and an ancestor journey line, both derived in the board/journeys view-model layer. Row activation reuses the board's existing deep-link/scroll/drawer machinery. Queries and invalidation follow `easi-frontend-data`; existing journey mutation effects already invalidate the bulk query this composition reads.

### Cross-Context Integration

None.

---

## Design Decisions

1. **Derive the relationship; never share milestones** — the duplicated milestone in the field feedback is a symptom of the child journey being invisible from the parent plan. A shared milestone entity across journeys was rejected: it breaks single-aggregate ownership (182), leaves edit/freeze semantics ambiguous once one journey is terminal, and duplicates what the child journey already records. A milestone that *references* a child journey was rejected: manual bookkeeping for a relationship the hierarchy already implies.
2. **Frontend composition, no rollup endpoint** — the board already fetches journeys in bulk and holds the hierarchy; a dedicated backend rollup endpoint was rejected as premature infrastructure for a pure presentation concern.
3. **Done shown, abandoned hidden** — a completed child journey is progress toward the parent story; an abandoned one is not part of it and remains reachable through the child's own journey history.
4. **Both directions in one slice** — the downward list and the upward ancestor line are the same derived relationship read from either end; shipping one without the other would leave half the mental model ("an L2 plan inherently affects the L1") invisible.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Milestones that mirror child journeys are not blocked | An architect can still duplicate by hand | The derived list makes duplication pointless; existing copies can simply be removed |
| Descendants at any depth | A deep hierarchy could yield a long list | Rows are compact, deterministically ordered, and journeys are sparse in practice |
| Frontend derivation | Each consumer surface must compose consistently | Derivation lives once in the view-model layer both drawer directions share |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
