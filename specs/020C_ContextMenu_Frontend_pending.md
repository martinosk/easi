# Context Menu Operations - Frontend

## Description
Implement context menu (right-click) functionality for components and relations in both the tree view and canvas, providing users with quick access to delete, rename, and other management operations.

## Purpose
Enable efficient component and relation management through intuitive context menus that surface relevant actions based on the user's current context (tree view vs canvas).

## Dependencies

### Backend Dependencies
This feature requires the following backend functionality to be implemented first:
- **Spec 020A:** DELETE API endpoints for components and relations
- **Spec 020B:** Cascade deletion and cross-context integration
- OpenAPI specification must include the new DELETE endpoints
- HATEOAS links must include delete operations

**BLOCKER:** Frontend implementation cannot proceed until backend endpoints are available.

## Integration Requirements

### API-First Approach
- Frontend must consume ONLY the backend API endpoints documented in OpenAPI specification
- Use HATEOAS links from API responses to discover available operations
- Do not hardcode API URLs; construct from HATEOAS links
- Handle eventual consistency (operations may not reflect immediately in read models)

## Functional Requirements

### Tree View - Component Context Menu

**Trigger:** User right-clicks a component in the navigation tree

**Menu Options:**
- **Rename:** Allows inline editing of component name
- **Delete from Model:** Removes component from entire model (destructive)

**Rename Behavior:**
- Clicking "Rename" enables inline editing in the tree
- Input field appears with current component name selected
- User can type new name
- Pressing Enter submits the change
- Pressing Escape cancels the edit
- Clicking outside the input cancels the edit
- After successful rename, tree updates to show new name
- Uses existing PUT /api/v1/components/{id} endpoint
- Validation errors displayed via toast notification

**Delete from Model Behavior:**
- Clicking "Delete from Model" shows confirmation dialog
- Confirmation dialog explains this will:
  - Delete the component from the model
  - Remove it from ALL views
  - Delete ALL relations involving this component
  - Cannot be undone
- User must explicitly confirm
- After deletion, component disappears from tree
- Toast notification confirms successful deletion
- Uses DELETE /api/v1/components/{id} endpoint

### Canvas - Component Context Menu

**Trigger:** User right-clicks a component node on the canvas

**Menu Options:**
- **Delete from View:** Removes component from current view only
- **Delete from Model:** Removes component from entire model (destructive)

**Delete from View Behavior:**
- Clicking "Delete from View" immediately removes component from canvas
- No confirmation required (non-destructive operation)
- Component remains in model and other views
- Uses existing DELETE /api/v1/views/{viewId}/components/{componentId} endpoint
- Toast notification confirms removal

**Delete from Model Behavior:**
- Same as tree view delete from model
- Shows confirmation dialog
- Explains cascade deletion consequences
- After deletion, component disappears from ALL open views

### Canvas - Relation/Connection Context Menu

**Trigger:** User right-clicks a relation/edge on the canvas

**Menu Options:**
- **Delete from View:** Removes relation from current view only
- **Delete from Model:** Removes relation from entire model (destructive)

**Delete from View Behavior:**
- Clicking "Delete from View" immediately removes relation from canvas
- No confirmation required (non-destructive operation)
- Relation remains in model and other views
- Uses DELETE /api/v1/views/{viewId}/relations/{relationId} endpoint
- Toast notification confirms removal

**Delete from Model Behavior:**
- Clicking "Delete from Model" shows confirmation dialog
- Confirmation explains this will:
  - Delete the relation from the model
  - Remove it from ALL views
  - Cannot be undone
- After deletion, relation disappears from ALL open views
- Uses DELETE /api/v1/relations/{id} endpoint

## User Experience Requirements

### Context Menu Display
- Menu appears at cursor position when user right-clicks
- Menu is visually distinct from surrounding content
- Menu has clear visual hierarchy (icons, separators, styling)
- Menu automatically repositions if near screen edge (prevents overflow)
- Menu closes when clicking outside
- Menu closes when pressing Escape key
- Menu closes after selecting an option
- Only one context menu can be open at a time

### Confirmation Dialogs
- Appear centered in viewport with backdrop overlay
- Clearly explain the consequences of the action
- Display item name being deleted
- Have prominent Cancel and Confirm buttons
- Confirm button uses warning/danger styling
- Clicking backdrop closes dialog (same as Cancel)
- Pressing Escape closes dialog (same as Cancel)
- Pressing Enter confirms action (when dialog is focused)

### Toast Notifications
- Success toasts for completed operations (green, auto-dismiss)
- Error toasts for failed operations (red, manual dismiss or longer timeout)
- Toast messages are concise and action-oriented
- Toast position does not obscure important UI elements

### Loading States
- Show loading indicator during API calls
- Disable UI interactions during delete operations
- Loading state appears on confirmation dialog during processing
- Prevent duplicate submissions while processing

### Optimistic Updates
- For non-destructive operations (delete from view), update UI immediately
- For destructive operations (delete from model), wait for API confirmation
- On API error, rollback optimistic changes
- Display error message explaining what went wrong

## State Management

### Application State Store
- Extend existing Zustand store with delete and rename operations
- Store maintains single source of truth for components, relations, and views
- State updates trigger automatic re-renders in all consuming components

### Required Store Actions
- `deleteComponent(id)`: Delete component from model
- `deleteComponentFromView(viewId, componentId)`: Remove component from view
- `deleteRelation(id)`: Delete relation from model
- `deleteRelationFromView(viewId, relationId)`: Remove relation from view
- `renameComponent(id, name)`: Update component name

### State Synchronization
- When component deleted from model, remove from all views in state
- When component deleted from model, remove all associated relations from state
- When relation deleted from model, remove from all views in state
- State changes automatically update tree view and all canvas views

### Handling Eventual Consistency
- After delete operation, optimistically update local state
- If read model has not yet updated, cached state serves UI
- Periodic refresh or event-based updates sync with backend
- Handle race conditions gracefully (e.g., operating on recently deleted item)

## Component Architecture

### Context Menu Components
- **ContextMenu:** Reusable base component for displaying context menus
- **ComponentContextMenu:** Specialized menu for components on canvas
- **ConnectionContextMenu:** Specialized menu for relations/edges on canvas
- **TreeComponentContextMenu:** Specialized menu for components in tree view
- Menus share common styling and behavior through base component
- Menus receive different options based on context

### Confirmation Dialog Component
- **ConfirmationDialog:** Reusable modal dialog for confirming destructive actions
- Accepts title, message, confirm button text, cancel button text
- Calls onConfirm or onCancel callbacks
- Handles keyboard shortcuts (Enter, Escape)
- Manages focus trapping for accessibility

### Integration Points

**NavigationTree Component:**
- Add right-click event handler to component items
- Render TreeComponentContextMenu when triggered
- Handle inline rename editing state
- Call store actions for delete/rename operations

**ComponentCanvas Component:**
- Use React Flow's onNodeContextMenu event for component right-clicks
- Use React Flow's onEdgeContextMenu event for relation right-clicks
- Render appropriate context menus based on event target
- Call store actions for delete operations
- Sync React Flow state with Zustand store state

## Accessibility Requirements

### Keyboard Support
- Shift+F10 or Context Menu key opens context menu on focused item
- Arrow keys navigate menu items
- Enter key selects menu item
- Escape key closes menu
- Tab key should not enter menu (menu is ephemeral)

### Screen Reader Support
- Context menu has role="menu"
- Menu items have role="menuitem"
- Menu is labeled with aria-label
- Destructive actions have aria-describedby with warning text
- Focus returns to trigger element when menu closes

### Visual Accessibility
- Sufficient color contrast for menu items
- Focus indicator visible on keyboard navigation
- Hover states distinct from selected states
- Icon-only actions have text labels or aria-labels

## Error Handling

### API Errors
- Network errors: Display error toast, do not update UI
- 404 Not Found: Item already deleted, update UI to reflect
- 409 Conflict: Display specific conflict message to user
- 500 Server Error: Display generic error, log details

### Validation Errors
- Rename with empty name: Show validation error, keep edit mode active
- Rename with duplicate name: Show error from API response
- Invalid characters in name: Show client-side validation error

### Edge Cases
- Deleting component while another user is viewing it: Handle gracefully on refresh
- Deleting component that was already deleted: Treat as success, update UI
- Network timeout during delete: Show timeout error, allow retry

## React Flow Integration

### Node Context Menu
- React Flow provides onNodeContextMenu callback
- Callback receives event and node object
- Use event.clientX and event.clientY for menu position
- Node object contains node ID for API calls

### Edge Context Menu
- React Flow provides onEdgeContextMenu callback
- Callback receives event and edge object
- Use event coordinates for menu positioning
- Edge object contains edge/relation ID for API calls

### State Synchronization with React Flow
- When Zustand store updates, sync to React Flow nodes/edges
- Use useEffect hooks to watch store changes
- Update React Flow's setNodes and setEdges functions
- Maintain React Flow's internal state consistency

## Visual Design

### Context Menu Styling
- Use existing context menu CSS classes from NavigationTree
- Match application's design system (colors, spacing, typography)
- Menu has subtle shadow and border for depth
- Menu items have hover states
- Destructive actions use warning/danger color

### Confirmation Dialog Styling
- Use existing dialog overlay and content styles
- Backdrop has blur effect
- Dialog is centered with white background
- Confirm button uses danger styling (red)
- Cancel button uses secondary styling (gray)

## Testing Requirements

### Unit Tests
- Context menu component renders with correct options
- Context menu positions correctly on screen edge
- Confirmation dialog accepts/rejects based on user action
- Store actions make correct API calls
- Store state updates correctly after operations

### Integration Tests
- Right-click on tree component shows correct menu
- Right-click on canvas component shows correct menu
- Right-click on canvas relation shows correct menu
- Delete from view removes only from current view
- Delete from model removes from all views
- Rename updates component name everywhere
- Cascade deletion removes component and relations

### End-to-End Tests
- User can right-click component in tree and delete from model
- User can right-click component in tree and rename
- User can right-click component on canvas and delete from view
- User can right-click component on canvas and delete from model
- User can right-click relation on canvas and delete
- Confirmation dialogs prevent accidental deletion
- Error handling displays appropriate messages

### Accessibility Tests
- Keyboard navigation works in context menus
- Screen readers announce menu items correctly
- Focus management works correctly
- Color contrast meets WCAG AA standards

## Checklist

### Prerequisites
- [ ] Backend DELETE endpoints implemented (Spec 020A)
- [ ] Cascade deletion implemented (Spec 020B)
- [ ] OpenAPI specification includes DELETE operations
- [ ] HATEOAS links include delete operations

### API Client
- [ ] Add deleteComponent method to API client
- [ ] Add deleteRelation method to API client
- [ ] Add deleteComponentFromView method (verify existing)
- [ ] Add deleteRelationFromView method
- [ ] Add updateComponent method for rename (verify existing)
- [ ] Handle error responses appropriately

### State Management
- [ ] Extend Zustand store with delete actions
- [ ] Extend Zustand store with rename action
- [ ] Implement cascade deletion in store (remove relations when component deleted)
- [ ] Implement multi-view synchronization
- [ ] Handle optimistic updates with rollback

### Components
- [ ] Create ContextMenu base component
- [ ] Create ComponentContextMenu component
- [ ] Create ConnectionContextMenu component
- [ ] Create TreeComponentContextMenu component
- [ ] Create or update ConfirmationDialog component
- [ ] Context menus position correctly on screen
- [ ] Context menus close on outside click
- [ ] Context menus close on Escape key

### Tree View Integration
- [ ] Add right-click handler to tree component items
- [ ] Render TreeComponentContextMenu on right-click
- [ ] Implement inline rename editing
- [ ] Call store actions for delete operations
- [ ] Update tree view after operations

### Canvas Integration
- [ ] Add onNodeContextMenu handler to React Flow
- [ ] Add onEdgeContextMenu handler to React Flow
- [ ] Render ComponentContextMenu on node right-click
- [ ] Render ConnectionContextMenu on edge right-click
- [ ] Sync Zustand state changes to React Flow
- [ ] Call store actions for delete operations

### User Feedback
- [ ] Toast notifications for success operations
- [ ] Toast notifications for error operations
- [ ] Loading states during API calls
- [ ] Confirmation dialogs for destructive operations
- [ ] Clear error messages for all failure scenarios

### Accessibility
- [ ] Keyboard navigation in context menus (Arrow keys, Enter, Escape)
- [ ] Keyboard shortcut to open context menu (Shift+F10)
- [ ] ARIA roles and labels on menus
- [ ] Focus management (return to trigger after close)
- [ ] Color contrast meets standards
- [ ] Screen reader testing completed

### Testing
- [ ] Unit tests for all context menu components
- [ ] Unit tests for store actions
- [ ] Integration tests for tree view operations
- [ ] Integration tests for canvas operations
- [ ] Integration tests for cascade deletion
- [ ] E2E test: Delete component from model via tree
- [ ] E2E test: Delete component from view via canvas
- [ ] E2E test: Rename component via tree
- [ ] E2E test: Delete relation from canvas
- [ ] E2E test: Cascade deletion updates all views
- [ ] Accessibility tests for keyboard navigation
- [ ] Error handling tests for all failure scenarios

### Documentation
- [ ] User guide for context menu operations
- [ ] Developer documentation for component architecture
- [ ] API integration documentation

### Final
- [ ] All tests passing
- [ ] Visual design matches application style
- [ ] Accessibility requirements met
- [ ] User acceptance testing completed
- [ ] User sign-off
