# Delete Operations - Backend Domain Model

## Description
Implement delete operations for application components and component relations in the ArchitectureModeling bounded context using CQRS with event sourcing patterns.

## Purpose
Enable deletion of components and relations from the model through proper domain commands and events, maintaining event sourcing integrity and audit trails.

## Bounded Context
**ArchitectureModeling** - Core domain responsible for managing application components and their relations.

## Business Rules

### Component Deletion
- A component can be deleted from the model
- Component deletion is permanent (but soft-deleted with audit trail)
- Deleting a component must cascade to all relations where it participates (either as source or target)
- Component deletion must propagate to other bounded contexts via events

### Relation Deletion
- A relation can be deleted independently from its components
- Relation deletion is permanent (but soft-deleted with audit trail)
- Relation deletion must propagate to other bounded contexts via events

### Deletion Strategy
- Use soft delete approach to preserve event sourcing audit trail
- Deleted aggregates maintain an `isDeleted` flag set to true
- Read models filter out deleted entities
- Event store preserves full history of all changes
- Deleted entities remain queryable for audit purposes but excluded from normal operations

## Domain Model Changes

### Commands Required

**DeleteApplicationComponent**
- Purpose: Remove an application component from the model
- Input: Component ID (globally unique identifier)
- Validation: Component must exist and not already be deleted
- Result: ApplicationComponentDeleted event raised

**DeleteComponentRelation**
- Purpose: Remove a relation between components
- Input: Relation ID (globally unique identifier)
- Validation: Relation must exist and not already be deleted
- Result: ComponentRelationDeleted event raised

### Events Required

**ApplicationComponentDeleted**
- Purpose: Signal that a component has been removed from the model
- Data: Component ID, deletion timestamp, component name (for audit)
- Subscribers: ArchitectureViews bounded context (for cascade removal from views)
- Outbound event from ArchitectureModeling context

**ComponentRelationDeleted**
- Purpose: Signal that a relation has been removed from the model
- Data: Relation ID, source component ID, target component ID, deletion timestamp
- Subscribers: ArchitectureViews bounded context (for cascade removal from views)
- Outbound event from ArchitectureModeling context

### Aggregate Changes

**ApplicationComponent Aggregate**
- Add Delete() method that raises ApplicationComponentDeleted event
- Maintain isDeleted boolean flag
- Prevent modifications to deleted aggregates

**ComponentRelation Aggregate**
- Add Delete() method that raises ComponentRelationDeleted event
- Maintain isDeleted boolean flag
- Prevent modifications to deleted aggregates

## API Endpoints

### Delete Component from Model

**Endpoint:** `DELETE /api/v1/components/{id}`

**Purpose:** Delete component from the entire model

**Request Parameters:**
- `id` (path parameter, UUID, required): Component identifier

**Success Response:**
- Status: 204 No Content
- No response body

**Error Responses:**
- 404 Not Found: Component does not exist or already deleted
- 500 Internal Server Error: Unhandled server error

**Behavior:**
- Issues DeleteApplicationComponent command
- Returns immediately after command accepted
- Cascade deletion handled asynchronously via events

**HATEOAS:**
- Component DTOs must include a `delete` link when component is deletable
- Link points to DELETE /api/v1/components/{id}

### Delete Relation from Model

**Endpoint:** `DELETE /api/v1/relations/{id}`

**Purpose:** Delete relation from the entire model

**Request Parameters:**
- `id` (path parameter, UUID, required): Relation identifier

**Success Response:**
- Status: 204 No Content
- No response body

**Error Responses:**
- 404 Not Found: Relation does not exist or already deleted
- 500 Internal Server Error: Unhandled server error

**Behavior:**
- Issues DeleteComponentRelation command
- Returns immediately after command accepted

**HATEOAS:**
- Relation DTOs must include a `delete` link when relation is deletable
- Link points to DELETE /api/v1/relations/{id}

## Command Handlers

### DeleteApplicationComponentHandler
- Receives DeleteApplicationComponent command
- Loads ApplicationComponent aggregate from event store
- Validates component exists and is not already deleted
- Calls Delete() method on aggregate
- Persists ApplicationComponentDeleted event to event store
- Returns success (no return value for delete operations)
- On error: throws appropriate exception (NotFound, etc.)

### DeleteComponentRelationHandler
- Receives DeleteComponentRelation command
- Loads ComponentRelation aggregate from event store
- Validates relation exists and is not already deleted
- Calls Delete() method on aggregate
- Persists ComponentRelationDeleted event to event store
- Returns success
- On error: throws appropriate exception

## Projector Updates

### ApplicationComponentProjector
- Must handle ApplicationComponentDeleted event
- Update read model to mark component as deleted
- Remove component from query results (filter by isDeleted = false)
- Maintain component record for audit trail purposes

### ComponentRelationProjector
- Must handle ComponentRelationDeleted event
- Update read model to mark relation as deleted
- Remove relation from query results (filter by isDeleted = false)
- Maintain relation record for audit trail purposes

## Read Model Changes

### Components Table
- Add `is_deleted` boolean column (default: false)
- Add `deleted_at` timestamp column (nullable)
- Update all queries to filter WHERE is_deleted = false

### Relations Table
- Add `is_deleted` boolean column (default: false)
- Add `deleted_at` timestamp column (nullable)
- Update all queries to filter WHERE is_deleted = false

## Integration Requirements

### Event Publishing
- ApplicationComponentDeleted event must be published to event bus
- ComponentRelationDeleted event must be published to event bus
- Events are consumed by ArchitectureViews bounded context
- Event schema must include all necessary data for subscribers

### Idempotency
- DELETE operations should be idempotent
- Deleting an already-deleted resource returns 204 No Content (not 404)
- Command handlers check isDeleted flag before processing

## Error Handling

### Domain Exceptions
- Map to appropriate HTTP status codes in API layer
- NotFound exceptions → 404 Not Found
- Validation exceptions → 400 Bad Request
- Unhandled exceptions → 500 Internal Server Error

### Error Response Format
- Use existing ErrorResponse DTO
- Include error type and descriptive message
- Optionally include details for debugging

## OpenAPI Specification Updates

### DELETE /api/v1/components/{id}
- Add endpoint definition
- Document request parameters
- Document response codes (204, 404, 500)
- Include example requests and responses
- Document cascade deletion behavior in description

### DELETE /api/v1/relations/{id}
- Add endpoint definition
- Document request parameters
- Document response codes (204, 404, 500)
- Include example requests and responses

### HATEOAS Links
- Update component schema to include delete link
- Update relation schema to include delete link
- Document link relations in OpenAPI spec

## Checklist

### Domain Model
- [x] DeleteApplicationComponent command created
- [x] DeleteComponentRelation command created
- [x] ApplicationComponentDeleted event created
- [x] ComponentRelationDeleted event created
- [x] ApplicationComponent.Delete() method implemented
- [x] ComponentRelation.Delete() method implemented
- [x] Aggregates maintain isDeleted flag

### Command Handlers
- [x] DeleteApplicationComponentHandler implemented
- [x] DeleteComponentRelationHandler implemented
- [x] Handlers validate aggregate existence
- [x] Handlers check isDeleted flag for idempotency
- [x] Handlers persist events to event store

### Projectors
- [x] ApplicationComponentProjector handles ApplicationComponentDeleted
- [x] ComponentRelationProjector handles ComponentRelationDeleted
- [x] Read models updated to mark entities as deleted
- [x] Queries filter out deleted entities

### Database Schema
- [x] Add is_deleted column to components table
- [x] Add deleted_at column to components table
- [x] Add is_deleted column to relations table
- [x] Add deleted_at column to relations table
- [x] Create database migration script

### API Layer
- [x] DELETE /api/v1/components/{id} endpoint implemented
- [x] DELETE /api/v1/relations/{id} endpoint implemented
- [x] Endpoints return correct status codes
- [x] Error responses follow standard format
- [x] HATEOAS delete links added to DTOs

### Event Bus
- [x] ApplicationComponentDeleted event published to bus
- [x] ComponentRelationDeleted event published to bus
- [x] Event schemas documented

### OpenAPI
- [ ] DELETE /api/v1/components/{id} documented
- [ ] DELETE /api/v1/relations/{id} documented
- [ ] Response schemas documented
- [ ] HATEOAS links documented
- [ ] OpenAPI spec generation script updated

### Testing
- [ ] Unit test: Delete component command handler
- [ ] Unit test: Delete relation command handler
- [ ] Unit test: Projector handles delete events
- [ ] Integration test: DELETE endpoint returns 204
- [ ] Integration test: DELETE non-existent returns 404
- [ ] Integration test: Idempotent deletion (delete twice returns 204)
- [ ] Integration test: Deleted entities excluded from queries
- [ ] Integration test: Read model updated after deletion
- [ ] Integration test: Events published to event bus

### Documentation
- [ ] API documentation updated with delete operations
- [ ] Event schemas documented
- [ ] Domain model documentation updated

### Final
- [ ] All tests passing
- [ ] OpenAPI spec generated successfully
- [ ] User sign-off
