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
- [ ] Button/action to create a new view
- [ ] Prompt for view name (required, 1-100 characters)
- [ ] New view starts empty (no components)

### View Management
- [ ] Rename existing view (add as functionality to the tree menu from spec 007)
- [ ] Delete view (with confirmation). Add as functionality to the tree menu from spec 007.
- [ ] Active view clearly indicated in treeview

### View Switching
- [ ] Switch between views via:
  - [ ] Navigation tree (Spec 007)
  - [ ] View selector tabs in main UI
- [ ] Switching saves current canvas state (pan, zoom)
- [ ] Canvas updates to show components in selected view
- [ ] View state persists across sessions

### Component-View Association
- [ ] Components can exist in zero, one, or multiple views
- [ ] When adding new component, user chooses which view(s) to add it to. It's always added to the view marked "default".
- [ ] Existing components can be added to additional views
- [ ] Components can be removed from a view without deleting the component
- [ ] Deleting a component removes it from all views

### Default View Behavior
- [ ] One view marked as "default" (initially the first created)

## Technical Requirements
#### Commands
- [ ] **CreateView**: ViewName
- [ ] **RenameView**: ViewId, NewName
- [ ] **DeleteView**: ViewId
- [ ] **AddComponentToView**: ViewId, ComponentId
- [ ] **RemoveComponentFromView**: ViewId, ComponentId
- [ ] **SetDefaultView**: ViewId

#### Events
- [ ] **ViewCreated**: ViewId, ViewName, IsDefault, Timestamp
- [ ] **ViewRenamed**: ViewId, OldName, NewName, Timestamp
- [ ] **ViewDeleted**: ViewId, Timestamp
- [ ] **ComponentAddedToView**: ViewId, ComponentId, Timestamp
- [ ] **ComponentRemovedFromView**: ViewId, ComponentId, Timestamp
- [ ] **DefaultViewChanged**: ViewId, Timestamp



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
