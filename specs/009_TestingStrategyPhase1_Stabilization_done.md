# Phase 1: Frontend Test Suite Stabilization

## Description
Immediately fix the critical memory leak issue causing frontend tests to crash with "JavaScript heap out of memory" errors. This phase removes the problematic React Flow component tests that leak memory in JSDOM and adds test isolation to prevent future memory issues.

## Problem Statement
Current frontend tests fail after ~60-70 seconds with heap out of memory errors. Root cause: React Flow components (`ComponentCanvas.test.tsx` and `ComponentCanvas.drag.test.tsx`) create canvas contexts and DOM subscriptions that JSDOM cannot properly garbage collect between test runs.

## Requirements

### 1. Remove Memory-Leaking Tests
- Delete `src/components/ComponentCanvas.test.tsx` (190 lines testing React Flow rendering)
- Delete `src/components/ComponentCanvas.drag.test.tsx` (131 lines testing drag interactions)
- Justification: These tests heavily mock React Flow internals, providing minimal value while causing crashes. Canvas interactions will be covered by Playwright E2E tests in Phase 3.

### 2. Update Vitest Configuration
- Add process isolation to `vitest.config.ts` to prevent memory leaks from spreading between test files
- Configure fork pool with `isolate: true`
- Add explicit exclusion for ComponentCanvas test files (safety measure)
- DO NOT increase Node.js heap size - fix the root cause instead

**Updated vitest.config.ts:**
```typescript
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',

    // Process isolation to prevent memory leaks
    pool: 'forks',
    poolOptions: {
      forks: {
        singleFork: false,  // Use multiple forks for parallelization
        isolate: true,      // Each test file runs in isolated process
      },
    },

    // Timeouts
    testTimeout: 10000,
    hookTimeout: 10000,

    // Exclude problematic and non-unit test files
    exclude: [
      '**/node_modules/**',
      '**/dist/**',
      '**/e2e/**',
      '**/ComponentCanvas*.test.tsx',  // Moved to E2E in Phase 3
    ],
  },
});
```

### 3. Verify Test Stability
- Run full test suite multiple times consecutively
- Verify no memory errors occur
- Verify all remaining tests pass
- Monitor test execution time (should remain under 30 seconds)

### 4. Update Documentation
- Add comment in vitest.config.ts explaining why ComponentCanvas tests are excluded
- Add TODO comment referencing Phase 3 spec for E2E replacement

## Expected Outcomes
- ✅ Tests run without crashes or memory errors
- ✅ CI pipeline becomes stable and green
- ✅ Test execution completes successfully every time
- ✅ No increase in test execution time despite isolation (parallelization compensates)
- ✅ Remaining 5 test files (dialogs, store, API) continue to pass

## Test Coverage Impact
**Before:**
- 7 test files, 44 test cases
- ComponentCanvas.test.tsx: 7 test cases
- ComponentCanvas.drag.test.tsx: 4 test cases

**After:**
- 5 test files, 33 test cases
- Lost: 11 test cases for canvas rendering/interaction
- Note: These will be replaced with superior Playwright E2E tests in Phase 3

## Files to Modify
- `frontend/src/components/ComponentCanvas.test.tsx` - DELETE
- `frontend/src/components/ComponentCanvas.drag.test.tsx` - DELETE
- `frontend/vitest.config.ts` - UPDATE (add isolation config)

## Dependencies
- None (this is the foundational phase)

## Follow-up Work
- Phase 2: Refactor existing tests (spec 010)
- Phase 3: Add Playwright E2E tests to cover canvas interactions (spec 011)

## Checklist
- [x] Specification ready
- [x] Delete ComponentCanvas.test.tsx
- [x] Delete ComponentCanvas.drag.test.tsx
- [x] Update vitest.config.ts with process isolation
- [x] Add exclusion for ComponentCanvas tests
- [x] Add explanatory comments in config
- [x] Run test suite 3+ times to verify stability
- [x] Verify no memory errors occur
- [ ] Verify all remaining tests pass (6 pre-existing failures unrelated to this phase)
- [x] Verify test execution time is acceptable (~17s)
- [ ] Update any CI configuration if needed
- [ ] Documentation updated
- [x] User sign-off
