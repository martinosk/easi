# Capability Dependencies on Canvas

## Description
Enable modeling capability dependencies (Requires, Enables, Supports) by connecting capabilities on canvas. Add view toggle to show/hide dependency edges separately from parent and realization edges.

## User Need
As a user, I need to model dependencies between capabilities so that I can document which capabilities depend on others and understand the capability dependency network.

## Visual Design

### Dependency Edge Types

**Requires** (strong dependency):
- Line style: Solid
- Color: Purple (#9333EA)
- Width: 2px
- Label: "Requires"

**Enables** (enabling relationship):
- Line style: Dashed
- Color: Blue (#3B82F6)
- Width: 2px
- Label: "Enables"

**Supports** (supporting relationship):
- Line style: Dotted
- Color: Orange (#F59E0B)
- Width: 2px
- Label: "Supports"

### View Toggle
Toolbar button to show/hide dependencies:
```
┌──────────────────────────────────────┐
│ [Auto Layout] [Dependencies ✓]      │
└──────────────────────────────────────┘
```

When off, only parent and realization edges visible.

## Functional Requirements

### Connection Type Selection
When connecting two capabilities, show dialog:
```
┌─────────────────────────────────────┐
│  Select Connection Type         [X] │
├─────────────────────────────────────┤
│  ○ Parent (hierarchy)               │
│  ○ Requires (strong dependency)     │
│  ○ Enables (enabling)               │
│  ○ Supports (supporting)            │
│           [Cancel]  [Create]        │
└─────────────────────────────────────┘
```

### Create Dependency
1. User connects two capability nodes
2. Dialog appears for type selection
3. User selects dependency type
4. Backend `POST /api/v1/capability-dependencies` called
5. Edge created with appropriate styling

### View Toggle
1. Toggle button in toolbar
2. When ON: All edges visible (parent, realization, dependencies)
3. When OFF: Only parent and realization edges visible
4. State persists during session

### Delete Dependency
1. Right-click dependency edge → Delete
2. Backend `DELETE /api/v1/capability-dependencies/:id` called
3. Edge removed from canvas

### Multiple Dependencies
- Multiple dependencies can exist between same capabilities (different types)
- Each shown as separate edge

## Details Panel Updates

When capability selected:
```
Capability: L2: Customer Management
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Depends On (Outgoing):
• Requires: L3: Identity Verification
• Enables: L2: Marketing Automation

Depended On By (Incoming):
• Required by: L3: Billing
• Supported by: L2: Analytics
```

## Edge Context Menu
Right-click on dependency edge:
```
┌──────────────┐
│ Delete       │
└──────────────┘
```

## Load Existing Dependencies
When canvas loads with capabilities:
1. Load all dependencies from backend
2. For each dependency where both capabilities are on canvas
3. Render dependency edge with appropriate styling

## Acceptance Criteria
- [ ] Connecting two capabilities shows connection type dialog
- [ ] User can select: Parent, Requires, Enables, or Supports
- [ ] Dependency edges created with correct styling per type (color, line style)
- [ ] Dependencies toggle button in toolbar
- [ ] Toggle hides/shows dependency edges
- [ ] Parent and realization edges always visible regardless of toggle
- [ ] Right-click dependency edge shows context menu with delete
- [ ] Dependencies can be deleted
- [ ] Capability details panel shows incoming and outgoing dependencies
- [ ] Existing dependencies loaded when canvas loads
- [ ] Multiple dependencies between same capabilities allowed (different types)

## Checklist
- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] User sign-off
