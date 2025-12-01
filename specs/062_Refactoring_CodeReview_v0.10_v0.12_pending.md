# Spec 062: Code Review Refactoring (v0.10.0 - v0.12.0)

## Overview
Findings from code review of changes between release v0.10.0 and v0.12.0. This spec documents code smells, structural issues, and proposed refactorings.

## Code Health Summary
| File | Before | After | Status |
|------|--------|-------|--------|
| import_session.go | 8.4 | 10.0 | ✅ Healthy |
| import_orchestrator.go | 8.54 | 10.0 | ✅ Healthy |
| layout_handlers.go | 8.24 | 8.82 | Improved |
| layout_container_repository.go | 8.01 | 9.38 | Improved |
| BusinessDomainsPage.tsx | 7.94 | 10.0 | ✅ Healthy |
| DomainVisualizationPage.tsx | 8.61 | 9.06 | Improved |
| useLayout.ts | 9.2 | 9.2 | Maintained |
| useImportSession.ts | 9.68 | 10.0 | ✅ Healthy |
| useCanvasNodes.ts | 9.68 | 10.0 | ✅ Healthy |
| client.ts | 10.0 | 10.0 | ✅ Healthy |

---

## Backend Refactorings

### BE-1: Extract Deserialization Helpers in ImportSession Aggregate
- **Location**: `backend/internal/importing/domain/aggregates/import_session.go:321-453`
- **Issue**: `deserializePreview` (cc=19) and `deserializeParsedData` (cc=19) have excessive cyclomatic complexity and bumpy road patterns with deep nesting
- **Fix**: Extract helper methods for each collection type (capabilities, components, relationships) and separate type conversion functions

### BE-2: Primitive Obsession in Import Domain Types
- **Location**: `backend/internal/importing/domain/aggregates/import_session.go:19-48`
- **Issue**: `ParsedElement`, `ParsedRelationship`, and `ImportResult` structs expose string/int primitives directly, violating DDD value object rules
- **Fix**: Convert to proper value objects with encapsulated validation

### BE-3: Extract Relationship Processing Strategy
- **Location**: `backend/internal/importing/application/orchestrator/import_orchestrator.go:156-314`
- **Issue**: `createRealizations` and `createComponentRelations` have identical structure (duplicated code smell) with bumps=2 each
- **Fix**: Extract generic relationship processor with strategy pattern for different relationship types

### BE-4: Reduce assignToDomain Parameter Count
- **Location**: `backend/internal/importing/application/orchestrator/import_orchestrator.go:219-245`
- **Issue**: Function has 5 parameters, exceeding the Go threshold of 4
- **Fix**: Create `DomainAssignmentContext` value object to encapsulate parameters

### BE-5: Simplify buildHierarchyLevels Complexity
- **Location**: `backend/internal/importing/application/orchestrator/import_orchestrator.go:326-362`
- **Issue**: Cyclomatic complexity of 11 with 2 bumps for child-finding logic
- **Fix**: Extract child-finding into separate method and simplify loop structure

### BE-6: Extract Repository Reconstitution Logic
- **Location**: `backend/internal/viewlayouts/infrastructure/repositories/layout_container_repository.go:29-165`
- **Issue**: `GetByContext` and `GetByID` have nearly identical structure (code duplication)
- **Fix**: Extract shared `reconstituteContainer` helper method

### BE-7: Simplify loadElements Method
- **Location**: `backend/internal/viewlayouts/infrastructure/repositories/layout_container_repository.go:167-227`
- **Issue**: cc=11 with bumpy road (2 bumps), null checking and value object construction deeply nested
- **Fix**: Extract `scanElementRow` helper to reduce nesting

### BE-8: Unify Element Position Persistence
- **Location**: `backend/internal/viewlayouts/infrastructure/repositories/layout_container_repository.go:290-418`
- **Issue**: `UpsertElementPosition` and `BatchUpdatePositions` contain identical null handling and SQL parameter building
- **Fix**: Extract `buildElementPositionParams` shared method

### BE-9: Split UpsertLayout Handler Logic
- **Location**: `backend/internal/viewlayouts/infrastructure/api/layout_handlers.go:204-261`
- **Issue**: cc=12, handles both create and update logic with multiple error paths
- **Fix**: Extract separate `createLayout` and `updateLayout` methods

### BE-10: Simplify BatchUpdateElements Handler
- **Location**: `backend/internal/viewlayouts/infrastructure/api/layout_handlers.go:448-522`
- **Issue**: cc=12 with 3 bumps, long method mixing validation, transformation, and response
- **Fix**: Extract `validateBatchItems`, `transformToPositions`, `buildBatchResponse` helpers

### BE-11: Remove Temporal Tracking from Aggregate
- **Location**: `backend/internal/viewlayouts/domain/aggregates/layout_container.go:104-122`
- **Issue**: Aggregate directly modifies `updatedAt` with `time.Now().UTC()`, mixing infrastructure concern with domain
- **Fix**: Move temporal tracking to repository layer

---

## Frontend Refactorings

### FE-1: Split BusinessDomainsPage Component
- **Location**: `frontend/src/features/business-domains/pages/BusinessDomainsPage.tsx:29-456`
- **Issue**: 381 LoC, cc=57 with 3 bumps - massive component doing too much
- **Fix**: Extract into separate components: DomainList, CapabilityManager, DragDropHandler, and domain-specific sub-pages

### FE-2: Extract Drag Handler Functions
- **Location**: `frontend/src/features/business-domains/pages/BusinessDomainsPage.tsx:160-205`
- **Issue**: Complex `handleDragEnd` mixes sorting, reassignment, and association logic
- **Fix**: Extract `handleSortDrag`, `handleReassignDrag`, `handleAssociateDrag` as separate functions

### FE-3: Create Shared Drag Handling Hook
- **Location**: Both `BusinessDomainsPage.tsx` and `DomainVisualizationPage.tsx`
- **Issue**: Similar drag handling logic duplicated across two pages
- **Fix**: Extract `useDomainDragDrop` custom hook for shared logic

### FE-4: Fix Stale Closure in useLayout
- **Location**: `frontend/src/hooks/useLayout.ts:80-142`
- **Issue**: `updateElementPosition` and `batchUpdatePositions` capture `positions` in closure for rollback, causing race conditions with rapid updates
- **Fix**: Use functional state updates `setPositions(prev => ...)` and store pre-update value in function scope

### FE-5: Move Polling Logic Outside useImportSession
- **Location**: `frontend/src/features/importing/hooks/useImportSession.ts:25-65`
- **Issue**: Axios instance created conditionally inside hook, recursive setTimeout creates memory leak risk on unmount
- **Fix**: Move axios to module scope, use proper interval with useEffect cleanup or polling library

### FE-6: Remove Duplicate Grid Rendering
- **Location**: `frontend/src/features/business-domains/components/DomainGrid.tsx:137-173` and `NestedCapabilityGrid.tsx:201-237`
- **Issue**: Entire grid structure duplicated with only sortable context difference
- **Fix**: Use conditional wrapping with SortableContext instead of duplicating grid structure

### FE-7: Extract Node Factory Functions
- **Location**: `frontend/src/features/canvas/hooks/useCanvasNodes.ts:10-58`
- **Issue**: Pure functions `createComponentNode` and `createCapabilityNode` defined inside hook file reduce testability
- **Fix**: Move to separate `/utils/nodeFactory.ts` for better separation of concerns

### FE-8: Add User Feedback for Errors
- **Location**: `BusinessDomainsPage.tsx:200-202`, `DomainVisualizationPage.tsx:123,139`
- **Issue**: Console.error without user-facing feedback for failed operations
- **Fix**: Replace console.error with toast.error notifications

### FE-9: Extract Imperative Dialog to Controlled
- **Location**: `frontend/src/features/importing/components/ImportDialog.tsx:35-44`
- **Issue**: Direct DOM manipulation with `showModal()`/`close()` bypasses React declarative model
- **Fix**: Use controlled dialog with `open` attribute and React state

### FE-10: Fix useCanvasNodes Effect Dependencies
- **Location**: `frontend/src/features/canvas/hooks/useCanvasNodes.ts:72-77`
- **Issue**: Effect depends on `currentView?.components.length` which changes object identity every render
- **Fix**: Use stable dependency like `currentView?.id` and memoize `loadRealizationsByComponent`

---

## Priority Matrix

| Priority | Refactoring | Impact | Effort |
|----------|-------------|--------|--------|
| High | FE-1 | High maintainability improvement | Medium |
| High | FE-4 | Fixes potential race condition bug | Low |
| High | BE-1 | Reduces complexity significantly | Medium |
| Medium | BE-3 | Removes duplication | Medium |
| Medium | BE-6 | Removes duplication | Low |
| Medium | FE-3 | Reduces code duplication | Medium |
| Medium | FE-5 | Fixes memory leak risk | Low |
| Low | BE-2 | DDD compliance | High |
| Low | BE-11 | Architectural purity | Low |
| Low | FE-7 | Testability improvement | Low |

---

## Implementation Checklist

### Backend
- [x] BE-1: Extract deserialization helpers (8.4 → 10.0)
- [ ] BE-2: Add value objects for parsed types (deferred - high effort, low priority)
- [x] BE-3: Create relationship processing strategy (8.54 → 10.0)
- [x] BE-4: Create DomainAssignmentContext VO (included in BE-3)
- [x] BE-5: Simplify buildHierarchyLevels (included in BE-3)
- [x] BE-6: Extract repository reconstitution (8.01 → 9.38)
- [x] BE-7: Simplify loadElements (included in BE-6)
- [x] BE-8: Unify element position persistence (included in BE-6)
- [x] BE-9: Split UpsertLayout handler (8.24 → 8.82)
- [x] BE-10: Simplify BatchUpdateElements (included in BE-9)
- [x] BE-11: Move temporal tracking to repository (9.68 maintained)

### Frontend
- [x] FE-1: Split BusinessDomainsPage (7.94 → 10.0)
- [x] FE-2: Extract drag handler functions (included in FE-1)
- [x] FE-3: Create shared drag handling hook (DomainVisualizationPage 8.61 → 9.06)
- [x] FE-4: Fix stale closure in useLayout (9.2 maintained)
- [x] FE-5: Fix polling in useImportSession (9.68 → 10.0)
- [x] FE-6: Remove duplicate grid rendering
- [x] FE-7: Extract node factory functions (useCanvasNodes 9.68 → 10.0)
- [x] FE-8: Add user feedback for errors
- [x] FE-9: Convert to controlled dialog
- [x] FE-10: Fix useCanvasNodes dependencies
