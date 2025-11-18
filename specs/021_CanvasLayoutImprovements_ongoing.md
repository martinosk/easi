# Canvas Layout Improvements

## Description
Enhance the React Flow canvas with configurable edge types and automatic layout capabilities using the Dagre algorithm, while maintaining manual positioning functionality.

## Purpose
Enable users to create cleaner, more readable architecture diagrams by:
- Selecting appropriate edge routing styles for different diagram types
- Automatically organizing components using hierarchical layout algorithms
- Fine-tuning layouts with manual adjustments

## Functional Requirements

### Edge Type Selection

Users should be able to choose between different edge routing styles for the entire canvas:

- **Bezier** (default): Smooth curved connections, good for organic layouts
- **Step**: Right-angle connections with steps, good for technical diagrams
- **Smooth Step**: Rounded right-angle connections, softer technical look
- **Straight**: Direct point-to-point lines, minimal visual clutter

**User Experience:**
- Edge type selector appears in the toolbar area
- Selection applies to all edges in the current view
- Edge type preference is saved per view (persists across sessions)
- Changing edge type immediately updates all relation edges on the canvas
- Default edge type is Bezier (matches current behavior)

### Automatic Layout with Dagre

Users should be able to automatically arrange components using the Dagre hierarchical layout algorithm:

**Layout Button:**
- "Auto Layout" button appears in the toolbar or canvas controls
- Clicking triggers automatic repositioning of all components in the current view
- Layout algorithm respects relation directions (top-to-bottom hierarchy)
- Components are positioned to minimize edge crossings and optimize readability

**Layout Behavior:**
- Calculates optimal positions for all components based on their relations
- Maintains component sizes from React Flow node dimensions
- Updates component positions via existing position update API
- Preserves viewport zoom level (does not reset zoom)
- Automatically fits view to show all components after layout
- Shows loading indicator during layout calculation
- Toast notification confirms layout completion

**Layout Options:**
- Direction: Top-to-bottom (TB), Left-to-right (LR), Bottom-to-top (BT), Right-to-left (RL)
- Node spacing: Configurable gap between components
- Rank spacing: Configurable gap between hierarchy levels
- Layout direction selector appears in toolbar

### Manual Positioning

Existing manual positioning must remain fully functional:

- Users can drag components to any position before or after auto-layout
- Manual position changes persist to backend via existing API
- Manual adjustments override auto-layout positions
- Drag-and-drop of new components from tree continues to work
- Position updates save to backend immediately on drag end

## Technical Requirements

### Dependencies

**New Package:**
- Install `dagre` library for graph layout algorithm
- Install `@types/dagre` for TypeScript support

**React Flow Integration:**
- React Flow 12.9.2 supports custom edge types out of the box
- React Flow provides `getBezierPath`, `getSmoothStepPath`, `getStraightPath` utility functions

### Frontend Implementation

**State Management:**
- Add `edgeType` to view state (new field in View interface)
- Add `layoutDirection` to view state (new field in View interface)
- Extend Zustand store with:
  - `setEdgeType(edgeType: EdgeType): Promise<void>` - updates current view's edge type
  - `setLayoutDirection(direction: LayoutDirection): Promise<void>` - updates layout direction
  - `applyAutoLayout(): Promise<void>` - calculates and applies Dagre layout

**Edge Type Support:**
- Update edge rendering in ComponentCanvas to use dynamic edge type
- Map edge type selection to React Flow edge type property
- Supported types: 'default' (Bezier), 'step', 'smoothstep', 'straight'

**Dagre Integration:**
- Create utility function `calculateDagreLayout(nodes, edges, direction, nodeSpacing, rankSpacing)`
- Function returns updated node positions
- Apply positions by calling existing `updatePosition` for each component
- Batch position updates to minimize API calls

**UI Components:**
- Add EdgeTypeSelector component to toolbar
- Add LayoutDirectionSelector component to toolbar
- Add AutoLayoutButton component to toolbar or Controls panel
- Components use existing design system styling

**View State Persistence:**
- Edge type persists to backend view state
- Layout direction persists to backend view state
- No persistence needed for node positions (already handled)

### Backend Integration

**API Changes:**
- Extend View aggregate to include `edgeType` property (string enum)
- Extend View aggregate to include `layoutDirection` property (string enum)
- Update PUT /api/v1/views/{id} endpoint to accept new properties
- Generate ViewEdgeTypeUpdated event when edge type changes
- Generate ViewLayoutDirectionUpdated event when layout direction changes
- Update View read model to include new properties
- Validate edge type values: "default", "step", "smoothstep", "straight"
- Validate layout direction values: "TB", "LR", "BT", "RL"

**Domain Model:**
- EdgeType value object with validation
- LayoutDirection value object with validation
- UpdateViewEdgeType command
- UpdateViewLayoutDirection command
- Handle backwards compatibility for views without edge type (default to "default")

## User Experience Requirements

### Edge Type Selector
- Dropdown or segmented control in toolbar
- Icons or labels for each edge type
- Visual preview of edge type (optional enhancement)
- Keyboard accessible
- Clear visual indication of current selection

### Auto Layout
- Button with appropriate icon (e.g., auto-layout grid icon)
- Loading state during calculation
- Success toast after completion
- Error toast if layout fails

### Layout Direction Selector
- Dropdown or segmented control
- Icons showing direction (arrows)
- Labels: "Top to Bottom", "Left to Right", etc.
- Current selection clearly indicated

### Performance
- Layout calculation completes within 2 seconds for diagrams up to 100 components
- Edge type changes apply instantly (no API delay for visual update)
- No visible lag when switching between views with different edge types

## Accessibility Requirements
- All toolbar controls have proper ARIA labels


## Error Handling

**Layout Failures:**
- Display error toast if Dagre calculation fails
- Log error details for debugging
- Leave component positions unchanged on error
- Provide actionable error message to user

**API Errors:**
- Handle network failures when saving edge type preference
- Display error toast with retry option
- Optimistically update UI, rollback on error
- Graceful degradation if backend doesn't support new properties

## Testing Requirements

### Unit Tests
- Test edge type mapping to React Flow types
- Test state updates for edge type and layout direction
- Test error handling in layout calculation

### Integration Tests
- Test edge type persistence to backend
- Test layout direction persistence to backend
- Test auto-layout with real component data
- Test view switching with different edge types

### End-to-End Tests
- User can select edge type and see edges update
- User can trigger auto-layout and components reposition
- User can change layout direction
- User can manually adjust positions after auto-layout
- Edge type and layout direction persist across page refresh

## Checklist

### Backend Implementation
- [ ] Add EdgeType value object to ArchitectureViews domain
- [ ] Add LayoutDirection value object to ArchitectureViews domain
- [ ] Add UpdateViewEdgeType command
- [ ] Add UpdateViewLayoutDirection command
- [ ] Add command handlers
- [ ] Generate ViewEdgeTypeUpdated event
- [ ] Generate ViewLayoutDirectionUpdated event
- [ ] Update View aggregate to store edge type
- [ ] Update View aggregate to store layout direction
- [ ] Update View read model projector
- [ ] Update PUT /api/v1/views/{id} API endpoint
- [ ] Add validation for edge type values
- [ ] Add validation for layout direction values
- [ ] Handle backwards compatibility
- [ ] Unit tests for value objects
- [ ] Unit tests for commands and handlers
- [ ] Integration tests for API endpoints

### Frontend Implementation
- [ ] Install dagre package
- [ ] Install @types/dagre package
- [ ] Create calculateDagreLayout utility function
- [ ] Update View TypeScript interface with edgeType property
- [ ] Update View TypeScript interface with layoutDirection property
- [ ] Add setEdgeType action to Zustand store
- [ ] Add setLayoutDirection action to Zustand store
- [ ] Add applyAutoLayout action to Zustand store
- [ ] Update ComponentCanvas to use dynamic edge type
- [ ] Create EdgeTypeSelector component
- [ ] Create LayoutDirectionSelector component
- [ ] Create AutoLayoutButton component
- [ ] Integrate selectors into Toolbar component
- [ ] Update API client to send edge type and layout direction
- [ ] Handle loading states during auto-layout
- [ ] Add toast notifications for success/error
- [ ] Persist edge type preference per view
- [ ] Persist layout direction preference per view

### Testing
- [ ] Unit tests for Dagre layout calculation
- [ ] Unit tests for edge type state management
- [ ] Unit tests for layout direction state management
- [ ] Integration tests for edge type persistence
- [ ] Integration tests for layout direction persistence
- [ ] Integration tests for auto-layout with real data
- [ ] E2E test: Select edge type and verify visual update
- [ ] E2E test: Trigger auto-layout and verify repositioning
- [ ] E2E test: Change layout direction and verify
- [ ] E2E test: Manual positioning after auto-layout
- [ ] E2E test: Edge type persists across page refresh
- [ ] Accessibility tests for toolbar controls

### Documentation
- [ ] Update user guide with edge type selection
- [ ] Update user guide with auto-layout feature
- [ ] Document layout direction options
- [ ] Document Dagre integration approach
- [ ] Add screenshots of different edge types

### Final
- [ ] All tests passing
- [ ] Visual design matches application style
- [ ] Accessibility requirements met
- [ ] Performance requirements met
- [ ] User acceptance testing completed
- [ ] User sign-off
