# Data Fetching Strategy

## Description
Establish a consistent data fetching pattern across the frontend, abstracting API calls from store actions and implementing proper loading/error states.

## Current State Analysis

### Data Fetching Patterns Found

#### Pattern 1: Store Actions Call API Directly
**Location**: Most store slices
```typescript
// componentSlice.ts
createComponent: async (data: ComponentData) => {
  const newComponent = await handleApiCall(
    () => apiClient.createComponent(data),
    'Failed to create component'
  );
  set({ components: [...components, newComponent] });
  toast.success(`Component "${data.name}" created`);
  return newComponent;
}
```
**Issues**:
- Tight coupling between state and API
- Side effects (toasts) mixed with state mutations
- Difficult to test without mocking apiClient

#### Pattern 2: Components Call API Directly
**Location**: NavigationTree.tsx
```typescript
const handleCreateView = async () => {
  await apiClient.createView({ name: createViewName, description: '' });
  await loadViews();
};
```
**Issues**:
- Inconsistent with store-based pattern
- Component handles loading/error states locally
- Duplicated API call patterns

#### Pattern 3: Custom Hooks Call API
**Location**: useViewOperations.ts
```typescript
const addComponentToView = useCallback(async (componentId: string, x: number, y: number) => {
  await apiClient.addComponentToView(currentView.id, { componentId, x, y });
  const updatedView = await apiClient.getViewById(currentView.id);
  useAppStore.setState({ currentView: updatedView });
});
```
**Issues**:
- Mixes direct store mutation with API calls
- No centralized error handling

#### Pattern 4: Effects Trigger Data Loading
**Location**: ComponentCanvas.tsx
```typescript
useEffect(() => {
  componentIdsOnCanvas.forEach((componentId) => {
    loadRealizationsByComponent(componentId);
  });
}, [currentView?.id, currentView?.components.length, loadRealizationsByComponent]);
```
**Issues**:
- N+1 query pattern (one request per component)
- No request deduplication
- No caching

### Current Utilities
- `handleApiCall`: Basic try/catch with toast error
- `optimisticUpdate`: Optimistic state updates with rollback

## Problems to Solve
1. No request caching or deduplication
2. Inconsistent loading state handling
3. No centralized error boundaries
4. Side effects mixed with state mutations
5. Multiple patterns for same operations
6. N+1 query patterns

## Proposed Solution

### Option A: TanStack Query Integration (Recommended)
Add TanStack Query (React Query) for server state management while keeping Zustand for UI state.

**Pros**:
- Built-in caching, deduplication, background refetching
- Automatic loading/error states
- Industry standard pattern
- Great DevTools for debugging

**Cons**:
- Additional dependency
- Learning curve for team
- Requires refactoring existing store patterns

### Option B: Custom Hooks with SWR-like Pattern
Create custom hooks that implement caching and loading states without external dependency.

**Pros**:
- No new dependency
- Full control over implementation
- Can be built incrementally

**Cons**:
- More code to maintain
- Reinventing well-solved problems
- Missing advanced features (background refetch, etc.)

## Requirements (Option A: TanStack Query)

### Phase 1: Setup and Configuration
- [ ] Add @tanstack/react-query dependency
- [ ] Create QueryClientProvider wrapper
- [ ] Configure default query/mutation options
- [ ] Add React Query DevTools (development only)

### Phase 2: Create Query Hooks
- [ ] Create `features/components/api/queries.ts`
  - useComponents(): Query hook for component list
  - useComponent(id): Query hook for single component

- [ ] Create `features/capabilities/api/queries.ts`
  - useCapabilities(): Query hook for capability list
  - useCapability(id): Query hook for single capability
  - useCapabilityRealizations(capabilityId): Query hook for realizations

- [ ] Create `features/views/api/queries.ts`
  - useViews(): Query hook for view list
  - useView(id): Query hook for single view

- [ ] Create `features/relations/api/queries.ts`
  - useRelations(): Query hook for relation list

### Phase 3: Create Mutation Hooks
- [ ] Create mutation hooks for CRUD operations
  - useCreateComponent()
  - useUpdateComponent()
  - useDeleteComponent()
  - (Similar for capabilities, relations, views)

- [ ] Implement optimistic updates where appropriate
- [ ] Configure cache invalidation strategies

### Phase 4: Refactor Store
- [ ] Move server state queries out of Zustand store
- [ ] Keep UI-only state in Zustand (selection, viewport, etc.)
- [ ] Update store slices to use query hooks or remove API logic

### Phase 5: Update Components
- [ ] Replace direct apiClient calls with query hooks
- [ ] Replace store selectors with query hooks where appropriate
- [ ] Add loading/error UI states at feature boundaries

## Example Implementation

### Query Hook Example
```typescript
// features/components/api/queries.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '@/api/client';
import type { Component, CreateComponentRequest } from '@/api/types';

export const componentKeys = {
  all: ['components'] as const,
  lists: () => [...componentKeys.all, 'list'] as const,
  details: () => [...componentKeys.all, 'detail'] as const,
  detail: (id: string) => [...componentKeys.details(), id] as const,
};

export function useComponents() {
  return useQuery({
    queryKey: componentKeys.lists(),
    queryFn: () => apiClient.getComponents(),
  });
}

export function useComponent(id: string) {
  return useQuery({
    queryKey: componentKeys.detail(id),
    queryFn: () => apiClient.getComponentById(id),
    enabled: !!id,
  });
}

export function useCreateComponent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateComponentRequest) => apiClient.createComponent(data),
    onSuccess: (newComponent) => {
      queryClient.invalidateQueries({ queryKey: componentKeys.lists() });
      toast.success(`Component "${newComponent.name}" created`);
    },
    onError: (error: ApiError) => {
      toast.error(error.message || 'Failed to create component');
    },
  });
}
```

### Zustand Store After Refactor
```typescript
// store/slices/selectionSlice.ts - UI state only
export interface SelectionState {
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
}

export interface SelectionActions {
  selectNode: (id: string | null) => void;
  selectEdge: (id: string | null) => void;
  clearSelection: () => void;
}

// No API calls in store - purely UI state
```

### Component Usage
```typescript
function ComponentList() {
  const { data: components, isLoading, error } = useComponents();
  const createMutation = useCreateComponent();

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorMessage error={error} />;

  return (
    <ul>
      {components?.map(c => <ComponentItem key={c.id} component={c} />)}
    </ul>
  );
}
```

## Migration Strategy
1. Add TanStack Query alongside existing implementation
2. Create query hooks for one feature (start with components)
3. Update components to use new hooks
4. Remove corresponding store logic
5. Repeat for other features
6. Clean up deprecated store code

## Checklist
- [ ] Specification ready
- [ ] TanStack Query installed and configured
- [ ] Component query hooks created
- [ ] Capability query hooks created
- [ ] View query hooks created
- [ ] Relation query hooks created
- [ ] Mutation hooks with cache invalidation
- [ ] Store refactored to UI-only state
- [ ] Components updated to use query hooks
- [ ] Loading/error states implemented
- [ ] Tests updated
- [ ] Documentation updated
- [ ] User sign-off
