# Cascade Deletion - Cross-Context Integration

## Description
Implement cascade deletion behavior when components or relations are deleted from the model, ensuring they are automatically removed from all architecture views through event-driven integration between bounded contexts.

## Purpose
Maintain consistency across the system by ensuring that when a component or relation is deleted from the ArchitectureModeling context, it is automatically removed from all views in the ArchitectureViews context.

## Bounded Contexts Involved

### ArchitectureModeling (Publisher)
- Core domain managing components and relations
- Publishes deletion events when entities are removed
- Outbound events: ApplicationComponentDeleted, ComponentRelationDeleted

### ArchitectureViews (Subscriber)
- Supporting domain managing visual representation in views
- Subscribes to deletion events from ArchitectureModeling
- Removes components/relations from all views when model entities are deleted

## Integration Pattern

### Event-Driven Integration (Publisher/Subscriber)
- Bounded contexts remain loosely coupled
- Integration happens asynchronously through events
- ArchitectureViews conforms to ArchitectureModeling's events
- No direct dependencies between contexts

## Business Rules

### Component Cascade Deletion
When a component is deleted from the model:
- Component must be removed from ALL architecture views that contain it
- Component positions in views are lost
- Relations in views involving the deleted component are handled separately (see relation cascade)
- Deletion is permanent and cannot be undone

### Relation Cascade Deletion
When a relation is deleted from the model:
- Relation must be removed from ALL architecture views that contain it
- Deletion is permanent and cannot be undone

### Component Deletion Triggers Relation Cascade
When a component is deleted:
- ALL relations where the component is source must be deleted
- ALL relations where the component is target must be deleted
- Each relation deletion triggers its own cascade to views
- This ensures referential integrity in the model

## Event Handling

### Listen to ApplicationComponentDeleted Event

**Event Source:** ArchitectureModeling bounded context

**Event Data:**
- Component ID
- Deletion timestamp
- Component name (for audit)

**Handler Behavior:**
- Query all architecture views to find which views contain the component
- For each view containing the component:
  - Issue RemoveComponentFromView command
  - Remove component position and visual properties
- Query all relations involving the component (source or target)
- For each relation:
  - Issue DeleteComponentRelation command
  - This triggers ComponentRelationDeleted event
- Update read models to reflect removal

**Outcome:**
- Component removed from all views
- All dependent relations deleted
- Views remain in consistent state

### Listen to ComponentRelationDeleted Event

**Event Source:** ArchitectureModeling bounded context

**Event Data:**
- Relation ID
- Source component ID
- Target component ID
- Deletion timestamp

**Handler Behavior:**
- Query all architecture views to find which views contain the relation
- For each view containing the relation:
  - Remove relation from view's visual representation
  - Update read model
- No commands needed (read model update only)

**Outcome:**
- Relation removed from all views
- Views remain in consistent state

## Commands in ArchitectureViews Context

### RemoveComponentFromView (Existing)
- Already implemented
- Purpose: Remove component from specific view
- Triggered by cascade deletion handler
- Raises ComponentRemovedFromView event

### RemoveRelationFromView (New)
- Purpose: Remove relation from specific view only (not from model)
- Input: View ID, Relation ID
- Result: RelationRemovedFromView event raised
- Used when user explicitly removes relation from view (not from model)

### DeleteComponentRelation (Enhancement)
- Purpose: Cascade delete relation when component is deleted
- Triggered by ApplicationComponentDeleted event handler
- Delegates to ArchitectureModeling DeleteComponentRelation command
- Maintains referential integrity

## Events in ArchitectureViews Context

### RelationRemovedFromView (New)
- Purpose: Signal that relation was removed from specific view
- Data: View ID, Relation ID, timestamp
- Raised when user removes relation from view (not model)
- Read model projector updates view state

## Read Model Changes

### ArchitectureViews Read Model

**Add Methods:**
- `GetViewsContainingComponent(componentID)`: Returns list of view IDs
- `GetViewsContainingRelation(relationID)`: Returns list of view IDs
- `RemoveRelationFromAllViews(relationID)`: Removes relation from all views

**Update Methods:**
- Read model must handle RelationRemovedFromView event
- Update view state to exclude removed relation

## Cascade Deletion Process Flow

### When Component is Deleted from Model

1. **ArchitectureModeling:** DeleteApplicationComponent command received
2. **ArchitectureModeling:** ApplicationComponentDeleted event raised and published
3. **ArchitectureViews:** Event handler receives ApplicationComponentDeleted
4. **ArchitectureViews:** Query views containing the component
5. **ArchitectureViews:** For each view, issue RemoveComponentFromView command
6. **ArchitectureViews:** Each RemoveComponentFromView raises ComponentRemovedFromView event
7. **ArchitectureViews:** Read model projector updates each view
8. **ArchitectureModeling:** Query relations involving the component
9. **ArchitectureModeling:** For each relation, issue DeleteComponentRelation command
10. **ArchitectureModeling:** Each deletion raises ComponentRelationDeleted event
11. **ArchitectureViews:** Handle each ComponentRelationDeleted event (see below)

### When Relation is Deleted from Model

1. **ArchitectureModeling:** DeleteComponentRelation command received
2. **ArchitectureModeling:** ComponentRelationDeleted event raised and published
3. **ArchitectureViews:** Event handler receives ComponentRelationDeleted
4. **ArchitectureViews:** Read model projector removes relation from all views
5. **ArchitectureViews:** Views updated to exclude the relation

## Event Bus Configuration

### Subscription Setup
- ArchitectureViews must subscribe to ApplicationComponentDeleted event
- ArchitectureViews must subscribe to ComponentRelationDeleted event
- Event handlers must be registered at application startup
- Event bus must support asynchronous event delivery

### Error Handling
- Event handlers must be idempotent (handle same event multiple times safely)
- Failed event processing should be retried with exponential backoff
- Dead letter queue for permanently failed events
- Logging for all event processing activities

## Consistency Considerations

### Eventual Consistency
- Deletion from model happens immediately (command processed)
- Removal from views happens asynchronously (via events)
- Brief window where component exists in views but not in model
- Read models eventually become consistent

### Handling Race Conditions
- User might interact with component in view while it's being deleted
- View operations on deleted components should fail gracefully
- Return appropriate error messages (e.g., "Component no longer exists")
- Frontend should refresh view state periodically or on errors

## API Endpoints for View Operations

### Remove Relation from View (New)

**Endpoint:** `DELETE /api/v1/views/{viewId}/relations/{relationId}`

**Purpose:** Remove relation from specific view only (not from model)

**Request Parameters:**
- `viewId` (path parameter, UUID, required): View identifier
- `relationId` (path parameter, UUID, required): Relation identifier

**Success Response:**
- Status: 204 No Content

**Error Responses:**
- 404 Not Found: View or relation not found in view
- 500 Internal Server Error: Unhandled server error

**Behavior:**
- Issues RemoveRelationFromView command
- Relation remains in model and other views
- Only visual representation in this view is removed

**HATEOAS:**
- Relation DTOs in view context include removeFromView link
- Link points to DELETE /api/v1/views/{viewId}/relations/{relationId}

### Remove Component from View (Existing)

**Endpoint:** `DELETE /api/v1/views/{viewId}/components/{componentId}`

**Status:** Already implemented, no changes required

**Purpose:** Remove component from specific view only (not from model)

## Checklist

### Event Handlers
- [ ] ApplicationComponentDeleted event handler created in ArchitectureViews
- [ ] ComponentRelationDeleted event handler created in ArchitectureViews
- [ ] Event handlers registered with event bus at startup
- [ ] Handlers query read model for affected views
- [ ] Handlers issue appropriate commands to remove from views

### Commands
- [ ] RemoveRelationFromView command created
- [ ] RemoveRelationFromView command handler implemented
- [ ] RelationRemovedFromView event created

### Cascade Logic
- [ ] ApplicationComponentDeleted handler removes component from all views
- [ ] ApplicationComponentDeleted handler triggers deletion of all relations
- [ ] ComponentRelationDeleted handler removes relation from all views
- [ ] Cascade deletion maintains referential integrity

### Read Model
- [ ] GetViewsContainingComponent method implemented
- [ ] GetViewsContainingRelation method implemented
- [ ] RemoveRelationFromAllViews method implemented
- [ ] Projector handles RelationRemovedFromView event

### API Layer
- [ ] DELETE /api/v1/views/{viewId}/relations/{relationId} endpoint implemented
- [ ] Endpoint returns correct status codes
- [ ] HATEOAS links include removeFromView for relations in views

### Event Bus
- [ ] Event subscriptions configured at startup
- [ ] Event handlers registered properly
- [ ] Error handling and retry logic implemented
- [ ] Event processing logged for debugging

### Testing
- [ ] Unit test: ApplicationComponentDeleted handler removes from all views
- [ ] Unit test: ComponentRelationDeleted handler removes from all views
- [ ] Integration test: Delete component cascades to relations
- [ ] Integration test: Delete component removes from multiple views
- [ ] Integration test: Delete relation removes from all views
- [ ] Integration test: Event bus delivers events to handlers
- [ ] Integration test: Idempotent event handling (duplicate events)
- [ ] Integration test: Race condition handling (operate on deleted component)
- [ ] End-to-end test: Delete component from model, verify removed from all views
- [ ] End-to-end test: Delete relation from model, verify removed from all views

### Documentation
- [ ] Event schemas documented
- [ ] Cascade deletion flow documented
- [ ] API documentation updated with new endpoint

### Final
- [ ] All tests passing
- [ ] Eventual consistency behavior verified
- [ ] User sign-off
