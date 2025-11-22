# Capability Tree in Sidebar Navigation

## Description
Add a "Capabilities" section to the existing NavigationTree component that displays the capability hierarchy in an expandable tree view with color coding by maturity level. Includes seed data for testing and validation.

## User Need
As a user, I need to browse the capability hierarchy in the sidebar so that I can understand the capability structure and select capabilities to work with on the canvas.

## UI Design

### Sidebar Section Structure
```
┌─────────────────────────┐
│ [Components]            │ ← Existing
│   └─ Component A        │
│                         │
│ [Capabilities]          │ ← New section
│   ▼ L1: Customer Mgmt   │ ← Color coded by maturity
│      ▶ L2: Onboarding   │
│      ▼ L2: Support      │
│         └─ L3: Ticketing│
│   ▼ L1: Finance         │
│      └─ L2: Billing     │
└─────────────────────────┘
```

### Visual Elements
- **Indentation**: 16px per level
- **Expand/Collapse Icons**: ▶ (collapsed) / ▼ (expanded)
- **Level Labels**: Show "L1:", "L2:", "L3:", "L4:" prefix before name
- **Color Coding** (background color based on maturity):
  - Initial: Gray
  - Developing: Yellow
  - Established: Green
  - Optimized: Blue
- **Hover State**: Subtle highlight
- **Selection State**: Bold text, darker background
- **Draggable**: Each node can be dragged onto canvas

## Functional Requirements

1. **Tree Structure**: Build hierarchical tree from flat capability list using parentId relationships
2. **Expand/Collapse**: Parent nodes can be expanded/collapsed
3. **Color Coding**: Node background reflects maturity level
4. **Drag Support**: Nodes are draggable onto the canvas
5. **Sorting**: Root nodes sorted alphabetically by name
6. **Loading**: Capabilities loaded on app mount
7. **Orphan Handling**: Capabilities without parents appear as roots (L1)

## Seed Data

Create SQL migration with sample capability hierarchy for testing:

**L1 Capabilities** (3):
- Customer Management (Developing)
- Finance (Established)
- Product Management (Initial)

**L2 Capabilities** (5):
- Customer Onboarding → under Customer Management
- Customer Support → under Customer Management
- Billing → under Finance
- Financial Reporting → under Finance
- Product Catalog → under Product Management

**L3 Capabilities** (3):
- Ticketing System → under Customer Support
- KYC Verification → under Customer Onboarding
- Pricing Management → under Product Catalog

## Acceptance Criteria
- [x] "Capabilities" section appears in NavigationTree sidebar
- [x] Hierarchical tree structure displays correctly (L1→L2→L3→L4)
- [x] Expand/collapse functionality works for parent nodes
- [x] Level labels (L1:, L2:, etc.) displayed before capability names
- [x] Color coding applied based on maturity level
- [x] Capabilities are draggable from tree
- [x] Seed data migration creates sample capability hierarchy
- [x] Tree correctly handles orphan capabilities as roots
- [x] Root nodes sorted alphabetically
- [x] Capabilities load on application startup

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] User sign-off
