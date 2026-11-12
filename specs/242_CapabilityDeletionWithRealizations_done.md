# 242 — Capability Deletion With Realizations

> **Status:** done
> **Depends on:** none
> **Roadmap alignment:** outside roadmap: defect repair inside the Capability Mapping context and its delete dialog. No bounded-context or capability change and no new events or commands. Respects the standing events-only integration invariant: deleting a capability never deletes applications (Architecture Modeling owns them; the cross-context deletion was removed on 2026-05-25, `c2bc39c4`).

---

## Problem Statement

Deleting a capability that applications realize is broken in three ways.

1. **The confirmed delete fails.** Cascade-deleting a capability fails with `404 Failed to delete capability` whenever an application realizes any capability below it. Realizations are inherited upwards automatically, so this covers almost every real subtree. The cascade deletes every realization stored on the capabilities in scope, inherited ones included. An inherited realization is a projection of a direct realization on a descendant and has no realization behind it, so deleting it fails as not found and the whole cascade rolls back. Found on 2026-11-12 while verifying spec 234. The cascade had no enclosing transaction, so a failed cascade left the realizations it deleted before the 404 deleted (see Persistence).
2. **The delete dialog misinforms.** It offers "Also delete N application realizations that only realise these capabilities". The backend has ignored that flag since 2026-05-25: the realizations are always removed and the applications never are, ticked or not. Its counts also include inherited realization rows, so one application realizing one child is counted twice.
3. **The canvas deletes without warning.** Deleting a capability from the model on the canvas skips the impact confirmation: a realized capability loses its realizations silently, and a capability with children fails with a conflict instead of offering the cascade.

Decision (2026-11-12): warn and confirm. The user sees which applications will stop realising which capabilities and confirms; the delete then removes those realizations and keeps the applications. Deletion is not blocked while realizations exist.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect** | Before deleting a capability, see exactly which applications will stop realising it or its descendants |
| **Architect** | Once confirmed, the capability, its descendants and their realizations are removed in one step, and the applications remain |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Deleting capabilities that applications realize

  Scenario: The confirmation lists the realizations that will be removed
    Given capability "Finance" with child "Invoicing"
    And application "Billing" realizes "Invoicing"
    When the user starts deleting "Finance"
    Then the confirmation says "Finance" and 1 child capability will be deleted
    And it lists "Billing" as stopping to realise "Invoicing"
    And it says the applications themselves are kept
    And it offers no option to delete applications

  Scenario: Each realization is listed once
    Given capability "Finance" with child "Accounting" with child "Invoicing"
    And application "Billing" realizes "Invoicing"
    When the user starts deleting "Finance"
    Then the confirmation lists exactly one realization

  Scenario: Confirming deletes the subtree and its realizations
    Given capability "Finance" with child "Accounting" with child "Invoicing"
    And application "Billing" realizes "Invoicing"
    And application "Ledger" realizes "Accounting"
    When the user confirms deleting "Finance"
    Then the deletion succeeds
    And no realization of "Finance", "Accounting" or "Invoicing" remains, direct or inherited
    And "Billing" and "Ledger" still exist

  Scenario: Ancestors outside the deleted subtree lose the inherited realization
    Given capability "Operations" with child "Finance" with child "Invoicing"
    And application "Billing" realizes "Invoicing"
    When the user confirms deleting "Finance"
    Then "Operations" no longer shows "Billing" as an inherited realization

  Scenario: Ancestors keep realizations sourced outside the deleted subtree
    Given capability "Operations" with children "Finance" and "Logistics"
    And application "Billing" realizes a child of "Finance"
    And application "Fleet" realizes "Logistics"
    When the user confirms deleting "Finance"
    Then "Operations" still shows "Fleet" as an inherited realization
    And "Logistics" still shows "Fleet" as a direct realization

  Scenario: Deleting a realized leaf
    Given capability "Invoicing" without children
    And application "Billing" realizes "Invoicing"
    When the user starts deleting "Invoicing"
    Then the confirmation lists "Billing" as stopping to realise "Invoicing"
    When the user confirms
    Then the deletion succeeds and "Billing" still exists

  Scenario: Deleting a capability from the canvas warns the same way
    Given capability "Finance" with child "Invoicing" shown on a view
    And application "Billing" realizes "Invoicing"
    When the user deletes "Finance" from the model on the canvas
    Then the same confirmation as in the capability tree is shown
    And confirming deletes "Finance", "Invoicing" and the realization
```

---

## Business Rules & Invariants

1. **Only direct realizations are deleted.** A cascade deletes the direct realizations of every capability in scope. Inherited realizations are never targeted; they disappear with the direct realization they were inherited from.
2. **Every inherited realization in scope has its source in scope.** Realizations are inherited from descendants to ancestors only, so deleting the subtree's direct realizations removes every realization in scope.
3. **Inherited realizations outside the subtree are cleaned up too.** Deleting a direct realization removes the realizations inherited from it on every ancestor (existing behaviour of `SystemRealizationDeleted`); inherited realizations from other sources stay. An ancestor that the same application still realizes through a surviving branch keeps its inherited realization, re-derived from the surviving source. Reparenting a capability follows the same rule.
4. **Deleting a capability never deletes an application.** Applications belong to Architecture Modeling; Capability Mapping only removes its own realizations.
5. **The impact shown is the impact executed.** The confirmation lists exactly the direct realizations the delete will remove, each once, and nothing the delete will not do. The delete request carries the realization IDs the user confirmed; if the realizations in scope differ when the delete runs, it is rejected with 409 and nothing is deleted.
6. **Every capability delete is confirmed with its impact.** The capability tree, business domain page and canvas all use the same impact confirmation, and so do the multi-select "Delete from Model" actions in the tree and on the canvas. A selected capability already inside another selected capability's cascade is deleted by that cascade, and its realizations are listed once. A multi-selection lists the realizations of every selected capability. Confirmation stays disabled until every impact has loaded; a failed load offers a retry.
7. **The cascade is all-or-nothing.** The whole cascade runs in one transaction, so any failure rolls back every delete it made. Before this spec each aggregate save committed on its own.

---

## Acceptance Criteria

- [x] `DELETE /capabilities/{id}` with `cascade: true` returns 204 for a capability whose descendants carry direct realizations, at one and at several levels
- [x] After such a delete, no realization row (direct or inherited) references a deleted capability; inherited realizations on outside ancestors sourced from the subtree are gone, those sourced elsewhere remain
- [x] Each deleted direct realization raises `SystemRealizationDeleted`; no command is dispatched for an inherited realization; no application is deleted
- [x] `GET /capabilities/{id}/delete-impact` returns the direct realizations in scope as one list, each with application and capability, and no inherited realizations
- [x] The delete request no longer documents or accepts a `deleteRealisingApplications` option; the frontend no longer sends it
- [x] The delete dialog shows the realizations that will be removed, states that applications are kept, and has no checkbox
- [x] Deleting a capability from the model on the canvas opens the same delete dialog, including the cascade for capabilities with children
- [x] Deleting a realized leaf capability still succeeds and removes its realization
- [x] When the same application realizes a deleted branch and a surviving branch, the common ancestor keeps its inherited realization, sourced from the surviving branch (also after reparenting)
- [x] A cascade that fails after earlier deletes leaves every realization and capability in place
- [x] `delete-impact` returns `"realizations": []`, never `null`, when nothing is realized
- [x] Deleting several selected capabilities lists the realizations of all of them; confirmation stays disabled until every impact has loaded, and a failed load offers a retry
- [x] Deleting a capability, linking a system and deleting a realization refresh application and capability one-pager completeness
- [x] A delete whose realizations changed after the impact was confirmed is rejected with 409, deletes nothing, and the dialog shows the refreshed impact
- [x] The tree and canvas multi-select "Delete from Model" confirmations list the realizations that will be removed and delete capabilities through the cascade with the confirmed realization IDs
- [x] Confirmation stays disabled while an impact is refreshing

---

## Architecture

### Ownership

Capability Mapping owns the change, backend and frontend. No other context is affected: consumers of `SystemRealizationDeleted` and `CapabilityDeleted` receive the same events a successful cascade produces today when its subtree has no inherited realizations.

### Domain Model

No new aggregates, events or commands. The cascade-delete handler dispatches `DeleteSystemRealization` only for realizations whose origin is `Direct`. Realization inheritance and its cleanup are unchanged. Because the direct realizations are deleted before the capabilities, the per-capability `CapabilityRealizationsUninherited` computation in the cascade could never find anything and was removed. The cascade now runs inside `TenantAwareDB.RunInTx`; the event store, projectors and read models join that transaction. The realization projection kept one inherited row per capability and application, so deleting or uninheriting its source dropped the row even when another branch still realized the application; after removing inherited rows, the projection now re-derives them from the application's remaining direct realizations (`RestoreInheritedRealizations`). The non-cascade `DeleteCapability` command was registered but never dispatched (the API always sends `CascadeDeleteCapability`); it was removed together with its handler, `CapabilityDeletionService` and the children-checker adapter.

### API Surface

- `DELETE /capabilities/{id}`: a cascade that failed with 404 now succeeds with 204. The Swagger description drops the `deleteRealisingApplications` option. The body adds `confirmedRealizationIds`: the realization IDs from the delete impact the user confirmed. A mismatch returns 409 with an `x-delete-impact` link.
- `GET /capabilities/{id}/delete-impact`: the split into `realizationsOnDeletedCapabilities` and `realizationsOnRetainedCapabilities` is replaced by one list of the direct realizations the delete will remove. The split classified applications by whether they could be deleted, which no longer happens. The delete dialog is the only consumer.

### Persistence

No schema change and no migration. Failed cascades were not rolled back before this spec: the direct realizations deleted before the 404 stayed deleted while the capabilities remained. That data cannot be repaired, because nothing records which realizations a failed cascade removed, and the user had asked for them to be removed anyway.

### Frontend

- The delete dialog lists the realizations to be removed (application and capability), states that the applications are kept, and drops the checkbox. The delete request and the cascade-delete mutation effects drop the application-deletion flag.
- The canvas "delete from model" action for a capability opens the delete dialog instead of deleting directly.
- Cache invalidation after a capability delete now always refreshes application one-pager completeness (an application can lose its last realized capability); linking a system and deleting a realization refresh application and capability completeness too. Views were already refreshed by spec 234.
- Every capability delete goes through one deletion plan (`useCapabilityDeletionPlan`): it loads the impact of each selected capability, leaves capabilities covered by a selected ancestor to that cascade, de-duplicates realizations, and sends each delete with its confirmed realization IDs. The delete dialog and the tree and canvas bulk confirmations share it and the realization list. The plain `useDeleteCapability` path, its API call and its mutation effects were removed as dead code.

### Cross-Context Integration

None beyond the existing published events.

---

## Design Decisions

1. **Warn and confirm, do not block.** The user sees the realizations that will be removed and confirms; removing a realization is reversible by re-linking. Rejected: refusing deletion while realizations exist, which forces unlinking realizations one by one before a subtree can go and changes the existing dialog's behaviour.
2. **Filter to direct realizations in the cascade, rather than make inherited deletes a no-op.** Inherited rows are projections and must never be targeted by a command. Rejected: treating "realization not found" as success in `DeleteSystemRealization`, which would also hide genuine not-found errors on the single-realization endpoint.
3. **Rely on `SystemRealizationDeleted` to remove inherited rows.** That projection already deletes every inherited row by source realization, on all ancestors. Rejected: deleting inherited rows explicitly in the cascade, which duplicates the projection.
4. **Drop the application-deletion option instead of restoring it.** Deleting applications from Capability Mapping was removed as a cross-context command. Restoring it would need Architecture Modeling to decide on its own whether to delete applications on `CapabilityDeleted`, a separate product decision. Rejected for this spec.
. **Re-derive inheritance in the projection after removals.** The read model keeps one inherited row per capability and application, so the projection re-derives missing rows from the remaining direct realizations instead of the write side computing replacements. Rejected: storing one inherited row per source, which changes the read model's unique key and every consumer that assumes one row per application.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Filter by origin in the handler | The handler depends on the read model's origin field | Unit test with mixed origins; integration test on a real hierarchy |
| Change the delete-impact response shape | Breaking change for any external client | The delete dialog is the only consumer; OpenAPI regenerated |
| Canvas delete opens the dialog | One more step on the canvas | Same step as everywhere else; it shows real consequences |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
