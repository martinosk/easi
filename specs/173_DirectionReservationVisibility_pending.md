# 173 — Direction Reservation Visibility: Make Reservation & Carve-Out Legible to Architects
> **Depends on:** 172 (DirectionIsTheAssociation — defines the domain rules this spec surfaces in the UI)

---

## Problem Statement

Spec 172 makes a domain decision with a non-obvious side effect: an **active direction** is any direction that is `draft`, `proposed`, or `agreed`, and **a draft reserves its sources the moment it is captured**. Capturing (or even drafting) a direction therefore:

- **Reserves** each source under R1 — no other EC may source the same capability.
- **Carves out** descendants of another EC's source under R2 — a subtree silently moves ownership.

172 deliberately scopes out the *experience* of this. Today an architect has no way to see **why** a capability is unavailable, **which EC and which direction** holds it, or that **their own abandoned draft** is the thing blocking a colleague. This spec covers making those moments legible and recoverable — it adds **no new domain rules**, only visibility and affordances over the rules 172 defines.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | Understand at a glance why a capability can't be sourced, who holds it, and how to resolve it. See what their draft is reserving before others are blocked. |
| **Enterprise Architect (blocked party)** | Discover that a *draft* (not just an agreed direction) is holding a capability, and have a path forward (request release, wait, escalate). |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Reservation and carve-out are legible to architects

  Scenario: A reserved capability explains who holds it
    Given "Customer Account Creation" is a source of an active direction on EC "Customer Identity"
    When the architect searches for it while capturing a direction on another EC
    Then it is shown as unavailable
    And the reason names the holding EC "Customer Identity" and the holding direction's status

  Scenario: A draft reservation is distinguished from a firm one
    Given a capability is reserved only by a "draft" direction
    When the architect views the reason it is unavailable
    Then the UI communicates the reservation is provisional (held by a draft, not yet proposed/agreed)

  Scenario: Capturing a direction previews its carve-outs before commit
    Given the architect is sourcing an ancestor whose subtree includes capabilities owned by other ECs
    When they review the direction before capturing
    Then the UI lists which descendants will be carved out and to which EC

  Scenario: An architect can see what their own active directions are reserving
    Given the architect has captured directions across several ECs
    When they open their reservations view
    Then they see each capability they hold and the direction (and status) holding it

  Scenario: Abandoned-draft reservation is recoverable
    Given a capability is blocked only by another architect's stale draft
    When the blocked architect views the conflict
    Then they are offered a path forward (e.g. request release / notify holder)
    # exact mechanism TBD in this spec
```

---

## Business Rules & Invariants

This spec introduces **no new domain rules**. It surfaces the rules defined in 172 (R1 same-node exclusivity, R2 subtree composition with most-specific-wins carve-out, and the "active direction" definition that makes drafts reserve). Any rule change belongs in 172, not here.

---

## Open Questions

- What is the resolution path for a reservation held by a stale/abandoned draft? (request release, auto-expiry of drafts, escalation, manual reject by holder)
- Should the carve-out preview at capture time be informational only, or block capture until acknowledged?
- Is a global "my reservations" view warranted, or is per-EC surfacing enough?

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
