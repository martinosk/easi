# Spec 123: Treeview Multi-Select

## Status
pending

## Description
Enable multi-selection of items in the navigation treeview across all component types (applications, capabilities, origin entities). Users can select a range of items via Shift+click and toggle individual items via Ctrl+click (Cmd+click on macOS). Multi-selected items show a context menu with "Delete from Model" when available via HATEOAS links. Multi-selected items can be dragged onto the active view, adding all selected items that are not already present.

## Dependencies
- Existing treeview sections (ApplicationsSection, CapabilitiesSection, AcquiredEntitiesSection, VendorsSection, InternalTeamsSection)
- Existing single-item context menus (`useTreeContextMenus`)
- Existing drag-and-drop to canvas (`useCanvasDragDrop`)
- HATEOAS-driven action resolution (entity `_links`)
- Bulk operations pattern from Spec 122

## Functional Requirements

### Multi-Select Behaviour

#### Ctrl+Click (Toggle Individual)
- Ctrl+click (Cmd+click on macOS) on a treeview item toggles its selected state without affecting other selections
- If the item is already selected, it becomes deselected
- If the item is not selected, it becomes selected in addition to any existing selections
- Works across all section types — a user can Ctrl+click an application, then Ctrl+click a capability, selecting both

#### Shift+Click (Select Range)
- Shift+click selects a contiguous range of items from the last-clicked item (anchor) to the Shift+clicked item
- The range is computed within a **single section** only — Shift+click does not span across sections (e.g., cannot range-select from Applications into Capabilities)
- Within the capabilities tree, the range follows the **visible flattened order** (respecting expanded/collapsed state) — collapsed children are not included in the range
- If no anchor exists (no prior click in that section), Shift+click selects from the first item in the section to the clicked item
- Shift+click replaces any existing multi-selection within that section but preserves selections in other sections

#### Plain Click (Reset)
- A plain click (no modifier keys) on a treeview item clears the entire multi-selection and selects only that item
- This restores the existing single-select behaviour: the item is highlighted, detail panels update, and the canvas pans to the node if it is in the current view

### Selection Visual Feedback
- Multi-selected items show the same `selected` CSS class as single-selected items
- The selection count is displayed at the bottom of the treeview when 2+ items are selected (e.g., "3 items selected")

### Context Menu on Multi-Selection

**Trigger:** Right-click on any selected item while 2+ items are multi-selected.

**Menu Resolution:**
1. Collect all currently multi-selected items with their HATEOAS `_links`
2. Compute the **intersection** of permitted actions across all selected items:
   - **Delete from Model** — shown only if every selected item has the `delete` link
3. If the intersection is empty, show no context menu

**Menu Items:**
| Action | Condition | Style |
|--------|-----------|-------|
| Delete from Model (N items) | All items have `delete` link | Danger |

where N is the count of selected items.

**Right-click on unselected item:** If the user right-clicks an item that is NOT part of the multi-selection, the multi-selection is cleared and the existing single-item context menu appears for that item.

### Bulk Delete from Model

**Trigger:** User clicks "Delete from Model (N items)" in the multi-select context menu.

**Behaviour:**
- Show a confirmation dialog with danger styling:
  - Title: "Delete N items from Model"
  - Message: "This will permanently delete N items from the entire model. They will be removed from ALL views and all associated relations will be deleted. This cannot be undone."
  - List the names of all items being deleted (scrollable if many)
  - Confirm button text: "Delete N items"
- On confirm: execute the delete mutation for each selected item sequentially (to respect event sourcing consistency), using each item's `delete` model link
- On cancel: close dialog, selection remains
- After completion, all deleted items disappear from the treeview and all views
- If any individual delete fails mid-batch, stop and show an error indicating which items were deleted and which failed

### Drag Multi-Selection onto Canvas

**Trigger:** User drags any multi-selected treeview item onto the canvas.

**Behaviour:**
1. All multi-selected items are included in the drop operation, not just the item under the cursor
2. Items that are already present in the active view are **skipped** (not re-added, no error)
3. Items not in the active view are added at positions offset from the drop point:
   - First item at the drop coordinates
   - Subsequent items offset vertically (stacked below) to avoid overlapping
4. The drop is only allowed if the user has edit permission on the current view (`canEdit(currentView)`)
5. Each item type uses its existing add-to-view mutation (component, capability, origin entity)
6. After the drop, the canvas shows all newly added items at their assigned positions

**Drag visual:** The drag ghost indicates multi-selection (e.g., a badge showing the count of items being dragged).

## Scenarios

### Scenario 1: Ctrl+click to select multiple applications
**Given** the Applications section contains App-A, App-B, and App-C
**When** the user clicks App-A (plain click)
**And** Ctrl+clicks App-C
**Then** both App-A and App-C are visually selected
**And** the selection count shows "2 items selected"

### Scenario 2: Shift+click to select a range of applications
**Given** the Applications section contains App-A, App-B, App-C, App-D (in order)
**When** the user clicks App-A (plain click)
**And** Shift+clicks App-C
**Then** App-A, App-B, and App-C are all selected
**And** App-D is not selected

### Scenario 3: Cross-section multi-select via Ctrl+click
**Given** the user Ctrl+clicks App-A in Applications
**And** Ctrl+clicks Capability-X in Capabilities
**And** Ctrl+clicks Vendor-1 in Vendors
**When** the user right-clicks on any of the three selected items
**Then** the context menu shows "Delete from Model (3 items)" if all three have the `delete` link

### Scenario 4: Shift+click range is scoped to a single section
**Given** the user clicks App-A in Applications
**When** the user Shift+clicks Capability-X in Capabilities
**Then** only Capability-X is selected (Shift range does not cross sections)
**And** App-A is deselected (plain selection replaced)

### Scenario 5: Shift+click in capabilities tree respects collapsed state
**Given** the Capabilities section shows:
  - Cap-1 (expanded)
    - Cap-1.1
    - Cap-1.2
  - Cap-2 (collapsed, has hidden children)
  - Cap-3
**When** the user clicks Cap-1.1
**And** Shift+clicks Cap-3
**Then** Cap-1.1, Cap-1.2, Cap-2, and Cap-3 are selected
**And** Cap-2's collapsed children are NOT selected

### Scenario 6: Context menu with mixed permissions
**Given** the user has multi-selected 2 applications and 1 capability
**And** the user has delete permission on the applications but NOT on the capability
**When** the user right-clicks on one of the selected items
**Then** no context menu appears
**Because** "Delete from Model" requires the `delete` link on **every** selected item, and no other bulk action is available

### Scenario 7: Right-click on unselected item clears multi-selection
**Given** 3 items are multi-selected
**When** the user right-clicks on an item that is NOT part of the selection
**Then** the multi-selection is cleared
**And** the standard single-item context menu appears for the right-clicked item

### Scenario 8: Drag multi-selected items onto canvas
**Given** the user has multi-selected App-A, App-B, and App-C in the treeview
**And** App-B is already in the current view
**And** the user has edit permission on the current view
**When** the user drags the selection onto the canvas
**Then** App-A and App-C are added to the view at offset positions near the drop point
**And** App-B is skipped (already in view)
**And** the canvas shows all three items

### Scenario 9: Drag multi-selection without edit permission
**Given** the user has multi-selected 2 items
**And** the current view is private and the user is NOT the owner
**When** the user attempts to drag the selection onto the canvas
**Then** the drop is rejected (no items are added)

### Scenario 10: Plain click resets multi-selection
**Given** the user has multi-selected App-A, App-B, and Cap-1
**When** the user plain-clicks App-C (no modifier keys)
**Then** the multi-selection is cleared
**And** only App-C is selected
**And** the detail panel updates to show App-C's details

## UX Details

### Selection State
- Multi-selection state is managed locally in the navigation tree (not in the global Zustand store)
- The global `selectedNodeId` / `selectedCapabilityId` in the app store is cleared when multi-selection is active (2+ items), since the detail panel cannot show multiple items
- When multi-selection is reduced to exactly 1 item (via Ctrl+click deselect), that item becomes the active single selection again and the detail panel updates

### Keyboard
- Ctrl (Windows/Linux) and Cmd (macOS) are treated equivalently as the toggle modifier
- Shift is the range modifier
- Escape clears the multi-selection

### Drag Behaviour
- Only items NOT already in the current view participate in the drag operation
- If all selected items are already in the view, the drag is a no-op
- The `dataTransfer` carries a JSON payload identifying all selected item IDs grouped by type, rather than the single-ID keys used for single-item drag

## HATEOAS Contract
No new API endpoints or link relations are needed. The feature reuses:
- `delete` link on model entities (components, capabilities, origin entities) for "Delete from Model"

The frontend determines available bulk actions by intersecting the `_links` present on all selected items. The backend remains the single source of truth for authorization.

## Checklist
- [x] Specification ready
- [x] Multi-select state management (Ctrl+click toggle, Shift+click range, plain click reset)
- [x] Visual selection feedback for multi-selected treeview items
- [x] Selection count indicator
- [x] Multi-select context menu with HATEOAS link intersection
- [x] Bulk "Delete from Model" with confirmation dialog and item list
- [x] Sequential execution for model deletes
- [x] Error handling for partial batch failures
- [x] Multi-item drag-and-drop onto canvas
- [x] Skip items already in view during drag-drop
- [x] Offset positioning for multiple dropped items
- [x] Right-click on unselected item clears multi-selection
- [x] Escape key clears multi-selection
- [x] Single-item context menu preserved for single-select
- [x] Unit tests for multi-select state logic
- [x] Unit tests for context menu link intersection
- [x] Unit tests for drag payload construction
- [ ] Build passes (`go test ./...`, `npm test -- --run`)
- [ ] User sign-off
