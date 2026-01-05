# Spec 108: Private Views

## Status
done

## Overview
Enable users to create private views that only they can edit. All users can view private views (visibility), but only the creator can modify them (editability). Public views remain fully editable by all users with appropriate role permissions.

## Business Context
As teams grow and more users interact with the architecture model, there's a need for:
- Personal workspaces where users can organize views without affecting others
- Protection of individual work from unintended modifications
- Clear ownership and accountability for views
- Flexibility to share work by making private views public

## Requirements Summary
| View Type | View (Read) | Edit | Delete |
|-----------|-------------|------|--------|
| Private   | All users | Creator only | Creator only |
| Public    | All users | All users (with role permission) | All users (with role permission) |

## Functional Requirements

### View Visibility
- [x] New views default to **private**
- [x] Private views are visible in the view list for all users
- [x] Private views display the creator's name/email next to the view name
- [x] UI shows a visual indicator (lock icon) for private views
- [x] Public views show no special indicator (or an "unlocked" icon for clarity)

### View Ownership
- [x] Every view has an owner (user ID + email)
- [x] Ownership is set to the creator when a view is created
- [x] Ownership can only transfer when an admin makes a private view public

### Edit Permissions
- [x] **Private views**: Only the owner can edit (rename, modify components, change settings)
- [x] **Public views**: Any user with `views:write` permission can edit
- [x] Stakeholders (read-only role) can create, edit, and delete their OWN private views
- [x] Edit/delete buttons hidden in UI for views the user cannot modify

### Visibility Toggle
- [x] Owner can convert their private view to public
- [x] Owner can convert their public view back to private
- [x] Toggle available via context menu or view settings
- [x] When owner makes view public, they remain the owner

### Admin Override
- [x] Admins can make any private view public (for cleanup when users leave)
- [x] When admin makes someone else's private view public, ownership transfers to the admin
- [x] Admins cannot edit/delete private views directly; they must make them public first

### Migration
- [x] Existing views become public with no owner (backward compatible)
- [x] No data loss or breaking changes

## Data Flow

### CQRS Element Types
```
Command: Create View
  INBOUND ← Screen: View Manager
  OUTBOUND → Event: View Created (visibility=Private, owner from actor metadata)

Command: Change View Visibility
  INBOUND ← Screen: View Settings / Context Menu
  OUTBOUND → Event: View Visibility Changed

Event: View Created
  Payload: viewId, name, description, isPrivate=true, createdAt
  Owner: Derived from event actor metadata (actor_id, actor_email)
  OUTBOUND → ReadModel: View List

Event: View Visibility Changed
  Payload: viewId, isPrivate
  Actor metadata: who performed the change
  Note: When admin makes private→public and admin != current owner, ownership transfers to actor
  OUTBOUND → ReadModel: View List

ReadModel: View List
  INBOUND ← Event: View Created
  INBOUND ← Event: View Visibility Changed
  INBOUND ← Event: View Renamed
  INBOUND ← Event: View Deleted
  OUTBOUND → Screen: Navigation Tree (with owner info)
  OUTBOUND → Screen: View Selector (with private/public indicator)
```

## Domain Model Changes

### New Value Objects
- **ViewOwner**: Encapsulates owner identity (userID + email)
  - Validation: userID required, non-empty
  - Methods: `UserID()`, `Email()`, `Equals()`

- **ViewVisibility**: Encapsulates visibility state
  - Values: `Private`, `Public`
  - Methods: `IsPrivate()`, `IsPublic()`

### Aggregate Changes (ArchitectureView)
- Add field: `owner ViewOwner`
- Add field: `visibility ViewVisibility`
- Add method: `CanBeEditedBy(actorID string, isAdmin bool) bool`
- Add method: `MakePublic(newOwner ViewOwner) error`
- Add method: `MakePrivate() error`
- Update constructor: Accept owner parameter, default visibility to Private

### New Events
- **ViewVisibilityChanged**: Captures visibility toggle and ownership transfer
  - Payload fields: ViewID, IsPrivate
  - Actor metadata (from event infrastructure): who performed the change
  - Ownership transfer logic: When admin (actor) makes another user's private view public, ownership transfers to actor

## API Changes

### Modified Endpoints
- `POST /api/v1/views` - Sets owner from actor context, defaults to private
- `GET /api/v1/views` - Returns `isPrivate`, `ownerUserId`, `ownerEmail` fields
- `GET /api/v1/views/{id}` - Returns `isPrivate`, `ownerUserId`, `ownerEmail` fields
- All write endpoints - Return 403 if not authorized based on ownership

### New Endpoint
- `PATCH /api/v1/views/{id}/visibility`
  - Request: `{ "isPrivate": boolean }`
  - Response: 204 No Content
  - Authorization: Owner can toggle; Admin can make public (transfers ownership)

### Response DTO Updates
```json
{
  "id": "view-123",
  "name": "My View",
  "isPrivate": true,
  "ownerUserId": "user-456",
  "ownerEmail": "user@example.com",
  "_links": {
    "self": "/api/v1/views/view-123",
    "update": "/api/v1/views/view-123/name",
    "changeVisibility": "/api/v1/views/view-123/visibility"
  }
}
```

### HATEOAS Link Changes
- Include `update`, `delete`, `changeVisibility` links only if user has permission
- Frontend uses link presence to show/hide action buttons

## Database Changes

### Migration: Add ownership columns
```sql
ALTER TABLE architecture_views
  ADD COLUMN is_private BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN owner_user_id VARCHAR(255),
  ADD COLUMN owner_email VARCHAR(500);

CREATE INDEX idx_architecture_views_owner ON architecture_views(owner_user_id);

-- Existing views become public with no owner
UPDATE architecture_views SET is_private = false WHERE owner_user_id IS NULL;
```

## Frontend Changes

### Type Updates (types.ts)
```typescript
export interface View {
  id: ViewId;
  name: string;
  description?: string;
  isDefault: boolean;
  isPrivate: boolean;        // NEW
  ownerUserId?: string;      // NEW
  ownerEmail?: string;       // NEW
  components: ViewComponent[];
  // ... rest unchanged
}
```

### UI Updates
- **ViewSelector**: Show lock icon for private views, owner email on hover
- **NavigationTree**: Context menu shows "Make Public" / "Make Private" based on ownership
- **Context Menu**: Hide edit/delete options if user cannot modify view
- **View List**: Display "(by user@example.com)" for private views

## Business Rules / Invariants
- [x] ViewOwner.userID must be non-empty
- [x] ViewVisibility can only be Private or Public
- [x] Only owner can make a view private
- [x] Only owner or admin can make a private view public
- [x] When admin makes private view public, ownership transfers to admin
- [x] Stakeholders can create/edit/delete their own private views (bypasses role permission for owned views)
- [x] Cannot delete the default view (regardless of ownership)

## Authorization Matrix

| Action | Private (Owner) | Private (Non-Owner) | Private (Admin) | Public (Any) |
|--------|-----------------|---------------------|-----------------|--------------|
| View | ✅ | ✅ | ✅ | ✅ |
| Edit | ✅ | ❌ | ❌ (must make public first) | ✅ (with role) |
| Delete | ✅ | ❌ | ❌ (must make public first) | ✅ (with role) |
| Make Public | ✅ | ❌ | ✅ (transfers ownership) | N/A |
| Make Private | ✅ | ❌ | ❌ | ✅ (owner only) |

## Test Plan

### Unit Tests
- [x] ViewOwner value object validation (empty userID rejected)
- [x] ViewVisibility value object (Private/Public states)
- [x] Aggregate: CanBeEditedBy logic for all scenarios
- [x] Aggregate: MakePublic/MakePrivate state transitions
- [x] Aggregate: Ownership transfer on admin MakePublic

### Integration Tests
- [x] Create view sets owner and defaults to private
- [x] Owner can edit private view
- [x] Non-owner cannot edit private view (403)
- [x] Admin cannot edit private view directly (403)
- [x] Owner can toggle visibility
- [x] Admin can make private view public (ownership transfers)
- [x] Stakeholder can create/edit/delete own private view
- [x] Stakeholder cannot edit others' private views

### Frontend Tests
- [x] Private indicator displays correctly
- [x] Owner email displays for private views
- [x] Edit/delete buttons hidden for non-owned private views
- [x] Visibility toggle works correctly

## Boy Scouting: Dead Code Cleanup

The architecture views bounded context has unused event infrastructure that should be cleaned up:

### Dead Events (never raised, handlers bypass event sourcing)
These events exist but are never raised. The corresponding handlers write directly to the database for performance reasons (high-frequency drag-and-drop operations).

| File | Status |
|------|--------|
| `domain/events/view_edge_type_updated.go` | Delete |
| `domain/events/view_layout_direction_updated.go` | Delete |
| `domain/events/component_position_updated.go` | Delete |

### Dead Projector Methods
| Location | Method | Action |
|----------|--------|--------|
| `projectors/architecture_view_projector.go` | `projectViewEdgeTypeUpdated` | Delete |
| `projectors/architecture_view_projector.go` | `projectViewLayoutDirectionUpdated` | Delete |
| `projectors/architecture_view_projector.go` | `projectComponentPositionUpdated` | Delete |
| `projectors/architecture_view_projector.go` | Switch cases for above events | Delete |

### Dead Event Bus Subscriptions
| File | Lines | Action |
|------|-------|--------|
| `infrastructure/api/routes.go` | `eventBus.Subscribe("ComponentPositionUpdated", ...)` | Delete |
| `infrastructure/api/routes.go` | `eventBus.Subscribe("ViewEdgeTypeUpdated", ...)` | Delete |
| `infrastructure/api/routes.go` | `eventBus.Subscribe("ViewLayoutDirectionUpdated", ...)` | Delete |

### Dead Aggregate Code
| File | Code | Action |
|------|------|--------|
| `domain/aggregates/architecture_view.go:172` | Empty case for `ComponentPositionUpdated, ViewEdgeTypeUpdated, ViewLayoutDirectionUpdated` | Delete |

### Dead Repository Code
| File | Code | Action |
|------|------|--------|
| `infrastructure/repositories/architecture_view_repository.go` | `ComponentPositionUpdated` deserializer | Delete |

## Implementation Checklist

### Phase 0: Boy Scouting (Dead Code Cleanup)
- [x] Delete unused event files (3 files)
- [x] Remove dead projector methods and switch cases
- [x] Remove dead event bus subscriptions
- [x] Remove empty aggregate case
- [x] Remove dead deserializer
- [x] Verify build passes after cleanup

### Phase 1: Domain Model
- [x] Create `ViewOwner` value object
- [x] Create `ViewVisibility` value object
- [x] Create `ViewVisibilityChanged` event
- [x] Update `ArchitectureView` aggregate with owner, visibility, methods
- [x] Unit tests for domain logic

### Phase 2: Application Layer
- [x] Update `CreateViewHandler` to set owner from actor context
- [x] Create `ChangeViewVisibilityHandler` with permission logic
- [x] Update `ArchitectureViewDTO` with owner/visibility fields
- [x] Update projector to handle new event
- [x] Add ownership checks to existing handlers (rename, delete, etc.)

### Phase 3: Database
- [x] Create migration for new columns
- [x] Run migration and verify schema

### Phase 4: API Layer
- [x] Add `PATCH /views/{id}/visibility` endpoint
- [x] Add permission checks to write endpoints
- [x] Update response DTOs with new fields
- [x] Update HATEOAS links based on permissions
- [x] Update Swagger documentation

### Phase 5: Frontend
- [x] Update `View` TypeScript interface
- [x] Add `changeVisibility` API call
- [x] Update ViewSelector with private indicator
- [x] Update NavigationTree context menu
- [x] Conditional rendering based on permissions

### Phase 6: Testing
- [x] Backend unit tests
- [x] Backend integration tests
- [x] Frontend component tests
- [x] Build verification (`go test ./...`, `npm test -- --run`)

## Sign-off
- [x] User: Approved for implementation
