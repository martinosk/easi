# Phase 4: Expand Test Coverage

## Description
Add tests for critical gaps identified in the current test suite: API client error handling, existing utility functions, and accessibility testing. 

## Problem Statement
After Phases 1-3, specific gaps remain:
- API client has minimal error scenario coverage
- Utility functions and helpers may lack tests
- No accessibility testing for critical user flows
- No testing documentation for future contributors

## Requirements

### 1. Expand API Client Tests
**Goal:** Test error scenarios and edge cases for existing API methods

Add tests for:
- **Error scenarios actually used in the app:** 400 validation, 404 not found, 409 conflict, 500 server error, network failures
- **Edge cases that can actually occur:** Empty responses, malformed JSON, network timeouts

### 2. Test Existing Utility Functions
- **Scan first:** Check `src/utils/**/*.ts`, `src/helpers/**/*.ts`, `src/lib/**/*.ts` for actual utility code
- **Only test what exists:** If there are no utility modules, skip this section
- **Focus on business logic:** Type guards, validation functions, data transformations that contain actual logic

Don't create tests for:
- Trivial utility functions (e.g., simple getters)
- Code that's already covered by integration tests
- Third-party libraries or library wrappers with no custom logic

### 3. Add Integration Tests for Critical Workflows
Create `src/test/integration/` directory only if needed:
- **Component lifecycle** - Create component → Add to view → Verify persistence
- **Relation management** - Create two components → Link them → Verify state updates
- **Error recovery** - API failure → Error display → Successful retry

Focus: Test actual user workflows that span multiple store actions, not individual component behavior.
Target: 5-10 meaningful integration tests (not an arbitrary number)

### 4. Add Accessibility Testing (Playwright Phase 3 dependency)
**Only after Phase 3 (Playwright) is complete:**
- Install `@axe-core/playwright`
- Create `e2e/accessibility.spec.ts` with tests for:
  - Main page has no critical violations
  - Dialogs are keyboard accessible
  - Focus management works correctly

Focus: Test critical accessibility issues, not achieve 100% compliance.

### 5. Add Testing Documentation
Create `frontend/docs/TESTING.md`:
- Testing philosophy (behavior over implementation, minimal mocking)
- When to write unit vs integration vs E2E tests
- How to run tests (commands, watch mode, debugging)
- Examples of good vs bad tests from this codebase
- How to add new tests for new features

Focus: Practical guide for developers, not comprehensive coverage targets.

### 6. Optional: Add Coverage Reporting (Not a Goal)
Only if useful for identifying gaps:
- Add `test:coverage` script to package.json
- Configure coverage reporters in vitest.config.ts
- **No coverage thresholds** - coverage is a tool for finding gaps, not a metric to optimize

## Expected Outcomes
- ✅ API client error handling properly tested
- ✅ Existing utility functions have tests (if any exist)
- ✅ 5-10 integration tests for critical workflows
- ✅ Basic accessibility testing in place (after Phase 3)
- ✅ Testing documentation helps future contributors
- ✅ Coverage reporting available (but not enforced)

## Files to Create
- `frontend/src/test/integration/component-lifecycle.test.tsx` (if needed)
- `frontend/src/test/integration/relation-management.test.tsx` (if needed)
- `frontend/src/utils/*.test.ts` (only for utilities that exist)
- `frontend/e2e/accessibility.spec.ts` (after Phase 3)
- `frontend/docs/TESTING.md`

## Files to Modify
- `frontend/src/api/client.test.ts` - Add error scenario tests
- `frontend/vitest.config.ts` - Add coverage configuration (optional)
- `frontend/package.json` - Add coverage script (optional)

## Dependencies
- Should complete Phase 1 (spec 009)
- Should complete Phase 2 (spec 010)
- Phase 3 (spec 011) required for accessibility testing

## Success Criteria
- [ ] API client handles all error scenarios the app uses
- [ ] Existing utility functions are tested
- [ ] Critical user workflows have integration tests
- [ ] Accessibility tests catch major violations (after Phase 3)
- [ ] TESTING.md provides clear guidance for contributors
- [ ] No artificial coverage targets met

## Checklist
- [ ] Specification ready
- [ ] Scan codebase for existing utility modules
- [ ] Add API client error scenario tests (validation, network, server errors)
- [ ] Create tests for utility functions that exist
- [ ] Create integration test directory
- [ ] Write component lifecycle integration test
- [ ] Write relation management integration test
- [ ] Write error recovery integration test (if relevant)
- [ ] Install @axe-core/playwright (after Phase 3)
- [ ] Create accessibility test suite (after Phase 3)
- [ ] Add optional coverage configuration
- [ ] Add optional coverage scripts
- [ ] Create TESTING.md documentation
- [ ] Review tests for relevance (remove any testing non-existent features)
- [ ] User sign-off
