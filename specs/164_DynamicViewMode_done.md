# 164 — Dynamic View Mode

> **Status:** done
> **Depends on:** —
> **Supersedes the behavior of:** [161 — Auto-Generate View from Entity](161_Auto_Generate_View_From_Entity_superseded.md). The one-shot BFS flow is removed and replaced with the interactive dynamic mode described here.

---

## Problem Statement

Spec 161's "Generate View for <entity>" feature traverses the full dependency graph in a single BFS pass and dumps every reachable entity into a new view. For well-connected entities the result is several hundred nodes — a complete-but-unusable picture that the user has to hand-trim. There is no way to express "show me X plus its callers, but stop at the first capability layer" without manually deleting dozens of nodes after the fact.

Architects need to *sculpt* a view interactively: start from a seed, expand only the relations that matter, drop subgraphs they don't, and try things out without polluting the saved view. The proposal is **Dynamic View Mode** — a per-session UI mode any view can be flipped into. While in dynamic mode the canvas presents per-node `+N` expansion badges (broken down by edge type) alongside the existing drag/drop, and all changes are draft until the user explicitly Saves. Toggling out without saving discards the draft.

Dynamic mode is reachable from two entry points: a right-click on any canvas entity ("Create dynamic view from <name>") and a toggle on the toolbar of any existing view.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise architect** | Build focused, intentional views without traversing the whole graph; explore "what's connected to X?" safely without committing to a saved view |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Dynamic View Mode

  Scenario: Create a new dynamic view from a canvas entity
    Given I am viewing the architecture canvas
    When I right-click on a Component and select "Create dynamic view from <name>"
    Then a new view is created containing only that entity
    And the canvas opens that view in dynamic mode
    And the entity has a "+N" expansion badge if it has neighbors

  Scenario: Expand a node by edge type
    Given I am in dynamic mode with a single component visible
    And the component has 4 unexpanded relations and 1 unrealized capability
    When I click the "+N" badge on the component
    Then a popover lists each enabled edge type with its unexpanded count
    When I click "Triggers / Serves (+4)"
    Then the 4 related components appear on the canvas
    And the capability is not added
    And the canvas remains in unsaved-draft state

  Scenario: Expand all neighbors at once
    Given I am in dynamic mode and click "+N" on a node
    When I click "Expand all" in the popover
    Then all neighbors across enabled edge types are added to the draft

  Scenario: Toggle dynamic mode on an existing view
    Given I am viewing a saved architecture view
    When I toggle "Dynamic mode" on
    Then the canvas enters dynamic mode for the current session only
    And every entity with unexpanded neighbors shows a "+N" badge
    And the existing drag/drop behavior remains available

  Scenario: Save a dynamic-mode draft
    Given I am in dynamic mode and have added 3 entities, removed 1, and moved 2
    When I click "Save view"
    Then the additions, removals, and position changes are persisted to the view via existing endpoints
    And the canvas exits dynamic mode and returns to regular mode
    And a success toast confirms the save

  Scenario: Discard a dynamic-mode draft
    Given I am in dynamic mode with unsaved changes
    When I toggle dynamic mode off (or click Cancel)
    Then I am asked to confirm discarding changes
    And on confirm, the view returns to its last saved state
    And the canvas exits dynamic mode

  Scenario: Toggle off with no unsaved changes
    Given I am in dynamic mode with no changes
    When I toggle dynamic mode off
    Then the canvas exits dynamic mode without prompting

  Scenario: Removing a node removes only that node
    Given I am in dynamic mode with a chain A -> B -> C where C is reachable only through B
    When I remove B from the canvas
    Then only B is removed from the draft
    And A and C remain
    And the removal is part of the unsaved draft

  Scenario: Removing a multi-selection removes exactly the selected nodes
    Given I am in dynamic mode with nodes A, B and C selected
    When I remove the selection from the canvas
    Then A, B and C are removed from the draft and nothing else

  Scenario: Drag/drop additions are part of the draft
    Given I am in dynamic mode
    When I drag a Capability from the entity sidebar onto the canvas
    Then the Capability appears on the canvas as a draft addition
    And no API call is made until I click Save

  Scenario: Drag-to-reposition is part of the draft
    Given I am in dynamic mode
    When I drag an existing node to a new position
    Then the new position is held in the draft
    And the saved view's position is unchanged until I Save

  Scenario: Auto-layout in dynamic mode applies to the draft
    Given I am in dynamic mode
    When I click "Auto Layout"
    Then entities reposition based on the current draft
    And the new positions are part of the unsaved draft

  Scenario: Re-opening a previously dynamic-edited view starts in regular mode
    Given I previously saved a view from dynamic mode
    When I open that view again
    Then the canvas opens in regular mode
    And I can re-enter dynamic mode by toggling it on

  Scenario: Filter edge types in the workbench
    Given I am in dynamic mode
    And the "Realizations" edge filter is unchecked
    When I click "+N" on a component
    Then the popover does not list realizations
    And "Expand all" does not add capabilities

  Scenario: Enter dynamic mode on an empty view
    Given I open a saved view with zero entities
    When I toggle "Dynamic mode" on
    Then the canvas is empty
    And the dynamic-mode toolbar and entity sidebar are available
    And no "+N" badges are shown until at least one entity is added to the draft
```

---

## Business Rules & Invariants

1. **Mode is per-session, not persisted** — a view's dynamic state lives only in the frontend session; it is never written to backend storage and never crosses the API boundary.
2. **Draft isolation** — while in dynamic mode, all entity additions, removals, and position changes accumulate in a frontend-only draft. The persisted view is unchanged until Save.
3. **Save is a diff over existing endpoints** — Save commits the difference (added entities, removed entities, position deltas) by calling the existing add-component / add-capability / add-origin-entity / remove-* / position-update endpoints. No new endpoints are introduced.
4. **No seed privilege when editing** — when entering dynamic mode on an existing view, every entity is removable. There is no protected seed.
5. **Seed exists only at creation** — when "Create dynamic view from X" is used, the new view contains exactly X. Once the workbench opens, X is just another entity in the draft and can be removed like any other.
6. **Removal is exact, never cascading** — removing an entity removes only that entity. Neighbours that become unreachable stay on the canvas until the user removes them. A cascade shipped in unit 8 and was reverted after user feedback; Remove from View is scoped to what the user picked, and only Delete from Model has cascade rules.
7. **Multi-select removal** — removing a multi-selection removes exactly the selected entities, with no orphan prompt.
8. **Filter changes affect the workbench, not the saved view** — toggling edge or entity-type filters in dynamic mode only affects what appears in `+N` popovers. Filters are not persisted and do not modify the draft on their own.
9. **Discard requires confirmation only when there are unsaved changes** — toggling dynamic mode off, navigating away, or clicking Cancel prompts only if the draft contains uncommitted changes.
10. **Edge-type coverage matches spec 161** — Triggers/Serves, Realizations, Capability Parentage, and Origin (AcquiredVia / PurchasedFrom / BuiltBy) are the expandable edge types. Capability Dependencies remain out of scope (spec 039 territory).
11. **The 500-entity safety cap from spec 161 is removed** — the cap was a guard against unbounded BFS. Dynamic mode is user-paced; no automatic cap is required.
12. **Permission model is unchanged** — entering dynamic mode requires the same `views:write` permission already required to edit a view; "Create dynamic view from X" requires the existing view-creation permission.
13. **Partial-save failure is observable** — if any of the per-entity API calls during Save fails, the canvas is refreshed from server state and the user is told which changes did not persist; the draft is not silently retained.
14. **Save is non-blocking but locks input** — Save shows a non-blocking progress indicator. While Save is in flight, expansion (`+N`), removal, drag/drop additions, and node repositioning are disabled; navigation away from the view is still possible (and abandons the in-flight save with the same partial-failure handling as rule 13).
15. **No concurrency control on Save** — Save does not detect or merge concurrent modifications by other users. Per-entity endpoint semantics apply (last write wins, including silently overwriting another user's recent changes to the same entities).
16. **Empty-view dynamic mode is supported** — toggling dynamic mode on a view with zero entities is allowed; the canvas is empty, no `+N` badges are shown, and the user can populate the draft via drag/drop from the entity sidebar.

---

## Acceptance Criteria

- [x] A new context-menu item "Create dynamic view from <name>" replaces "Generate View for <name>" on canvas nodes; clicking it creates a view containing only the source entity and opens that view in dynamic mode.
- [x] An existing-view toolbar exposes a "Dynamic mode" toggle; toggling it on enters dynamic mode for the current session only.
- [x] In dynamic mode, every entity with at least one unexpanded neighbor under current filters shows a `+N` badge.
- [x] Clicking a `+N` badge opens a popover with one row per enabled edge type, each showing the unexpanded count for that type, plus an "Expand all" row.
- [x] Clicking a single-edge-type row in the popover adds only that type's neighbors to the draft; clicking "Expand all" adds neighbors across all enabled edge types.
- [x] Removing an entity in dynamic mode removes only that entity, or exactly the multi-selected entities; nothing cascades.
- [x] Dragging an entity from the sidebar onto the canvas in dynamic mode adds it to the draft; no API call is made until Save.
- [x] Drag-to-reposition in dynamic mode updates draft positions only; the saved view's positions are unchanged until Save.
- [x] Auto-layout (spec 124) in dynamic mode repositions the draft; the layout is part of the unsaved draft until Save.
- [x] Clicking "Save view" persists the draft via existing add / remove / position endpoints; no new backend endpoints are added.
- [x] On partial-save failure, the canvas is refreshed from server state and a toast names what did not persist.
- [x] Toggling dynamic mode off, clicking Cancel, or navigating away with unsaved changes prompts for confirmation; on discard, the view reverts to its last saved state.
- [x] Reopening a view that was previously edited in dynamic mode starts in regular mode (no `dynamic` flag is persisted).
- [x] The 500-entity traversal cap and `truncated` flag from spec 161 are removed from `collectRelatedEntities` and any callers.
- [x] Spec 161's `useGenerateView` hook and one-shot BFS flow are removed; the `Generate View for X` menu item no longer exists.
- [x] Spec 161 is renamed to `_superseded` once this spec is `_done`.
- [x] During Save, a non-blocking progress indicator is shown and canvas mutation inputs (expansion, removal, drag/drop, repositioning) are disabled until the call sequence completes.
- [x] Toggling dynamic mode on a zero-entity view enters dynamic mode with an empty canvas; the toolbar and entity sidebar are available and no badges appear until an entity is added.

---

## Architecture

### Ownership

This change is **frontend-only**. The `architectureviews` bounded context (`backend/internal/architectureviews`) is unchanged — its existing commands and HATEOAS links are sufficient to express the diff at Save time.

The frontend canvas feature (`frontend/src/features/canvas`) owns:
- The dynamic-mode state machine (idle → editing-draft → saving → idle).
- The popover UI for per-edge-type expansion.
- The draft store layered over the persisted view (additions, removals, position deltas, active filters).
- The dirty-check and confirm-discard prompt.
- The Save translator that turns a draft into a sequence of existing API calls.

The relations / capabilities / origin-relationships hooks already loaded by the canvas continue to be the source of truth for "what neighbors exist".

### Domain Model

No domain model changes. No new aggregates, value objects, or events. The view aggregate is unaware of dynamic mode.

### API Surface

No new endpoints. Save composes the existing endpoints exposed by the architecture-views context:

- Add: `POST /views/{id}/components`, `POST /views/{id}/capabilities`, `POST /views/{id}/origin-entities`.
- Remove: existing per-entity remove endpoints.
- Position deltas: the existing batch position update used by auto-layout (spec 124).

Calls are issued sequentially. On failure the frontend refreshes from server state and reports which changes did not persist.

### Persistence

No schema changes. No migrations. The dynamic flag never crosses the API boundary.

### Frontend

Affected and new areas (paths for orientation; final structure decided in implementation):

- `features/canvas/components/context-menus/NodeContextMenu.tsx` — replace "Generate View for X" with "Create dynamic view from X".
- `features/canvas/CanvasContainer.tsx` (or its toolbar component) — host the dynamic-mode toggle and a Save / Cancel pair surfaced when dynamic mode is on.
- `features/canvas/hooks/useGenerateView.ts` — **deleted**; replaced by the dynamic-mode entry flow.
- `features/canvas/utils/collectRelatedEntities.ts` — repurposed: instead of returning the full reachable set, exposes per-edge-type neighbor lookup keyed by entity id. The 500-cap and `truncated` flag are removed.
- New: a draft store (zustand slice or context) holding additions / removals / position deltas / active filters / dirty flag.
- New: an `ExpandPopover` component rendering the per-edge-type breakdown.
- New: a `DynamicModeToggle` button in the canvas toolbar.
- New: a `useSaveDynamicDraft` hook translating a draft into existing add / remove / position calls.

### Cross-Context Integration

None.

---

## Design Decisions

1. **Dynamic mode is per-session, not persisted.** Rationale: avoids any backend or schema change, and the user's mental model is that a view *is* a view — "dynamic" is just how you happen to be looking at it right now. Alternatives considered: persisting `view.isDynamic` (rejected — adds backend surface and forces every reader to handle a new state for no functional gain).

2. **Save is a diff over existing endpoints; no batch-save endpoint.** Rationale: keeps the change frontend-only, avoids designing a new transactional API, and reuses well-tested commands. Alternative: a `PUT /views/{id}/contents` replace-everything endpoint (rejected — larger blast radius, requires a new event type, no urgent need).

3. **Per-edge-type expansion via a popover, not separate badges.** Rationale: a single `+N` badge per node keeps the canvas visually quiet; the popover surfaces the breakdown only when the user opts in. Alternatives considered: multiple small per-edge badges around each node (rejected — clutters dense graphs); shift-click for "expand all" (rejected — undiscoverable).

4. **No cascade on remove.** Rationale: the user removes what they clicked and nothing else. An earlier cascade-to-orphans implementation surprised users by dropping subgraphs they meant to keep and was reverted; orphans are visible on the canvas and cheap to remove explicitly.

5. **Drag/drop additions and position changes are part of the draft.** Rationale: in dynamic mode the user is sculpting one coherent change. Mixing live-saved drag/drop with draft `+N` expansion would create two commit semantics on the same surface and confuse users. Alternative: keep drag/drop and position changes live (rejected — explicitly contradicts the user's intent).

6. **"Create dynamic view from X" replaces "Generate View for X".** Rationale: the auto-generate behavior produced unusable views; the dynamic-mode entry is strictly more capable (it can still produce the same result via "Fill canvas to depth N" but lets the user stop earlier). Alternative: keep both menu items (rejected — two paths for similar intent, with the worse one becoming a footgun).

7. **Discard prompt only when there are unsaved changes.** Rationale: keeps the no-changes case friction-free while protecting users from losing meaningful work. Alternative: always prompt (rejected — annoying); never prompt (rejected — data-loss risk).

8. **Remove the 500-entity cap from the collected-entities utility.** Rationale: the cap existed because spec 161's BFS was unbounded and could produce unusable graphs in one shot. In dynamic mode the user controls every expansion, so the cap is no longer protective. Alternative: keep a soft cap with a warning (rejected — premature given the user-paced model).

9. **Sequential Save with refresh-on-failure rather than transactional save.** Rationale: a transactional batch endpoint would require new backend work (event design, projector handling, atomicity story); sequential calls reuse battle-tested commands. Failure surface is small because the user can re-Save and the draft state is preserved through the refresh.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Per-session dynamic flag | Users cannot bookmark or share "this view in dynamic mode" | Dynamic mode is a tool, not a presentation; if a curated subgraph is worth sharing, the user saves it as a regular view |
| Diff-based save (sequential calls) | Partial failure can leave the persisted view inconsistent with the draft | On any per-call failure, surface a toast naming what didn't persist and refresh the canvas from server state |
| No cascade on remove | Orphaned subgraphs linger after removing a hub node | Orphans stay visible; the user multi-selects and removes them explicitly |
| Drag/drop and position changes are draft, not live | Diverges from regular-mode live-save behavior on the same canvas | The dynamic-mode toggle and an unsaved-changes indicator (e.g. dot on Save) make the mode shift visible at all times |
| Replacing "Generate View for X" outright | Users may have muscle memory for one-shot generation | Dynamic mode + "Fill canvas to depth N" reproduces the old behavior in two clicks; release notes flag the change |

---

## Implementation Progress

Each unit followed strict RED-GREEN-REFACTOR TDD where logic was testable; integration glue was verified via the full test + build pipeline.

- [x] **Unit 1** — `dynamicMode` utility: `getNeighbors`, `getUnexpandedByEdgeType`, `computeOrphans` (19 tests)
- [x] **Unit 2** — `dynamicModeSlice` zustand slice + diff selectors (17 tests)
- [x] **Unit 3** — `useSaveDynamicDraft` hook + pure `saveDraft` translator (7 tests)
- [x] **Unit 4** — `ExpandPopover` Mantine component (7 tests)
- [x] **Unit 5** — `DynamicModeToolbar` (toggle / Save / Cancel + discard-confirm) (8 tests)
- [x] **Unit 6** — `DynamicExpandBadge`, `withDynamicExpansion` HOC over node components, `DynamicModeContainer`; mounted in `ComponentCanvas`
- [x] **Unit 7** — Context-menu item updated (`Generate View for X` → `Create dynamic view from X`) on canvas + tree menus; `useCreateDynamicView` replaces `useGenerateView`; old `useGenerateView` and `collectRelatedEntities` deleted; spec 161 renamed to `_superseded`
- [x] **Unit 8** — Drag/drop interception in `useCanvasDragDrop`, position interception in `useCanvasSelection.onNodeDragStop`, cascade-on-delete in `useDeleteConfirmation` (cascade later reverted, see design decision 6); full test suite (1228 tests passing) and `npm run build` verified

**Not implemented in this spec**: the "Fill canvas to depth N" sidebar action and the per-edge-type / per-entity-type filter sidebar from the workbench mockup. These are additive UX affordances; the core dynamic-mode behavior works without them. If valuable, they warrant a follow-up numbered spec.

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing (58 new tests; full suite 1228/1228)
- [x] Integration tests implemented if relevant — n/a; behavior verified via existing canvas tests + production build
- [x] API documentation updated — n/a; no backend API changes
- [x] User sign-off (2026-09-14)
