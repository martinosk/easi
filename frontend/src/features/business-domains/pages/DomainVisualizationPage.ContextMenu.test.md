# Context Menu Test Suite - Slice 1

## Testing Strategy

This test suite validates the **Capability Grid Context Menu (Slice 1)** feature using behavior-driven testing principles. The tests focus on **business rules and user interactions** rather than implementation details, ensuring they remain stable as the code evolves.

### Test Philosophy

1. **Minimal Mocking**: Only mock external dependencies (hooks, store). The component tree renders real DOM elements.
2. **Behavior Over Implementation**: Tests validate what the user sees and does, not how the code works internally.
3. **Specification Translation**: Each test maps directly to a requirement in spec 075.
4. **L1 Ancestor Resolution**: Tests verify that operations on L2/L3/L4 capabilities correctly resolve to their L1 ancestor.

## Test Categories

### 1. Opening Context Menu (4 tests)

**Business Rule**: Right-clicking any capability (L1-L4) opens a context menu at the click position.

- **Test L1**: Validates context menu opens at correct position when right-clicking L1 capability
- **Test L2**: Validates context menu opens when right-clicking L2 capability
- **Test L3**: Validates context menu opens when right-clicking L3 capability
- **Test L4**: Validates context menu opens when right-clicking L4 capability

**Why**: Ensures the context menu trigger works consistently across all capability levels.

### 2. Context Menu Options (3 tests)

**Business Rule**: Context menu displays two options: "Remove from Business Domain" and "Delete from Model".

- **Test Remove Option**: Verifies "Remove from Business Domain" option is present
- **Test Delete Option**: Verifies "Delete from Model" option is present
- **Test Option Order**: Validates both options appear in correct order

**Why**: Ensures users see the correct action choices as specified.

### 3. Remove from Business Domain Action (5 tests)

**Business Rule**: "Remove from Business Domain" dissociates the L1 capability (and all children) from the current domain.

- **Test L1 Remove**: Calls `dissociateCapability` with L1 when removing L1
- **Test L2 Remove**: Calls `dissociateCapability` with L1 ancestor when removing L2
- **Test L3 Remove**: Calls `dissociateCapability` with L1 ancestor when removing L3
- **Test L4 Remove**: Calls `dissociateCapability` with L1 ancestor when removing L4
- **Test Refetch**: Verifies grid refreshes after successful dissociate
- **Test Menu Close**: Verifies context menu closes after action completes

**Why**: Validates the core business rule that dissociation always targets the L1 ancestor, not the clicked capability.

### 4. Delete from Model Action (9 tests)

**Business Rule**: "Delete from Model" permanently deletes the L1 capability and all children from the entire model after confirmation.

- **Test Confirmation Dialog**: Opens confirmation dialog when clicking delete
- **Test Capability Name**: Shows capability name in confirmation dialog
- **Test L1 Delete**: Calls `deleteCapability` with L1 ID when confirming delete on L1
- **Test L2 Delete**: Calls `deleteCapability` with L1 ancestor ID when confirming delete on L2
- **Test L3 Delete**: Calls `deleteCapability` with L1 ancestor ID when confirming delete on L3
- **Test L4 Delete**: Calls `deleteCapability` with L1 ancestor ID when confirming delete on L4
- **Test Cancel**: Does not call `deleteCapability` when canceling confirmation
- **Test Refetch**: Verifies grid refreshes after successful delete
- **Test Menu Close**: Verifies context menu closes after opening confirmation

**Why**: Ensures destructive delete action requires confirmation and always targets the L1 ancestor.

### 5. Closing Context Menu (3 tests)

**Business Rule**: Context menu closes on action completion, outside click, or Escape key.

- **Test Outside Click**: Closes when clicking outside menu
- **Test Escape Key**: Closes when pressing Escape
- **Test Inside Click**: Does NOT close when clicking inside menu

**Why**: Validates proper UX for dismissing the context menu.

### 6. L1 Ancestor Resolution (4 tests)

**Business Rule**: When right-clicking any capability (L1-L4), operations resolve to the L1 parent.

- **Test L2 Resolution**: Resolves L2 → L1 (one level up)
- **Test L3 Resolution**: Resolves L3 → L2 → L1 (two levels up)
- **Test L4 Resolution**: Resolves L4 → L3 → L2 → L1 (three levels up)
- **Test L1 Identity**: L1 capability returns itself

**Why**: Core domain logic - ensures operations maintain domain assignment consistency by always targeting L1.

### 7. Context Menu State Management (4 tests)

**Business Rule**: Context menu tracks position and target capability correctly.

- **Test Position Tracking**: Verifies context menu appears at click coordinates
- **Test Target Tracking**: Verifies correct capability is targeted for operations
- **Test State Clear**: Verifies state clears when menu closes
- **Test New Menu**: Verifies new position when opening menu on different capability

**Why**: Ensures context menu state is properly managed across interactions.

### 8. No Interference with Existing Functionality (2 tests)

**Business Rule**: Context menu does not break existing capability click behavior.

- **Test Detail Panel**: Left-click still opens detail panel
- **Test No Menu on Left Click**: Left-click does not open context menu

**Why**: Regression testing to ensure new feature doesn't break existing behavior.

## Test Data Structure

### Mock Capabilities
- **L1-1**: Financial Management (L1)
  - **L2-1**: Accounting (L2) → parent: L1-1
    - **L3-1**: General Ledger (L3) → parent: L2-1
      - **L4-1**: Journal Entries (L4) → parent: L3-1
- **L1-2**: Treasury (L1)

This hierarchy allows testing L1 ancestor resolution across all four levels.

### Mock Dependencies
- `useBusinessDomains`: Returns Finance and HR domains
- `useDomainCapabilities`: Returns capabilities with `dissociateCapability` and `refetch` functions
- `useAppStore`: Returns `deleteCapability` function
- Other hooks: Mocked with minimal implementations

## Expected Test Results (RED Phase)

**Status**: All 33 feature tests should FAIL initially because the context menu functionality is not yet implemented.

**Passing Tests**: The 2 tests in "No Interference" category should PASS because they test existing functionality.

**Total**: 33 failures + 2 passes = 35 tests

## Implementation Checklist (to make tests GREEN)

From spec 075, Slice 1:

- [ ] Add `onContextMenu` handler to `NestedCapabilityItem` component
- [ ] Track context menu state (position, target capability) in `DomainVisualizationPage`
- [ ] Render `ContextMenu` component with two options when triggered
- [ ] Implement "Remove from Business Domain" action calling `dissociateCapability` for L1 ancestor
- [ ] Implement "Delete from Model" action opening confirmation dialog, then calling `deleteCapability` for L1 ancestor
- [ ] Implement L1 ancestor resolution logic (traverse parentId chain until level === 'L1')
- [ ] Call `refetch` after successful remove/delete operations
- [ ] Close context menu on action completion, outside click, or Escape key

## Notes

1. **Why Test L1 Resolution Separately?**: While the remove/delete tests verify L1 resolution works, dedicated resolution tests document this critical business logic explicitly.

2. **Why Test Position Tracking?**: Context menu position is part of the user experience - it should appear where the user right-clicked.

3. **Why Test Both Cancel and Confirm?**: Delete is destructive - we must ensure users can cancel without triggering the delete.

4. **Why Mock Hooks Instead of API?**: This is a component integration test. We trust the hooks work correctly (they have their own tests). We validate that the page component orchestrates them properly.

5. **Why 35 Tests for One Feature?**: Each test validates a single business rule or edge case. Small, focused tests make debugging easier and document the specification clearly.
