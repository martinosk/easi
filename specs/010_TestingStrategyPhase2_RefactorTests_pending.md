# Phase 2: Refactor and Consolidate Existing Tests

## Description
Refactor existing frontend tests to improve organization, reduce duplication, and align with testing best practices. Consolidate misnamed "e2e" tests into proper store unit tests, create reusable test helpers, and simplify dialog component tests.

## Problem Statement
- `e2e.test.tsx` and `error-handling.test.tsx` aren't actually E2E tests - they're store integration tests
- Duplicate mock setup across multiple test files
- Improper async handling with store state assertions
- Tests scattered without coherent structure

## Requirements

### 1. Create Test Helper Utilities
Create `src/test/helpers/` directory with three helper files:
- **`testStore.ts`** - Factory function for test store instances with mocked methods and default state
- **`mockApiClient.ts`** - Factory for mocked API client with all methods as vi.fn()
- **`dialogTestUtils.ts`** - Shared dialog test setup (mock useAppStore with common methods)

### 2. Consolidate Store Tests
- **Move** `src/test/e2e.test.tsx` → `src/store/appStore.test.ts`
- **Merge** `src/test/error-handling.test.tsx` into `src/store/appStore.test.ts`
- **Delete** both original files
- **Organize** tests by feature area: Component Management, Relation Management, Selection State, Error Handling
- **Fix** async patterns: use `store.getState().method()` then `expect(store.getState().property)` (not immediate checks)

### 3. Simplify Dialog Tests
- Extract common mock setup into `dialogTestUtils.ts`
- Update `CreateComponentDialog.test.tsx` to use shared helper
- Update `CreateRelationDialog.test.tsx` to use shared helper
- Reduce boilerplate while maintaining same coverage

### 4. Test Quality Standards
- Follow Arrange-Act-Assert pattern
- Test behavior, not implementation
- Only mock API boundary (not domain logic)
- Clear, descriptive test names
- Deterministic tests (no random data, no timing dependencies)

## Expected Outcomes
- ✅ Clear test organization aligned with codebase structure
- ✅ Store tests live with store code
- ✅ Reusable helpers eliminate duplication
- ✅ All async assertions properly handled
- ✅ No loss of coverage (33 test cases maintained)

## Files to Create
- `frontend/src/test/helpers/testStore.ts`
- `frontend/src/test/helpers/mockApiClient.ts`
- `frontend/src/test/helpers/dialogTestUtils.ts`
- `frontend/src/store/appStore.test.ts`

## Files to Modify
- `frontend/src/components/CreateComponentDialog.test.tsx`
- `frontend/src/components/CreateRelationDialog.test.tsx`

## Files to Delete
- `frontend/src/test/e2e.test.tsx`
- `frontend/src/test/error-handling.test.tsx`

## Dependencies
- Must complete Phase 1 (spec 009) first

## Checklist
- [ ] Specification ready
- [ ] Create test helper utilities (testStore, mockApiClient, dialogTestUtils)
- [ ] Create consolidated appStore.test.ts
- [ ] Migrate and organize tests by feature area
- [ ] Fix async state assertion patterns
- [ ] Delete original e2e.test.tsx and error-handling.test.tsx
- [ ] Update dialog tests to use shared helpers
- [ ] Run full test suite to verify all tests pass
- [ ] Verify no test coverage loss (33 test cases)
- [ ] Verify test execution time acceptable
- [ ] User sign-off
