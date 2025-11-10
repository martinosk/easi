# Spec 008: Multiple Views Management

## Status
pending

## Overview
Enable users to create, manage, and switch between multiple views of their application model. Each view represents a different perspective or subset of components, allowing users to organize complex models into manageable, focused workspaces.

## Business Context
As application models grow in complexity, users need the ability to:
- Create multiple views to focus on different aspects (e.g., "User Management", "Payment Flow", "Reporting")
- Switch between views quickly
- Organize components into logical groupings
- Reduce visual clutter by showing only relevant components per view

This follows the same mental model as tabs in a browser, views in a CAD tool, or diagrams in a modeling tool.

## Functional Requirements

### View Creation
- [x] Add "+" button in Views section header of navigation tree to create a new view
- [x] Prompt for view name (required, 1-100 characters)
- [x] New view starts empty (no components)

### View Management
- [x] Rename existing view inline in the tree menu (double-click or F2 to edit)
- [x] Delete view via context menu (right-click) with confirmation dialog
- [x] Set view as default via context menu (right-click)
- [x] Active view clearly indicated in treeview
- [x] Default view shows star indicator (⭐)

### View Switching
- [x] Switch between views via:
  - [x] Navigation tree (Spec 007)
  - [ ] View selector tabs in main UI
- [ ] Switching saves current canvas state (pan, zoom)
- [x] Canvas updates to show components in selected view
- [ ] View state persists across sessions

### Component-View Association
- [x] Components can exist in zero, one, or multiple views
- [x] When adding new component, it's automatically added to the current active view
- [x] Canvas only displays components that are in the current view
- [x] Components not in the current view can be dragged from the navigation tree onto the canvas to add them to the view
- [x] Existing components can be added to additional views via drag-and-drop from tree to canvas
- [x] Components can be removed from a view without deleting the component (via context menu or delete key while selected)
- [x] Deleting a component removes it from all views

### Default View Behavior
- [ ] One view marked as "default" (initially the first created)


### Data Flow
```
Command: Create View
  INBOUND ← Screen: View Manager
  OUTBOUND → Event: View Created

Command: Rename View
  INBOUND ← Screen: View Manager
  OUTBOUND → Event: View Renamed

Command: Delete View
  INBOUND ← Screen: View Manager
  OUTBOUND → Event: View Deleted

Command: Add Component To View
  INBOUND ← Screen: Canvas / Component Manager
  OUTBOUND → Event: Component Added To View

Command: Remove Component From View
  INBOUND ← Screen: Canvas / Component Manager
  OUTBOUND → Event: Component Removed From View

Command: Set Default View
  INBOUND ← Screen: View Manager
  OUTBOUND → Event: Default View Changed

Event: View Created
  OUTBOUND → ReadModel: View List

ReadModel: View List
  INBOUND ← Event: View Created
  INBOUND ← Event: View Renamed
  INBOUND ← Event: View Deleted
  OUTBOUND → Screen: Navigation Tree
  OUTBOUND → Screen: View Selector

ReadModel: View Detail
  INBOUND ← Event: View Created
  INBOUND ← Event: Component Added To View
  INBOUND ← Event: Component Removed From View
  OUTBOUND → Screen: Canvas
```

## Business Rules / Invariants
- [ ] View name must be 1-100 characters
- [ ] View name must be unique
- [ ] Cannot delete the default view
- [ ] Only one view can be default

## Non-Functional Requirements
- [ ] View switching completes within 200ms for models with 100+ components
- [ ] Support up to 50 views
- [ ] Support up to 500 components per view

## Test Plan

### Unit Tests
- [ ] Test view aggregate invariants
- [ ] Test view name validation
- [ ] Test cannot delete last view
- [ ] Test default view logic
- [ ] Test component-view associations

### Integration Tests
- [ ] Test create view via API
- [ ] Test rename view via API
- [ ] Test delete view via API (success and conflict cases)
- [ ] Test add component to view
- [ ] Test remove component from view
- [ ] Test set default view
- [ ] Test get views
- [ ] Test view events update read models correctly

### UI Tests
- [ ] Test create new view workflow
- [ ] Test switch between views
- [ ] Test rename view
- [ ] Test delete view with confirmation
- [ ] Test cannot delete last view
- [ ] Test add component to multiple views
- [ ] Test remove component from view
- [ ] Test default view indicator
- [ ] Test view state persists across refresh

## Sign-off
- [ ] Developer: Implementation complete
- [ ] User: Approved for completion
