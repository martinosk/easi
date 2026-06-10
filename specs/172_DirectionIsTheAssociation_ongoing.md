# 172 — Direction Is the Association: Collapse Linking and Inclusion into Direction Sources
> **Depends on:** 101 (EnterpriseCapability_Linking_UI — done, superseded by this spec), 167 (Direction_Aggregate_Capture — done)

---

## Problem Statement

Architects today must learn two distinct concepts on an Enterprise Capability (EC):

1. **Linking** — a standalone act that says "this domain capability falls under this EC."
2. **Setting direction** — a workflow-driven `Direction` aggregate that captures consolidate/decompose/stay intent with type, horizon, narrative, sources, and placements.

This is redundant: *"If a domain capability gets a direction in an EC, it's implicitly linked anyway."* 

Fix: **collapse capability linking into direction.**


---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | One act, one place: capture the direction over an EC, and its sources — with their subtrees — compose the EC. Correct mistakes by editing the direction. Recompose by rejecting and recapturing. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Enterprise Capability composition via Direction

  Scenario: A direction's source set defines the EC's included capabilities
    Given an active Enterprise Capability "Customer Identity" with no active direction
    And a domain capability "Customer Account Creation" not associated with any EC
    When the architect captures a "consolidate" direction in draft on "Customer Identity" with source "Customer Account Creation"
    Then the draft is accepted even though it has a single source
    And the EC overview lists "Customer Account Creation" under its included capabilities
    And the included-capability count is 1, since the source has no descendants
    And the EC's composition is derived from the direction's sources and their subtrees

  Scenario: A draft direction may carry a single source regardless of type (R8)
    Given an active EC "Customer Identity" with no active direction
    When the architect captures a "consolidate" direction in draft with a single source
    Then the draft is accepted
    But advancing the draft to "proposed" is rejected until the type's source cardinality is satisfied

  Scenario: Sources may be domain capabilities at any hierarchy level
    Given an active EC "Order Management" with no active direction
    And four capabilities at levels L1, L2, L3, and L4 that are neither ancestors nor descendants of one another and have no further descendants
    When the architect captures a direction sourcing all four
    Then the direction is accepted
    And exactly those four capabilities are listed under the EC's included capabilities

  Scenario: Excluding a capability removes it from the active direction's source set
    Given an EC "Customer Identity" with an editable direction sourcing "A", "B", "C"
    When the architect excludes "B" from the EC overview
    Then the direction now sources only "A" and "C"
    And the EC overview no longer lists "B"
    And no separate "exclusion" event is emitted; this is a Direction source-set change

  Scenario: Excluding from an agreed direction is rejected
    Given an EC "Customer Identity" with an `agreed` direction sourcing "A" and "B"
    When the architect attempts to exclude "A"
    Then the action is rejected because agreed directions are immutable
    And the EC overview communicates that recomposition requires rejecting the direction

  Scenario: Direction is rejected when the same capability is already sourced by an active direction on a different EC (R1)
    Given a domain capability "Customer Account Creation" is a source of an active direction on EC "Customer Identity"
    When the architect attempts to capture a direction on EC "Identity Platform" sourcing the same "Customer Account Creation"
    Then the direction is rejected with an explanation referencing "Customer Identity"

  Scenario: Sourcing an ancestor is accepted; a descendant sourced elsewhere is carved out (R2)
    Given a domain capability "Customer Identity Mgmt" with children "Customer Consent" and "Customer Fraud Prevention"
    And "Customer Fraud Prevention" is a source of an active direction on EC "Take Payment"
    When the architect captures a direction on EC "CRM" sourcing the ancestor "Customer Identity Mgmt"
    Then the direction is accepted
    And "CRM" includes "Customer Identity Mgmt" and "Customer Consent"
    And "Customer Fraud Prevention" remains owned by "Take Payment"
    And "CRM" does not include "Customer Fraud Prevention"

  Scenario: A carve-out carries its entire subtree (R2)
    Given an active direction on EC "Take Payment" sourcing "Customer Fraud Prevention"
    And "Customer Fraud Prevention" has a descendant "Chargeback Handling"
    When an active direction on EC "CRM" sources the ancestor "Customer Identity Mgmt"
    Then "Customer Fraud Prevention" and "Chargeback Handling" are both owned by "Take Payment"
    And neither is included under "CRM"

  Scenario: A still-more-specific source carves a deeper node out of a carve-out (R2)
    Given "Customer Fraud Prevention" (owned by "Take Payment") has a descendant "Chargeback Handling"
    When an active direction on EC "Disputes" sources "Chargeback Handling" specifically
    Then "Chargeback Handling" and its descendants are owned by "Disputes"
    And "Customer Fraud Prevention" remains owned by "Take Payment" without "Chargeback Handling"

  Scenario: Deleting an EC releases its carve-outs back to the ancestor source (R2, R6)
    Given EC "CRM" has an active direction sourcing "Customer Identity Mgmt"
    And EC "Take Payment" has an active direction sourcing the descendant "Customer Fraud Prevention", carving it out of "CRM"
    When "Take Payment" is deleted
    Then its direction is rejected as part of deletion
    And "Customer Fraud Prevention" is no longer carved out
    And "Customer Fraud Prevention" is again included under "CRM" via its ancestor source

  Scenario: Direction cannot be captured on an inactive EC (R4)
    Given an inactive Enterprise Capability "Legacy Customer Identity"
    When the architect attempts to capture any direction on it
    Then the capture is rejected

  Scenario: EC deletion removes all associations
    Given an EC with an active direction sourcing two capabilities
    When the EC is deleted
    Then the direction is rejected as part of deletion
    And no capability remains "under" the deleted EC

  Scenario: EC overview when no direction is set
    Given an EC with no active direction
    When the architect opens the EC overview
    Then they see "No direction set" with a call to action to capture one
    And the included-capabilities list is empty
```

---

## Business Rules & Invariants

**Definition — "active direction."** A direction is *active* while its status is `draft`, `proposed`, or `agreed`, and its EC is not deleted; a `rejected` direction is inactive. R1 and R2 are evaluated against active directions. **A draft therefore reserves its sources** the moment it is captured — exclusivity and carve-outs apply from capture onward, not only from `proposed`. This is consistent with enforcing R1/R2 at capture time (capture produces a draft). Communicating *when and why* a capability becomes reserved or carved out to the architect (and handling an abandoned draft holding a reservation) is **out of scope here and deferred to a follow-up UI spec** (see [[173_DirectionReservationVisibility]]).

**R1 — Same-node exclusivity.** A given domain capability may be the explicit source of at most one active direction across all ECs. Sourcing the exact same capability on a second EC is rejected, with an explanation referencing the EC that already sources it.
**R2 — Subtree composition, resolved most-specific-first.** An active direction's composition is its explicit sources plus every descendant of those sources, computed recursively. Where a capability would be included via an ancestor source on one EC but is itself the explicit source — or lies within the subtree of a nearer explicit source — of an active direction on a *different* EC, the more specific source wins: that capability and its entire subtree are owned by the nearer source's EC and carved out of the ancestor's composition. Carve-outs are transitive — a carved-out subtree carries all its descendants with it, unless a still-more-specific source carves a deeper node out again. Consequently, sourcing an ancestor or descendant of another EC's source is *accepted* (with carve-out), not rejected; only R1 (same node) rejects. Only explicit sources can be excluded from a direction (excluding a source removes it and its implicit subtree); an implicitly included descendant is removed only by excluding its source or by a more-specific direction carving it out.
**R3 — No level restriction.** Sources may be at any level L1–L4.
**R4 — EC must be active.** Directions can only be captured on active ECs.
**R5 — Agreed-direction immutability.** Once a direction is `agreed`, its source set is frozen. Recomposing requires rejecting the direction and capturing a new one.
**R6 — Cascade on EC delete.** Deleting an EC rejects its active direction (if any) before emitting the EC-deletion event. No source remains "under" the deleted EC.
**R7 — Auditability.** Source-set changes flow through `DirectionSourceCapabilitiesChanged` events, which carry a timestamp today. Capturing the acting architect on these events is net-new work (see R9).
**R8 — Draft relaxes source cardinality.** While a direction is in `draft`, it may have a single source regardless of type. Full type cardinality (e.g. `consolidate` requires ≥2 sources) is enforced only when advancing the direction from `draft` to `proposed`.
**R9 — Actor capture on source changes.** `DirectionSourceCapabilitiesChanged` must record the acting architect. Neither the event nor the shared `BaseEvent` carries an actor today, so an actor field must be added to satisfy R7.

---

### Test coverage and gates

- [ ] Every BDD scenario in this spec has a corresponding test.
- [ ] Every business rule has a unit test on the aggregate or domain service.

---

## Architecture

### Ownership

The change remains in the **`enterprisearchitecture`** bounded context for the eligibility policy and the included-capabilities read view. The **`architecturedirection`** context is where the actual writes happen — direction capture and source-set edits — because the source set is intrinsic state of the `Direction` aggregate.

Cross-context responsibility split:
- `enterprisearchitecture` owns: EC aggregate (lifecycle, identity, target maturity, strategic importance), the eligibility policy/service, and the read view "what's under this EC" — computed as the subtree of each source minus any most-specific-wins carve-outs (R2).
- `architecturedirection` owns: the Direction aggregate, including its source set. Direction commands consult `InclusionEligibilityService` (a service owned by `enterprisearchitecture` and exposed cross-context) to enforce R1 (same-node exclusivity) at capture and source-change time, and to resolve effective subtree ownership and carve-outs (R2).

This is upstream/downstream in the same direction as before: `enterprisearchitecture` is upstream supplier of EC identity and eligibility; `architecturedirection` is downstream and reacts to EC deletion by rejecting active directions.

### Linking removal

The standalone "linking" concept is removed entirely — directions are now the only way a domain capability becomes associated with an EC. The following must be deleted, not migrated:

- The `EnterpriseCapabilityLink` aggregate and its `EnterpriseCapabilityLinked` / `EnterpriseCapabilityUnlinked` events.
- The link command/query handlers and the `POST` / `DELETE` / `GET` `.../enterprise-capabilities/{id}/links` endpoints.
- The link read model(s) and the database table(s) backing links.
- The frontend linking UI (the linked-capabilities section, its hooks, and any "Manage Links" surface).

Existing link rows are **hard-deleted** via migration — there is no migration of historical links into directions. After this change there is no residual linking data, schema, or code.


## Implementation status

**Frontend slice (done):** Implemented against an MSW stub API whose contract was designed and validated by the api-design-expert. Endpoints stubbed: `GET .../composition`, `GET /capabilities/source-candidates`, `POST .../direction/composition-preview`, `POST .../direction/sources`, `DELETE .../direction/sources/{capabilityId}`, plus capture/transition with R1 (409) and R5 (immutable) enforcement, and the R1/R2 most-specific-wins carve-out resolution. Standalone linking UI and its hooks/api/types were deleted. Built TDD; all unit tests pass; per-file code health 10.0 (stub `store.ts` 9.68 — inherent HATEOAS/DTO builder).

Frontend decisions taken from the mockup (source of truth for the FE):
- Capture modal drops the placements section; uses a search-driven source picker with R1 eligibility + an R2 carve-out preview, and a draft-cardinality hint (R8).
- Source list/placements removed from the Direction panel — sources now surface only in the Included-capabilities composition view; the panel shows type/status/horizon/narrative/actions and an agreed-immutability callout.
- Source-set edits use granular sub-resources (`POST .../sources`, `DELETE .../sources/{id}`); `sourceCapabilityIds` removed from the `PUT .../direction` request.

**Remaining:** Go backend (aggregate, events, read model, migrations hard-deleting link rows, actor on `DirectionSourceCapabilitiesChanged` per R9), Swagger docs.

## Checklist

- [x] Specification ready
- [ ] Implementation done (frontend slice done; backend pending)
- [ ] Unit tests implemented and passing (frontend done; backend pending)
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
