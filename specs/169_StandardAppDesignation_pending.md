# 169 — Standard Application Designation

> **Status:** pending
> **Depends on:** [166 — Logical Capability Rename](166_LogicalCapability_Rename_pending.md), [167 — Direction on a Logical Capability](167_Direction_Aggregate_Capture_pending.md)
> **Conceptual basis:** [`mockups/architecture-direction-model.md`](../mockups/architecture-direction-model.md)

---

## Problem Statement

A daily decision an engineer or product manager faces routinely is: *"For this capability, which application should I use?"* Today, that question lives in tribal knowledge — different domains use different apps for similar work, and the architecture group's view on which one *should* be the standard is communicated in slides at best.

This slice introduces **Standard Application Designation** — a structured statement attached to a Logical Capability that says *"the agreed standard application for this capability is X."* It is the Type-2 path from the conceptual model: physical capabilities can stay distributed across domains while the application landscape consolidates. The two are independent decisions and the model treats them as such.

After this slice ships, anyone at DFDS can navigate to a Logical Capability and answer "which application should I be using" in five seconds. They get either an agreed standard, a proposed standard under discussion, or an explicit "no standard yet" — every state is a valid five-second answer.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Product Manager / Engineer** | Look up the agreed standard application for a capability they're about to invest in or build against, in seconds. |
| **Enterprise Architect** | Capture the group's agreed standard for a Logical Capability and progress it through the same draft / proposed / agreed flow Directions use. |
| **Domain Owner** | See whether the application currently used in their domain matches the agreed standard, so they can plan migrations. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Standard Application Designation

  Scenario: A logical capability with no standard surfaces that explicitly
    Given I am viewing a Logical Capability with no Standard App designation
    Then the surface shows "no standard yet"
    And the state is distinguishable from a draft

  Scenario: An architect designates a draft standard
    Given I am an architect viewing a Logical Capability
    When I designate an application as the proposed standard with a narrative
    Then a Standard App designation is created in draft

  Scenario: An architect advances the designation toward agreement
    Given a Standard App designation exists in draft
    When I advance the status to proposed
    Then the designation's status updates and the change is recorded as a discrete event
    When I advance the status to agreed
    Then the designation's status updates and the change is recorded as a discrete event

  Scenario: A standard can be replaced
    Given a Logical Capability has an agreed Standard App designation for application A
    When the architect designates application B as the new standard
    Then the new designation supersedes the old one
    And the old designation is preserved in history with status superseded
    And the new designation starts in draft

  Scenario: A reader sees the standard in context
    Given I have read access to a Logical Capability with an agreed Standard App designation
    When I open the Logical Capability
    Then I see the agreed standard application named clearly
    And I see whether it is agreed or proposed
    And I cannot designate, advance, or supersede

  Scenario: Browsing the application portfolio
    Given the system has Logical Capabilities with various designation states
    When I open the Application Portfolio surface
    Then I see every Logical Capability with its Standard App state — agreed standard, proposed standard, or none
    And I can identify the candidates the architecture group still has to decide

  Scenario: A standard references an application that becomes stale
    Given a designation references an application that has since been deleted
    When I view the designation
    Then the missing reference is marked stale
    And the designation otherwise renders normally
```

---

## Business Rules & Invariants

1. **A Logical Capability has at most one *active* Standard App designation at a time.** A designation is *active* if its status is `draft`, `proposed`, or `agreed`. `Superseded` and `rejected` designations are preserved but not active.
2. **A designation references one application by ID.** The reference is eventually consistent with the application catalog.
3. **A designation has a status workflow:** `draft` → `proposed` → `agreed`, with `rejected` reachable from `draft` or `proposed`, and `superseded` set automatically when a new designation replaces an `agreed` one.
4. **Replacing an agreed standard supersedes it.** The old designation transitions to `superseded` (preserved for audit); the new designation begins at `draft`. Both events are atomic — replacement does not leave the Logical Capability with two active designations or with none.
5. **Status transitions are recorded as discrete past-tense events**, one per transition. Replay reconstructs status. No generic StatusChanged event.
6. **A designation has a stakeholder-readable narrative** explaining why this application and what it covers (e.g. "covers the operational and reporting layers; excludes legacy COBOL flows"). The narrative is required before a designation can advance from `draft` to `proposed`.
7. **A Standard App designation is independent of any Direction on the same Logical Capability.** The two can co-exist: a Logical can simultaneously have an agreed Direction (consolidating physically into one domain) AND an agreed Standard App (the application the consolidated capability will use). They can also exist apart — a Logical can have a Standard App without a Direction (the Type-2 case).
8. **Application references can become stale.** When the referenced application is deleted, the designation surfaces a stale-reference indicator but does not block reading.
9. **Authorisation is gated.** Designation, advancement, supersession, and rejection require an architect-level permission consistent with `architecturedirection`'s scheme. Reading follows the parent Logical Capability's read permission.
10. **Designation events are part of the published language** of `architecturedirection`. Downstream contexts (e.g. the Application Portfolio surface, the Open Discussions inbox in spec 172, the Target Architecture view in spec 171) subscribe to them.

---

## Acceptance Criteria

- [ ] An architect can designate an application as the standard for a Logical Capability with a narrative; the designation starts in `draft`
- [ ] An architect can advance a designation `draft` → `proposed` → `agreed`, and reject from `draft` or `proposed`
- [ ] Each transition is its own past-tense event; replay reconstructs status
- [ ] At most one active designation per Logical Capability is enforced at the command boundary
- [ ] Replacing an agreed designation supersedes the old one atomically; both transitions are visible in the event log
- [ ] Viewing a Logical Capability surfaces the active Standard App designation (or "no standard yet") within the same view; readable in five seconds
- [ ] An Application Portfolio surface lists every Logical Capability with its current designation state; usable by non-architects
- [ ] A designation referencing a deleted application renders with a stale-reference indicator
- [ ] HATEOAS affordances on a Logical Capability response advertise designate / advance / supersede / reject only when the user is authorised
- [ ] Designation events are documented as published-language and are subscribable by downstream contexts
- [ ] All BDD scenarios have at least one corresponding test
- [ ] CodeScene `pre_commit_code_health_safeguard` passes on every modified file

---

## Architecture

### Ownership

`StandardAppDesignation` is an aggregate in the `architecturedirection` bounded context, parallel to `Direction`. Same authorisation, event-sourcing, and publishing patterns.

### Domain Model

The aggregate carries: an identity, a Logical Capability ID, an application ID, a horizon (`now` / `next` / `later`), a status, and a narrative. Status reconstruction is event-driven; transitions are discrete events.

The "at most one active per Logical Capability" invariant is enforced at the command-handler level via the read model (the established EASI pattern), not as an in-aggregate cross-aggregate constraint.

### API Surface

The Standard App is exposed under the Logical Capability's resource tree — designation operations target a Logical Capability and the active designation surfaces inline (or via HATEOAS link) on the Logical Capability response. A separate Application Portfolio resource provides the cross-Logical view: list every Logical Capability with its current designation state.

The contract obligation: a Logical Capability response makes the standard-application question answerable in one round trip. Exact route shapes settled at implementation time per the API standards skill.

### Persistence

Event-sourced. Read-side projections drive the per-Logical detail panel and the Application Portfolio list.

### Frontend

A Standard App panel surfaces on the existing Logical Capability detail surface, alongside the Direction panel from spec 167. Both panels answer different alignment questions; they coexist without blocking each other.

A new Application Portfolio surface lands under the architecture-direction area of the UI, scannable as a list. Where it sits in the navigation and the column shape are settled during implementation; the constraint is that a non-architect can scan it and answer "which logicals still need a standard?" without training.

### Cross-Context Integration

`architecturedirection` subscribes to the application catalog (existing `componentregistry` or equivalent context) for application existence and to detect stale references. The integration is read-only — designation does not write into the application catalog.

---

## Design Decisions

1. **StandardAppDesignation as its own aggregate.** Per the DDD memo. Independent lifecycle (a designation can be drafted and superseded without touching the Logical Capability), distinct authorisation surface, and an explicit audit trail of "this app was the standard from / until." Embedding the standard as a property of Logical Capability would lose the supersession history.

2. **Supersession as a distinct status.** The "old standard" is not the same as "rejected" — the group did once agree on it; it was simply replaced. Modelling supersession explicitly preserves the why-and-when of the change. Alternative considered: hard-delete the old designation when replaced — rejected because it loses audit history.

3. **Narrative required before `proposed`.** A designation without a stakeholder narrative cannot advance — it forces the author to articulate "why this app and what it covers" before the group debates it. The friction is intentional.

4. **Designation independent of Direction on the same Logical.** The two answer different questions: Direction = where does the capability live; Standard App = which application realises it. The model insists they decouple; the aggregate boundary respects the model. A future cross-aggregate query can join them for synthesis views (the Target Architecture view in spec 171 is exactly that).

5. **Application Portfolio as a read-side projection only, not a new aggregate.** The Portfolio is a *view* over many designations; it doesn't own state of its own. Modelling it as a projection keeps the aggregate count honest.

6. **Permission scheme inherited from `architecturedirection` (settled in spec 167).** Designations and Directions live in the same bounded context and share the architect-level write surface.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Standard App as separate aggregate | One more aggregate to author, test, and document | The audit-trail and lifecycle independence are worth the cost; the aggregate is small |
| Supersession as a distinct status | Two status values mean "no longer active" (`rejected`, `superseded`); UI must distinguish | Both are read-only states; the distinction is meaningful (one is a no-go, one is a replaced yes) |
| Narrative required before `proposed` | Author cannot skip the why-statement | Intentional friction; the daily-alignment use case depends on the narrative being present |
| Independent of Direction | Two separate panels on the Logical Capability detail | The composite Target Architecture surface (spec 171) joins them when synthesis is needed |
| Application Portfolio as projection | Recomputation on every relevant event | Standard EASI pattern; the projection is small |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
