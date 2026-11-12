# 234 — View Membership Integrity

> **Status:** done
> **Depends on:** none
> **Roadmap alignment:** outside roadmap: defect repair inside the Architecture Views context. No bounded-context or capability change. Standing invariant 1 (events-only integration) is respected; the one-time repair migration relies on the spec 209 backfill exception and also appends events to the event store, both of which need explicit human approval (see Design Decision 6).

---

## Problem Statement

A view can show applications, capabilities and origin entities that no longer exist, and applications that the view itself believes it does not contain. A production tenant reviewed on 2026-11-12 had 31 views holding 405 application placements; 53 of those placements (in 11 views) were unknown to the view aggregate, 27 of them pointing at deleted applications. The same tenant had 56 placements of deleted capabilities.

Three defects produce this:

1. **Saving a position creates membership.** Every position endpoint (`PATCH /views/{id}/components/{componentId}/position`, `PATCH /views/{id}/layout`, `PATCH /views/{id}/capabilities/{capabilityId}/position`, `PATCH /views/{id}/origin-entities/{originEntityId}/position`) inserts a placement row when none exists. For applications this writes membership to the placement table without any event, so the table and the view's event history disagree. The audit history of one affected view shows an application added, removed two seconds later, and still stored in the view eight months on.
2. **Application cleanup swallows failures.** When an application is deleted, the views context removes it from each view through the view aggregate, but logs any failure and moves on. A view that is soft-deleted, or a placement the aggregate does not list, makes the removal fail, and the placement stays.
3. **Deleted capabilities and origin entities are never cleaned up.** The views context subscribes to application deletion only. Nothing reacts when a capability, vendor, acquired entity or internal team is deleted.

The effect for users is canvases with ghost nodes, nodes that cannot be removed, and views that reference things that return 404.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect** | A view shows exactly what they placed on it, and removing a node removes it for good |
| **View owner** | Their private view stays consistent when someone else deletes an entity it shows |
| **Admin** | Deleting an entity leaves no trace of it in any view, including private views of other users, and the view's history records the removal |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: View membership integrity

  Scenario: Saving a position does not add an application to a view
    Given a view that does not contain application "Billing"
    When a position is saved for "Billing" in that view
    Then the request is rejected as not found
    And the view still does not contain "Billing"

  Scenario: A layout save skips applications that are not in the view
    Given a view containing application "Orders" but not "Billing"
    When a layout is saved with positions for "Orders" and "Billing"
    Then the position of "Orders" is updated
    And the view still does not contain "Billing"

  Scenario: A late position save does not resurrect a removed application
    Given a view containing application "Billing"
    When "Billing" is removed from the view
    And a position save for "Billing" arrives afterwards
    Then the view does not contain "Billing"

  Scenario: Saving a position does not add a capability or origin entity to a view
    Given a view that does not contain capability "Invoicing"
    When a position is saved for "Invoicing" in that view
    Then the request is rejected as not found
    And the view still does not contain "Invoicing"

  Scenario: Deleting an application removes it from every view and records the removal
    Given application "Billing" is shown in a public view and in another user's private view
    When "Billing" is deleted
    Then neither view contains "Billing"
    And each view's history records "Billing" as removed

  Scenario: Deleting an application shown in a deleted view
    Given application "Billing" is shown in a view that has since been deleted
    When "Billing" is deleted
    Then the deletion succeeds

  Scenario: Deleting a view removes its placements
    Given a view containing application "Orders" and capability "Invoicing"
    When the view is deleted
    Then no placement of that view remains stored

  Scenario: Deleting a capability removes it from every view
    Given capability "Invoicing" is shown in a public view and in another user's private view
    When "Invoicing" is deleted by a user who owns neither view
    Then neither view contains "Invoicing"

  Scenario: Deleting a parent capability removes its deleted children from views
    Given capability "Finance" with child "Invoicing", both shown in a view
    When "Finance" is deleted together with its children
    Then the view contains neither "Finance" nor "Invoicing"

  Scenario Outline: Deleting an origin entity removes it from every view
    Given <kind> "Acme" is shown in another user's private view
    When "Acme" is deleted by a user who does not own that view
    Then the view no longer contains "Acme"

    Examples:
      | kind            |
      | vendor          |
      | acquired entity |
      | internal team   |

  Scenario: Existing placements of deleted entities are repaired on upgrade
    Given a tenant whose views hold placements of deleted applications, capabilities and origin entities
    When the release containing this change is deployed
    Then no view holds a placement of an entity that no longer exists

  Scenario: Deleted applications the history still lists are removed through the history
    Given a view whose history lists application "Billing" as a current member
    And "Billing" has been deleted
    When the release containing this change is deployed
    Then the view's history records "Billing" as removed
    And the view no longer contains "Billing"

  Scenario: Applications that came back after being removed are purged on upgrade
    Given a view that stores a placement for application "Billing"
    And the last membership event for "Billing" in that view's history is its removal
    When the release containing this change is deployed
    Then the view no longer contains "Billing"

  Scenario: Applications that were never added are purged on upgrade
    Given a view that stores a placement for application "Billing"
    And the view's history has no membership event for "Billing"
    When the release containing this change is deployed
    Then the view no longer contains "Billing"

  Scenario: Applications the view legitimately contains survive the upgrade
    Given a view whose history lists existing application "Orders" as a current member
    When the release containing this change is deployed
    Then the view still contains "Orders" at the same position
```

---

## Business Rules & Invariants

1. **The view history decides application membership.** A view contains an application exactly when the view aggregate's history lists it as a current member. The stored application placements are a projection of that history plus a position, and must always match it. Removing an application the aggregate does not list stays a 404; adding one it already lists stays a 409.
2. **The placement table is the primary record for capabilities and origin entities.** They have no membership events; adding, moving and removing them write the placement table directly. This is a deliberate, stated exception, not a read model (Design Decision 3).
3. **Positions never create membership.** A position write for an element that has no placement in the view changes nothing. The single-element endpoints answer 404; the bulk layout endpoint skips the element and still applies the rest.
4. **Deletion cascades to every view.** When an application, capability, vendor, acquired entity or internal team is deleted, it is removed from all non-deleted views of the tenant, private or public, whoever owns them. The cleanup is a consequence of the deletion, authorised by the right to delete the entity; view ownership and view edit permission are not consulted.
5. **Application cleanup goes through the history.** Removing a deleted application from a view raises `ComponentRemovedFromView` on that view, so the history and the placements stay in agreement. Capability and origin-entity cleanup deletes placements directly (rule 2).
6. **Cleanup cannot fail silently.** A failure while removing a deleted entity from a view fails the deletion's transaction; it is never logged and skipped.
7. **Deleted views hold no placements.** Projecting `ViewDeleted` removes all of that view's placements. Cleanup skips deleted views, which have nothing left to clean up.
8. **Removal never cascades.** Removing or cleaning up one element removes only that element's placement; related or orphaned elements stay.
9. **Repair restores agreement and what users asked for.** After the upgrade repair:
   - every non-deleted view's application placements are exactly the existing applications its history lists as current members;
   - for every deleted application its history listed as a current member, the history ends with a `ComponentRemovedFromView`;
   - no placement references a capability or origin entity that no longer exists;
   - no deleted view holds placements;
   - placements of existing applications the history lists, and of existing capabilities and origin entities, keep their position and colour.

---

## Acceptance Criteria

- [x] `PATCH /views/{id}/components/{componentId}/position` returns 404 and stores nothing when the view has no placement for that application
- [x] `PATCH /views/{id}/capabilities/{capabilityId}/position` and `PATCH /views/{id}/origin-entities/{originEntityId}/position` return 404 and store nothing when the view has no placement for that element
- [x] `PATCH /views/{id}/layout` updates existing placements, creates none, and returns 204 when some listed applications are not in the view
- [x] After remove-then-position for the same application, `GET /views/{id}` does not list it
- [x] Deleting an application raises `ComponentRemovedFromView` on every non-deleted view whose history lists it, and removes its placement from all of them, including private views of other users
- [x] Deleting an application that is shown in a soft-deleted view succeeds
- [x] Deleting a view removes all its placements
- [x] Deleting a capability removes its placement from every view, including private views of other users; deleting a parent with its children removes the children's placements too
- [x] Deleting a vendor, an acquired entity or an internal team removes its placement from every view, including private views of other users
- [x] A forced failure in view cleanup rolls back the entity deletion
- [x] After the repair migration, for every non-deleted view the set of application placements equals the set of existing applications the view aggregate, loaded from the event store, lists as members
- [x] After the repair migration, loading any affected view aggregate succeeds, its version equals its event count, and no deleted application is listed as a member
- [x] After the repair migration, no placement references a deleted or non-existent capability, vendor, acquired entity or internal team, and no deleted view holds placements
- [x] The repair migration leaves placements of existing, legitimately contained elements unchanged, including position and colour
- [x] The repair migration is idempotent: a second run inserts no events and deletes no rows
- [x] Adding via the canvas, auto-layout, drag-to-move and dynamic-view draft save still work end to end for all three element types, including a drag immediately after a drop

---

## Architecture

### Ownership

Architecture Views owns the change. Architecture Modeling and Capability Mapping are affected only as publishers: the views context consumes their already-published deletion events (`ApplicationComponentDeleted`, `CapabilityDeleted`, `VendorDeleted`, `AcquiredEntityDeleted`, `InternalTeamDeleted`). No new events, commands or cross-context reads at runtime.

### Domain Model

No new aggregates or events. The `ArchitectureView` aggregate keeps recording application membership through `ComponentAddedToView` and `ComponentRemovedFromView`, and its membership guards stay as they are: they are the gate for application membership. Its refusal to change a deleted view also stays.

Capabilities and origin entities remain placement-only (no aggregate membership), as today.

### API Surface

No new resources. Behaviour changes on existing operations:

| Operation | Today | After |
|-----------|-------|-------|
| Save single element position, element not in view | 204, placement created | 404, nothing stored |
| Save layout with applications not in view | 204, placements created | 204, those applications skipped |

Swagger annotations for the three single-element position endpoints gain the 404 response. The layout endpoint's description states that unknown applications are skipped.

### Persistence

Placements stay in `architectureviews.view_element_positions`. Position writes become update-only (`UPDATE … WHERE` the placement key, 404 when no row matched for the single-element endpoints). Inserts happen only through the add operations. The bulk layout write joins the caller's transaction via `database.WithTx` instead of opening its own.

Projections:

- `ComponentAddedToView` inserts the placement; `ComponentRemovedFromView` deletes it (as today).
- `ViewDeleted` additionally deletes every placement of the view.

Deletion cleanup runs inside the publishing transaction (in-transaction projection) and stays database-only:

- **Application deleted:** for each non-deleted view whose aggregate lists the application, load the view, remove the component and save. The `ComponentRemovedFromView` projection deletes the placement. The views to process are found from the stored placements of non-deleted views; because rule 1 holds after the repair, that set equals the set of views whose history lists the application. Any error is returned, failing the deletion.
- **Capability, vendor, acquired entity or internal team deleted:** delete the placements with that element id and type (`capability` or `origin_entity`) across all views of the tenant. Origin entities are stored without a subtype, keyed by their UUID, so the id alone identifies the placement.

### Repair migration

One migration, named so that it passes the spec 209 cross-schema guard (filename contains `backfill`, DML only). It runs in the migrate init container of the new backend pod, in one transaction together with its `schema_migrations` record. The backend Deployment uses the default `RollingUpdate`, so the previous release's pod keeps serving during the migration and until the new pod is ready. This window is accepted (decided 2026-11-12 on PR #91, instead of a `Recreate` rollout with downtime on every deploy):

- A view write by the old pod that takes the same event version as an appended removal fails one of the two on the unique `(tenant_id, aggregate_id, version)` constraint. If the migration loses, it rolls back as a whole and the init container retries it.
- An old-release position save or failed application cleanup in the window, during or after the migration, can recreate a ghost placement. The migration does not run again, so such a placement stays until a later repair. The window is the length of one rollout, about a minute.

It is tenant-scoped throughout: every join matches `tenant_id` on both sides, and event-store reads match `tenant_id` and `aggregate_id` = view id.

"Latest membership event" for a view and application means the `ComponentAddedToView` or `ComponentRemovedFromView` row in `infrastructure.events` with that view as `aggregate_id`, that application as `event_data->>'componentId'`, and the highest `version`.

"No longer exists" means the owning context has no row for the id, or, for the Architecture Modeling tables, the row has `is_deleted = TRUE`. Capabilities are hard-deleted and checked by absence from `capabilitymapping.capabilities`; applications against `architecturemodeling.application_components`; origin entities against the union of `architecturemodeling.vendors`, `architecturemodeling.acquired_entities` and `architecturemodeling.internal_teams`.

Steps, in order:

1. **Append removals for deleted applications.** For every non-deleted view and application where the latest membership event is an addition and the application no longer exists, insert a `ComponentRemovedFromView` event with `event_data` `{"viewId", "componentId", "removedAt"}`, `version` continuing that view's highest version (several removals for one view get consecutive versions), `occurred_at` now, and the same actor columns as events saved with no actor in context.
2. **Delete application placements the history does not list.** Delete every application placement in a non-deleted view whose latest membership event is a removal (including those appended in step 1) or for which no membership event exists.
2a. **Restore placements the history lists but the table lacks.** For every non-deleted view and application whose latest membership event is an addition and that has no placement, insert the placement at the position carried by that event, as projecting the history would. This is what makes the application placements *equal* the listed members (rule 9), not only a subset.
3. **Delete placements of entities that no longer exist** for capabilities and origin entities.
4. **Delete all placements of deleted views**, and of view ids the views context has no record of.

`is_deleted` is compared with `IS NOT TRUE` throughout, so a row with a NULL flag counts as existing.

Step 1 is idempotent because the appended removal becomes the latest event; steps 2 to 4 are idempotent deletes. The migration contains no DDL.

### Frontend

- A single-element position 404 is a no-op, not an error toast.
- A position save for an element whose add request is still in flight waits for that add to finish, so a drag immediately after a drop keeps its position.
- Auto-layout tolerates the layout endpoint skipping elements.
- The dynamic-view draft save and the canvas keep treating add-409 and remove-404 as idempotent success.

No new UI.

### Cross-Context Integration

Subscriptions only, to events already in the published language of Architecture Modeling and Capability Mapping. The views context does not query other contexts at runtime to decide whether an element exists.

---

## Design Decisions

1. **Update-only position writes, not client-side fixes.** The canvas has several position writers (drag, auto-layout, draft save) and any of them can race a removal or run against a stale node set. Making the server refuse to create membership from a position closes every such path at once. Rejected: debouncing or cancelling in-flight position saves in the frontend, which leaves the API unsafe for other clients.
2. **The event history stays authoritative for application membership.** All 53 ghost placements in the reviewed tenant were cases where the history was right (the user removed the application) and the placement table was wrong (a position save put it back). Fixing the writes that bypass the history, and cleaning up through the aggregate, keeps the placement table a faithful projection of the ledger, so a rebuild from events gives the same application membership. Rejected: letting the stored placement decide removal and cleanup, and deleting application placements by SQL on deletion, which both make the ledger and the views disagree and would bring deleted applications back on any rebuild.
3. **Capabilities and origin entities stay event-free, and the spec says so.** Their membership, and every element's position and colour, have never been in the ledger; the placement table is their primary record, and cleaning them up by direct delete loses nothing a rebuild could have restored. Making all of view state rebuildable needs membership and position events for every element type, which is a modelling change well beyond this defect. Rejected for this spec; recorded as a roadmap gap: the Architecture Views context is not fully rebuildable from events.
4. **Deleted views drop their placements on `ViewDeleted`.** The aggregate rightly refuses changes to a deleted view, so cleanup through it would fail and, under Design Decision 7, block the entity's deletion. Clearing placements as part of the view's own deletion projection makes deleted views irrelevant to cleanup and keeps a rebuild consistent. Rejected: allowing membership changes on deleted views, which weakens an aggregate invariant for data nobody can see.
5. **Purge application placements the history does not list.** Every one of the 53 such placements in the reviewed tenant had a removal as its last membership event; purging them restores what the user asked for and what the ledger already says. Rejected: keeping them because users currently see them, which preserves the defect's output.
6. **Repair by a one-time migration that appends events.** Finding placements of non-existent entities needs the owning contexts' tables, which only a migration may read (spec 209 backfill exception, DML only). Deleted applications that the history still lists as members cannot be fixed by deleting placements alone without breaking rule 1, so the migration appends the `ComponentRemovedFromView` events the failed cleanup should have produced. The event shape is pinned above, and an integration test loads every affected aggregate after the migration. This uses the backfill exception to delete rows and to write to the event store, so it needs explicit human approval. Rejected: replaying deletion events from the event store into the views context at startup, or a startup repair job, both of which add permanent machinery for a one-off repair.
7. **Cleanup failures fail the deletion.** Projectors run in the write transaction; a cleanup that cannot complete must roll the deletion back so the model and the views never disagree. Rejected: log-and-continue, which is the current behaviour and the reason deleted applications remain in views.

---

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
