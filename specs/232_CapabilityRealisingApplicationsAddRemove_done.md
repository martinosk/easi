# 232 — Capability Details: Add and Remove Realising Applications

> **Status:** done
> **Depends on:** 026 (capability system realization), 131B (realization inheritance read model), 220 (capability details edited in place)
> **Roadmap alignment:** outside roadmap: presentation affordance on the existing capability realization; no domain change. Respects SD6's principle that every affordance is a HATEOAS link.

---

## Problem Statement

Clicking a capability on the Business Domains board opens the capability drawer. Its "realising applications" section lists the applications that realise the capability, but offers no way to change that list. Today a realization can only be created by drawing an edge on the Architecture Canvas or by importing, and removed only from the canvas. A user curating a domain must leave the board, find both elements on a view, and draw or delete an edge.

The backend already accepts `POST /capabilities/{id}/systems` and `DELETE /capability-realizations/{id}`, but the capability resource does not advertise the create operation as a link, so the frontend cannot gate an "add" control by HATEOAS. The realization resource also advertises `edit` and `delete` on inherited realizations, which are read-model projections that cannot be deleted directly.

This spec adds in-place add and remove of realising applications to the shared realising-applications section, gated by links.

---

## User Personas

| Persona                                                                    | Needs                                                                             |
| -------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| **Architecture maintainer** (`capabilities:write` / `capabilities:delete`) | Map or unmap an application against a capability without leaving the domain board |
| **Invited editor (190 grantee) / read-only viewer**                        | A clean list with no add or remove controls they cannot use                       |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Add and remove realising applications in capability details

  Scenario: Add a realising application
    Given the capability "Order Management" carries an "x-add-realization" link
    And "Order Management" is opened in the Business Domains capability drawer
    When I choose "Add application", pick "SAP ERP", choose realization level "Partial" and confirm
    Then "SAP ERP" appears under realising applications with level "Partial"
    And the domain board card for "Order Management" shows the "SAP ERP" chip

  Scenario: The picker offers only applications not already directly realising the capability
    Given "CRM" directly realises "Order Management"
    When I open the add-application picker
    Then "CRM" is not offered

  Scenario: Create a new application and add it as realising
    Given the capability "Order Management" carries an "x-add-realization" link
    And its "x-related" links contain a POST "capability-realization" entry
    When I choose "Add application", choose realization level "Planned", choose "Create new application", enter the name "Fleet Planner" and confirm
    Then the application "Fleet Planner" is created
    And "Fleet Planner" appears under realising applications with level "Planned"

  Scenario: No create-new control without the related link
    Given the capability carries "x-add-realization" but no POST "capability-realization" entry in "x-related"
    When I open the add-application dialog
    Then no "Create new application" control is shown

  Scenario: No add control without the link
    Given the capability resource carries no "x-add-realization" link
    When I open its details
    Then no "Add application" control is shown

  Scenario: Remove a direct realising application
    Given "CRM" directly realises "Order Management" and that realization carries a "delete" link
    When I choose remove on "CRM" and confirm
    Then "CRM" disappears from realising applications
    And the domain board card for "Order Management" no longer shows the "CRM" chip

  Scenario: Inherited realisations cannot be removed here
    Given "Billing" realises "Order Management" only by inheritance from a child capability
    When I open the details of "Order Management"
    Then "Billing" is listed as inherited with no remove control

  Scenario: No remove control without the link
    Given a direct realization carries no "delete" link
    Then no remove control is shown on it

  Scenario: The behaviour is the same on every capability surface
    When I open the same capability on the Architecture Canvas or the one-pager subject drawer
    Then the same add and remove controls are shown under the same link rules
```

---

## Business Rules & Invariants

1. **Add is link-gated** — the capability resource carries `x-add-realization` (`POST /capabilities/{id}/systems`) iff the actor has `capabilities:write`, matching the route's permission.
2. **Remove is link-gated** — a realization carries `delete` iff it is `Direct` and the actor has `capabilities:delete`.
3. **Inherited realizations are read-only** — an `Inherited` realization carries neither `edit` nor `delete`.
4. **No duplicate direct realizations from the picker** — the picker excludes applications that already directly realise the capability.
5. **Level is required** — an add always sends one of `Full`, `Partial`, `Planned`; default `Full`.
6. **Remove is confirmed** — removal asks for confirmation before calling `delete`.
7. **Create-new is link-gated** — the add dialog offers "Create new application" iff the capability's `x-related` contains a POST entry with relation type `capability-realization` (today: `components:write` and `capabilities:write`).
8. **Create-new links with the chosen level** — the newly created application is linked via `x-add-realization` with the level selected in the add dialog.

---

## Acceptance Criteria

- [x] `GET /capabilities/{id}` includes `x-add-realization` (POST `/api/v1/capabilities/{id}/systems`) for an actor with `capabilities:write`, and omits it otherwise (including edit-grant holders without the permission)
- [x] Realization links for an `Inherited` realization contain no `edit` or `delete`; for a `Direct` one they follow the existing permission rules — on every endpoint returning realizations
- [x] The realising-applications section shows "Add application" only when the capability has `x-add-realization`; adding posts to that link's href and the new realization appears without reload
- [x] The picker lists all applications minus those directly realising the capability, and requires a realization level
- [x] Each realization row shows a remove control only when it has a `delete` link; confirming deletes via that link and the row disappears without reload
- [x] After add or remove, the domain board's realization chips update (cache invalidation covers the per-capability, per-component, all-realizations and business-domain realization queries)
- [x] The add dialog shows "Create new application" only when the capability's `x-related` has a POST `capability-realization` entry; creating an application links it with the chosen level and it appears without reload
- [x] The empty state ("no realising application mapped") still shows the add control when the link is present
- [x] Swagger docs regenerated

---

## Architecture

### Ownership

Capability Mapping owns realizations. No other context changes.

### Domain Model

No change. `LinkSystemToCapability` and `DeleteSystemRealization` commands, the `CapabilityRealization` aggregate and inheritance propagation are reused as is.

### API Surface

- Capability resource: new link `x-add-realization` → `POST /capabilities/{id}/systems`, added under `capabilities:write` (not under edit grants, because the route requires `capabilities:write`).
- Realization resource: `RealizationLinksForActor` takes the realization origin; `edit`/`delete` emitted only for `Direct`.
- No new endpoints.

### Persistence

None.

### Frontend

- `CapabilityRealizationsSection` (shared by the canvas details pane, the Business Domains drawer and the one-pager subject drawer) gains an "Add application" control and per-row remove control.
- New `capabilitiesApi.addRealization(capability, request)` / `useAddRealization` follow the capability's `x-add-realization` link. The canvas edge-drawing and x-related flows keep `capabilitiesApi.linkSystem` by capability ID, because they hold only the ID (and, for x-related, the capability may have just been created), so rewriting them is left out of this spec.
- Add dialog: Mantine `Modal` with a searchable `Select` of applications (`useComponents`) and a realization-level `Select`.
- Add dialog's "Create new application" opens the existing `CreateComponentDialog`; its `onCreated` hand-off calls `useAddRealization` with the selected level, then both dialogs close. Gated by `getPostableRelated(capability)` containing relation type `capability-realization`.
- Remove uses the existing `useDeleteRealization` with a confirmation.
- Mutation effects of `useLinkSystemToCapability` / `useAddRealization` / `useDeleteRealization` audited so the domain board (business-domain realizations query) refreshes.

### Cross-Context Integration

None beyond existing events (`SystemLinkedToCapability`, `SystemRealizationDeleted`), which already drive downstream projections.

---

## Design Decisions

1. **Controls live in the shared section** — spec 220 requires identical sections on all capability surfaces, so the add/remove controls appear everywhere the section renders. Alternative: Business Domains drawer only (rejected: breaks 220's single-surface rule).
2. **New `x-add-realization` link on the capability** rather than a `create` link on the `/systems` collection — the drawer already holds the capability resource; fetching the collection's links would add a round trip-dependency. Alternative: collection `create` link (rejected for that reason).
3. **Hide `edit`/`delete` on inherited realizations** — the links currently advertise operations that cannot succeed. Alternative: keep the links and filter by `origin` in the UI (rejected: UI would second-guess HATEOAS).
4. **Exclude existing direct realisers in the picker** rather than adding a domain uniqueness rule — keeps the change presentation-only. Alternative: domain invariant rejecting duplicates (rejected here; would be a behaviour change to imports and canvas, warrants its own spec).
5. **Editing realization level/notes is not part of this spec** — the user asked for add/remove only.
6. **`x-add-realization` instead of reusing the canvas `x-related` "Component (realization)" entry** — that entry targets `POST /components` (create a new application, then link it) and requires both `components:write` and `capabilities:write`; linking an existing application needs only `capabilities:write` and targets `POST /capabilities/{id}/systems`.
7. **Create-new reuses the `x-related` "Component (realization)" entry** — that entry already advertises exactly "create an application related to this capability as realization" with the right permission pair; reusing it keeps the canvas and the drawer under the same rule. Alternative: a new capability link (rejected: duplicates an existing affordance).

---

## Trade-offs

| Decision                        | Trade-off                                                          | Mitigation                                                         |
| ------------------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------ |
| Picker-side duplicate exclusion | API can still create duplicates via other clients                  | Existing behaviour; unchanged risk                                 |
| Controls on all three surfaces  | Canvas users get a second way to create realizations besides edges | Same command and events; canvas edges refresh from the same caches |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
