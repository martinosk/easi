# 165 — Create Related Entity from Canvas Handle

> **Status:** pending
> **Depends on:** —
> **Coordinates with:** [164 — Dynamic View Mode](164_DynamicViewMode_ongoing.md) (in flight). The handle-click flow must behave correctly in both regular and dynamic modes; see Cross-Context Integration below.
> **Related:** [039 — Capability Dependencies on Canvas](039_Capability_Dependencies_Canvas_pending.md). When 039 lands, capability→capability dependency creation surfaces automatically through the same HATEOAS mechanism without further changes here.

---

## Problem Statement

Today, an architect who wants to add a related entity to a canvas view must leave the canvas, find the right `+` button on the treeview sidebar, fill out a context-agnostic create dialog, then go back and connect the new entity to its source node. The dialog does not know which entity types are valid in relation to the source, does not pre-link the new entity, and does not place it near the source on the canvas. The architect ends up doing four steps (open dialog, fill form, connect edge, drag into place) for what is conceptually one action: "create a related X off this node".

The architecture canvas already renders React Flow handles on every node side (Top / Left / Right / Bottom), and these handles already accept drag-to-existing-node connections to create relations. They are the natural anchor for the inverse flow: starting at the source, declaring the relation, and producing a brand-new related entity in one gesture.

This spec adds a **click-on-handle** entry point on every canvas-renderable entity type. Clicking a free handle (not dragging it onto an existing node) opens a small picker offering each valid related entity type. Selecting a type opens the **existing full create dialog** for that entity type; on submit, the new entity is linked to the source via the appropriate relation, placed at an offset from the clicked handle, and added to the current view. The treeview `+` button is unchanged and remains the entry point for entities that have no canvas anchor yet.

**Validity is HATEOAS-driven.** The backend must expose, on every canvas-renderable entity, a single `x-related` link relation in `_links` — an array of link objects, one per available relation. Each entry declares the HTTP `methods` it supports. The picker filters to entries that advertise `POST` (i.e., creating a new related entity). `GET` (listing existing related entities) is allowed on the same entry and reserved for future use; this spec does not require it. The frontend drives the picker exclusively from these entries — no client-side validity tables, no duplication of domain rules. If a relation type is added to the domain model later, it surfaces in the picker automatically once the backend includes a `POST`-capable entry for it.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise architect** | Quickly extend an architecture view by adding related entities directly off existing nodes, without leaving the canvas |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Create Related Entity from Canvas Handle

  Scenario: Click a handle on a Component to create a related Component
    Given I am viewing the architecture canvas
    And the canvas contains a Component "Order Service"
    And the response for "Order Service" includes an "x-related" entry for component-relation advertising the POST method
    When I click the right handle on "Order Service" without dragging
    Then a picker appears next to the handle listing every "x-related" entry advertising POST
    When I select "Component (related)"
    Then the existing Create Component dialog opens
    When I fill in the dialog and submit
    Then a new Component is created
    And a Component-to-Component relation links "Order Service" to the new Component
    And the new Component is added to the current view, positioned to the right of "Order Service"
    And both the new node and the relation edge are visible without page reload

  Scenario: Click a handle on a Capability to create a child Capability
    Given the canvas contains a Capability "Customer Management" at level L1
    And the "x-related" array on the response includes POST entries for capability-parent and capability-realization, and (post-039) capability-requires / capability-enables / capability-supports
    When I click a handle on "Customer Management" and select "Capability (child of)"
    Then the existing Create Capability dialog opens with level pre-filled to L2
    When I fill in the dialog and submit
    Then a new Capability is created at level L2
    And a parent edge connects "Customer Management" to the new Capability
    And the new Capability appears on the canvas

  Scenario: Click a handle on a Capability to create a realizing Component
    Given the canvas contains a Capability "Order Fulfillment"
    When I click a handle on "Order Fulfillment" and select "Component (realization)"
    Then the existing Create Component dialog opens
    When I fill in the dialog and submit
    Then a new Component is created
    And a realization edge connects "Order Fulfillment" to the new Component
    And the new Component appears on the canvas

  Scenario: Click a handle on an Acquired Entity offers only Component
    Given the canvas contains an Acquired Entity "Acme Corp"
    When I click a handle on "Acme Corp"
    Then the picker offers "Component (acquired-via)" as the only related entity type
    When I select it and submit the Create Component dialog
    Then a new Component is created with an "AcquiredVia" origin relation to "Acme Corp"

  Scenario: Click a handle on a Vendor offers only Component
    Given the canvas contains a Vendor "Acme Vendor"
    When I click a handle on "Acme Vendor"
    Then the picker offers "Component (purchased-from)" as the only related entity type
    When I select it and submit the Create Component dialog
    Then a new Component is created with a "PurchasedFrom" origin relation to "Acme Vendor"

  Scenario: Click a handle on an Internal Team offers only Component
    Given the canvas contains an Internal Team "Platform Team"
    When I click a handle on "Platform Team"
    Then the picker offers "Component (built-by)" as the only related entity type
    When I select it and submit the Create Component dialog
    Then a new Component is created with a "BuiltBy" origin relation to "Platform Team"

  Scenario: Picker shows only what the backend advertises
    Given I am a viewer with read-only permission on the current entity
    When I click a handle on a node
    Then the "x-related" array contains no entries advertising POST for that entity
    And no picker opens

  Scenario: Read-only relations do not show in the picker
    Given an "x-related" entry advertises only the GET method (the user can list, not create)
    When I click a handle on the entity
    Then the picker does not include that entry
    And only entries advertising POST appear

  Scenario: Picker offerings change when the backend adds new relation types
    Given the backend adds a new POST-capable entry to the Component "x-related" array
    When I click a handle on a Component
    Then the new entry appears in the picker without any frontend code change

  Scenario: Cancel the picker without creating anything
    Given the picker is open after clicking a handle
    When I press Escape, click outside the picker, or click "Cancel"
    Then the picker closes
    And no entity, relation, or view change is made

  Scenario: Cancel the create dialog after selecting a target type
    Given the picker is open and I select "Component (related)"
    And the existing Create Component dialog is open
    When I close the dialog without submitting
    Then no entity is created
    And no relation is created
    And the canvas is unchanged

  Scenario: Drag-to-existing-node behavior is preserved
    Given the canvas contains "Order Service" and "Payment Service"
    When I drag from a handle on "Order Service" onto "Payment Service"
    Then a Component-to-Component relation is created using the existing connection flow
    And no picker appears

  Scenario: Click vs drag is disambiguated by movement, not by mouse button
    Given I press down on a handle on "Order Service"
    When I release the mouse without moving more than a small threshold
    Then this is treated as a click and the picker opens
    When I press down on the same handle and drag onto another node before releasing
    Then this is treated as a drag and the existing connection flow runs

  Scenario: Creating from a handle while in Dynamic View Mode adds to the draft
    Given I am in Dynamic View Mode (per spec 164) with unsaved changes
    When I click a handle, select a target type, fill the dialog, and submit
    Then the new entity, the new relation, and the new position appear in the draft only
    And no backend mutations occur until I click "Save view"
    And discarding the draft removes the new entity and its relation from the canvas

  Scenario: Entity-create failure leaves the canvas unchanged
    Given I have submitted the create dialog from a handle-driven flow
    When the backend rejects the entity creation
    Then the dialog shows an inline error
    And no relation is created
    And the canvas is unchanged

  Scenario: Relation-create failure surfaces an actionable error but keeps the entity
    Given the entity is created successfully
    When the relation-create call fails
    Then the user is shown an actionable error naming the relation that failed
    And the just-created entity remains on the canvas as an orphan
    And the user can retry the relation manually (e.g., by drag-connecting from the source handle to the orphan)
```

---

## Business Rules & Invariants

1. **HATEOAS is the single source of truth for picker contents** — The frontend MUST build the picker exclusively from the source entity's `x-related` array, filtered to entries whose `methods` array includes `POST`. The frontend MUST NOT consult `CONNECTION_TYPE_MAP` or any other client-side validity table when populating the picker.
2. **Related link shape** — `_links["x-related"]` is an **array** of link objects. Each entry MUST carry: `href` (where the create dialog will POST), `methods` (the HTTP methods supported on this relation; at minimum `POST` for picker eligibility), `title` (picker label), `targetType` (which create dialog to open: `component` / `capability` / `acquiredEntity` / `vendor` / `internalTeam`), and `relationType` (a stable string identifying which existing relation endpoint the orchestrator will call after the entity is created). The link rel names the resource (related entities), not the action — affordances are signaled by `methods`.
3. **GET in the same entry is reserved for future use** — Backends MAY include `GET` in `methods` to advertise listing existing related entities. This spec does not require any new GET endpoints; it only requires that the frontend correctly ignore `GET`-only entries when building the picker.
4. **Permission and invariant gating live in the backend** — A `POST` affordance is only advertised when the calling user has both (a) permission to create the target entity type and (b) permission to mutate the source entity / current view. Domain invariants that exclude a relation (e.g., a Capability at L4 cannot have child capabilities) MUST be enforced by the backend omitting `POST` from `methods` (or omitting the entry entirely), not by client-side checks.
5. **Click vs drag disambiguation** — A handle interaction is treated as a click only if the pointer moves less than a small pixel threshold between mousedown and mouseup. Above the threshold, the existing drag-connect flow runs and the picker does not open.
6. **All entity types support click-create** — Every node kind rendered on the canvas (Component, Capability, AcquiredEntity, Vendor, InternalTeam) MUST support handle-click. If the `x-related` array contains zero `POST`-capable entries for the source entity, the click is a no-op (no picker opens).
7. **Reuse existing create dialogs** — Selecting an entry in the picker MUST open the same full create dialog used by the treeview `+` button (`CreateComponentDialog`, `CreateCapabilityDialog`, `CreateAcquiredEntityDialog`, `CreateVendorDialog`, `CreateInternalTeamDialog`). No new dialog or quick-create form is introduced. Dialogs MAY be parameterized with pre-fill (e.g., capability level for a parent relation) so domain invariants are respected, but their UX is unchanged.
8. **Source–target ordering matches the existing connection flow** — For asymmetric relations, the new entity occupies whichever role the existing `onConnectHandler` would assign when dragging from the source handle to a target node. The `relationType` carried in the link tells the orchestrator which existing endpoint to call after the entity is created.
9. **Position of the new entity** — The new entity is positioned at a deterministic offset from the source node in the direction of the clicked handle (Top → above, Right → right of, Bottom → below, Left → left of). The offset MUST NOT overlap the source node and MUST be deterministic so auto-layout (spec 124) can re-arrange later.
10. **Add-to-view is implicit** — The new entity is added to the current view as part of the same flow. The user does not need a separate "add to view" step.
11. **One vertical slice per click** — A single handle-click produces at most one new entity and one new relation. Bulk creation is out of scope; users wanting to add many should use Dynamic View Mode (164).
12. **Dynamic mode draft semantics** — When Dynamic View Mode is active, handle-driven creation goes into the draft (no backend calls until Save), reusing the same draft store that drag-to-canvas additions use.
13. **Failure surfacing without rollback** — In regular mode, if entity creation succeeds but relation creation fails, the user is shown an actionable error naming the failed relation, and the just-created entity is **kept** on the canvas. The user can manually create the relation afterwards (e.g., by drag-connecting from the source handle to the orphan). The orchestrator MUST NOT silently swallow the failure or auto-delete the entity.
14. **The treeview `+` button is unchanged** — This spec only adds a new entry point; it does not modify or remove the treeview create flow.
15. **Discoverability is cursor-only** — No tooltip, badge, or hover indicator is added to the handle. The cursor changes to indicate clickability when over a handle. This minimises canvas chrome.

---

## Acceptance Criteria

- [x] Backend exposes `_links["x-related"]` as an array on every canvas-renderable entity (Component, Capability, AcquiredEntity, Vendor, InternalTeam), with one entry per valid relation; entries advertise `POST` in `methods` when the user is authorized to create that related entity
- [x] Each `x-related` array entry carries `href`, `methods`, `title`, `targetType`, and `relationType`
- [x] Backend omits `POST` from `methods` (or omits the entry) when the user lacks entity-create or source-mutation permission, or when domain invariants exclude the relation (e.g., Capability L4 has no POST-capable child entry)
- [x] Backend HATEOAS contract (link rel `x-related`, array shape, entry fields, method semantics) is documented in Swagger
- [ ] Frontend picker is built exclusively from `x-related` array entries advertising `POST` — manual inspection confirms `CONNECTION_TYPE_MAP` is not referenced when building the picker
- [ ] `GET`-only `x-related` entries do not appear in the picker
- [ ] Clicking a handle on any of the 5 entity types opens the picker if at least one `POST`-capable entry exists, or is a no-op if none exist
- [ ] Dragging from a handle onto an existing node creates a relation via the existing flow and does NOT open the picker
- [ ] Click-vs-drag movement threshold prevents accidental picker opens during slow drags and accidental drags during slow clicks
- [ ] Selecting a target type opens the existing full create dialog for that entity type, with appropriate pre-fill (e.g., capability level for parent relation)
- [ ] Submitting the dialog creates the entity, the relation, and the view-addition in that order, using existing endpoints only
- [ ] The new node is rendered on the canvas at a deterministic offset from the source handle, in the direction of the clicked side, without overlapping the source
- [ ] The new edge connects the source and the new entity using the existing edge-rendering pipeline (no new edge type)
- [ ] Pressing Escape, clicking outside, or clicking Cancel closes the picker without side effects
- [ ] Closing the create dialog without submitting leaves the canvas unchanged
- [ ] In Dynamic View Mode, handle-driven creation adds to the draft and produces no backend calls until Save; discarding the draft removes the new entity and its relation
- [ ] In regular mode, a failure during relation creation surfaces an actionable error naming the failed relation; the just-created entity is kept on the canvas (no auto-rollback) so the user can retry the relation manually
- [x] No new entity-mutation endpoints are introduced (the only backend additions are HATEOAS link advertisements)
- [ ] Cursor changes to a click affordance when over a handle; no tooltip, badge, or indicator is added
- [ ] All scenarios in the BDD section have at least one corresponding automated test (unit or integration), with at least one Playwright E2E covering the happy path

---

## Architecture

### Ownership

This spec spans **backend** (HATEOAS link generation across Application Components, Capabilities, and Origin Entities bounded contexts) and **frontend** (canvas feature module). The handle-click UX is owned by the canvas feature module; the affordance contract is owned by each entity's bounded context.

No new aggregates, no new domain events, no migrations. The backend changes are purely additions to the link-generation layer (`links.go` files in each bounded context).

### Domain Model

Unchanged. Existing relation invariants (component-to-component direction, capability parent level cap, realization inheritance, origin relation kinds) continue to be enforced by the existing command handlers. The HATEOAS affordances **reflect** those invariants; they do not duplicate them.

### API Surface

#### `x-related` HATEOAS contract (new)

Every canvas-renderable entity response MUST expose, in `_links`, a single `x-related` entry that is an **array** of link objects — one per available relation from the source entity. Multi-instance link rels are permitted by HAL; this spec uses that mechanism so the rel name (`x-related`) describes the resource (related entities), and the array enumerates the available relations.

```jsonc
"_links": {
  "x-related": [
    {
      "href": "/api/v1/capabilities",          // where the create dialog will POST
      "methods": ["POST"],                      // HTTP methods supported on this relation
      "title": "Child Capability",              // picker label
      "targetType": "capability",               // which create dialog to open
      "relationType": "capability-parent"       // which existing relation endpoint the orchestrator calls
    },
    {
      "href": "/api/v1/components",
      "methods": ["POST"],
      "title": "Component (realization)",
      "targetType": "component",
      "relationType": "capability-realization"
    }
    // GET-only example, reserved for future "list existing related" UI:
    // {
    //   "href": "/api/v1/capabilities/{id}/requires",
    //   "methods": ["GET"],
    //   "title": "Required Capabilities",
    //   "targetType": "capability",
    //   "relationType": "capability-requires"
    // }
  ]
}
```

**Entry fields** — `href`, `methods`, `title`, `targetType`, `relationType`. `relationType` is a stable string identifier the backend chooses per relation; suggested values: `component-relation`, `capability-parent`, `capability-realization`, `origin-acquired-via`, `origin-purchased-from`, `origin-built-by`, and post-039 `capability-requires`, `capability-enables`, `capability-supports`.

**Method semantics** — `POST` advertises "create a new entity related to me of this type". `GET` advertises "list existing entities related to me of this type" and is reserved for future use. The picker filters to entries whose `methods` include `POST`.

**No new mutation endpoints** — `href` points at the existing flat collection POST endpoint (e.g., `/api/v1/capabilities`). The orchestrator performs the entity-create, then dispatches the relation-create against the existing endpoint indicated by `relationType`. The link advertises the affordance; URL design for atomic create+link sub-resources is out of scope here.

#### Composed flow at runtime

1. **Create entity** — `POST` to the link's `href` using the existing entity-create endpoint.
2. **Create relation** — chosen by `relationType`:
   - `component-relation` → existing component-relation endpoint
   - `capability-parent` → `PATCH /api/v1/capabilities/:id/parent`
   - `capability-realization` → existing realization endpoint
   - `origin-acquired-via` / `origin-purchased-from` / `origin-built-by` → existing origin-relation endpoints
   - `capability-requires` / `capability-enables` / `capability-supports` → endpoints owned by spec 039 (only relevant once 039 is `_done`)
3. **Add to view** — existing add-to-view endpoint for the target entity kind, with the deterministic offset position from rule 9.

If a `relationType` does not map to an existing relation endpoint, the implementer MUST stop and either re-scope this spec or open a follow-up spec. No new mutation endpoints are added under cover of this spec.

### Persistence

Unchanged. No new persisted state.

### Frontend

**Affected:**
- `features/canvas/components/ComponentNode.tsx`, `CapabilityNode.tsx`, `OriginEntityNode.tsx` — handle click-vs-drag detection; cursor styling on handle hover.
- New picker component (e.g., `features/canvas/components/HandleCreatePicker.tsx`) — popover anchored to the clicked handle, populated entirely from the source entity's `x-related` array filtered to entries advertising `POST`. Each entry's `title` is the picker label; `targetType` selects the dialog; `relationType` is passed forward to the orchestrator.
- New orchestrator hook (e.g., `features/canvas/hooks/useCreateRelatedEntity.ts`) — opens the existing create dialog selected by `targetType`, then on success runs the relation-create (using `relationType`) and add-to-view (or writes all three to the draft store in dynamic mode).
- Existing dialogs (`CreateComponentDialog`, `CreateCapabilityDialog`, and the three origin-entity create dialogs) — accept a pre-fill / context prop so the orchestrator can set capability level (parent relation) or any other invariant-driven default. No UX changes.
- Existing `mutationEffects.ts` files for components, capabilities, origin entities, and view layouts — verified to invalidate the right caches; updated only if a new cache pair is touched.

**Not affected:**
- TreeView `+` button and its dialogs (other than the new pre-fill prop, which is opt-in).
- Drag-to-existing-node connection flow.
- Edge rendering and styling.
- Auto-layout (spec 124).
- `CONNECTION_TYPE_MAP` — remains the source of truth for drag-connect validity only; the picker MUST NOT consult it.

### Cross-Context Integration

No cross-bounded-context events. The only cross-feature integration is with **Dynamic View Mode (spec 164)**: this spec must not bypass the draft store. The orchestrator hook checks `dynamicModeSlice.draftActive` and, when active, writes the new entity, the new relation, and the new position into the draft instead of calling the backend. Save semantics are owned by spec 164 and unchanged.

---

## Implementation Plan

This section sequences the work so the backend contract lands before any frontend code depends on it.

### Phase A — Backend `x-related` HATEOAS links (must merge first)

A1. Define the `x-related` array entry shape and document it once (e.g., a published-language note plus a Swagger schema component reused across entity types). Document the `methods` semantics: `POST` advertises create-related, `GET` advertises list-existing-related (reserved for future).

A2. For each canvas-renderable entity type, extend the relevant `links.go` to generate the `x-related` array on every entity response. Coverage by `relationType`:

  - **Component** — entry with `relationType: component-relation`. Verify the existing relation directions before advertising any inverse-creation entry toward Capability or Origin entities; advertise only what the existing endpoints actually support today.
  - **Capability** — entries with `relationType: capability-parent` and `capability-realization`. Add `capability-requires`, `capability-enables`, `capability-supports` once spec 039 is `_done`.
  - **AcquiredEntity** — entry with `relationType: origin-acquired-via` (target Component).
  - **Vendor** — entry with `relationType: origin-purchased-from` (target Component).
  - **InternalTeam** — entry with `relationType: origin-built-by` (target Component).

A3. Each entry's `methods` array reflects what the calling user is authorized to do on that relation, gated by:
  - Permission to create the target entity type (existing permission check).
  - Permission to mutate the source entity / view.
  - Domain invariants (e.g., source Capability level == L4 → omit `POST` from the `capability-parent` entry, or omit the entry entirely).

  When the user is authorized for nothing on a relation, omit `POST` (and, if there's also no `GET` to advertise, omit the entry).

A4. Update Swagger docs for each entity DTO to describe the `x-related` array shape and entry fields (`href`, `methods`, `title`, `targetType`, `relationType`).

A5. Backend tests:
  - Unit test per entity type: `x-related` array present, entries and `methods` correct when authorized; `POST` omitted when user lacks permission; entry omitted when invariant-blocked.
  - Integration test: full create-via-entry round trip for at least one representative relation (Component → Component) — POST to the entry's `href`, then call the relation endpoint indicated by `relationType`, then add-to-view; verify the advertised contract holds end-to-end.

### Phase B — Frontend handle-click + picker (depends on Phase A)

B1. Handle click-vs-drag detection on `ComponentNode`, `CapabilityNode`, `OriginEntityNode`. Cursor styling on hover. Pixel threshold for click vs drag.

B2. `HandleCreatePicker` component — popover anchored to the clicked handle, list built from the source entity's `x-related` array filtered to entries advertising `POST`. Empty filtered list → no popover, click is a no-op. Escape / outside-click / Cancel closes without side effect.

B3. `useCreateRelatedEntity` orchestrator hook —
  - On entry selection: opens the existing create dialog selected by `targetType`, passing pre-fill where applicable (capability level for parent relation; any other invariants surfaced by the entry).
  - On dialog submit success: runs relation-create + add-to-view via existing endpoints, mapped from `relationType`.
  - On failure during relation-create in regular mode: surfaces an actionable error naming the failed relation; the just-created entity is kept on the canvas as an orphan; the user can retry the relation manually via the existing drag-connect flow.
  - In dynamic mode: writes entity, relation, and position into the draft store; no backend calls.

B4. Pre-fill prop on existing create dialogs — additive, opt-in; treeview `+` button continues to call them with no pre-fill.

B5. Frontend tests:
  - Unit tests for click-vs-drag threshold, picker rendering from affordance links, orchestrator branching (regular vs dynamic mode), rollback on relation failure.
  - Playwright E2E: one happy-path scenario per source kind (5 total).

### Phase C — Verification

C1. Manual canvas walkthrough for all 5 source kinds, both modes (regular + dynamic).

C2. Confirm no regressions in drag-to-existing-node, treeview `+` button, or auto-layout.

C3. `mcp__codescene__pre_commit_code_health_safeguard` on all changed files; iterate to 10.0 per repo standard.

---

## Implementation Notes (added during build)

### How `_links["x-related"]` is wired in the JSON output

The shared `types.Links` map type is unchanged (still `map[string]Link`) — touching it would have rippled through ~30 call sites that read `links["self"].Href` and similar. Instead, each canvas-renderable DTO carries an additional `XRelated []types.RelatedLink` field tagged `json:"-"` plus a custom `MarshalJSON` that calls `types.SpliceXRelated(base, d.XRelated)`. The helper parses the default JSON, splices `x-related` into the `_links` subtree, and returns the merged bytes. This keeps the JSON wire shape exactly as the spec describes (`{"_links":{"self":{...},"x-related":[...]}}`) without mutating the shared link type.

Trade-off: per-DTO `MarshalJSON` adds five small custom marshalers (one per canvas DTO) and a single shared splice helper, in exchange for zero churn to existing handlers and tests. The helper is unit-tested (3 cases: existing links, no `_links` field, empty related → input unchanged).

### Swagger discoverability of `RelatedLink`

Because `XRelated` is `json:"-"` it is invisible to swag. To make `RelatedLink` reachable in the generated OpenAPI schema, a new doc-only handler `GET /reference/x-related-links` was added (alongside the existing `/reference/{relations,components}` handlers) returning an `XRelatedReferenceDoc{Title, Description, Example []RelatedLink}`. This both (a) gives clients a single discovery URL describing the contract and (b) causes swag to emit `RelatedLink` and `XRelatedReferenceDoc` definitions in `docs/swagger.json`. The handler is annotated with `@Failure 401`/`@Failure 403` to match the auth posture of the rest of the API.

### Canonical `relationType` → backend endpoint mapping

`types.LookupRelationEndpoint(relationType)` lives next to `RelatedLink` and resolves each stable `relationType` (`component-relation`, `capability-parent`, `capability-realization`, `origin-acquired-via`, `origin-purchased-from`, `origin-built-by`) to the existing endpoint the orchestrator must call (path + HTTP method). This is the single source of truth for the relation step today: the integration test consumes it instead of hard-coding `/api/v1/relations`, so a route rename or a `relationType` rename surfaces as a test failure rather than silent drift. Phase B's frontend orchestrator will mirror the same table.

### JSON encoding

`SpliceXRelated` uses `json.Encoder.SetEscapeHTML(false)` for all marshals so `&`, `<`, `>` in advertised hrefs (e.g., future query-string-bearing URLs) round-trip unchanged.

### Backend completion summary

Phase A (acceptance criteria 1–4 and 16) is implemented and green:
- `types.RelatedLink` + `types.SpliceXRelated` helper, with unit tests.
- `XRelated`/`MarshalJSON` on `ApplicationComponentDTO`, `CapabilityDTO`, `AcquiredEntityDTO`, `VendorDTO`, `InternalTeamDTO`, with unit tests.
- `*XRelatedForActor` link generators on `ArchitectureModelingLinks` and `CapabilityMappingLinks`, gated by:
  - `components:write` for the component-relation/origin flows;
  - `capabilities:write` + L1/L2/L3 source level for the `capability-parent` entry (L4 omits POST per the spec invariant);
  - `components:write` AND `capabilities:write` for the `capability-realization` entry.
- Handler enrichment wired in 5 places: `ComponentHandlers.enrichWithLinks`, `CapabilityHandlers.addLinksToCapability`, `AcquiredEntityHandlers.enrichWithLinks`, `VendorHandlers.enrichWithLinks`, `InternalTeamHandlers.enrichWithLinks`.
- Integration round-trip test (`TestXRelated_ComponentToComponent_RoundTrip_Integration`) exercises the advertised contract end-to-end through the real handlers + projections: POST source component → wait for projection → GET source → assert `_links["x-related"]` contains the `component-relation` POST entry → POST to the entry's `href` to create the target → resolve the relation endpoint via `types.LookupRelationEndpoint(relationType)` → POST to the resolved endpoint → assert the relation appears in the read model.
- Swagger regenerated; `RelatedLink`, `XRelatedReferenceDoc`, and the `/reference/x-related-links` route (with 401/403) are first-class definitions.

Phase B (frontend) and Phase C (verification) are not yet started; their work remains as described above.

---

## Design Decisions

1. **HATEOAS as the single source of validity** — Per user direction. Keeps relation rules in one place (the backend domain), removes the risk of frontend/backend drift, and lets new relation types surface in the picker without frontend code changes. Alternative considered: reuse the frontend `CONNECTION_TYPE_MAP` — rejected because it duplicates domain rules and has already drifted from backend reality (e.g., capability dependencies are not yet in the map even though spec 039 defines them).

   Link rel naming follows REST resource conventions: a single `x-related` rel names "related entities" as a resource concept, and the array enumerates the available relations (HAL allows multi-instance link rels). Affordances are signaled by each entry's `methods` array (`POST` = create-related, `GET` = list-related). Alternatives considered: (a) action-named keys like `x-create-related-*` — rejected because the key would conflate verb with resource, and adding sibling operations (list, add-existing, remove) would proliferate keys instead of reusing `methods`; (b) per-relation keys like `x-related-children`, `x-related-realizations` — rejected because it introduces a naming convention to maintain in addition to `relationType`, while the array form keeps the whole contract in one place.
2. **All entity types support click-create uniformly** — Per user direction. Even Origin entities, where the only target is Component, get the click-create gesture for consistency. The cost is one click + one picker entry; the benefit is no special-cased UI.
3. **Reuse existing full create dialogs** — Per user direction. Invariants are identical across entry points; a separate quick-create form would duplicate validation and drift over time. Dialogs gain only an additive pre-fill prop.
4. **Click on handle, drag is unchanged** — Click is the new gesture; drag-to-existing-node keeps its current behavior. Click-vs-drag disambiguated by a small movement threshold. Alternative considered: drag-to-empty-canvas — rejected because it conflates "abort connection" with "create entity" and is ambiguous on touch / trackpad.
5. **Cursor change is the only discoverability cue** — Per user direction. Avoids canvas chrome and matches the minimal-affordance aesthetic of the rest of the canvas.
6. **Position by clicked-handle direction, not by auto-layout** — A predictable, click-aligned placement keeps the user's mental model stable. Auto-layout (spec 124) can rearrange later. Alternative considered: run auto-layout after each create — rejected as too disruptive to the user's spatial memory.
7. **Add-to-view is implicit** — Asking the user "do you want to add this to the view?" after creation defeats the point of the gesture. The view they're looking at is by definition where the new entity belongs.
8. **Dynamic mode integration via the existing draft store** — Avoids divergence with spec 164's Save / Discard flows. New entities created from a handle are indistinguishable from drag-from-sidebar additions inside the draft.
9. **No new mutation endpoints; HATEOAS links are the only backend addition** — The feature is a UI affordance over an already-complete domain API plus an exposure layer. If the implementer discovers a missing mutation endpoint, that's a domain gap that warrants its own spec, not an undeclared backend addition.
10. **Treeview `+` button is preserved** — It's the only entry point for creating entities that have no canvas anchor (e.g., the very first entity in a new tenant) and remains useful when users are working in the treeview rather than the canvas.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| HATEOAS-driven validity | Backend work required across 5 entity types before any frontend work can ship | Implementation Plan sequences the backend phase first; backend changes are additive (new links) and shippable independently of the frontend |
| Click on handle as the gesture | Some users may instinctively drag for "more" | Drag-to-existing flow is unchanged so learned muscle memory still works; cursor change indicates clickability |
| Cursor-only discoverability | Lower visibility for new users | Acceptable for power-user-oriented modeling tool; can be revisited if telemetry shows low adoption |
| Position by handle direction | Multiple creates from the same handle pile up in the same spot | Deterministic offset includes a small jitter / stack offset for consecutive creates from the same handle; auto-layout can be invoked manually |
| No new mutation endpoints | Implementer is forced to stop if a composition gap is found | Encoded as a hard rule in Acceptance Criteria so gaps surface during planning, not mid-implementation |
| Reuse existing dialogs across two entry points | Dialog state must accept optional pre-fill from the orchestrator | Pre-fill prop is additive and opt-in; treeview path unchanged |

---

## Checklist

- [x] Specification ready
- [ ] Implementation done <!-- backend (Phase A) complete; frontend (Phase B) not started -->
- [ ] Unit tests implemented and passing <!-- backend unit tests in place; frontend pending -->
- [x] Integration tests implemented if relevant <!-- Component-to-Component round-trip test under `integration` build tag -->
- [x] API documentation updated <!-- Swagger regenerated with RelatedLink, XRelatedReferenceDoc, and /reference/x-related-links route -->
- [ ] User sign-off
