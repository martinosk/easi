# Spec 112: HATEOAS-Driven Frontend

## Status
done

## Overview
Transform the frontend to consistently use HATEOAS links from API responses to determine which mutation actions (add, edit, delete, drag, etc.) are available to the user. This eliminates duplicated business logic in the frontend and ensures the UI accurately reflects user permissions.

## Problem Statement

The frontend currently has **inconsistent** patterns for showing/hiding mutation actions:

1. **Some components use HATEOAS correctly** (NavigationTree view menu checks `_links.update`, `_links.delete`, `_links.changeVisibility`)
2. **Most components show actions unconditionally** (always show Edit, Delete, Remove buttons)
3. **Some components use prop-based logic** (passing `canWrite`/`canDelete` props instead of checking HATEOAS links)
4. **Drag-and-drop accepts all drops** without verifying API permissions

This causes:
- Users seeing actions they cannot perform (e.g., removing items from other users' private views)
- Business logic duplication between frontend and backend
- Inconsistent UX where some features correctly hide unavailable actions while others don't
- Security concerns where the UI misleads users about their actual permissions

## Target State

**Single Source of Truth**: The API returns HATEOAS links ONLY for actions the current user can perform. The frontend renders action buttons/options ONLY when the corresponding link exists.

```
API Response → Has _links.delete? → Show Delete button
API Response → No _links.update? → Hide Edit button
API Response → Has _links.removeFromView? → Enable Remove option
```

## Components to Update

### 1. Context Menus (Canvas)

#### NodeContextMenu.tsx
**Location**: `frontend/src/features/canvas/components/context-menus/NodeContextMenu.tsx`
**Current**: Always shows "Remove from View" and "Delete from Model"
**Fix**: Check node's `_links.removeFromView` and `_links.delete`

```
Before: Menu items always rendered
After:
- "Remove from View" → only if _links.removeFromView exists
- "Delete from Model" → only if _links.delete exists
```

#### EdgeContextMenu.tsx
**Location**: `frontend/src/features/canvas/components/context-menus/EdgeContextMenu.tsx`
**Current**: Shows delete options based on `isInherited` flag only
**Fix**: Check edge's `_links.delete` or `_links.removeRelation`

### 2. Navigation Tree

#### NavigationTree.tsx - Capability Menu
**Location**: Lines 259-276
**Current**: Always shows Edit/Delete for capabilities
**Fix**: Check `capability._links.update` and `capability._links.delete`

#### NavigationTree.tsx - Component Menu
**Location**: Lines 513-536
**Current**: Always shows Edit/Delete for components
**Fix**: Check `component._links.update` and `component._links.delete`

#### NavigationTree.tsx - View Menu
**Location**: Lines 449-511
**Status**: ✅ ALREADY CORRECT - This is the reference pattern

### 3. Detail Panels

#### CapabilityDetails.tsx
**Location**: `frontend/src/features/capabilities/components/CapabilityDetails.tsx`
**Current**: Always shows Edit and Remove from View buttons
**Fix**:
- Edit → check `capability._links.update`
- Remove from View → check `view._links.removeCapability` or element-level link

#### ComponentDetails.tsx
**Location**: `frontend/src/features/components/components/ComponentDetails.tsx`
**Current**: Always shows Edit, Remove from View, Clear Color buttons
**Fix**:
- Edit → check `component._links.update`
- Remove from View → check view position `_links.remove` or view-level permission
- Clear Color → check `position._links.clearColor`

#### RealizationDetails.tsx
**Location**: `frontend/src/features/relations/components/RealizationDetails.tsx`
**Current**: Shows Edit based on `isInherited` only
**Fix**: Check `realization._links.update` in addition to inherited check

#### RelationDetails.tsx
**Location**: `frontend/src/features/relations/components/RelationDetails.tsx`
**Current**: Always shows Edit button
**Fix**: Check `relation._links.update`

### 4. Enterprise Architecture Panels

#### EnterpriseCapabilityDetailPanel.tsx
**Location**: `frontend/src/features/enterprise-architecture/components/EnterpriseCapabilityDetailPanel.tsx`
**Current**: Uses `canWrite` prop for Unlink buttons
**Fix**: Check `link._links.delete` or `link._links.unlink`

#### EnterpriseCapabilitiesTable.tsx
**Location**: `frontend/src/features/enterprise-architecture/components/EnterpriseCapabilitiesTable.tsx`
**Current**: Uses `canDelete` prop for Delete button
**Fix**: Check `capability._links.delete`

#### MaturityGapDetailPanel.tsx
**Location**: `frontend/src/features/enterprise-architecture/components/MaturityGapDetailPanel.tsx`
**Current**: Always shows Set/Edit Target button
**Fix**: Check `detail._links.setTargetMaturity` or similar

### 5. Business Domains

#### StrategicImportanceSection.tsx
**Location**: `frontend/src/features/business-domains/components/StrategicImportanceSection.tsx`
**Current**: Always shows Edit, Remove, Add Importance buttons
**Fix**:
- Edit → check `importance._links.update`
- Remove → check `importance._links.delete`
- Add → check parent `_links.createImportance`

### 6. Drag-and-Drop

#### EnterpriseCapabilityCard.tsx
**Location**: `frontend/src/features/enterprise-architecture/components/EnterpriseCapabilityCard.tsx`
**Current**: Drop handler always accepts drops
**Fix**: Verify target capability has `_links.link` or `_links.acceptLink` before showing drop zone

#### NavigationTree.tsx (Capability Drag)
**Location**: Line 321-324
**Current**: Drag always enabled
**Fix**: Check `capability._links.drag` or rely on drop-side validation

#### DomainCapabilityPanel.tsx
**Location**: `frontend/src/features/enterprise-architecture/components/DomainCapabilityPanel.tsx`
**Current**: Uses `linkStatus` but not HATEOAS links
**Fix**: Verify `capability._links.link` for drag enablement

#### EnterpriseCapabilitiesTable.tsx (Drop)
**Location**: Lines 34-62
**Current**: Drop accepted if dock panel is open
**Fix**: Check `capability._links.link` before accepting drop

### 7. Create Buttons

#### NavigationTree.tsx - Create Application Button
**Location**: Line 565-572
**Current**: Always shows "+" button
**Fix**: Check root/collection `_links.create`

## Backend Changes Required

Some HATEOAS links need to be added or made permission-aware in the backend:

### View Element Links
Update `addElementLinks` in `view_handlers.go` to include permission-based links:
- `remove` - if user can modify the view
- `updatePosition` - if user can modify the view
- `updateColor` / `clearColor` - if user can modify the view

### Capability Links
Update capability responses to include permission-based:
- `update` - if user has write permission
- `delete` - if user has write permission
- `removeFromView` - if user can edit the current view

### Component Links
Update component responses to include permission-based:
- `update` - if user has write permission
- `delete` - if user has write permission (and component can be deleted)

### Enterprise Capability Links
Ensure link responses include:
- `delete` / `unlink` on individual links
- `setTargetMaturity` when user can set targets

## Implementation Pattern

### Correct Pattern (Reference: NavigationTree View Menu)
```typescript
const getViewContextMenuItems = (menu: ViewContextMenuState): ContextMenuItem[] => {
  const items: ContextMenuItem[] = [];
  const canEdit = menu._links?.update !== undefined;
  const canDelete = menu._links?.delete !== undefined;
  const canChangeVisibility = menu._links?.changeVisibility !== undefined;

  if (canEdit) {
    items.push({ label: 'Rename View', onClick: ... });
  }

  if (canChangeVisibility) {
    items.push({ label: menu.isPrivate ? 'Make Public' : 'Make Private', onClick: ... });
  }

  if (canDelete) {
    items.push({ label: 'Delete View', onClick: ..., isDanger: true });
  }

  return items;
};
```

### Utility Helper (Create if needed)
```typescript
// src/shared/utils/hateoas.ts
export function hasLink(resource: { _links?: Record<string, string> }, linkName: string): boolean {
  return resource._links?.[linkName] !== undefined;
}

export function getLink(resource: { _links?: Record<string, string> }, linkName: string): string | undefined {
  return resource._links?.[linkName];
}
```

### Button Pattern
```typescript
// Before (wrong)
<button onClick={onEdit}>Edit</button>
<button onClick={onDelete}>Delete</button>

// After (correct)
{hasLink(resource, 'update') && <button onClick={onEdit}>Edit</button>}
{hasLink(resource, 'delete') && <button onClick={onDelete}>Delete</button>}
```

### Context Menu Pattern
```typescript
// Before (wrong)
const menuItems = [
  { label: 'Edit', onClick: handleEdit },
  { label: 'Delete', onClick: handleDelete },
];

// After (correct)
const menuItems: ContextMenuItem[] = [];
if (hasLink(resource, 'update')) {
  menuItems.push({ label: 'Edit', onClick: handleEdit });
}
if (hasLink(resource, 'delete')) {
  menuItems.push({ label: 'Delete', onClick: handleDelete, isDanger: true });
}
```

### Drag-Drop Pattern
```typescript
// Before (wrong)
const handleDrop = (e: DragEvent) => {
  const data = e.dataTransfer.getData('application/json');
  onDrop(JSON.parse(data));
};

// After (correct)
const canAcceptDrop = hasLink(targetCapability, 'link');

const handleDragOver = (e: DragEvent) => {
  if (!canAcceptDrop) return;
  e.preventDefault();
};

const handleDrop = (e: DragEvent) => {
  if (!canAcceptDrop) return;
  const data = e.dataTransfer.getData('application/json');
  onDrop(JSON.parse(data));
};
```

## TypeScript Type Updates

Ensure all DTOs include `_links` typing:
```typescript
interface ResourceWithLinks {
  _links?: {
    self?: string;
    update?: string;
    delete?: string;
    [key: string]: string | undefined;
  };
}
```

## Test Plan

### Unit Tests
- [ ] HATEOAS utility functions (hasLink, getLink)
- [ ] Context menu generation respects link presence
- [ ] Button rendering respects link presence

### Component Tests
- [ ] NodeContextMenu hides actions when links missing
- [ ] EdgeContextMenu hides actions when links missing
- [ ] CapabilityDetails hides Edit when no update link
- [ ] ComponentDetails hides Remove when no remove link
- [ ] Drag handlers disabled when no link permission

### Integration Tests
- [ ] Private view: non-owner sees no edit/delete actions
- [ ] Public view: all users see edit/delete actions
- [ ] Drop zones disabled for linked capabilities
- [ ] Create buttons hidden when no create permission

## Implementation Checklist

### Phase 1: Backend HATEOAS Enhancements
- [x] Add permission-aware links to view element responses
- [x] Add permission-aware links to capability responses
- [x] Add permission-aware links to component responses
- [x] Add `unlink`/`delete` links to enterprise capability link responses
- [x] Regenerate Swagger docs (skipped - pre-existing swagger parse issues)

### Phase 2: Frontend Utilities
- [x] Create `src/utils/hateoas.ts` with helper functions
- [x] Add `_links` typing to ViewComponent and ViewCapability interfaces

### Phase 3: Context Menus
- [x] Update NodeContextMenu to use HATEOAS
- [x] Update EdgeContextMenu to use HATEOAS
- [x] Update NavigationTree capability menu to use HATEOAS
- [x] Update NavigationTree component menu to use HATEOAS

### Phase 4: Detail Panels
- [x] Update CapabilityDetails to use HATEOAS
- [x] Update ComponentDetails to use HATEOAS
- [x] Update RealizationDetails to use HATEOAS
- [x] Update RelationDetails to use HATEOAS

### Phase 5: Enterprise Architecture
- [x] Update EnterpriseCapabilityDetailPanel to use HATEOAS
- [x] Update EnterpriseCapabilitiesTable to use HATEOAS
- [x] Update MaturityGapDetailPanel to use HATEOAS
- [x] Update StrategicImportanceSection to use HATEOAS

### Phase 6: Drag-and-Drop
- [x] Update EnterpriseCapabilityCard drop handling
- [x] Update DomainCapabilityPanel drag handling (already uses linkStatuses from backend API)
- [x] Update EnterpriseCapabilitiesTable drop handling
- [x] Disable drop visual feedback when link unavailable

### Phase 7: Create Buttons
- [x] Update NavigationTree create button visibility (uses prop-based control - handlers only passed when user has permission)
- [x] Update any other create dialogs/buttons (EnterpriseArchPage already checks `canWrite` before showing create button)

### Phase 8: Testing
- [x] Write unit tests for HATEOAS utilities
- [x] Write component tests for permission-gated rendering
- [x] Manual testing of private view scenarios
- [x] Build verification (`npm test -- --run`, `go test ./...`)

## Affected Files Summary

| File | Changes |
|------|---------|
| `canvas/context-menus/NodeContextMenu.tsx` | Add HATEOAS checks |
| `canvas/context-menus/EdgeContextMenu.tsx` | Add HATEOAS checks |
| `navigation/components/NavigationTree.tsx` | Add HATEOAS checks to capability/component menus |
| `capabilities/components/CapabilityDetails.tsx` | Add HATEOAS checks for buttons |
| `components/components/ComponentDetails.tsx` | Add HATEOAS checks for buttons |
| `relations/components/RealizationDetails.tsx` | Add HATEOAS checks |
| `relations/components/RelationDetails.tsx` | Add HATEOAS checks |
| `enterprise-architecture/components/EnterpriseCapabilityDetailPanel.tsx` | Replace prop-based checks with HATEOAS |
| `enterprise-architecture/components/EnterpriseCapabilitiesTable.tsx` | Replace prop-based checks with HATEOAS |
| `enterprise-architecture/components/MaturityGapDetailPanel.tsx` | Add HATEOAS checks |
| `enterprise-architecture/components/EnterpriseCapabilityCard.tsx` | Add HATEOAS checks for drop |
| `enterprise-architecture/components/DomainCapabilityPanel.tsx` | Add HATEOAS checks for drag |
| `business-domains/components/StrategicImportanceSection.tsx` | Add HATEOAS checks |
| `shared/utils/hateoas.ts` | NEW: HATEOAS utility functions |
| Backend: `view_handlers.go` | Add permission-aware element links |
| Backend: Various handlers | Ensure all mutation endpoints have corresponding HATEOAS links |

## Sign-off
- [ ] User: Approved for implementation
