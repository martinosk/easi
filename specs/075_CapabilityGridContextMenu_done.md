# Capability Grid Context Menu

## User Need
Enterprise architects need to remove or delete capabilities directly from the business domain grid visualization without navigating to separate management screens. When managing multiple capabilities, they need efficient multi-select operations to reduce repetitive actions.

## Success Criteria
- Right-click on a capability in the grid opens a context menu with "Remove from Business Domain" and "Delete from Model" options
- "Remove from Business Domain" dissociates the L1 capability (and all children) from the current domain without deleting it
- "Delete from Model" permanently deletes the L1 capability and all children from the entire model
- Multi-select via Shift-click allows selecting multiple L1 capabilities
- Ctrl+A selects all capabilities in the active domain
- Context menu operations apply to all selected capabilities
- Confirmation dialogs clearly explain the scope of the action (including children)

## Existing Patterns to Reuse
- `ContextMenu` component (`/frontend/src/components/shared/ContextMenu.tsx`)
- `ConfirmationDialog` component for destructive actions
- `NodeContextMenu` pattern from canvas (`/frontend/src/features/canvas/components/context-menus/NodeContextMenu.tsx`)
- `useDomainCapabilities` hook with `dissociateCapability` method
- `deleteCapability` from app store (handles API call to `DELETE /api/v1/capabilities/{id}`)

## API Endpoints (Already Available)
- Dissociate: `DELETE {capability._links.dissociate}` - removes capability from domain
- Delete: `DELETE /api/v1/capabilities/{id}` - deletes capability from model

## Vertical Slices

### Slice 1: Single-Select Context Menu
Add right-click context menu to capability items in the grid.

- [x] Add `onContextMenu` handler to `NestedCapabilityItem` component
- [x] Track context menu state (position, target capability) in `DomainVisualizationPage`
- [x] Render `ContextMenu` with two options when triggered
- [x] "Remove from Business Domain" calls `dissociateCapability` for the L1 ancestor
- [x] "Delete from Model" opens confirmation dialog, then calls `deleteCapability` for the L1 ancestor
- [x] Resolve L1 ancestor: when right-clicking any capability (L1-L4), find its L1 parent
- [x] After remove/delete, refresh grid to reflect changes
- [x] Close context menu on action completion, outside click, or Escape

### Slice 2: Confirmation Dialog for Delete
Show confirmation before permanently deleting capabilities.

- [x] Create or reuse `DeleteConfirmationDialog` for capability deletion
- [x] Dialog message explains: "This will permanently delete [capability name] and all child capabilities from the model"
- [x] Show count of affected capabilities (L1 + all descendants)
- [x] Confirm button uses danger styling
- [x] Loading state during deletion
- [x] Toast notification on success/failure

### Slice 3: Multi-Select with Shift-Click
Enable selecting multiple capabilities for batch operations.

- [x] Add `selectedCapabilities` state (Set of CapabilityId) to `DomainVisualizationPage`
- [x] Shift-click on capability toggles selection
- [x] Visual indicator (border, highlight) for selected capabilities
- [x] Regular click (without Shift) clears selection and opens detail panel
- [x] Context menu on selected capability applies to all selected L1 capabilities
- [x] Confirmation dialog shows count of capabilities being removed/deleted

### Slice 4: Select All with Ctrl+A
Add keyboard shortcut to select all capabilities in active domain.

- [x] Add `keydown` event listener for Ctrl+A when grid is focused
- [x] Prevent default browser behavior (select all text)
- [x] Select all L1 capabilities in current domain
- [x] Visual feedback shows all capabilities selected
- [x] Escape key clears selection

## Out of Scope
- Drag-to-remove (dragging capability out of grid to remove)
- Cut/copy/paste operations
- Undo functionality
- Removing individual child capabilities (L2-L4) from domain - always operates on L1
- Context menu in the Capability Explorer sidebar
- Batch reassign to different domain

## Technical Notes
- Operations always target the L1 ancestor to maintain domain assignment consistency
- When deleting, the backend cascade-deletes all child capabilities
- When dissociating, only the domain assignment is removed; the capability tree remains intact
- Multi-select state is local to the visualization page (not persisted)
