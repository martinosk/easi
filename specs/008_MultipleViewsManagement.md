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
- [ ] System generates unique view ID
- [ ] First view created automatically when first component is added

### View Management
- [ ] List all views for the current bounded context
- [ ] Rename existing view
- [ ] Delete view (with confirmation)
- [ ] Cannot delete the last remaining view
- [ ] Active view clearly indicated in UI

### View Switching
- [ ] Switch between views via:
  - [ ] Navigation tree (Spec 007)
  - [ ] View selector dropdown/tabs in main UI
- [ ] Switching saves current canvas state (pan, zoom)
- [ ] Canvas updates to show components in selected view
- [ ] View state persists across sessions

### Component-View Association
- [ ] Components can exist in zero, one, or multiple views
- [ ] When adding new component, user chooses which view(s) to add it to
- [ ] Existing components can be added to additional views
- [ ] Components can be removed from a view without deleting the component
- [ ] Deleting a component removes it from all views

### Default View Behavior
- [ ] One view marked as "default" (initially the first created)
- [ ] Default view opens when user navigates to bounded context
- [ ] User can change which view is default

## Technical Requirements

### Domain Model (Event Sourcing)

#### Aggregates
```
View Aggregate:
- ViewId (value object, GUID)
- BoundedContextId (value object, GUID)
- ViewName (value object, string 1-100 chars)
- ComponentIds (collection of ComponentId value objects)
- IsDefault (bool)
- CreatedAt (DateTime)
- UpdatedAt (DateTime)
```

#### Commands
- [ ] **CreateView**: BoundedContextId, ViewName
- [ ] **RenameView**: ViewId, NewName
- [ ] **DeleteView**: ViewId
- [ ] **AddComponentToView**: ViewId, ComponentId
- [ ] **RemoveComponentFromView**: ViewId, ComponentId
- [ ] **SetDefaultView**: BoundedContextId, ViewId

#### Events
- [ ] **ViewCreated**: ViewId, BoundedContextId, ViewName, IsDefault, Timestamp
- [ ] **ViewRenamed**: ViewId, OldName, NewName, Timestamp
- [ ] **ViewDeleted**: ViewId, Timestamp
- [ ] **ComponentAddedToView**: ViewId, ComponentId, Timestamp
- [ ] **ComponentRemovedFromView**: ViewId, ComponentId, Timestamp
- [ ] **DefaultViewChanged**: BoundedContextId, ViewId, Timestamp

#### Read Models
- [ ] **ViewList**: ViewId, ViewName, IsDefault, ComponentCount, UpdatedAt
- [ ] **ViewDetail**: ViewId, ViewName, IsDefault, ComponentIds[], CreatedAt, UpdatedAt

### API Endpoints

#### View Management
```
POST /api/bounded-contexts/{contextId}/views
  Body: { name: string }
  Returns: 201 Created, { id: string, name: string, isDefault: bool }
  Links: self, components, set-default

GET /api/bounded-contexts/{contextId}/views
  Returns: 200 OK, [{ id, name, isDefault, componentCount, updatedAt, _links }]
  Links: self, create

GET /api/views/{viewId}
  Returns: 200 OK, { id, name, isDefault, componentIds, createdAt, updatedAt }
  Links: self, components, rename, delete, add-component, set-default

PUT /api/views/{viewId}/name
  Body: { name: string }
  Returns: 200 OK

DELETE /api/views/{viewId}
  Returns: 204 No Content
  Error: 409 Conflict if last view

PUT /api/views/{viewId}/default
  Returns: 200 OK

POST /api/views/{viewId}/components
  Body: { componentId: string }
  Returns: 201 Created

DELETE /api/views/{viewId}/components/{componentId}
  Returns: 204 No Content
```

### Frontend Implementation
- [ ] New React component: `ViewSelector`
- [ ] New React component: `ViewManager` (create, rename, delete UI)
- [ ] Update canvas to filter components by active view
- [ ] State management for active view
- [ ] localStorage for last active view per context
- [ ] Integrate with NavigationTree (Spec 007)

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
- [ ] View name must be unique within a bounded context
- [ ] Cannot delete the last view in a bounded context
- [ ] Only one view can be default per bounded context
- [ ] View must belong to exactly one bounded context
- [ ] Component can only be added to view if component exists
- [ ] Component can only be added to view if component is in same bounded context

## Non-Functional Requirements
- [ ] View switching completes within 200ms for models with 100+ components
- [ ] Support up to 50 views per bounded context
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
- [ ] Test get views for bounded context
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

## Future Enhancements
- View templates (pre-configured view types)
- Copy/duplicate view
- View descriptions
- View sharing/permissions
- Auto-organize components in view
- View bookmarks/favorites
- Bulk add components to view

## Dependencies
- Requires existing Component aggregate and events
- Required by Spec 007 (Hidable Tree Menu)
- May interact with relation drawing (components in different views)

## Sign-off
- [ ] Developer: Implementation complete
- [ ] User: Approved for completion
