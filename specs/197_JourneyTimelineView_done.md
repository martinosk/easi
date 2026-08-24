# 197 — Journey Timeline View

> **Status:** done
> **Depends on:** 182 (journeys), 183 (board toolbar and deep-link machinery). Pairs with 196 — the stored milestone order becomes architect-controlled — but does not require it.
> **Design doc:** [`docs/specs/capability-journeys.md`](../docs/specs/capability-journeys.md)
> **Mockup:** [`197_JourneyTimelineView_mockup.html`](197_JourneyTimelineView_mockup.html)

---

## Problem Statement

Journeys and milestones were built timeline-ready by design — structured sortable target periods, discrete evented history (design doc D8, 182 decision 6) — but the only rendering is a per-capability list inside the drawer. Whether the landscape's plans are on track is invisible until someone opens every drawer and compares quarters by hand. Field feedback (2026-08) asks for exactly the view D8 anticipated: plans and milestones on a time axis, expected dates visible, behind-schedule work flagged, milestones shown in order.

This slice adds a Timeline presentation to the Business Domains page: every active journey as a row on a quarter axis, milestones and target periods placed at their quarters, overdue work highlighted, with click-through to the capability.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | One view of schedule health across all plans — what lands when, what is behind. |
| **Domain Architect** | See whether the journeys in their domains are on track without opening each drawer. |
| **Engineer / Product Manager / Stakeholder** | Read when a capability's change is expected to land and what steps remain. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Journey timeline

  Scenario: Active journeys render on a quarter axis
    Given active journeys exist with target periods and milestones
    When a user switches the Business Domains page to the Timeline view
    Then each active journey renders as a row grouped under its business domain
    And each row shows the capability name, journey kind, status, and progress
    And each dated milestone renders as a labelled row placed at its target quarter, showing its label, status, and period
    And the journey's target period is marked at its quarter

  Scenario: A milestone behind schedule is flagged
    Given the current quarter is Q3 2026
    And a milestone has target period Q1 2026 and status in-flight
    Then that milestone is marked overdue
    And a done milestone with target period Q1 2026 is not marked overdue

  Scenario: A journey past its target period is flagged
    Given the current quarter is Q3 2026
    And an in-flight journey has target period Q2 2026
    Then that journey's row is marked overdue

  Scenario: Undated milestones are shown without a quarter
    Given a journey with milestones lacking a target period
    Then those milestones render as the journey's last rows, marked as having no date, in stored order
    And they are never marked overdue

  Scenario: The current quarter is marked
    Then the axis includes the current quarter with a visible today marker
    And the axis spans from the earliest to the latest period among the shown journeys and milestones

  Scenario: Schedule health is summarised
    Given three overdue milestones across two overdue journeys
    Then the Timeline view shows a summary of overdue journeys and milestones

  Scenario: Milestone rows follow the stored order
    Given a journey with two milestones targeting Q4 2026
    Then its milestone rows render in the journey's stored milestone order, dated before undated

  Scenario: A row navigates to the capability
    When the user activates a journey row
    Then that capability's drawer opens with its board card scrolled into view

  Scenario: The view is addressable
    Given a user on the Timeline view
    Then the URL identifies the view and restores it on reload and deep link

  Scenario: Empty state
    Given no active journeys exist
    Then the Timeline view states that no journeys are planned or in flight

  Scenario: Read-only users can use the view
    Given a user without the architect permission
    Then the Timeline view renders identically and carries no write affordances
```

---

## Business Rules & Invariants

1. **Content** — all active journeys (`planned`, `in-flight`); terminal journeys never appear.
2. **Overdue milestone** — has a target period, the period is before the current quarter, and status is not `done`.
3. **Overdue journey** — has a target period, the period is before the current quarter, and status is not `done`.
4. **Done and undated items are never overdue.**
5. **Milestones are self-describing** — every milestone renders its label, status, and target period in the view; identifying a milestone must not require hover or opening the drawer.
6. **Axis** — spans the earliest to the latest target period across shown journeys and milestones, always includes the current quarter, and marks it.
7. **Deterministic ordering** — domains alphabetically; within a domain, journeys by target period ascending (undated last), ties by capability name; within a journey, milestone rows in the stored order, dated before undated.
8. **Computed, never persisted** — overdue state, axis bounds, and summaries are derived at read time from target periods and statuses; no schedule state is written anywhere (184 posture).
9. **Quarter granularity** — the view invents no calendar dates; every placement is a year + quarter (182 rule 9).
10. **Permission** — the view follows board read permission and carries no write affordances.

---

## Acceptance Criteria

- [x] The Timeline view renders every active journey per rules 1 and 5–7, with kind, status, progress, labelled milestone rows at their quarters, and the target period marked
- [x] Overdue marking follows rules 2–4 exactly, covered by unit tests including boundary cases (current-quarter items are not overdue; done past-period items are not overdue)
- [x] Undated milestones render as the journey's last rows, marked as having no date, in stored order
- [x] The current-quarter marker and the overdue summary render; the empty state renders when no active journeys exist
- [x] Activating a row opens the capability's drawer with its card scrolled into view; the view state is URL-addressable
- [x] No schedule state is persisted; changing a milestone status or period changes the next render with no other bookkeeping
- [x] Read-only users see the identical view with no write affordances
- [x] Every BDD scenario has at least one corresponding test
- [x] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

Read-side only. `architecturedirection`'s existing bulk journey query already carries kind, status, progress, target periods, and milestones; extend its shape only if a field the view needs is missing. All schedule computation is presentation-layer derivation.

### Domain Model

None.

### API Surface

None new beyond, at most, completing the bulk journey query's read shape.

### Persistence

None.

### Frontend

A Timeline presentation on the Business Domains page, entered from the board toolbar beside the lens switcher and reflected in the URL like a lens (183). Rows group by domain on a shared quarter axis; overdue styling uses the existing status/danger tokens; row activation reuses the board's deep-link/scroll/drawer machinery. Queries and invalidation follow `easi-frontend-data`, sharing the bulk journey query and its existing mutation invalidation.

### Cross-Context Integration

None.

---

## Design Decisions

1. **A view of the Business Domains page, not a new page and not a fourth lens** — D5 rejects a second landscape surface; the lenses recolor the same card grid while the timeline answers "when" with a different geometry, so it sits beside the lens switcher as its own URL-addressed presentation.
2. **Quarter markers, not Gantt bars** — journeys carry no start period, and inventing one was rejected as data-entry burden and false precision; quarter-placed markers reproduce how architects actually plan (182 decision 6).
3. **Overdue is computed, never persisted** — the 184/D9 posture: schedule health is a question derived from the data, with no acknowledge/dismiss state; updating the milestone is the only way to clear the flag.
4. **Active journeys only** — terminal journeys carry no schedule question; they remain in the drawer's journey history, and mixing them in was rejected as burying the on-track/behind answer the view exists for.
5. **Milestone dependencies rejected** — field feedback floated dependency links between milestones; stored order (196), dated periods, and overdue flags answer "show them in order and show what is behind", while dependency edges across journey aggregates would couple aggregates and drift into project-management tooling outside an EA landscape record.
6. **Labelled milestone rows, not bare markers** — field feedback: a dot gives no context. Each milestone is a row of its own — label, status, period, overdue flag — placed at its quarter, so the journey band grows vertically with its milestone count. A dots-per-cell band with hover tooltips was rejected: identification would require pointer hover, unusable on touch and useless when scanning the whole view.
7. **URL parameter is `presentation`** (implementation discovery) — `?presentation=timeline` addresses the view; `view` was unavailable as it is a globally registered canvas deep-link parameter. Board and map keep their existing localStorage preference, so visiting the timeline never overwrites the user's chosen board/map presentation.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Quarter granularity | No finer scheduling than a quarter | Matches planning practice; milestone labels carry any finer nuance |
| Markers, not bars | Rows read sparser than a Gantt | Status colours, progress, and overdue flags carry the health story |
| One row per milestone | Journey bands grow with milestone count | Compact single-line rows; milestone lists are short in practice (182 keeps phased work as milestones, not sub-plans) |
| Tenant-wide fetch | Cost grows with landscape size | Same indexed bulk query the board lenses already use |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (component-integration tests over the timeline view and page presentation switch; no backend changes, so no server integration tests)
- [x] API documentation updated (none — read-side frontend slice over the existing bulk journey query, no new endpoints)
- [x] User sign-off
