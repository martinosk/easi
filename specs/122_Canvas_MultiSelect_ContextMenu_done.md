# Spec 122: Canvas Multi-Select Context Menu

## Status
done

## Description
Enable bulk operations on multiple selected canvas objects via React Flow's built-in shift+drag rectangle selection. When multiple nodes are selected and the user right-clicks, a context menu shows only the actions that are applicable to **all** selected objects. Confirming an action executes it for every selected object.

## Dependencies
- React Flow built-in multi-select (shift+drag rectangle)
- Existing single-node context menu (Spec 020C)
- Existing delete confirmation flow (`useDeleteConfirmation`)
- HATEOAS-driven action resolution

## Functional Requirements

### Multi-Select Behaviour (React Flow Built-In)
- User holds Shift and drags a rectangle on the canvas
- All nodes whose bounding box intersects the rectangle become selected
- React Flow marks these nodes with `selected: true`
- Edges are **not** included in rectangle selection (React Flow default)

### Context Menu on Multi-Selection

**Trigger:** User right-clicks anywhere on the canvas while 2+ nodes are selected.

**Menu Resolution:**
1. Collect all currently selected nodes from React Flow state
2. For each selected node, resolve its HATEOAS links (both `modelLinks` and `viewElementLinks`) using the same lookup logic as the single-node context menu
3. Compute the **intersection** of permitted actions across all selected nodes:
   - **Remove from View** — shown only if every selected node has the `x-remove` view element link
   - **Delete from Model** — shown only if every selected node has the `delete` model link
4. If the intersection is empty (no common actions), show no context menu

**Menu Items:**
| Action | Condition | Style |
|--------|-----------|-------|
| Remove from View (N items) | All nodes have `x-remove` link | Normal |
| Delete from Model (N items) | All nodes have `delete` link | Danger |

where N is the count of selected nodes.

### Scenarios

#### Scenario 1: All selected objects can be removed and deleted
**Given** the user has write access to the current view and delete access to all model entities
**And** 5 nodes are selected via shift+drag
**When** the user right-clicks on the canvas
**Then** the context menu shows:
  - "Remove from View (5 items)"
  - "Delete from Model (5 items)"

#### Scenario 2: Mixed permissions — some objects cannot be deleted from model
**Given** the selection contains 3 components and 2 capabilities
**And** the user has delete permission on the components but NOT on the capabilities
**And** the user can remove all 5 from the current view
**When** the user right-clicks on the canvas
**Then** the context menu shows only:
  - "Remove from View (5 items)"
**Because** "Delete from Model" requires the `delete` link on **every** selected node

#### Scenario 3: Private view — non-owner selects multiple nodes
**Given** the current view is private and the user is NOT the owner
**And** 3 nodes are selected
**When** the user right-clicks on the canvas
**Then** no context menu appears
**Because** the view element links do not include `x-remove` (non-owner cannot edit private view) and model delete links depend on the user's role

#### Scenario 4: Single node selected — existing behaviour preserved
**Given** exactly 1 node is selected (or the user right-clicks a specific node)
**When** the user right-clicks that node
**Then** the existing single-node context menu appears (unchanged behaviour from Spec 020C)

#### Scenario 5: No nodes selected
**Given** no nodes are selected
**When** the user right-clicks on the empty canvas
**Then** no context menu appears

### Remove from View (Bulk)

**Trigger:** User clicks "Remove from View (N items)" in the multi-select context menu.

**Behaviour:**
- Show a confirmation dialog: "Remove N items from the current view? The items will remain in the model."
- On confirm: execute the remove-from-view mutation for each selected node, using the `x-remove` link from each node's view element links
- On cancel: close dialog, selection remains
- Non-destructive operation — items stay in the model and other views
- After completion, all removed nodes disappear from the canvas

### Delete from Model (Bulk)

**Trigger:** User clicks "Delete from Model (N items)" in the multi-select context menu.

**Behaviour:**
- Show a confirmation dialog with danger styling:
  - Title: "Delete N items from Model"
  - Message: "This will permanently delete N items from the entire model. They will be removed from ALL views and all associated relations will be deleted. This cannot be undone."
  - List the names of all items being deleted (scrollable if many)
  - Confirm button text: "Delete N items"
- On confirm: execute the delete mutation for each selected node sequentially (to avoid overwhelming the backend with concurrent writes), using each node's `delete` model link
- On cancel: close dialog, selection remains
- After completion, all deleted nodes disappear from all views
- If any individual delete fails mid-batch, stop and show an error indicating which items were deleted and which failed

## UX Details

### Context Menu Appearance
- Uses the existing shared `ContextMenu` component
- Positioned at the right-click cursor location
- Closes on outside click, Escape, or after selecting an item

### Confirmation Dialog
- Uses the existing shared `ConfirmationDialog` component
- For "Remove from View": standard styling, confirm button says "Remove"
- For "Delete from Model": danger styling, confirm button says "Delete N items"
- Shows loading spinner while operations execute
- Disables confirm button during execution

### Selection Visual Feedback
- React Flow's built-in selection rectangle (blue dashed border) during drag
- Selected nodes show React Flow's default selected styling
- Selection persists after context menu closes (unless an action was performed)
- Clicking the canvas pane clears selection (existing behaviour)

## HATEOAS Contract
No new API endpoints or link relations are needed. The feature reuses:
- `x-remove` link on view elements (components, capabilities, origin entities) for "Remove from View"
- `delete` link on model entities (components, capabilities, origin entities) for "Delete from Model"

The frontend determines available bulk actions by intersecting the links present on all selected nodes. This preserves the backend as the single source of truth for authorization.

## Implementation Notes

### Hook: `useMultiSelectContextMenu`
- Reads selected nodes from React Flow via `useStore` or `getNodes().filter(n => n.selected)`
- On right-click with 2+ selected nodes, resolves HATEOAS links for each node
- Computes action intersection
- Returns menu state (position, items, selected node details)

### Distinguishing Single vs Multi Context Menu
- `onNodeContextMenu` fires when right-clicking a specific node — use single-node menu (existing)
- `onPaneContextMenu` or a right-click handler on the canvas that checks selection count — if 2+ nodes selected, show multi-select menu instead of nothing
- When right-clicking a node that is part of a multi-selection, show the multi-select menu (not the single-node menu)

### Bulk Execution Strategy
- "Remove from View": can execute in parallel (independent mutations, non-destructive)
- "Delete from Model": execute sequentially to respect event sourcing consistency and avoid race conditions with cascade deletions

## Checklist
- [ ] Specification ready
- [ ] `useMultiSelectContextMenu` hook implemented
- [ ] Multi-select context menu component implemented
- [ ] HATEOAS link intersection logic for bulk actions
- [ ] Bulk "Remove from View" with confirmation dialog
- [ ] Bulk "Delete from Model" with confirmation dialog and item list
- [ ] Sequential execution for model deletes
- [ ] Error handling for partial batch failures
- [ ] Single-node right-click still works when only 1 node selected
- [ ] Right-click on multi-selected node shows multi-select menu
- [ ] Unit tests for link intersection logic
- [ ] Unit tests for multi-select context menu rendering
- [ ] Integration: verify HATEOAS permissions respected
- [ ] Build passes (`go test ./...`, `npm test -- --run`)
- [ ] User sign-off
