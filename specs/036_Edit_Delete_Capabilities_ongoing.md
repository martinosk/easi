# Edit and Delete Capabilities

## Description
Add capability editing and deletion functionality with dialogs for updating capability properties, metadata, experts, and tags. Users can modify capabilities through context menu or detail panel actions.

## User Need
As a user, I need to edit and delete capabilities so that I can keep my capability model accurate and up-to-date.

## UI Design

### Context Menu Actions
Right-click on capability in tree or canvas:
```
┌──────────────┐
│ Edit         │
│ Delete       │
│ Add Expert   │
│ Add Tag      │
└──────────────┘
```

### Edit Capability Dialog
```
┌─────────────────────────────────────┐
│  Edit Capability                [X] │
├─────────────────────────────────────┤
│  Name: *                            │
│  [Customer Management           ]   │
│                                     │
│  Description:                       │
│  [Manage customer lifecycle     ]   │
│                                     │
│  Status: [Active ▼]                 │
│  Maturity Level: [Genesis ▼]    │
│  Ownership Model: [Shared ▼]        │
│  Primary Owner: [John Doe       ]   │
│  EA Owner: [Jane Smith          ]   │
│  Strategy Pillar: [Customer Exp ▼]  │
│  Pillar Weight: [75             ]   │
│                                     │
│  Experts:                           │
│  • Alice (Architect) - alice@...    │
│  • Bob (Lead) - bob@...             │
│  [+ Add Expert]                     │
│                                     │
│  Tags: [Critical] [Core] [Strategic]│
│  [+ Add Tag]                        │
│                                     │
│           [Cancel]  [Save]          │
└─────────────────────────────────────┘
```

### Delete Confirmation
```
┌─────────────────────────────────────┐
│  Delete Capability?             [X] │
├─────────────────────────────────────┤
│  Are you sure you want to delete    │
│  "Customer Management"?             │
│                                     │
│  This action cannot be undone.      │
│                                     │
│           [Cancel]  [Delete]        │
└─────────────────────────────────────┘
```

## Functional Requirements

### Edit
1. Context menu "Edit" opens edit dialog
2. Dialog pre-populated with existing capability data
3. Update capability name and description via UpdateCapability API
4. Update metadata via UpdateCapabilityMetadata API
5. List existing experts with ability to add new ones
6. List existing tags with ability to add new ones
7. Tree/canvas refreshes after save

### Delete
1. Context menu "Delete" shows confirmation dialog
2. Confirmation shows capability name
3. Successful delete removes capability from tree/canvas
4. Backend returns error if capability has children
5. Error message: "Cannot delete capability with children. Delete child capabilities first."

### Add Expert Dialog
```
┌─────────────────────────────────────┐
│  Add Expert                     [X] │
├─────────────────────────────────────┤
│  Name: *    [________________]      │
│  Role: *    [________________]      │
│  Contact: * [________________]      │
│           [Cancel]  [Add]           │
└─────────────────────────────────────┘
```

### Add Tag
Simple input field or dialog for entering tag name.

## Form Fields

| Field | API | Validation |
|-------|-----|------------|
| Name | UpdateCapability | Required, 1-200 chars |
| Description | UpdateCapability | Max 1000 chars |
| Status | UpdateCapabilityMetadata | Valid enum |
| Maturity Level | UpdateCapabilityMetadata | Valid enum |
| Ownership Model | UpdateCapabilityMetadata | Valid enum |
| Primary Owner | UpdateCapabilityMetadata | Text |
| EA Owner | UpdateCapabilityMetadata | Text |
| Strategy Pillar | UpdateCapabilityMetadata | Valid enum |
| Pillar Weight | UpdateCapabilityMetadata | 0-100 |

## Acceptance Criteria
- [x] Context menu appears on right-click of capability in tree and canvas
- [x] Edit action opens dialog with pre-populated data
- [x] All fields can be edited (name, description, metadata)
- [x] Experts can be added and are listed
- [x] Tags can be added and are displayed as chips/badges
- [x] Save button updates capability via API
- [x] Validation errors displayed inline
- [x] Delete action shows confirmation dialog
- [x] Successful delete removes capability from tree/canvas
- [x] Delete fails with error if capability has children
- [x] All changes refresh capability tree
- [x] Backend errors displayed to user
- [x] Cancel closes dialogs without changes

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] User sign-off
