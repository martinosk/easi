# Capability Nodes on Canvas with Parent Relations

## Description
Enable dragging capabilities from the sidebar tree onto the canvas and connecting them with parent-child relationships. Parent edges trigger backend auto-level calculation. Capability nodes are visually distinct from application nodes.

## User Need
As a user, I need to visualize and model capability hierarchies on the canvas by connecting nodes, so that I can intuitively build and understand the capability structure.

## Visual Design

### Capability Node
```
┌──────────────────────┐
│  ◆ L1: Finance       │  ← Diamond icon, level prefix
│                      │
│  Maturity: Established│  ← Color-coded badge
└──────────────────────┘
```

**Distinctions from Application Nodes:**
| Attribute | Capability | Application |
|-----------|------------|-------------|
| Icon | Diamond (◆) | Circle (●) |
| Border | Solid | Dashed |
| Width | 200px | 180px |
| Background | Maturity color | White |

### Parent Edge
```
    ┌────────┐
    │  L1    │
    └───┬────┘
        │ ← Thick solid gray line
        ▼
    ┌────────┐
    │  L2    │
    └────────┘
```

**Edge Styling:**
- Line style: Solid
- Color: Gray (#374151)
- Width: 3px
- Label: "Parent"
- Arrow: Standard arrowhead

## Functional Requirements

### Drag from Sidebar
1. Drag capability node from sidebar tree onto canvas
2. Node appears at drop position
3. Node displays: level prefix, name, maturity badge
4. Node background colored by maturity level

### Parent Connection
1. Connect source capability to target capability
2. Source becomes parent, target becomes child
3. Backend `PATCH /api/v1/capabilities/:id/parent` called
4. Levels auto-recalculated
5. Edge displayed with parent styling

### Level Validation
1. If connection would create L5+, backend returns error
2. Error displayed to user: "Cannot create this parent relationship: would result in hierarchy deeper than L4"
3. Edge not created

### Remove Parent Relationship
1. Delete parent edge on canvas
2. Backend called with empty parentId
3. Child capability becomes orphan (L1)
4. Levels recalculated

### Canvas Operations
- **Select**: Click node to select, show details panel
- **Move**: Drag to reposition
- **Remove**: Remove from canvas (does not delete capability)
- **Connect**: Drag from source to target to create parent relationship

## Details Panel (when capability selected)

Shows:
- Name, level, description
- Maturity level
- Status
- Ownership information
- Experts list
- Tags list

Actions:
- Edit button
- Delete button
- Remove from Canvas button

## Acceptance Criteria
- [x] Capabilities can be dragged from sidebar onto canvas
- [x] Capability nodes visually distinct from applications (diamond icon, maturity color, solid border)
- [x] Level prefix shown (L1:, L2:, etc.)
- [x] Connecting two capabilities creates parent-child relationship
- [x] Backend API called to update parent
- [x] Capability levels automatically recalculated and displayed
- [x] Error shown if connection would create L5+
- [x] Deleting parent edge makes child orphan (L1)
- [x] Capability nodes can be selected
- [x] Selected capability shows details in panel
- [x] Capabilities can be removed from canvas (not deleted)
- [x] Parent edges visually distinct (thick gray solid line with label)

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] User sign-off
