# Frontend Architecture Refactoring

## Description
Comprehensive refactoring initiative to address architectural concerns, improve maintainability, and establish scalable patterns for the frontend application.

## Current State Analysis

### Project Overview
- **Framework**: React 19.1 with TypeScript 5.9
- **Build Tool**: Vite 7.1
- **State Management**: Zustand 5.0 with slice pattern
- **Visualization**: React Flow (@xyflow/react)
- **HTTP Client**: Axios
- **Testing**: Vitest (unit), Playwright (E2E)

### Current Directory Structure
```
frontend/src/
├── api/                    # API client and types
├── assets/                 # Static assets
├── components/             # All UI components (flat structure)
├── contexts/               # Feature context (releases only)
├── hooks/                  # Custom hooks
├── store/                  # Zustand store with slices
│   ├── slices/            # State slices
│   ├── types/             # Store types
│   └── utils/             # Store utilities
├── test/                   # Test setup and helpers
└── utils/                  # Utility functions
```

### Metrics
- **Component Files**: 51 files
- **Total Component Lines**: ~11,645 lines
- **Store Slices**: 8 slices in single store
- **Custom Hooks**: 6 hooks

## Identified Problems

### 1. Monolithic Component Directory (HIGH PRIORITY)
**Problem**: All 51 components live in a flat `/components` directory with no organization by feature or responsibility.

**Impact**:
- Difficult to understand component relationships
- Hard to locate related components
- No clear feature boundaries
- Poor code discoverability

**Evidence**:
- ComponentCanvas.tsx: 868 lines handling rendering, selection, context menus, drag-drop, and confirmations
- NavigationTree.tsx: 780 lines handling tree state, context menus, dialogs, and multiple entity types
- Mixing of generic UI components (ColorPicker, ConfirmationDialog) with domain-specific components

### 2. Oversized Components with Mixed Concerns (HIGH PRIORITY)
**Problem**: Several components exceed reasonable size limits and handle multiple responsibilities.

**Examples**:
- `ComponentCanvas.tsx` (868 lines): Handles node rendering, edge rendering, selection, context menus, drag-drop, viewport state, delete confirmations
- `NavigationTree.tsx` (780 lines): Handles tree rendering for multiple entity types (components, views, capabilities), context menus, inline editing, dialogs
- `App.tsx` (178 lines): Manages all dialog state, coordinates multiple features

**Anti-patterns**:
- Functions defined inside components that could be extracted
- Multiple useState hooks managing complex related state
- Business logic mixed with presentation

### 3. God Store Pattern (MEDIUM PRIORITY)
**Problem**: Single Zustand store combining 8 slices creates a large unified state object.

**Current Slices**:
- componentSlice, relationSlice, viewSlice
- capabilitySlice, canvasCapabilitySlice
- selectionSlice, viewportSlice, layoutSlice

**Issues**:
- Cross-slice dependencies create implicit coupling
- Some slices directly access state from other slices (e.g., componentSlice needs currentView, relations)
- No clear domain boundaries in state management
- Testing individual slices requires mocking entire store

### 4. Tight Coupling Between API and Store (MEDIUM PRIORITY)
**Problem**: Store slices directly import and call apiClient, making them difficult to test and creating tight coupling.

**Evidence**:
- Every slice imports `apiClient` directly
- `toast` notifications are called from within store actions
- No abstraction layer between data fetching and state management

### 5. Inconsistent Data Fetching Patterns (MEDIUM PRIORITY)
**Problem**: Data fetching happens in multiple ways without a consistent strategy.

**Patterns Found**:
- Store actions calling API directly (most slices)
- Components calling API directly (NavigationTree.tsx)
- Custom hooks calling API (useViewOperations.ts)
- Effects triggering data loads (ComponentCanvas useEffect for realizations)

### 6. Missing Feature Module Structure (MEDIUM PRIORITY)
**Problem**: Only one feature has proper module structure (`contexts/releases/`). Other features are scattered.

**Evidence**:
- `contexts/releases/` has proper structure: api/, components/, store/
- Capability-related code spread across: components/, store/slices/, api/
- No clear bounded contexts despite backend using DDD

### 7. Poor Type Reuse (LOW PRIORITY)
**Problem**: Types are defined in multiple locations without clear organization.

**Locations**:
- `api/types.ts`: API request/response types (305 lines)
- `store/types/storeTypes.ts`: Store-related types
- Component files: Inline interface definitions

### 8. Dialog State Management in App.tsx (LOW PRIORITY)
**Problem**: App.tsx manages state for 6+ dialogs, creating a large coordinating component.

**Evidence**:
- Multiple useDialogState hooks
- Multiple useState for edit targets
- Complex callback chains passed through component tree

## Requirements

### Phase 1: Component Organization (Recommended Start)
- [ ] Create feature-based directory structure
- [ ] Extract shared/generic UI components to `components/shared/`
- [ ] Group domain components by feature (capability, component, relation, view)
- [ ] Establish component composition patterns
- [ ] Document component organization conventions

### Phase 2: Component Decomposition
- [ ] Break down ComponentCanvas.tsx into smaller, focused components
- [ ] Break down NavigationTree.tsx into composable tree components
- [ ] Extract context menu logic to reusable hook or component
- [ ] Extract inline dialog management from App.tsx

### Phase 3: State Management Refinement
- [ ] Consider feature-based stores or maintain single store with clearer boundaries
- [ ] Abstract API calls behind repository pattern or hooks
- [ ] Separate side effects (toast notifications) from state actions
- [ ] Add proper error boundaries

### Phase 4: Data Fetching Strategy
- [ ] Establish consistent data fetching pattern (recommend TanStack Query or custom hooks)
- [ ] Implement loading/error states at feature boundaries
- [ ] Add request deduplication and caching

## Checklist
- [ ] Specification ready
- [ ] Phase 1 implementation done
- [ ] Phase 2 implementation done
- [ ] Phase 3 implementation done
- [ ] Phase 4 implementation done
- [ ] Unit tests updated
- [ ] Documentation updated
- [ ] User sign-off

## Notes
- This is a large-scale refactoring effort that should be broken into smaller, incremental changes
- Prioritize changes that provide immediate value while not breaking existing functionality
- Consider creating sub-specs for each phase
