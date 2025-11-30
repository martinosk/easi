# ViewLayouts Bounded Context

## Status
**Done** - New bounded context to separate presentation concerns (layout/positions) from domain concerns (view membership).

## Problem Statement

The ArchitectureViews bounded context is currently overloaded:
1. **Domain concern**: Managing named views and component membership (Architecture Canvas)
2. **Presentation concern**: Storing element positions, colors, and layout preferences

This causes Business Domain grid layouts to appear in the Architecture Canvas view selector because both use the same `architecture_views` table and API.

## User Need

- Architecture Canvas users need a view selector that only shows architecture views
- Business Domain users need layout persistence without creating "views"
- Both need concurrent editing support (multiple users editing simultaneously)
- Both need automatic cleanup when domain entities are deleted

## Solution

Create a new **ViewLayouts** bounded context focused purely on presentation concerns:
- Layout containers identified by `contextType` + `contextRef` (not view IDs)
- Element positions stored per-element for concurrent editing
- Event handlers for model-view synchronization
- Clear separation from ArchitectureViews (which keeps view membership)

## Dependencies

- Spec 008: Multiple Views Management (existing ArchitectureViews context)
- Spec 058: Business Domain Visualization UI (current consumer)

---

## Vertical Slices

### Slice 1: Backend - Database Schema

Create database tables for the ViewLayouts context.

- [ ] Create migration for `layout_containers` table
- [ ] Create migration for `element_positions` table
- [ ] Add indexes for efficient queries
- [ ] No foreign key constraints (event-sourced read model pattern)

**Schema:**
```sql
CREATE TABLE layout_containers (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    context_type VARCHAR(50) NOT NULL,
    context_ref VARCHAR(255) NOT NULL,
    preferences JSONB DEFAULT '{}',
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(tenant_id, context_type, context_ref)
);

CREATE TABLE element_positions (
    container_id VARCHAR(255) NOT NULL,
    element_id VARCHAR(255) NOT NULL,
    x DOUBLE PRECISION NOT NULL,
    y DOUBLE PRECISION NOT NULL,
    width DOUBLE PRECISION,
    height DOUBLE PRECISION,
    custom_color VARCHAR(50),
    sort_order INTEGER,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (container_id, element_id)
);

CREATE INDEX idx_element_positions_container ON element_positions(container_id);
```

### Slice 2: Backend - Domain Model

Create the ViewLayouts domain model (supporting subdomain - simple CRUD).

- [ ] Create `LayoutContainerID` value object
- [ ] Create `LayoutContextType` value object (architecture-canvas, business-domain-grid)
- [ ] Create `ElementPosition` value object
- [ ] Create `LayoutPreferences` value object
- [ ] Create `LayoutContainer` aggregate
- [ ] Create repository interface `LayoutContainerRepository`

### Slice 3: Backend - Repository Implementation

Implement PostgreSQL repository.

- [ ] Implement `LayoutContainerRepository` with JSONB support
- [ ] Implement `GetByContext(tenantID, contextType, contextRef)`
- [ ] Implement `Save(container)` with optimistic locking
- [ ] Implement `Delete(id)`
- [ ] Implement `UpsertElementPosition(containerID, position)`
- [ ] Implement `DeleteElementPosition(containerID, elementID)`
- [ ] Implement `BatchUpdatePositions(containerID, positions)`
- [ ] Unit tests for repository

### Slice 4: Backend - REST API

Implement REST Level 3 API endpoints.

- [ ] `GET /api/v1/layouts/{contextType}/{contextRef}` - Get layout with positions
- [ ] `PUT /api/v1/layouts/{contextType}/{contextRef}` - Create/update container
- [ ] `DELETE /api/v1/layouts/{contextType}/{contextRef}` - Delete container
- [ ] `PATCH /api/v1/layouts/{contextType}/{contextRef}/preferences` - Update preferences
- [ ] `PUT /api/v1/layouts/{contextType}/{contextRef}/elements/{elementId}` - Upsert position
- [ ] `DELETE /api/v1/layouts/{contextType}/{contextRef}/elements/{elementId}` - Remove element
- [ ] `PATCH /api/v1/layouts/{contextType}/{contextRef}/elements` - Batch update
- [ ] HATEOAS links on all responses
- [ ] OpenAPI documentation
- [ ] Integration tests

**Response Structure (GET):**
```json
{
  "id": "container-uuid",
  "contextType": "business-domain-grid",
  "contextRef": "domain-finance",
  "preferences": {
    "colorScheme": "pastel",
    "layoutDirection": "TB"
  },
  "elements": [
    { "elementId": "cap-123", "x": 100, "y": 200, "customColor": "#3b82f6" },
    { "elementId": "cap-456", "x": 300, "y": 150, "sortOrder": 2 }
  ],
  "version": 5,
  "_links": {
    "self": { "href": "/api/v1/layouts/business-domain-grid/domain-finance" },
    "updatePreferences": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/preferences", "method": "PATCH" },
    "batchUpdate": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements", "method": "PATCH" }
  }
}
```

### Slice 5: Backend - Event Handlers for Model Sync

Handle domain entity deletions to clean up orphaned positions.

- [ ] Create `ComponentDeletedHandler` - removes component positions from all layouts
- [ ] Create `CapabilityDeletedHandler` - removes capability positions from all layouts
- [ ] Create `BusinessDomainDeletedHandler` - deletes business-domain-grid layouts for that domain
- [ ] Create `ViewDeletedHandler` - deletes architecture-canvas layouts for that view
- [ ] Integration tests for event handlers

### Slice 6: Frontend - Types and API Client

Add TypeScript types and API client methods.

- [ ] Add `LayoutContextType` type
- [ ] Add `Layout`, `ElementPosition`, `LayoutPreferences` interfaces
- [ ] Add `CreateLayoutRequest`, `UpdatePreferencesRequest`, `BatchUpdateRequest` types
- [ ] Add `getLayout(contextType, contextRef)` method
- [ ] Add `createOrUpdateLayout(contextType, contextRef, request)` method
- [ ] Add `deleteLayout(contextType, contextRef)` method
- [ ] Add `updateLayoutPreferences(contextType, contextRef, preferences)` method
- [ ] Add `upsertElementPosition(contextType, contextRef, elementId, position)` method
- [ ] Add `removeElementPosition(contextType, contextRef, elementId)` method
- [ ] Add `batchUpdateElementPositions(contextType, contextRef, elements)` method

### Slice 7: Frontend - useLayout Hook

Create generic layout management hook.

- [x] Create `useLayout(contextType, contextRef)` hook
- [x] Auto-create layout if 404 returned
- [x] Return `{ layout, positions, preferences, isLoading, error }`
- [x] Implement `updateElementPosition()` with optimistic updates
- [x] Implement `batchUpdatePositions()` with optimistic updates
- [x] Implement `updatePreferences()` with optimistic updates
- [x] Implement rollback on API error
- [x] Unit tests for hook

### Slice 8: Migrate Business Domain Grid

Update Business Domain visualization to use new API.

- [x] Update `useGridPositions` hook to use `useLayout('business-domain-grid', domainId)`
- [x] Maintain backward-compatible interface (no component changes needed)
- [x] Remove view creation logic (no longer needed)
- [x] Test drag-and-drop still works
- [x] Test position persistence works
- [x] Test concurrent editing works

### Slice 9: Migrate Architecture Canvas

Update Architecture Canvas to use new API for positions.

- [x] Update layout slice to use new layout API
- [x] Keep ArchitectureViews for view membership (name, components list)
- [x] Positions now stored via ViewLayouts API
- [x] Link layout to view via `contextType='architecture-canvas'`, `contextRef=viewId`
- [x] Test view switching loads correct positions
- [x] Test position updates persist correctly

### Slice 10: Cleanup

Remove deprecated code and old tables.

- [ ] Remove view-related position methods from old API client
- [ ] Remove `view_element_positions` table (migration)
- [ ] Remove `view_preferences` table (migration)
- [ ] Remove position-related events from ArchitectureViews aggregate
- [ ] Update documentation

---

## Technical Requirements

### Concurrency Handling

**Element-level updates (concurrent-safe):**
- Each element position is a separate row
- Two users moving different elements don't conflict
- Last-write-wins for same element (positions self-correct visually)

**Container-level updates (optimistic locking):**
- Version field incremented on each update
- `412 Precondition Failed` if version mismatch
- Used for preferences updates, not individual positions

### Context Types

| Context Type | Context Ref | Use Case |
|-------------|-------------|----------|
| `architecture-canvas` | View ID | Component positions on canvas |
| `business-domain-grid` | Domain ID | Capability positions in domain grid |

### API Concurrency Pattern

```
Single element: PUT (last-write-wins, no version check)
Batch elements: PATCH (atomic, all-or-nothing)
Preferences: PATCH (requires version in ETag/If-Match)
```

### Model-View Sync Events

| Domain Event | Handler Action |
|--------------|---------------|
| ComponentDeleted | Remove from all `architecture-canvas` layouts |
| CapabilityDeleted | Remove from all `business-domain-grid` layouts |
| BusinessDomainDeleted | Delete entire layout for that domain |
| ViewDeleted | Delete entire layout for that view |

---

## Out of Scope

- Per-user layouts (all users see same layout)
- Undo/redo for position changes
- Real-time collaboration (WebSocket sync)
- Layout templates or presets

---

## Acceptance Criteria

- [ ] Business Domain grid positions persist correctly
- [ ] Architecture Canvas positions persist correctly
- [ ] Business Domain layouts do NOT appear in Architecture Canvas view selector
- [ ] Concurrent editing by multiple users works without data loss
- [ ] Deleting a component removes its position from all layouts
- [ ] Deleting a capability removes its position from all layouts
- [ ] Deleting a business domain removes its layout
- [ ] Deleting a view removes its layout
- [ ] All endpoints return proper HATEOAS links
- [ ] All endpoints documented in OpenAPI spec

---

## Checklist

- [x] Specification approved
- [x] Backend implementation complete
- [x] Frontend implementation complete
- [x] Migration of Business Domains complete
- [x] Migration of Architecture Canvas complete
- [x] Event handlers for model sync complete
- [x] Unit tests passing
- [ ] Integration tests passing
- [ ] Old code cleaned up (deferred - verify in production first)
- [ ] User sign-off
