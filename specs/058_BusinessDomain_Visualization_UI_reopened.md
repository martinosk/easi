# Business Domain Visualization UI

## Status
**Reopened** - Original treemap visualization completed. Adding interactive grid-based visualization with drag-and-drop capabilities.

## User Need
Enterprise architects need a spatial, interactive visualization of business domain composition that allows them to arrange capabilities visually, assign L1 capabilities to domains via drag-and-drop, and navigate a nested hierarchy down to L4 depth.

## Dependencies
- Spec 056: Business Domain REST API
- Spec 057: Business Domain Management UI

---

## Completed Work (Previous Implementation)

The following components and features have been implemented:

- [x] DomainVisualizationPage with three-column layout
- [x] DomainFilter sidebar for domain selection
- [x] DomainTreemap using Recharts (superseded by grid view)
- [x] CapabilityHierarchyView tree visualization (kept as alternative)
- [x] OrphanedCapabilitiesPanel showing unassigned L1 capabilities
- [x] CapabilityDetailPanel for capability inspection
- [x] ViewModeToggle component
- [x] useCapabilityTree, useUnassignedCapabilities, useDomainComposition hooks
- [x] Color scheme: L1=Blue, L2=Purple, L3=Pink, L4=Orange

---

## New Work: Grid-Based Visualization

### User Stories

1. As an architect, I can view a business domain's capabilities in a nested grid layout
2. As an architect, I can drag L1 capabilities from the explorer onto a domain to assign them
3. As an architect, I can rearrange capabilities within the grid to organize the visualization
4. As an architect, I can drag capabilities between parents with confirmation
5. As an architect, I can control visualization depth (L1 only through L1-L4)
6. As an architect, I can identify which L1 capabilities are assigned to multiple domains

### Success Criteria

- Grid displays capabilities in nested structure by domain
- Drag from explorer assigns L1 capability to currently selected domain
- Drag within grid updates layout (positions persist across sessions)
- All users see the same layout (organization-wide persistence)
- Depth selector controls visible hierarchy levels
- Domain selector in sidebar switches between domain views
- Click on capability opens existing detail panel
- Capabilities assigned to multiple domains show visual indicator

### Out of Scope

- Per-user custom layouts (design for future, but not implemented)
- View-based layouts (architecture prepared with optional viewId field)
- Cross-domain drag operations (use explorer menu only)
- Mobile/tablet support (desktop only)
- Virtualization (not needed for expected data scale)
- Role-based permissions (always editable, architect for later)

---

## Vertical Slices

### Slice 1: Domain Selector + Read-Only L1 Grid

Display L1 capabilities for a selected domain in a static grid.

- [ ] Add "grid" view mode to ViewModeToggle (treemap | tree | grid)
- [ ] Create DomainGrid component using @dnd-kit DndContext
- [ ] Display L1 capabilities only (no nesting yet)
- [ ] Auto-arrange by capability code (alphabetical)
- [ ] Fixed size: L1 = 4x4 grid units
- [ ] Apply L1 color (Blue #3b82f6)
- [ ] Domain selector in left sidebar switches grid content
- [ ] Orphaned capabilities NOT shown in grid (explorer only)

### Slice 2: Capability Explorer + Drag to Assign

Right sidebar explorer showing all L1 capabilities for drag-and-drop assignment.

- [ ] Create CapabilityExplorer component (right sidebar)
- [ ] Display full L1-L4 capability tree in explorer
- [ ] Only L1 items draggable from explorer
- [ ] Visual indicator for L1s already assigned to other domains
- [ ] Drag L1 from explorer onto grid assigns to current domain
- [ ] API call: assign capability to domain (existing endpoint)
- [ ] Grid refreshes after assignment
- [ ] Many-to-many: same L1 can appear in multiple domains

### Slice 3: Drag Within Grid + Position Persistence

Allow rearranging items within the grid with persisted positions.

- [ ] Enable @dnd-kit sortable drag-within-grid functionality
- [ ] Items shift to make space when dragging
- [ ] Create BusinessDomainView on entering visualization (POST /api/v1/views)
- [ ] Load view layout: GET /api/v1/views/{viewId}/layout
- [ ] Save positions: PUT /api/v1/views/{viewId}/layout
- [ ] Auto-save positions on drag end
- [ ] Load persisted positions on page load
- [ ] Last-write-wins conflict resolution

### Slice 4: Nested Grids (L2, L3, L4) + Depth Control

Render child capabilities inside parent grid cells with depth selector.

- [ ] Create NestedCapabilityGrid component
- [ ] L1 contains nested grid for L2 children
- [ ] L2 contains nested grid for L3 children
- [ ] L3 contains nested grid for L4 children
- [ ] Fixed sizes: L2=3x3, L3=2x2, L4=1x1 grid units
- [ ] Apply level colors (L2=Purple, L3=Pink, L4=Orange)
- [ ] Overflow: scroll within parent cell if children exceed space
- [ ] Add DepthSelector component (L1 | L1-L2 | L1-L2-L3 | L1-L2-L3-L4)
- [ ] Depth applies globally to all L1s uniformly
- [ ] Collapse/hide children smoothly when depth reduced
- [ ] Preserve positions when depth changes

### Slice 5: Drag to Reassign Parent + Confirmation

Allow dragging capabilities to different parents with confirmation dialog.

- [ ] Enable dragging capability onto another capability (new parent)
- [ ] Show simple confirmation dialog before reassignment
- [ ] API call: change capability parent (existing endpoint)
- [ ] Grid refreshes after parent change
- [ ] Level auto-adjusts based on new parent (existing backend logic)

### Slice 6: Detail Panel Integration

Click capability to open existing detail panel.

- [ ] Wire click handler on grid items
- [ ] Open CapabilityDetailPanel on click (reuse existing component)
- [ ] Panel shows full capability information
- [ ] Close panel returns focus to grid

### Slice 7: Visual Polish

Final styling and animations.

- [ ] Apply domain-specific background color to grid area
- [ ] Visual indicator for L1s assigned to multiple domains (badge/icon)
- [ ] Smooth animation for depth collapse/expand
- [ ] DragOverlay for visual feedback during drag
- [ ] Hover states on grid items
- [ ] Loading states during API calls

---

## Technical Requirements

### Frontend

**Dependencies:**
```
npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities
```

**Why @dnd-kit over gridstack.js:**
- First-class support for nested drag-and-drop (Multiple Containers pattern)
- Better TypeScript support
- Smaller bundle size
- Follows React patterns (DndContext, SortableContext)
- gridstack.js designed for flat dashboard grids, not nested hierarchies

**Component Architecture:**
- DomainGrid: Main container with DndContext
- SortableCapabilityContainer: Wrapper providing SortableContext per level
- SortableCapability: Recursive component (useSortable hook)
- DragOverlay: Visual feedback during drag operations
- CapabilityExplorer: Right sidebar with draggable L1 tree
- DepthSelector: Global depth control (1-4 levels)
- ReassignConfirmDialog: Simple confirmation for parent changes

**@dnd-kit Pattern (Multiple Containers):**
- DndContext wraps entire grid
- Each container level (L1, L2, L3, L4) has its own SortableContext
- SortableCapability renders children recursively with nested SortableContext
- DragOverlay renders dragged item outside DOM hierarchy
- Reference: @dnd-kit "Multiple Containers" example

**Grid Layout:**
- @dnd-kit handles drag-and-drop, not layout
- Use CSS Grid for visual arrangement
- Custom position calculation for auto-arrange
- Database stores logical positions (x, y, w, h) for grid-like arrangement
- Convert logical positions to CSS Grid placement

**State Management:**
- React Query for API state (positions, capabilities)
- Grid positions loaded from API per domain
- Auto-save on position change (debounced)
- Depth selection stored in component state (not persisted)
- Optimistic updates for smooth drag experience

### Backend

**Architecture: Integration with Views Bounded Context**

Grid layout positions are a **presentation concern**, not a domain concern. Following DDD bounded context principles, layout persistence belongs in the **Architecture Views** context, not Capability Mapping.

**Why Views Context?**
- Layout positions represent "how we visualize" (presentation), not "what capabilities belong to domains" (domain)
- Capability Mapping owns: domain-to-capability assignments, business invariants
- Architecture Views owns: visual layout, spatial arrangement, presentation state
- Clear separation of concerns: domain model stays pure, presentation is separate

**Existing Infrastructure (Already Implemented):**

The system already has the necessary infrastructure:

1. **Database Table** (Migration 006, 013):
```sql
-- view_element_positions table already supports capabilities
CREATE TABLE view_element_positions (
  id SERIAL PRIMARY KEY,
  tenant_id VARCHAR(50) NOT NULL,
  view_id VARCHAR(255) NOT NULL,
  element_id VARCHAR(255) NOT NULL,
  element_type VARCHAR(50) NOT NULL,  -- 'capability', 'component', 'relationship'
  x DOUBLE PRECISION NOT NULL,
  y DOUBLE PRECISION NOT NULL,
  width DOUBLE PRECISION,
  height DOUBLE PRECISION,
  UNIQUE(tenant_id, view_id, element_id)
);
```

2. **ViewLayoutRepository** (ArchitectureViews context):
- `GetCapabilityPositions(ctx, viewID)` - Load positions
- `SetCapabilityPositions(ctx, viewID, positions)` - Save positions
- Already implemented and tested

3. **View Aggregate Properties**:
- `ViewType` - Can be "architecture", "businessDomain", etc.
- `Metadata` - JSON field for view-specific data (e.g., `{"businessDomainId": "bd-123"}`)

**Business Domain View Concept:**

A Business Domain visualization is a specialized type of View:

```json
{
  "id": "view-abc123",
  "name": "Finance Domain Layout",
  "viewType": "businessDomain",
  "canvasType": "grid",
  "metadata": {
    "businessDomainId": "bd-finance"
  },
  "createdAt": "2025-01-15T10:00:00Z"
}
```

**API Integration Pattern:**

Frontend uses **Views context API** (already exists):

1. **Create or Get View**: `POST /api/v1/views` or `GET /api/v1/views?viewType=businessDomain&metadata.businessDomainId=bd-123`
2. **Load Layout**: `GET /api/v1/views/{viewId}/layout`
3. **Save Layout**: `PUT /api/v1/views/{viewId}/layout`

**No New Backend Code Required:**
- Reuse existing View aggregate and ViewLayoutRepository
- No new tables, migrations, or endpoints
- Business Domain context provides read-only data (domain names, capability assignments)
- Views context handles all layout persistence

**Cross-Context Data Flow:**

```
User Action → Frontend
              ↓
        Views API (save positions)
              ↓
        ViewLayoutRepository
              ↓
        view_element_positions table

User Opens Visualization → Frontend
                            ↓
                      GET capabilities (CapabilityMapping API)
                            ↓
                      GET positions (Views API)
                            ↓
                      Merge data client-side
```

### Grid Sizing

| Level | Grid Units | Visual Size |
|-------|-----------|-------------|
| L1    | 4x4       | Large block |
| L2    | 3x3       | Medium block |
| L3    | 2x2       | Small block |
| L4    | 1x1       | Minimum block |

### Color Scheme (unchanged)

| Level | Color  | Hex      |
|-------|--------|----------|
| L1    | Blue   | #3b82f6  |
| L2    | Purple | #8b5cf6  |
| L3    | Pink   | #ec4899  |
| L4    | Orange | #f97316  |

### Expected Scale

- ~5 business domains
- 5-10 L1 capabilities per domain
- 5-10 children per level (up to L4)
- No virtualization required

---

## Acceptance Criteria Summary

- [ ] Grid view mode added to existing toggle
- [ ] Domain selector switches between domain grids
- [ ] BusinessDomainView created on entering visualization (Views API)
- [ ] L1 capabilities display in grid at 4x4 size
- [ ] Capability explorer shows full L1-L4 tree
- [ ] Drag L1 from explorer assigns to domain (CapabilityMapping API)
- [ ] Drag within grid rearranges layout
- [ ] Positions persist via Views API (organization-wide)
- [ ] Nested grids show L2/L3/L4 inside parents
- [ ] Depth selector controls visible levels
- [ ] Drag between parents shows confirmation
- [ ] Click opens detail panel
- [ ] Level colors match existing scheme
- [ ] Multiple-domain indicator on L1s
