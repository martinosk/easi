# Spec 007: Hidable Tree Structure Menu

## Status
pending

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
- [ ] Tree menu positioned on the left side of the canvas area
- [ ] Menu is collapsible/expandable via toggle button
- [ ] Menu state (open/closed) persists in browser session
- [ ] Default state: open

### Tree Categories
- [ ] Top-level category: "Models"
  - [ ] Shows all application components from the current context
  - [ ] Flat list (no nesting) for initial implementation
  - [ ] Each item shows component name
- [ ] Top-level category: "Views"
  - [ ] Shows all available views
  - [ ] Flat list (no nesting) for initial implementation
  - [ ] Each item shows view name

### Interaction Behavior
- [ ] Clicking a component in "Models" section:
  - [ ] Highlights the component on the canvas if visible
  - [ ] Centers/pans the canvas to show the component
  - [ ] If component not in current view, does nothing (future: switch to view containing it)
- [ ] Clicking a view in "Views" section:
  - [ ] Switches to that view (see Spec 008)
- [ ] Categories ("Models", "Views") are collapsible independently

### Visual Design
- [ ] Menu width: 250-300px when open
- [ ] Smooth collapse/expand animation (200-300ms)
- [ ] Toggle button clearly visible when collapsed
- [ ] Visual hierarchy: categories bold, items regular weight
- [ ] Hover states for interactive elements
- [ ] Active/selected item highlighted

## Technical Requirements

### Frontend Implementation
- [ ] New React component: `NavigationTree`
- [ ] Component state manages:
  - [ ] Menu open/closed state
  - [ ] Category expanded/collapsed states
  - [ ] Selected item
- [ ] Use localStorage for menu state persistence
- [ ] Integration with existing canvas component layout

### Backend/API Requirements
- [ ] Use existing GET /components endpoint for Models section
- [ ] Use existing GET /views endpoint for Views section (from Spec 008)
- [ ] No new endpoints required initially

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
- [ ] Menu open/close animation smooth (60fps)
- [ ] Menu state persists across browser refresh
- [ ] Responsive: hide menu automatically on mobile/small screens
- [ ] Keyboard accessible (tab navigation, enter to select)

## Test Plan
- [ ] Test menu toggle functionality
- [ ] Test category expand/collapse
- [ ] Test component selection highlights on canvas
- [ ] Test view switching (when Spec 008 implemented)
- [ ] Test state persistence after page refresh
- [ ] Test with empty Models list
- [ ] Test with empty Views list
- [ ] Test with many items (50+ components)
- [ ] Test keyboard navigation
- [ ] Test on mobile viewport

## Future Enhancements
- Hierarchical grouping of components (by bounded context, aggregate, etc.)
- Search/filter within tree
- Drag items to create new relations
- Context menu for items (delete, rename, etc.)
- Icons for different component types
- Collapsible subsections

## Dependencies
- Depends on Spec 008 for Views functionality
- Extends existing canvas layout

## Sign-off
- [ ] Developer: Implementation complete
- [ ] User: Approved for completion
