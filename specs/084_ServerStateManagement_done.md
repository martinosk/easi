# Server State Management with React Query

## Description
Introduce React Query (TanStack Query) to manage server state separately from client state. This simplifies data fetching, provides automatic caching, and reduces complexity in Zustand store slices.

## Rationale
- Current Zustand slices mix server data (fetched from API) with client state (UI state, selections)
- Manual loading/error state management in each slice
- No automatic cache invalidation or background refetching
- Optimistic updates require manual rollback logic
- Duplicate data fetching when components remount

## Target State
- React Query handles all server state (components, capabilities, relations, etc.)
- Zustand handles only client state (selections, UI state, viewport)
- Automatic caching reduces API calls
- Built-in loading and error states
- Simplified optimistic updates with automatic rollback

## Requirements

### Phase 1: Infrastructure Setup
- [x] Add React Query dependency (@tanstack/react-query)
- [x] Configure QueryClient with appropriate defaults (src/lib/queryClient.ts)
- [x] Add QueryClientProvider to app root (main.tsx)
- [x] Set up React Query DevTools for development

### Phase 2: Read Operations Migration (First Domain)
- [x] Migrate components list fetching to useQuery (useComponents)
- [x] Migrate component by ID fetching to useQuery (useComponent)
- [x] React Query hooks available alongside existing store
- [x] Caching configured with 5 min stale time

### Phase 3: Write Operations Migration (First Domain)
- [x] Migrate create component to useMutation with cache invalidation (useCreateComponent)
- [x] Migrate update component to useMutation with cache update (useUpdateComponent)
- [x] Migrate delete component to useMutation with cache invalidation (useDeleteComponent)
- [x] Toast notifications for success/error feedback

### Phase 4: Remaining Domains
- [x] Migrate capabilities queries and mutations (useCapabilities hooks)
- [x] Migrate relations queries and mutations (useRelations hooks)
- [x] Migrate views queries and mutations (useViews hooks)
- [x] Migrate business domains queries and mutations (useBusinessDomains hooks)
- [x] Migrate layouts queries and mutations (useLayouts hooks)
- [x] Migrate metadata queries (useMetadata hooks)

### Phase 5: Client State Cleanup (REOPENED)

#### 5.1 Migrate Component Reads to React Query
Replace all `useAppStore((state) => state.components)` with `useComponents()`:
- [x] `useCanvasNodes.ts` - use `useComponents()` instead of Zustand
- [x] `useContextMenu.ts` - use `useComponents()` instead of Zustand
- [x] `NavigationTree.tsx` - use `useComponents()` instead of Zustand
- [x] `ComponentDetails.tsx` - use `useComponents()` instead of Zustand
- [x] `CapabilityDetails.tsx` - use `useComponents()` instead of Zustand
- [x] `CreateRelationDialog.tsx` - use `useComponents()` instead of Zustand
- [x] `EditRealizationDialog.tsx` - use `useComponents()` instead of Zustand
- [x] `RealizationDetails.tsx` - use `useComponents()` instead of Zustand
- [x] `RelationDetails.tsx` - use `useComponents()` instead of Zustand
- [x] `DetailsSidebar.tsx` - use `useComponents()` instead of Zustand

#### 5.2 Migrate Relation Reads to React Query
Create and use `useRelations()` hook:
- [x] Create `useRelations()` hook in `features/relations/hooks/useRelations.ts`
- [x] `useCanvasEdges.ts` - use `useRelations()` instead of Zustand
- [x] `useContextMenu.ts` - use `useRelations()` instead of Zustand
- [x] `RelationDetails.tsx` - use `useRelations()` instead of Zustand
- [x] `App.tsx` - use `useRelations()` instead of Zustand

#### 5.3 Migrate Capability Realizations to React Query
- [x] Create `useRealizationsForComponents()` hook to fetch realizations for multiple components
- [x] `useCanvasEdges.ts` - use React Query instead of `useAppStore((state) => state.capabilityRealizations)`
- [x] `useContextMenu.ts` - use React Query instead of Zustand
- [x] `ComponentDetails.tsx` - use `useCapabilitiesByComponent()` instead of Zustand
- [x] `CapabilityDetails.tsx` - use `useCapabilityRealizations()` instead of Zustand
- [x] `DetailsSidebar.tsx` - use `useCapabilitiesByComponent()` instead of Zustand
- [x] `RealizationDetails.tsx` - use `useRealizationsForComponents()` instead of Zustand

#### 5.4 Remove Server State from Zustand Slices
- [x] Remove `components: Component[]` from `componentSlice.ts`
- [x] Remove component CRUD actions that duplicate React Query mutations
- [x] Remove `capabilities: Capability[]` from `capabilitySlice.ts`
- [x] Remove `loadCapabilities()` and capability CRUD actions that duplicate React Query
- [x] Remove `capabilityDependencies` from `capabilitySlice.ts`
- [x] Remove `capabilityRealizations` from `capabilitySlice.ts`
- [x] Remove `relations: Relation[]` from `relationSlice.ts`
- [x] Remove relation CRUD actions that duplicate React Query mutations

#### 5.5 Simplify Store Composition
- [x] Remove `componentSlice.ts` if empty after cleanup
- [x] Remove `capabilitySlice.ts` if empty after cleanup
- [x] Remove `relationSlice.ts` if empty after cleanup
- [x] Update `appStore.ts` to remove deleted slices
- [x] Verify remaining Zustand state is UI-only (selections, viewport, layout)

## Dependencies
- Spec 082 (API Client Split) provides clean API modules for React Query hooks
- Spec 083 (Test Utilities) should include React Query test utilities

## Incremental Delivery
1. First: Infrastructure setup (no breaking changes) ✅
2. Second: Migrate one domain completely (components) as proof of concept ✅
3. Third: Migrate remaining domains one at a time ✅
4. Fourth: Client state cleanup ✅

## Implementation Notes

### Pattern for Migration
When migrating a component from Zustand to React Query:

```tsx
// Before (Zustand - server state in client store)
const components = useAppStore((state) => state.components);

// After (React Query - server state managed by React Query)
const { data: components = [] } = useComponents();
```

### Handling Loading States
Components should handle loading states from React Query:
```tsx
const { data: components = [], isLoading } = useComponents();
if (isLoading) return <Spinner />;
```

### What Stays in Zustand
Only UI/client state should remain:
- `selectedNodeId`, `selectedEdgeId`, `selectedCapabilityId` (selection state)
- `currentView` (active view selection)
- `viewport` (canvas pan/zoom state)
- Layout preferences

## Checklist
- [x] Specification ready
- [x] React Query infrastructure configured
- [x] Components domain migrated (hooks created)
- [x] Capabilities domain migrated (hooks created)
- [x] Relations domain migrated (hooks created)
- [x] Views domain migrated
- [x] Business domains migrated
- [x] Components reads migrated to React Query (Phase 5.1)
- [x] Relations reads migrated to React Query (Phase 5.2)
- [x] Capability realizations migrated to React Query (Phase 5.3)
- [x] Server state removed from Zustand slices (Phase 5.4)
- [x] Store composition simplified (Phase 5.5)
- [x] Test mocks updated for dialog tests (React Query mutation assertions)
- [x] All tests passing (519 tests)
- [x] User sign-off
