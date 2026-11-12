# 239 — Guided Capability Creation

> **Status:** pending
> **Depends on:** 237 (similar-entity suggestions), 200 (EA owner as user id)
> **Roadmap alignment:** outside roadmap: a creation-flow improvement inside Capability Mapping's existing commands; no bounded-context or capability change — slice E of [`docs/specs/data-quality.md`](../docs/specs/data-quality.md) (decision Q3)

---

## Problem Statement

A new L1 capability belongs to no business domain and has no EA owner until someone assigns them from the domain board and the details panel. A reviewed production tenant had 20 L1 capabilities in no domain, 9 of them single-node stubs, and 51 L1s without a description. A capability outside every domain is invisible on the domain boards, belongs to nobody's stewardship and drops out of every per-domain analysis.

This slice makes L1 creation ask for the business domain and the EA owner in the same flow, and tells the user where a new child capability lands.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect** | Create an L1 that is immediately on a domain board with someone answering for it |
| **Architect extending a tree** | See that a new child inherits its place from the parent |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Guided capability creation

  Scenario: Creating an L1 continues into placing it
    When an architect creates the L1 capability "Tracking" from the create dialog
    Then the capability exists
    And the dialog shows "Place the capability" with business domain and EA owner, neither set

  Scenario: Assigning a domain
    When the architect picks the domain "Logistics"
    Then "Tracking" is assigned to "Logistics"
    And the step shows business domain as done

  Scenario: Assigning more than one domain
    When the architect also picks "Ferry Freight Operations"
    Then "Tracking" is assigned to both domains

  Scenario: Setting the EA owner
    When the architect picks the user "Alice Smith" as EA owner
    Then "Alice Smith" is the EA owner of "Tracking"

  Scenario: The domain architect is proposed as EA owner
    Given "Logistics" has the domain architect "Alice Smith"
    When the architect has assigned "Tracking" to "Logistics"
    Then "Alice Smith" is preselected as EA owner and not yet saved

  Scenario: Creating an L1 from a domain board preassigns it
    Given the architect is on the board of "Logistics"
    When they create an L1 capability there
    Then it is assigned to "Logistics" and the step shows business domain as done

  Scenario: Finishing with gaps says so
    Given no domain is assigned
    Then the finish button reads "Finish later"
    And the step states "This capability is not in any business domain and will not appear on a domain board"

  Scenario: A child capability needs no placing
    When an architect creates "Invoicing" under "Billing"
    Then the dialog closes after creation
    And a notification states the domains "Invoicing" belongs to through "Billing"

  Scenario: A child of an unplaced L1 is told so
    Given the L1 "Tracking" is in no business domain
    When an architect creates a child under "Tracking"
    Then the notification states that "Tracking" is in no business domain and offers to place it

  Scenario: Placing later from the details panel
    Given the L1 "Tracking" is in no business domain
    When an architect opens its details
    Then a "Place the capability" prompt names what is missing
    And selecting it opens the placing step

  Scenario: Only what the user may do is offered
    Given the caller may create capabilities but may not assign domains
    Then the placing step offers only the EA owner
```

---

## Business Rules & Invariants

1. **Create first, then place** — the capability is created by the existing command; domain assignment and EA owner are the existing commands on the created capability.
2. **Placing items, L1 only** — at least one business domain; an EA owner that is a user of the tenant. L2–L4 capabilities have no placing items: their domains are those of their L1 ancestor and spec 230 checks EA ownership on L1 only.
3. **Each item is offered only by its affordance** on the capability and domain resources; an item without its link is not shown and not counted as missing.
4. **Items save individually and immediately**; a failure affects only that item.
5. **Nothing is mandatory.**
6. **The proposed EA owner is a suggestion** — preselected from the domain architect of the first assigned domain, saved only when the user confirms.
7. **Derived** — whether an L1 is placed is computed from its current state; nothing records that the step was shown or skipped.

---

## Acceptance Criteria

- [ ] After creating an L1 from `CreateCapabilityDialog`, the dialog shows the placing step for it
- [ ] Domain assignment and EA owner call the existing commands through the capability's and domains' links; no new backend endpoint is added by this spec
- [ ] Creating an L1 from a domain board assigns that domain as part of the flow and shows it done
- [ ] The EA owner is preselected per rule 6 and not saved until confirmed
- [ ] The finish label and missing-items text follow the scenario
- [ ] Creating a child closes the dialog and shows the notification with the inherited domains, or the unplaced-L1 notice with a link that opens the placing step for the L1
- [ ] The capability details panel shows the prompt for an L1 while an offered item is missing
- [ ] An item whose affordance is absent is neither rendered nor counted
- [ ] Mutation effects invalidate the capability, business-domain, domain-board and home queries
- [ ] Every BDD scenario has at least one corresponding test
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

Frontend only, in the `capabilities` feature, using Capability Mapping's existing commands.

### Domain Model

None.

### API Surface

No change.

### Persistence

None.

### Frontend

`PlaceCapabilityStep` in `CreateCapabilityDialog`, a child-creation notification, and a prompt in the capability details panel, sharing one placement function. The step reuses the item component of spec 238 if that spec has landed and otherwise introduces it.

### Cross-Context Integration

None.

---

## Design Decisions

1. **Domain is asked for, not required** — an L1 may be drafted before its domain is agreed. An unplaced L1 is reported (spec 236 `capability-outside-domain`) to its creator (spec 235). Alternative: reject an L1 without a domain (rejected: a domain invariant that existing data violates, and Capability Mapping deliberately models assignment as a separate fact).
2. **Only L1s are placed** — children inherit; asking at every level repeats a question with one answer.
3. **Domain architect proposed, not applied** — the likeliest EA owner, but accountability should not be assigned by default.
4. **Description stays on the create form** — required-ness of the description is the one-pager configuration's decision.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Skippable | Unplaced L1s can still be created | Stated in the dialog at the moment of skipping, then reported by spec 236 |
| Separate specs for application and capability guidance | Two similar flows | One shared item component; each ships and is validated alone |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
