# Capability UI Consistency

## Description
Resolve UI inconsistencies between capability and application component interactions to provide a unified, predictable user experience.

## Current Issues

### 1. Dialog Management Inconsistency
- **Capability edit dialog**: Rendered inline within NavigationTree component, appears to open from the tree
- **Component edit dialog**: Managed via centralized DialogManager, opens as a proper modal overlay
- **Impact**: Users experience different dialog behaviors for similar operations

### 2. Treeview Visibility Inconsistency
- **Application components**: Always visible in treeview; items not in current view are grayed out (opacity 0.5, italic) with "(not in current view)" tooltip
- **Capabilities**: Only visible when the Capabilities section is expanded; no visual distinction for capabilities not on canvas
- **Impact**: Users cannot easily see which capabilities exist but aren't placed on the canvas

### 3. Context Menu Inconsistency
- **Capability (tree)**: Edit, Add Expert, Add Tag, Delete
- **Capability (canvas)**: Only "Remove from Canvas"
- **Component (tree)**: Rename, Delete from Model
- **Component (canvas)**: Delete from View, Delete from Model
- **Impact**: Different mental models required for similar operations

## Requirements

### Dialog Management
- [x] Move EditCapabilityDialog rendering from NavigationTree to DialogManager
- [x] Use the same dialog opening pattern as EditComponentDialog
- [x] Both dialogs should render as centered modal overlays

### Treeview Visibility
- [x] Show all capabilities in tree regardless of canvas presence
- [x] Apply grayed-out styling (opacity 0.5, italic) to capabilities not on current canvas
- [x] Add "(not on canvas)" tooltip suffix for capabilities not on current view
- [x] Capabilities not on canvas remain draggable (same as current behavior)

### Canvas Focus on Selection
- [x] When a capability is selected in the treeview, center the canvas view on that capability node
- [x] Use the same focus behavior as component selection (pan and highlight)

### Context Menu Consistency
- [x] Standardize canvas context menu structure for both capabilities and components:
  - "Remove from View" (removes from current view/canvas only)
  - "Delete from Model" (deletes entirely - marked as danger)
- [x] Standardize tree context menu structure:
  - "Edit" (opens edit dialog)
  - "Delete from Model" (deletes entirely - marked as danger)
- [x] Move capability-specific actions (Add Expert, Add Tag) to the edit dialog instead of context menu

## Checklist
- [x] Specification ready
- [x] Dialog management refactored
- [x] Treeview visibility implemented
- [x] Canvas focus on selection implemented
- [x] Context menus unified
- [x] Unit tests implemented and passing
- [x] User sign-off
