# 172 — Open Discussions Inbox

> **Status:** pending
> **Depends on:** [167 — Direction on a Logical Capability](167_Direction_Aggregate_Capture_pending.md), [169 — Standard Application Designation](169_StandardAppDesignation_pending.md)
> **Conceptual basis:** [`mockups/architecture-direction-model.md`](../mockups/architecture-direction-model.md), [`mockups/architecture-direction.html`](../mockups/architecture-direction.html)

---

## Problem Statement

The architecture group meets regularly to advance Directions and Standard App designations through their decision flow. Today, building the agenda for those meetings is manual — someone scans through Logicals, remembers what's outstanding, and assembles a list. This makes it easy to miss items in flight; the agenda is only as good as one person's memory.

This slice introduces the **Open Discussions inbox** — a single read-only surface that lists, in one place, every Direction and every Standard App designation currently in `draft` or `proposed` status across the whole system. The list is the agenda, generated rather than maintained.

After this slice ships, the architecture group can open one tab before a meeting and see exactly what's outstanding. An empty inbox reads "Nothing on the agenda" — itself a five-second alignment answer.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect (preparing for a working session)** | One surface listing every outstanding decision so meeting prep is mechanical, not memory-bound. |
| **Enterprise Architect (in session)** | A live surface during the meeting that updates as items advance, so the group can see "what's left" shrink. |
| **Domain Owner / Stakeholder** | Visibility into what the architecture group is currently deciding, so they can flag items relevant to their domain before they're agreed. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Open Discussions Inbox

  Scenario: Browse everything outstanding
    Given the system has Directions and Standard App designations in various statuses
    When I open the Open Discussions surface
    Then I see every Direction whose status is draft or proposed
    And every Standard App designation whose status is draft or proposed
    And nothing whose status is agreed, rejected, or superseded

  Scenario: Each entry carries enough context to triage
    Given I am viewing the Open Discussions list
    Then each entry shows: its kind (Direction or Standard App), the affected Logical Capability, the current status, a one-line narrative summary, and the last-touched timestamp

  Scenario: Live updates as items advance
    Given I am viewing the Open Discussions list during a session
    When an architect advances a Direction from proposed to agreed
    Then the Direction disappears from the list within a second
    When an architect advances a Direction from draft to proposed
    Then the Direction's row updates within a second to reflect the new status

  Scenario: Click into an entry to act on it
    Given I am viewing the Open Discussions list
    When I select an entry
    Then I am taken to the appropriate edit surface for that aggregate
    And edits there reflect back into the list

  Scenario: An empty inbox is itself a useful answer
    Given the system has no Direction or Standard App designation in draft or proposed status
    When I open the Open Discussions surface
    Then I see an explicit "Nothing on the agenda" state
    And the state is unambiguous

  Scenario: Read-only access shows the same list
    Given I have read-only access
    When I open the Open Discussions surface
    Then I see the same list an architect would see, scoped to entries I can read
    And the click-through surfaces are read-only

  Scenario: Authorisation scopes the visible list
    Given I cannot read certain Logical Capabilities
    When I open the Open Discussions surface
    Then entries for those Logical Capabilities are not in the list
```

---

## Business Rules & Invariants

1. **The inbox is a read-side surface.** It owns no aggregates. Every row is an aggregate owned elsewhere (Direction or Standard App designation).
2. **Inclusion criterion: status ∈ {draft, proposed}.** Agreed, rejected, and superseded entries are excluded by default. They remain accessible from their respective home surfaces.
3. **Two aggregate kinds are unified in one list.** Directions and Standard App designations both contribute. The user sees them in one queue because they are both decisions the group has to make.
4. **Each entry carries enough context for triage** without click-through: kind, parent Logical Capability, current status, a one-line narrative summary, last-touched timestamp.
5. **The list is sortable by recency by default.** Last-touched first. Other sort orders (by Logical name, by status) are nice-to-haves; the implementation may include them but does not have to.
6. **Updates are live.** When an aggregate's status changes, the list reflects the change in under a second. The mechanism (websocket / polling / projection-driven push) is settled during implementation; the user-facing requirement is freshness.
7. **Empty state is meaningful.** "Nothing on the agenda" is a positive answer to a daily question, not a degenerate state. The empty surface reads cleanly.
8. **The surface respects read authorisation.** Entries for aggregates the caller cannot read are not in the list.
9. **Click-through routes to canonical edit surfaces.** Selecting an entry navigates to the spec 167 Direction editor or the spec 169 designation editor, as appropriate. The inbox does not offer in-place editing.
10. **No new aggregate, no new write surface.** The slice is purely additive on the read side.

---

## Acceptance Criteria

- [ ] The Open Discussions surface lists every Direction with status `draft` or `proposed`, plus every Standard App designation with status `draft` or `proposed`
- [ ] Each entry shows kind, affected Logical Capability, current status, a one-line narrative summary, and last-touched timestamp
- [ ] The list updates within one second when any included aggregate transitions status
- [ ] Aggregates that transition out of the included statuses (to `agreed`, `rejected`, `superseded`) leave the list within one second
- [ ] Selecting an entry navigates to the appropriate editor (the spec 167 Direction editor or the spec 169 designation editor); edits there propagate back to the list
- [ ] An empty inbox renders an explicit "Nothing on the agenda" state
- [ ] Read-only users see the same list (scoped to readable entries) and read-only click-through
- [ ] The list honours read authorisation; out-of-scope entries are excluded
- [ ] All BDD scenarios above have at least one corresponding test
- [ ] CodeScene `pre_commit_code_health_safeguard` passes on every modified file

---

## Architecture

### Ownership

A read-side feature of `architecturedirection`. No new aggregate.

### Domain Model

No new aggregates. The slice introduces a read model that flat-projects across Direction events and StandardAppDesignation events. The projection's record shape is the row shape required to populate the list (kind, parent Logical, status, narrative, last-touched).

The read model is rebuildable from the event store of `architecturedirection`.

### API Surface

A single read endpoint returns the open-discussions list for the calling user, honouring read authorisation. Pagination, if needed at production scale, is added during implementation; the unfiltered list is small enough that a single response is the default expectation.

The endpoint shape: a list of entries, each carrying every field listed in Business Rule 4.

### Persistence

A flat read model in `architecturedirection`, projected from Direction and StandardAppDesignation events. Live updates use the established EASI projection / push mechanism (settled during implementation).

### Frontend

A new Open Discussions surface in the architecture-direction area of the UI. Single list, last-touched-first by default. Click-through routes to the canonical editors. Empty state reads "Nothing on the agenda" with no further chrome.

### Cross-Context Integration

None outside `architecturedirection`. The slice consumes only events from within its own context.

---

## Design Decisions

1. **One list, two aggregate kinds.** Directions and Standard App designations are different things to model but the same thing to discuss — items the group has to converge on. Splitting into two lists would force the user to look in two places before a meeting. Alternative considered: tabbed or grouped — rejected; the value is the unified queue.

2. **Read-only with click-through.** Editing in place would mean replicating the edit forms across surfaces. Click-through to the canonical editor stays consistent.

3. **Live updates are a hard requirement.** A list that goes stale during a meeting is worse than no list — the group would re-build it from memory anyway. The implementation chooses the mechanism (push / poll); the user-facing freshness budget is one second.

4. **Empty state is a positive answer.** "Nothing on the agenda" is a useful daily-alignment answer for an architect or a stakeholder peering in. Treating empty as a success state, not as a degenerate state, matches the model's intent.

5. **Default sort by recency.** The most-recently-touched item is the most likely current focus. Other orderings are nice-to-haves; the spec does not require them.

6. **No comment threads, no notifications, no voting.** This surface is the agenda, not the discussion forum. Anything that turns it into a project-management tool is out by the model's principles.

7. **Authorisation is a hard constraint.** Stakeholders may legitimately have read access to some Logicals and not others. The list must not leak entries for unreachable aggregates.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Unified list of two aggregate kinds | Mixed-type rendering needs slight differentiation | Each row shows its kind; the unified queue is the primary value |
| Read-only with click-through | Cost per edit is one click | Worth it for editor consistency |
| Live update budget of one second | Implementation must use a real push or fast-poll mechanism | Established EASI patterns exist; the slice is small enough that the mechanism choice doesn't block scope |
| No comment threads or notifications | Some groups expect inbox-like features | Out of model scope by the user's principles; can be reconsidered with a separate spec if the principle changes |
| Default sort by recency only | Some users may want by-Logical or by-domain | Cheap to add later if needed; not blocking |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
