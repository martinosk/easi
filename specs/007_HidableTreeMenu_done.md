# Spec 007: Hidable Tree Structure Menu

## Status
ongoing

## Overview
Add a collapsible/hidable tree structure menu to the left side of the UI that displays:
- **Models**: All application components (flat hierarchy for now)
- **Views**: All existing views (flat hierarchy for now)

The menu should be toggleable to maximize workspace when not needed.

## Business Context
Users need quick navigation and overview of:
- All components in their system model
- All available views they can work with

This menu provides a project explorer-style navigation that's familiar from IDEs and makes it easier to work with larger models.

## Functional Requirements

### Menu Structure
- [x] Tree menu positioned on the left side of the canvas area
- [x] Menu is collapsible/expandable via toggle button
- [x] Menu state (open/closed) persists in browser session
- [x] Default state: open

### Tree Categories
- [x] Top-level category: "Models"
  - [x] Shows all application components from the current context
  - [x] Flat list (no nesting) for initial implementation
  - [x] Each item shows component name
- [x] Top-level category: "Views"
  - [x] Shows all available views
  - [x] Flat list (no nesting) for initial implementation
  - [x] Each item shows view name

### Interaction Behavior
- [x] Clicking a component in "Models" section:
  - [x] Highlights the component on the canvas if visible
  - [x] Centers/pans the canvas to show the component
  - [x] If component not in current view, does nothing (future: tooltip showing wich views containing it)
- [x] Clicking a view in "Views" section:
  - [x] Switches to that view (see Spec 008)
- [x] Categories ("Models", "Views") are collapsible independently

### Visual Design
- [x] Menu width: 250-300px when open (280px)
- [x] Smooth collapse/expand animation (200-300ms)
- [x] Toggle button clearly visible when collapsed
- [x] Visual hierarchy: categories bold, items regular weight
- [x] Hover states for interactive elements
- [x] Active/selected item highlighted

## Technical Requirements

### Frontend Implementation
- [x] New React component: `NavigationTree`
- [x] Component state manages:
  - [x] Menu open/closed state
  - [x] Category expanded/collapsed states
  - [x] Selected item
- [x] Use localStorage for menu state persistence
- [x] Integration with existing canvas component layout

### Backend/API Requirements
- [x] Use existing GET /components endpoint for Models section
- [x] Use existing GET /views endpoint for Views section
- [x] No new endpoints required initially

### Data Flow
```
Screen: Navigation Tree
  OUTBOUND → Command: Select Component
  OUTBOUND → Command: Select View

ReadModel: Component List (existing)
  OUTBOUND → Screen: Navigation Tree

ReadModel: View List (from Spec 008)
  OUTBOUND → Screen: Navigation Tree
```

## Non-Functional Requirements
- [x] Menu open/close animation smooth (60fps)
- [x] Menu state persists across browser refresh
- [x] Responsive: hide menu automatically on mobile/small screens

## Test Plan
- [x] Test menu toggle functionality
- [x] Test category expand/collapse
- [x] Test component selection highlights on canvas
- [x] Test state persistence after page refresh
- [x] Test with empty Models list
- [x] Test with empty Views list
- [x] Test on mobile viewport

## Implementation Notes
- Created `NavigationTree.tsx` component with full feature set
- Integrated with existing `ComponentCanvas` using React Flow's `setCenter` API for smooth panning
- Added comprehensive CSS styling matching the existing design system
- View switching works with existing views API
- Components not in current view are shown with reduced opacity and italic style

## Sign-off
- [x] Developer: Implementation complete
- [x] User: Approved for completion
