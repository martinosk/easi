# Capability Grouping Visualization

## Description
Implement visual grouping of applications by the capabilities they realize. When enabled, applications automatically appear within capability boundary regions, creating intuitive visual clusters.

## User Need
As a user, I need to see which applications belong to which capabilities so that I can understand capability coverage and identify capability-application relationships at a glance.

## Visual Design

### Grouped Layout
```
┌─────────────────────────────────────────────┐
│ ▢ L1: Finance (Maturity: Established)      │ ← Capability group
│ ┌─────────────────────────────────────────┐ │
│ │  ● SAP Finance                          │ │ ← Application nodes
│ │  ● Legacy Billing                       │ │   inside group
│ │  ● Payment Gateway                      │ │
│ └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│ ▢ L2: Customer Support (Initial)           │
│ ┌─────────────────────────────────────────┐ │
│ │  ● Zendesk                              │ │
│ │  ● Ticketing System                     │ │
│ └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
```

### Group Visual Elements
- **Background**: Capability's maturity color with low opacity (15%)
- **Border**: Solid gray border
- **Header**: Capability level, name, and maturity badge
- **Padding**: Space around contained applications

### Empty Group
```
┌─────────────────────────────────────────────┐
│ ▢ L2: Risk Management (Initial)            │
│ ┌─────────────────────────────────────────┐ │
│ │  No applications realize this           │ │
│ │  capability yet.                        │ │
│ └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
```

## Functional Requirements

### Toggle Grouping
1. Toolbar toggle button: "Group by Capability"
2. When ON: Applications grouped within capability regions
3. When OFF: Free layout (default)
4. Smooth transition between modes

### Auto-Positioning
1. Group capabilities visible on canvas
2. Calculate group size based on number of applications
3. Position applications in grid within group
4. Groups arranged vertically on canvas

### Hierarchical Realization Handling
When application realizes multiple levels:
- Show application in **deepest level only** (e.g., L3 if realizes L3, L2, L1)
- Cleaner visualization, avoids duplication

### Overlapping Memberships
When application realizes multiple capabilities at same level:
- Application appears in first capability group (alphabetically)
- Or: Show with connector lines to multiple groups

### Drag Between Groups
1. User drags application from one group to another
2. Old realization deleted
3. New realization created
4. Grouping updated

### Group Header
Each group displays:
- Level badge (L1, L2, etc.)
- Capability name
- Maturity level badge

## Layout Algorithm

1. Filter capabilities visible on canvas
2. For each capability, find applications that realize it
3. Apply hierarchical rule (show at deepest level)
4. Calculate group size: based on application count (3 per row)
5. Position groups vertically with spacing
6. Position applications within groups in grid layout

## Acceptance Criteria
- [ ] Toggle button enables/disables grouping view
- [ ] Applications automatically grouped by realized capabilities
- [ ] Groups show capability name, level, and maturity
- [ ] Groups styled with maturity color (subtle background)
- [ ] Applications positioned within capability boundaries
- [ ] Empty groups show placeholder message
- [ ] Overlapping memberships handled (show at deepest level)
- [ ] Dragging application between groups updates realizations
- [ ] Layout automatically adjusts when realizations change
- [ ] Toggle smoothly switches between grouped and free layout
- [ ] Performance acceptable with 20+ capabilities and 50+ applications

## Checklist
- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] User sign-off
