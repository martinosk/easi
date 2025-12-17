# 077: Application Details in Business Domains View

## Description
Enable viewing and editing application details when clicking on application chips in the business domains view. The feature reuses the existing `ComponentDetails` component from the architecture canvas for consistency. Additionally, the details pane will be visible by default with a placeholder message to avoid UI resizing.

## User Story
As a user viewing business domains, I want to click on an application chip and see the same detailed information about that application that I see in the architecture canvas, including the ability to edit it.

## Requirements

### Functional Requirements
1. Clicking an application chip in the capability grid shows application details in the right-side details pane
2. The details pane shows the same information as the architecture canvas:
   - Application name, description, created date, ID
   - Type badge ("Application Component")
   - Realized capabilities (direct and inherited)
   - Reference documentation link
   - Edit button
3. Edit button opens the same `EditComponentDialog` used in the architecture canvas
4. Details pane is visible by default with placeholder: "Select a capability or application to view details"
5. Close button (X) clears the selection and returns to placeholder state

### Non-Functional Requirements
1. Reuse existing `ComponentDetails` presentation logic to avoid code duplication
2. Standardize data fetching to maximize cache hits across views
3. Use existing CSS classes (`.detail-panel`, `.detail-header`, `.detail-content`) for styling consistency

### Out of Scope
- Color picker (only applicable in architecture views with custom color scheme)
- "Remove from View" button (applications are not placed in business domain views)

## Technical Design

### Architecture Overview
```
BusinessDomainsPage
├── VisualizationArea (capability grid with application chips)
├── CapabilityDetailSidebar (existing - for capability selection)
└── DetailsSidebar (NEW - unified details pane)
    ├── EmptyState (when nothing selected)
    ├── CapabilityDetailsContent (when capability selected)
    └── ApplicationDetailsContent (when application selected)
        └── EditComponentDialog (when edit clicked)
```

### Component Changes

#### 1. Refactor `ComponentDetails.tsx`
Extract `ComponentDetailsContent` as a separate exportable component that accepts data via props:

```typescript
export interface ComponentDetailsContentProps {
  component: Component;
  realizations: CapabilityRealization[];
  capabilities: Capability[];
  onEdit: (componentId: string) => void;
  onClose: () => void;
}

export const ComponentDetailsContent: React.FC<ComponentDetailsContentProps>
```

The existing `ComponentDetails` becomes a wrapper that fetches from Zustand store and renders `ComponentDetailsContent`.

#### 2. Create `DetailsSidebar.tsx` (NEW)
Unified details sidebar for business domains that:
- Shows empty state placeholder when nothing selected
- Shows capability details when capability selected
- Shows application details when application selected
- Manages edit dialog state locally

```typescript
interface DetailsSidebarProps {
  selectedCapability: Capability | null;
  selectedComponentId: ComponentId | null;
  onCloseCapability: () => void;
  onCloseApplication: () => void;
}
```

#### 3. Update `BusinessDomainsPage.tsx`
- Replace separate `CapabilityDetailSidebar` and `ApplicationDetailSidebar` with unified `DetailsSidebar`
- Remove conditional rendering (sidebar handles empty state internally)

#### 4. Delete Obsolete Files
- `ApplicationDetailSidebar.tsx` - replaced by DetailsSidebar
- Keep `CapabilityDetailSidebar.tsx` content but merge into DetailsSidebar

### Data Flow

```
User clicks application chip
  → handleApplicationClick(componentId)
  → setSelectedComponentId(componentId), setSelectedCapability(null)
  → DetailsSidebar receives selectedComponentId
  → Fetches component from store (or API fallback)
  → Fetches realizations from store
  → Renders ComponentDetailsContent

User clicks Edit
  → Opens EditComponentDialog
  → On save: updateComponent via store
  → UI updates with new data

User clicks Close
  → clearSelectedComponent()
  → Returns to empty state placeholder
```

### Data Standardization
Per project requirements, data fetching should be standardized:
1. Components: Read from global Zustand store (`useAppStore.components`)
2. Capabilities: Read from global Zustand store (`useAppStore.capabilities`)
3. Realizations: Read from global Zustand store (`useAppStore.capabilityRealizations`)
4. Fallback: If component not in store, fetch via API and optionally add to store

## Checklist
- [x] Specification ready
- [x] Refactor ComponentDetails to extract ComponentDetailsContent
- [x] Create DetailsSidebar component with empty state
- [x] Integrate EditComponentDialog for application editing
- [x] Update BusinessDomainsPage to use DetailsSidebar
- [x] Delete obsolete ApplicationDetailSidebar
- [x] Merge CapabilityDetailSidebar into DetailsSidebar
- [x] Build passes with no TypeScript errors
- [x] All existing tests pass (548 tests)
- [x] Manual testing of application click → details → edit flow
- [x] User sign-off
