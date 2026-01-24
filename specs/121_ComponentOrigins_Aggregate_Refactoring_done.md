# ComponentOrigins Aggregate Refactoring

**Status**: done

**Related**: Spec 117 (Portfolio Metadata Foundation)

## Description

Refactor origin relationships (acquired-via, purchased-from, built-by) from three separate per-relationship aggregates to a single `ComponentOrigins` aggregate per component. This fixes architectural issues where domain invariants are enforced in the wrong layer.

## Purpose

The current design has fundamental DDD violations:

1. **Invariant in wrong layer**: "One origin relationship per type per component" is enforced by handlers querying read models, not by aggregates
2. **Bug**: Changing from Entity A to Entity B fails with 500 error when a soft-deleted relationship to Entity B exists (unique constraint includes soft-deleted records)
3. **Anemic aggregates**: Each relationship aggregate only knows about itself, cannot enforce cross-relationship invariants

## Dependencies

- Spec 117: Portfolio Metadata Foundation (existing implementation to refactor)

## Architecture

### Current Design (Problems)

```
AcquiredViaRelationship (aggregate per relationship)
├── id: UUID (random)
├── acquiredEntityId
├── componentId
└── notes

Handler:
1. Query read model for existing relationships  ← WRONG: domain logic using read model
2. If exists && !replaceExisting → error
3. If exists && replaceExisting → delete via repository
4. Create new aggregate
5. Save
```

**Issues:**
- Handler contains domain logic
- Read model queried for invariant enforcement
- Soft-delete in read model causes unique constraint violation on re-link

### New Design (Solution)

```
ComponentOrigins (aggregate per component)
├── id: ComponentID (aggregate ID = component ID)
├── acquiredVia: OriginLink | empty     ← at most one (structural invariant)
├── purchasedFrom: OriginLink | empty   ← at most one (structural invariant)
└── builtBy: OriginLink | empty         ← at most one (structural invariant)

Handler:
1. Load aggregate by componentID (or create if not exists)
2. Call aggregate.SetAcquiredVia(entityID, notes)  ← domain logic in aggregate
3. Save
```

**Benefits:**
- Invariant enforced by structure (single fields, not slices)
- No read model queries in domain logic
- Idempotency built-in
- No soft-delete complexity

### Domain Model

**Bounded Context:** `architecturemodeling`

**OriginLink Value Object:** `backend/internal/architecturemodeling/domain/valueobjects/origin_link.go`

```go
type OriginLink struct {
    entityID string
    notes    Notes
    linkedAt time.Time
}

func NewOriginLink(entityID string, notes Notes, linkedAt time.Time) OriginLink
func EmptyOriginLink() OriginLink
func (o OriginLink) EntityID() string
func (o OriginLink) Notes() Notes
func (o OriginLink) LinkedAt() time.Time
func (o OriginLink) IsEmpty() bool
```

**ComponentOrigins Aggregate:** `backend/internal/architecturemodeling/domain/aggregates/component_origins.go`

```go
type ComponentOrigins struct {
    AggregateRoot
    componentID   ComponentID
    acquiredVia   OriginLink  // at most one
    purchasedFrom OriginLink  // at most one
    builtBy       OriginLink  // at most one
    createdAt     time.Time
    isDeleted     bool
}

// Aggregate ID is the componentId - natural scoping
func NewComponentOrigins(componentID ComponentID) (*ComponentOrigins, error)

// Set or replace relationship - invariant enforced by single field
func (co *ComponentOrigins) SetAcquiredVia(entityID AcquiredEntityID, notes Notes) error
func (co *ComponentOrigins) SetPurchasedFrom(vendorID VendorID, notes Notes) error
func (co *ComponentOrigins) SetBuiltBy(teamID InternalTeamID, notes Notes) error

// Clear relationship
func (co *ComponentOrigins) ClearAcquiredVia() error
func (co *ComponentOrigins) ClearPurchasedFrom() error
func (co *ComponentOrigins) ClearBuiltBy() error

// Cascade from component deletion
func (co *ComponentOrigins) Delete() error
```

### Events

| Event | Description |
|-------|-------------|
| `ComponentOriginsCreated` | First origin set for a component |
| `AcquiredViaRelationshipSet` | Acquired-via set (no previous) |
| `AcquiredViaRelationshipReplaced` | Changed to different entity |
| `AcquiredViaNotesUpdated` | Same entity, notes changed |
| `AcquiredViaRelationshipCleared` | Relationship removed |
| `PurchasedFromRelationshipSet` | Purchased-from set (no previous) |
| `PurchasedFromRelationshipReplaced` | Changed to different vendor |
| `PurchasedFromNotesUpdated` | Same vendor, notes changed |
| `PurchasedFromRelationshipCleared` | Relationship removed |
| `BuiltByRelationshipSet` | Built-by set (no previous) |
| `BuiltByRelationshipReplaced` | Changed to different team |
| `BuiltByNotesUpdated` | Same team, notes changed |
| `BuiltByRelationshipCleared` | Relationship removed |
| `ComponentOriginsDeleted` | Component deleted (cascade) |

### Commands

| Command | Description |
|---------|-------------|
| `SetAcquiredVia` | Set or replace acquired-via relationship |
| `ClearAcquiredVia` | Remove acquired-via relationship |
| `SetPurchasedFrom` | Set or replace purchased-from relationship |
| `ClearPurchasedFrom` | Remove purchased-from relationship |
| `SetBuiltBy` | Set or replace built-by relationship |
| `ClearBuiltBy` | Remove built-by relationship |

### API Changes

**Existing endpoints remain the same** - only internal implementation changes:

| Method | Path | Handler Change |
|--------|------|----------------|
| `POST` | `/components/{id}/origin/acquired-via` | Use `SetAcquiredVia` command |
| `DELETE` | `/origin-relationships/acquired-via/{id}` | Use `ClearAcquiredVia` command |
| `POST` | `/components/{id}/origin/purchased-from` | Use `SetPurchasedFrom` command |
| `DELETE` | `/origin-relationships/purchased-from/{id}` | Use `ClearPurchasedFrom` command |
| `POST` | `/components/{id}/origin/built-by` | Use `SetBuiltBy` command |
| `DELETE` | `/origin-relationships/built-by/{id}` | Use `ClearBuiltBy` command |

**Note:** Delete endpoints change semantically - the `{id}` parameter becomes the componentId, not a relationship UUID.

### Read Model Changes

**Tables remain the same**, but projection logic changes:

```go
// On AcquiredViaRelationshipSet or AcquiredViaRelationshipReplaced:
// UPSERT by componentId (handles soft-delete issue)
func (rm *AcquiredViaRelationshipReadModel) Upsert(ctx, dto) error

// On AcquiredViaRelationshipCleared:
// Hard delete (no soft-delete needed)
func (rm *AcquiredViaRelationshipReadModel) DeleteByComponentID(ctx, componentID) error
```

### Repository

**New:** `ComponentOriginsRepository`

```go
type ComponentOriginsRepository struct {
    EventSourcedRepository[*ComponentOrigins]
}

func (r *ComponentOriginsRepository) GetByID(ctx, componentID string) (*ComponentOrigins, error)
func (r *ComponentOriginsRepository) Save(ctx, aggregate *ComponentOrigins) error
```

**Key:** The aggregate ID is the componentId, so `GetByID(componentID)` loads all origin relationships for that component.

## Behaviour

### Setting First Origin Relationship

**Given** component "CRM System" has no origin relationships
**When** I set acquired-via to "TechCorp" with notes "Acquired in 2021 merger"
**Then** a ComponentOriginsCreated event is raised
**And** an AcquiredViaRelationshipSet event is raised
**And** the component shows "TechCorp" as its acquired-via origin

### Replacing Origin Relationship

**Given** component "CRM System" has acquired-via "TechCorp"
**When** I set acquired-via to "AcmeCo" with notes "Corrected origin"
**Then** an AcquiredViaRelationshipReplaced event is raised with old="TechCorp", new="AcmeCo"
**And** the component shows "AcmeCo" as its acquired-via origin
**And** "TechCorp" is no longer linked

### Replacing with Previously Linked Entity (Bug Fix)

**Given** component "CRM System" had acquired-via "TechCorp" (now cleared)
**And** component "CRM System" currently has acquired-via "AcmeCo"
**When** I set acquired-via back to "TechCorp"
**Then** an AcquiredViaRelationshipReplaced event is raised
**And** the component shows "TechCorp" as its acquired-via origin
**And** NO 500 error occurs (bug fixed)

### Idempotent Set (Same Entity, Same Notes)

**Given** component "CRM System" has acquired-via "TechCorp" with notes "Acquired 2021"
**When** I set acquired-via to "TechCorp" with notes "Acquired 2021"
**Then** no event is raised (idempotent, no change)

### Update Notes Only

**Given** component "CRM System" has acquired-via "TechCorp" with notes "Acquired 2021"
**When** I set acquired-via to "TechCorp" with notes "Acquired in Q1 2021 merger"
**Then** an AcquiredViaNotesUpdated event is raised
**And** the notes are updated

### Clearing Relationship

**Given** component "CRM System" has acquired-via "TechCorp"
**When** I clear the acquired-via relationship
**Then** an AcquiredViaRelationshipCleared event is raised
**And** the component shows no acquired-via origin

### Clearing Non-Existent Relationship

**Given** component "CRM System" has no acquired-via relationship
**When** I try to clear the acquired-via relationship
**Then** an error "no acquired-via relationship exists" is returned

### Multiple Origin Types

**Given** component "CRM System" has no origin relationships
**When** I set acquired-via to "TechCorp"
**And** I set built-by to "Platform Team"
**Then** the component shows both origins
**And** each type has at most one relationship (invariant maintained)

### Component Deletion Cascade

**Given** component "CRM System" has acquired-via "TechCorp" and built-by "Platform Team"
**When** the component is deleted
**Then** a ComponentOriginsDeleted event is raised
**And** all origin relationships are cleared from read model

## Implementation Strategy

Since the old events have never been used in production, no migration is needed. Simply replace the old implementation:

1. **Delete old code**: Remove old aggregates, events, commands, handlers, repository
2. **Implement new code**: ComponentOrigins aggregate with proper invariants
3. **Update projector**: Listen to new events, use Upsert/Delete instead of Insert/MarkAsDeleted

### Files to Delete

```
backend/internal/architecturemodeling/domain/aggregates/acquired_via_relationship.go
backend/internal/architecturemodeling/domain/aggregates/purchased_from_relationship.go
backend/internal/architecturemodeling/domain/aggregates/built_by_relationship.go
backend/internal/architecturemodeling/domain/events/acquired_via_relationship_created.go
backend/internal/architecturemodeling/domain/events/acquired_via_relationship_deleted.go
backend/internal/architecturemodeling/domain/events/purchased_from_relationship_created.go
backend/internal/architecturemodeling/domain/events/purchased_from_relationship_deleted.go
backend/internal/architecturemodeling/domain/events/built_by_relationship_created.go
backend/internal/architecturemodeling/domain/events/built_by_relationship_deleted.go
backend/internal/architecturemodeling/domain/valueobjects/acquired_via_relationship_id.go
backend/internal/architecturemodeling/domain/valueobjects/purchased_from_relationship_id.go
backend/internal/architecturemodeling/domain/valueobjects/built_by_relationship_id.go
backend/internal/architecturemodeling/application/commands/create_acquired_via_relationship.go
backend/internal/architecturemodeling/application/commands/delete_acquired_via_relationship.go
backend/internal/architecturemodeling/application/commands/create_purchased_from_relationship.go
backend/internal/architecturemodeling/application/commands/delete_purchased_from_relationship.go
backend/internal/architecturemodeling/application/commands/create_built_by_relationship.go
backend/internal/architecturemodeling/application/commands/delete_built_by_relationship.go
backend/internal/architecturemodeling/application/handlers/create_acquired_via_relationship_handler.go
backend/internal/architecturemodeling/application/handlers/delete_acquired_via_relationship_handler.go
backend/internal/architecturemodeling/application/handlers/create_purchased_from_relationship_handler.go
backend/internal/architecturemodeling/application/handlers/delete_purchased_from_relationship_handler.go
backend/internal/architecturemodeling/application/handlers/create_built_by_relationship_handler.go
backend/internal/architecturemodeling/application/handlers/delete_built_by_relationship_handler.go
backend/internal/architecturemodeling/infrastructure/repositories/acquired_via_relationship_repository.go
backend/internal/architecturemodeling/infrastructure/repositories/purchased_from_relationship_repository.go
backend/internal/architecturemodeling/infrastructure/repositories/built_by_relationship_repository.go
```

### Files to Create

```
backend/internal/architecturemodeling/domain/valueobjects/origin_link.go
backend/internal/architecturemodeling/domain/aggregates/component_origins.go
backend/internal/architecturemodeling/domain/aggregates/component_origins_test.go
backend/internal/architecturemodeling/domain/events/component_origins_events.go
backend/internal/architecturemodeling/application/commands/origin_commands.go
backend/internal/architecturemodeling/application/handlers/set_acquired_via_handler.go
backend/internal/architecturemodeling/application/handlers/clear_acquired_via_handler.go
backend/internal/architecturemodeling/application/handlers/set_purchased_from_handler.go
backend/internal/architecturemodeling/application/handlers/clear_purchased_from_handler.go
backend/internal/architecturemodeling/application/handlers/set_built_by_handler.go
backend/internal/architecturemodeling/application/handlers/clear_built_by_handler.go
backend/internal/architecturemodeling/infrastructure/repositories/component_origins_repository.go
```

### Files to Modify

```
backend/internal/architecturemodeling/application/projectors/origin_relationship_projector.go
backend/internal/architecturemodeling/application/readmodels/acquired_via_relationship_read_model.go
backend/internal/architecturemodeling/application/readmodels/purchased_from_relationship_read_model.go
backend/internal/architecturemodeling/application/readmodels/built_by_relationship_read_model.go
backend/internal/architecturemodeling/infrastructure/api/origin_relationship_handlers.go
backend/internal/architecturemodeling/module.go (wire up new handlers)
```

## Checklist

- [x] Specification approved
- [x] Old aggregates, events, commands, handlers, repositories deleted
- [x] OriginLink value object implemented
- [x] ComponentOrigins aggregate implemented
- [x] ComponentOrigins aggregate unit tests implemented
- [x] New events implemented (Set, Replaced, NotesUpdated, Cleared for each type)
- [x] New commands implemented
- [x] New handlers implemented
- [x] ComponentOriginsRepository implemented
- [x] Read model Upsert/DeleteByComponentID methods added
- [x] Projector updated for new events
- [x] API handlers updated to use new commands
- [x] All tests passing
- [x] Implementation complete
