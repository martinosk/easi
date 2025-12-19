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
- [ ] Add React Query dependency
- [ ] Configure QueryClient with appropriate defaults
- [ ] Add QueryClientProvider to app root
- [ ] Set up React Query DevTools for development

### Phase 2: Read Operations Migration (First Domain)
- [ ] Migrate components list fetching to useQuery
- [ ] Migrate component by ID fetching to useQuery
- [ ] Remove corresponding state and actions from componentSlice
- [ ] Verify caching behavior works correctly

### Phase 3: Write Operations Migration (First Domain)
- [ ] Migrate create component to useMutation with cache invalidation
- [ ] Migrate update component to useMutation with optimistic update
- [ ] Migrate delete component to useMutation with cache invalidation
- [ ] Remove corresponding actions from componentSlice

### Phase 4: Remaining Domains
- [ ] Migrate capabilities queries and mutations
- [ ] Migrate relations queries and mutations
- [ ] Migrate views queries and mutations
- [ ] Migrate business domains queries and mutations

### Phase 5: Client State Cleanup
- [ ] Consolidate remaining Zustand state to UI-only concerns
- [ ] Remove empty or near-empty slices
- [ ] Simplify store composition

### Phase 6: Advanced Patterns
- [ ] Implement prefetching for predictable navigation
- [ ] Add background refetching configuration
- [ ] Consider infinite queries for paginated endpoints

## Dependencies
- Spec 082 (API Client Split) provides clean API modules for React Query hooks
- Spec 083 (Test Utilities) should include React Query test utilities

## Incremental Delivery
1. First: Infrastructure setup (no breaking changes)
2. Second: Migrate one domain completely (components) as proof of concept
3. Third: Migrate remaining domains one at a time
4. Fourth: Client state cleanup
5. Fifth: Advanced patterns

## Checklist
- [ ] Specification ready
- [ ] React Query infrastructure configured
- [ ] Components domain migrated
- [ ] Capabilities domain migrated
- [ ] Relations domain migrated
- [ ] Views domain migrated
- [ ] Business domains migrated
- [ ] Client state simplified
- [ ] All tests passing
- [ ] User sign-off
