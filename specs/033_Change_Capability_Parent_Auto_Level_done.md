# Change Capability Parent with Auto-Level Calculation

## Description
Implement backend command and domain logic to change a capability's parent and automatically recalculate capability levels for the entire affected subtree. This enables intuitive visual hierarchy modeling on the canvas where users connect capabilities and the system automatically determines appropriate levels (L1→L2→L3→L4).

## User Need
As a user, I need to create parent-child relationships between capabilities on the canvas without manually managing level assignments, so that I can intuitively model capability hierarchies by connecting nodes.

## Business Rules

1. **Auto-Level Calculation**: When a capability's parent changes, recalculate levels for the capability and its entire subtree
   - Root capabilities (no parent) are L1
   - Children of L1 are L2
   - Children of L2 are L3
   - Children of L3 are L4

2. **Orphan Capabilities Allowed**: Any capability can exist without a parent (will be L1 until connected)

3. **Maximum Depth Validation**: Block operations that would create L5 or deeper hierarchies

4. **Subtree Recalculation**: When moving a subtree, all descendants must recalculate their levels

5. **Circular Reference Prevention**: Cannot create cycles (A→B→C→A)

6. **Self-Reference Prevention**: Cannot set a capability as its own parent

## Domain Changes Required

**New Command**: `ChangeCapabilityParent`
- Input: capabilityId, newParentId (empty string means make it root/L1)

**New Domain Event**: `CapabilityParentChanged`
- Contains: capabilityId, oldParentId, newParentId, oldLevel, newLevel, timestamp

**New Aggregate Method**: `ChangeParent`
- Validates maximum depth
- Raises event
- Updates internal state

**Domain Model Relaxation**:
- Remove strict validation that non-L1 must have parent
- Allow capabilities to be created as orphans

## API Endpoint

**Request**: `PATCH /api/v1/capabilities/{id}/parent`

Body: `{ "parentId": "uuid-of-new-parent" }` (empty string for root)

**Response**: `204 No Content` on success

**Error Responses**:
- `400 Bad Request`: Operation would create L5+ hierarchy
- `400 Bad Request`: Circular reference detected
- `400 Bad Request`: Self-reference attempted
- `404 Not Found`: Capability or parent not found

## Acceptance Criteria
- [x] `ChangeCapabilityParent` command implemented
- [x] `CapabilityParentChanged` event defined and projected
- [x] Command handler implements recursive level recalculation for all descendants
- [x] Maximum depth validation prevents L5+ hierarchies with clear error message
- [x] Circular reference detection prevents invalid hierarchies
- [x] Self-reference validation prevents capability being its own parent
- [x] Read model updated with parent and level changes
- [x] API endpoint `PATCH /api/v1/capabilities/:id/parent` implemented
- [x] Domain model relaxed to allow orphan capabilities at creation
- [x] Setting parentId to empty string makes capability L1 (root)

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] Documentation updated if needed
- [x] User sign-off
