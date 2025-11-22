# Create Capability from Tree View

## Description
Add "Add Capability" button to the capabilities section in the sidebar tree view. Clicking opens a dialog where users can create new capabilities. Newly created capabilities start as L1 (root) and can later be connected on the canvas to establish parent-child relationships.

## User Need
As a user, I need to create new capabilities directly from the sidebar so that I can quickly add capabilities to my architecture without navigating away from the main view.

## UI Design

### Button Placement
```
┌─────────────────────────┐
│ [Capabilities]  [+ Add] │ ← Button next to section header
│   ▼ L1: Customer Mgmt   │
└─────────────────────────┘
```

### Dialog Layout
```
┌─────────────────────────────────────┐
│  Create Capability              [X] │
├─────────────────────────────────────┤
│  Name: *                            │
│  [________________________]         │
│                                     │
│  Description:                       │
│  [________________________]         │
│  [________________________]         │
│                                     │
│  Status:                            │
│  [Active              ▼]            │
│                                     │
│  Maturity Level:                    │
│  [Initial             ▼]            │
│                                     │
│           [Cancel]  [Create]        │
└─────────────────────────────────────┘
```

## Functional Requirements

1. **Button**: "+ Add" button appears next to "Capabilities" section header
2. **Dialog**: Opens modal dialog with form fields
3. **Level Assignment**: Capabilities created as L1 (orphan) - level adjusts when parent connected later
4. **Refresh**: Tree refreshes after successful creation

## Form Fields

| Field | Required | Validation | Default |
|-------|----------|------------|---------|
| Name | Yes | 1-200 characters | - |
| Description | No | Max 1000 characters | - |
| Status | Yes | Enum: Active, Planned, Deprecated, Retired | Active |
| Maturity Level | Yes | Enum: Initial, Developing, Established, Optimized | Initial |

## User Flow

1. User clicks "+ Add" button in Capabilities section
2. Dialog opens with empty form
3. User enters capability name (required)
4. User optionally enters description
5. User selects status (defaults to Active)
6. User selects maturity level (defaults to Initial)
7. User clicks "Create"
8. System creates capability as L1 (root)
9. System updates metadata (status, maturity)
10. System reloads capability tree
11. New capability appears in tree as L1
12. Dialog closes

## Error Handling

- Validation errors displayed inline below fields
- Backend errors displayed at bottom of form
- Submit button disabled during submission
- "Creating..." text shown during submission

## Acceptance Criteria
- [ ] "+ Add" button appears next to "Capabilities" header in sidebar
- [ ] Clicking button opens CreateCapabilityDialog
- [ ] Dialog contains all fields: name, description, status, maturity level
- [ ] Name field is required and validated (1-200 chars)
- [ ] Description is optional and validated (max 1000 chars)
- [ ] Status dropdown has valid options with Active as default
- [ ] Maturity level dropdown has valid options with Initial as default
- [ ] Form validation shows errors inline
- [ ] Successful creation refreshes capability tree
- [ ] New capability appears in tree as L1
- [ ] Dialog closes after successful creation
- [ ] Cancel button closes dialog without creating
- [ ] Backend errors are displayed to user

## Checklist
- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] User sign-off
