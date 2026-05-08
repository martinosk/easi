# 167 — Direction on a Logical Capability

> **Status:** pending
> **Depends on:** [166 — Logical Capability Rename](166_LogicalCapability_Rename_pending.md)
> **Conceptual basis:** [`mockups/architecture-direction-model.md`](../mockups/architecture-direction-model.md), [`mockups/architecture-direction-ddd.md`](../mockups/architecture-direction-ddd.md)

---

## Problem Statement

Today a Logical Capability carries no information about whether the architecture group has a direction on it — the decisions live in conversations, slides, and shared documents, not in the tool. Anyone outside the room cannot see the direction without asking an architect. That is the exact alignment-decision friction the model exists to remove.

This slice introduces the **Direction** concept — a structured statement attached to a Logical Capability that says: *what the group intends to do here* (consolidate / decompose / stay), *where it is in the group's decision process* (draft / proposed / agreed / rejected), and *why* (a stakeholder-readable narrative). After this slice ships, an individual making a daily decision can open a Logical Capability and answer "is there a direction on this?" in five seconds.

Direction is the load-bearing addition. Subsequent slices (Discover, Direction Map, Target Architecture, Open Discussions) are read-side surfaces over the same aggregate. This spec defines the aggregate and the simplest write-side flow: capture, advance, reject, view in context.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect (author)** | Capture and progress a direction on a Logical Capability without leaving its detail surface; advance status as the group reaches consensus. |
| **Enterprise Architect (reader)** | See the current direction on any Logical Capability they navigate to, with enough context to know whether it is final, under discussion, or still being shaped. |
| **Domain Owner / Product Manager** | When making an investment or design decision involving a Logical Capability, see at a glance whether the architecture group has a direction, what it is, and how settled it is. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Direction on a Logical Capability

  Scenario: A logical capability with no direction shows that explicitly
    Given I am viewing a Logical Capability with no Direction
    Then the detail surface shows an explicit "no direction set" state
    And the state is distinguishable from a Direction in draft

  Scenario: An architect captures a draft direction
    Given I am an architect viewing a Logical Capability
    When I capture a Direction with a type, one or more source physical capabilities, and a narrative
    Then the Direction is created in draft status
    And it appears on the Logical Capability detail surface within five seconds

  Scenario: An architect advances a direction toward agreement
    Given a Direction exists in draft
    When I advance the status to proposed
    Then the Direction's status updates and the change is recorded as a discrete event
    When I advance the status to agreed
    Then the Direction's status updates and the change is recorded as a discrete event

  Scenario: A direction can be rejected from any non-terminal status
    Given a Direction exists in draft or proposed
    When I reject the Direction
    Then the Direction's status becomes rejected and is preserved for audit
    And the Direction no longer presents as the active alignment answer for the Logical Capability

  Scenario: A reader sees the current direction in context
    Given I have read-only access to a Logical Capability with a Direction
    When I open the Logical Capability
    Then I see the Direction's type, status, and narrative
    And I cannot create, advance, or reject the Direction

  Scenario: A direction references physical capabilities that may change
    Given a Direction references a physical capability that has since been deleted
    When I view the Direction
    Then the missing reference is marked as stale
    And the Direction otherwise renders normally

  Scenario: At most one active direction per logical capability
    Given a Logical Capability has an agreed Direction
    When an architect attempts to capture a new Direction on the same Logical Capability
    Then the system prevents the second active Direction
    And the architect is offered the path of rejecting the existing Direction first
```

---

## Business Rules & Invariants

1. **A Logical Capability has at most one active Direction at a time.** A Direction is *active* if its status is `draft`, `proposed`, or `agreed`. `Rejected` Directions are preserved but not active.
2. **A Direction is its own aggregate**, owned by the new `architecturedirection` bounded context. Its lifecycle is independent of the Logical Capability and of any physical capability it references.
3. **A Direction has a type.** One of: `consolidate` (multiple physicals merge into one), `decompose` (one physical splits into multiple), `stay` (explicitly confirmed no change). Type is immutable once the Direction is created — to change intent, reject and capture a new Direction.
4. **A Direction references one or more physical capabilities by ID.** These references are eventually consistent with the `capabilitymapping` context. The Direction does not embed physical capability state; it carries only the references and any local annotations needed for the narrative.
5. **A Direction has a status workflow:** `draft` → `proposed` → `agreed`, with `rejected` reachable from `draft` or `proposed` (terminal). Forward-only on the agreement axis: a Direction does not transition from `agreed` back to `proposed` — to revisit, reject and replace.
6. **Status transitions are recorded as discrete past-tense domain events**, one event per transition (e.g. an event for proposing, an event for agreeing, an event for rejecting). Replay reconstructs the current status. There is no generic "status changed" event.
7. **A Direction has a stakeholder-readable narrative.** One to two sentences naming what the group decided and why. The narrative is required before a Direction can advance from `draft` to `proposed`.
8. **A Direction's source-capability references can become stale.** When a referenced physical capability is deleted, the Direction surfaces a stale-reference indicator but does not block reading or further status transitions. The architect can edit the source list to remove the stale reference.
9. **Authorisation is gated.** Capture, advance, and reject require an architect-level permission. Reading a Direction follows the same read permission as the underlying Logical Capability. The exact permission key is settled at implementation time (see Design Decisions).
10. **Direction is a published-language concept of `architecturedirection`.** Other contexts that need to know about Directions (notably the Discover view in slice 168 and the Direction Map in slice 170) subscribe to its events. No context outside `architecturedirection` writes Directions.

---

## Acceptance Criteria

- [ ] An architect can create a Direction on any Logical Capability with type, source physical capabilities (one or more), and narrative; the Direction starts in `draft`
- [ ] An architect can advance a Direction's status from `draft` → `proposed` → `agreed`, and reject from `draft` or `proposed`
- [ ] Each status transition is persisted as its own past-tense domain event; replaying the event store reconstructs the current status
- [ ] A Logical Capability cannot host two simultaneously-active Directions; the second creation attempt is rejected with a clear error
- [ ] Viewing a Logical Capability with a Direction surfaces, within the same view, the Direction's type, current status, and narrative — readable by any user with view permission on the Logical Capability
- [ ] Viewing a Logical Capability without a Direction surfaces an explicit "no direction" state distinguishable from a draft
- [ ] A Direction whose source list contains a deleted physical capability still renders, with the stale references clearly marked
- [ ] Read-only users see the Direction but cannot create, advance, or reject
- [ ] HATEOAS affordances on a Logical Capability's response advertise create-direction and advance-direction operations only when the calling user is authorised; readers see no such affordances
- [ ] Other bounded contexts can subscribe to `architecturedirection` events for read-side use; the published-language event contract is documented
- [ ] All BDD scenarios above have at least one corresponding test
- [ ] CodeScene `pre_commit_code_health_safeguard` passes on every modified file

---

## Architecture

### Ownership

A new bounded context: `architecturedirection`. It owns the `Direction` aggregate end-to-end — write-side, read-side projections, API surface, frontend integration. The `enterprisearchitecture` context (after spec 166) owns Logical Capabilities and is referenced read-only. The `capabilitymapping` context owns physical capabilities and is also referenced read-only via event subscriptions.

### Domain Model

The `Direction` aggregate carries: an identity, a Logical Capability ID, a type (`consolidate` / `decompose` / `stay`), one or more source physical capability IDs, an optional set of target placements (target domain references; relevant for `consolidate` and `decompose`), a horizon (`now` / `next` / `later`), a status, and a narrative. Status is reconstructed from the event log; status transitions are individual past-tense events.

Invariants are listed in Business Rules; the aggregate enforces them at the command boundary. Cross-aggregate invariants (uniqueness of active Direction per Logical Capability) are handled with the established pattern in EASI for one-active-per-parent (verified at the command-handler level via the read model).

### API Surface

Direction is exposed under the Logical Capability's resource tree — a Direction belongs to a Logical Capability and is fetched and written through routes that begin with `/api/v1/logical-capabilities/{id}/...`. The exact route shape is settled at implementation time per the API standards skill, but the contract obligation is: a single Logical Capability response surfaces (a) any active Direction inline or via a HATEOAS link, and (b) HATEOAS affordances for any operation the calling user is authorised to perform on it.

Status transitions are exposed as discrete operations (one for advancing, one for rejecting) rather than a free-form PATCH on status — this keeps the wire format honest about which transitions are valid.

### Persistence

Event-sourced, following the established EASI pattern. The aggregate's events stream is its source of truth; read models are projected from it. A read-side projection joins Directions with their parent Logical Capability for fast "show me the direction on this Logical" queries.

### Frontend

A `Direction` panel surfaces on the existing Logical Capability detail page, occupying enough room for type, status, narrative, and source-capability list at a glance. Where the panel sits and how it composes with existing detail content is settled during implementation; the constraint is that the alignment question — *is there a Direction on this; what type; what status* — must be answerable in five seconds without scrolling or drill-down.

Status transitions are exposed as actions on the panel and gated by HATEOAS affordances per EASI's existing pattern.

### Cross-Context Integration

`architecturedirection` subscribes to:
- `enterprisearchitecture` Logical Capability events, to know which capabilities exist and which it can host a Direction on.
- `capabilitymapping` physical capability events, to know which sources are valid and to detect when a referenced source has been deleted (driving the stale-reference indicator).

`architecturedirection` publishes Direction lifecycle events for downstream contexts. No outbound write commands cross context boundaries.

---

## Design Decisions

1. **Direction as its own aggregate, not embedded in Logical Capability.** Independent lifecycle (a Direction can be drafted, debated, and rejected without touching the Logical Capability), distinct authorisation surface (architect-only writes vs broader Logical reads), and clean separation between the steady-state classification (Logical Capability) and the change proposal (Direction). The DDD memo committed to this; the spec follows.

2. **Status transitions as discrete past-tense events, not a generic StatusChanged.** Aligns with the established EASI event-sourcing pattern; lets read-side projections subscribe to specific transitions (e.g. "every time a Direction is agreed, recompute the daily-alignment heat map") without filtering. Rejected because the alternative (one generic event with a status field) reads as less truthful in the event log and forces every consumer to know the status transition table.

3. **At most one active Direction per Logical Capability.** Multiple in-flight Directions on the same Logical lead to ambiguous alignment answers. Reject-and-replace is cleaner than concurrent drafts. Alternative considered: allow multiple drafts with one designated as "primary" — rejected as a needless concept that fails the five-second test (which Direction is the answer?).

4. **Stale references surface but do not block.** A Direction whose source capability has been deleted is still meaningful (it carries a recorded group decision). Hiding it would lose history; blocking deletion of physical capabilities to protect Directions is the wrong tail wagging the dog. Alternative considered: hard-delete the Direction when a source is deleted — rejected because it loses the historical record of a decision the group made.

5. **Type is immutable; reject-and-replace to change.** A Direction's *type* (consolidate / decompose / stay) is the central commitment. Changing it mid-flight obscures the audit trail of what the group decided when. Reject-and-replace makes the change explicit. Alternative considered: type as mutable until `agreed` — rejected because it muddles the meaning of the draft → proposed transition.

6. **Permission scheme deferred until implementation.** The DDD memo flags this as an open decision. Default presumption: a new `architecture-direction:*` permission family scoped per tenant, mirroring the existing `enterprise-arch:*` pattern. Confirmed during implementation against the existing permission infrastructure.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Direction is its own aggregate | Cross-aggregate invariant (one active per Logical) needs handler-level enforcement, not aggregate-internal | Established EASI pattern; verified by read-model lookup at command time |
| Status transitions as discrete events | More event types in the published language | Each carries clearer semantics for projections; the count is bounded (4 transitions) |
| Stale references render | Direction may show degraded data | Stale indicator is explicit; user can edit the source list |
| At most one active Direction | A group exploring two alternatives at once cannot model both in the tool | Two alternatives can be discussed in the existing draft's narrative; if they diverge enough to need separate aggregates, that is itself a signal the group is past discussion and into decision |
| Type immutability | Editing a misclassified Direction requires reject-and-replace | The cost (one extra event in the log) is small relative to the audit-trail clarity gained |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
