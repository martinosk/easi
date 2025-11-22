# Rename Components to Applications

## Description
Update all frontend user-facing text from "Component" to "Application" to better align with enterprise architecture terminology. Backend API and internal naming remain unchanged.

## User Need
As a user, I need consistent terminology that matches enterprise architecture standards, where systems are referred to as "Applications" rather than "Components."

## Scope

### UI Text Changes

| Location | Before | After |
|----------|--------|-------|
| Navigation tree section | "Components" | "Applications" |
| Add button | "Add Component" | "Add Application" |
| Create dialog title | "Create Component" | "Create Application" |
| Edit dialog title | "Edit Component" | "Edit Application" |
| Details panel header | "Component Details" | "Application Details" |
| Relation dialogs | "Source Component" | "Source Application" |
| Relation dialogs | "Target Component" | "Target Application" |
| Context menu | "Add Component" | "Add Application" |
| Context menu | "Remove Component from View" | "Remove Application from View" |
| Error messages | "Failed to create component" | "Failed to create application" |
| Error messages | "Component not found" | "Application not found" |

### Files to Update
- NavigationTree component
- CreateComponentDialog → CreateApplicationDialog
- EditComponentDialog → EditApplicationDialog
- ComponentDetails → ApplicationDetails
- MainLayout (labels and props)
- Toolbar (button labels)
- ContextMenu (menu item labels)
- DialogManager
- Related test files

### What Stays the Same
- Backend API endpoints (`/api/v1/components`)
- API client method names
- TypeScript interfaces (Component type)
- Import paths for API types
- Internal variable names (unless user-facing)

## Implementation Strategy

**Phase 1: File Renames**
- Rename component files to use "Application" naming
- Update imports across codebase

**Phase 2: UI Text Updates**
- Update all visible labels in JSX
- Update button text
- Update dialog titles
- Update form field labels
- Update placeholder text
- Update error messages

**Phase 3: Testing**
- Manual testing of all dialogs and forms
- Verify all labels updated
- Run automated tests
- Check for missed "Component" references in UI

## Backward Compatibility

API layer unchanged - frontend continues to call "components" endpoints. Only user-facing text is updated.

## Testing Checklist
- Navigation tree shows "Applications" section
- "Add Application" button appears in tree
- Create dialog title is "Create Application"
- Edit dialog title is "Edit Application"
- Details panel header shows "Application Details"
- Form labels use "application" terminology
- Relation dialogs show "Source Application" / "Target Application"
- Context menus use "application" terminology
- Error messages use "application" terminology
- No user-visible "component" references remain
- All tests pass
- API calls still work

## Acceptance Criteria
- [ ] All UI text updated from "Component" to "Application"
- [ ] Navigation tree section renamed
- [ ] All dialog titles updated
- [ ] All button labels updated
- [ ] All form field labels updated
- [ ] Context menus updated
- [ ] Error messages updated
- [ ] Backend API calls unchanged (still work)
- [ ] No breaking changes to functionality
- [ ] All tests pass
- [ ] No visible "Component" references in UI

## Checklist
- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] User sign-off
