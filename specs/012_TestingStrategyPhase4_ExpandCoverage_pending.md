# Phase 4: Expand Test Coverage

## Description
Expand frontend test coverage to include untested or under-tested areas: API client edge cases, utility functions, complex workflows, and accessibility testing.

## Problem Statement
After Phases 1-3, gaps remain:
- API client has minimal coverage (32 lines, basic happy path only)
- No tests for utility functions, type guards, or helper modules
- Limited integration tests for complex cross-component workflows
- No accessibility testing
- No documentation on testing approach

## Requirements

### 1. Expand API Client Tests
**Goal:** Increase `src/api/client.test.ts` from 32 lines to comprehensive coverage

Test areas to add:
- **All HTTP methods:** GET, POST, PUT, DELETE for components, relations, views
- **Error scenarios:** 400 validation, 404 not found, 409 conflict, 500 server error, network failures
- **Edge cases:** Empty responses, malformed JSON, timeouts, CORS errors
- **Headers:** Authorization tokens, custom headers
- **Target coverage:** 90%+

### 2. Add Utility Function Tests
- Scan for utility modules: `src/utils/**/*.ts`, `src/helpers/**/*.ts`, `src/lib/**/*.ts`
- Create test files for:
  - Type guards (e.g., `isComponent`, `isValidId`)
  - Validation functions
  - Data transformation utilities
  - Formatters and parsers
- **Target coverage:** 90%+ for utilities

### 3. Add Integration Tests
Create `src/test/integration/` directory with:
- **`component-lifecycle.test.tsx`** - Full create → edit → delete workflow
- **`error-recovery.test.tsx`** - Failed API calls, retry logic, error display
- **`relation-creation.test.tsx`** - Create two components, link them, verify state

Test cross-component interactions and complete user workflows (not single component behavior).
**Target:** 15+ integration test cases

### 4. Add Accessibility Testing
- Install `@axe-core/playwright`
- Create `e2e/accessibility.spec.ts` with:
  - No violations on main page
  - Accessible dialogs (ARIA labels, roles)
  - Keyboard navigation support (Tab, Enter)
  - Focus visibility
- **Target:** 5+ accessibility test cases

### 5. Add Code Coverage Configuration
Update `vitest.config.ts`:
- Add coverage provider (v8)
- Configure reporters (text, html, lcov)
- Set thresholds: 70% lines/functions/branches/statements
- Exclude test files, setup files, type definitions

Add scripts to `package.json`:
- `test:coverage` - Run with coverage report
- `test:coverage:ui` - Interactive coverage UI

### 6. Create Testing Documentation
Create `frontend/docs/TESTING.md`:
- Testing philosophy (behavior over implementation, minimal mocking)
- When to write unit vs integration vs E2E tests
- Test structure (Arrange-Act-Assert)
- Running tests (commands, watch mode, debugging)
- Coverage targets by module type

## Expected Outcomes
- ✅ API client: 90%+ coverage
- ✅ All utility functions tested
- ✅ 15+ integration tests for complex workflows
- ✅ 5+ accessibility tests catch WCAG violations
- ✅ Testing documentation guides future development
- ✅ Overall coverage: 70%+

## Coverage Targets

| Module | Target | Priority |
|--------|--------|----------|
| API Client | 90% | High |
| Store | 80% | High |
| Components | 60% | Medium |
| Utilities | 90% | High |

## Files to Create
- `frontend/src/test/integration/component-lifecycle.test.tsx`
- `frontend/src/test/integration/error-recovery.test.tsx`
- `frontend/src/test/integration/relation-creation.test.tsx`
- `frontend/src/utils/*.test.ts` (for each utility module found)
- `frontend/e2e/accessibility.spec.ts`
- `frontend/docs/TESTING.md`

## Files to Modify
- `frontend/src/api/client.test.ts` - EXPAND from 32 lines to 200+ lines
- `frontend/vitest.config.ts` - ADD coverage configuration
- `frontend/package.json` - ADD coverage scripts

## Dependencies
- Should complete Phase 1 (spec 009)
- Should complete Phase 2 (spec 010)
- Should complete Phase 3 (spec 011)
- Install `@axe-core/playwright` for accessibility testing

## Success Metrics
- [ ] Overall code coverage reaches 70%+
- [ ] API client coverage reaches 90%+
- [ ] Zero accessibility violations in main user flows
- [ ] All utility modules have tests
- [ ] At least 15 integration tests
- [ ] TESTING.md documentation complete

## Checklist
- [ ] Specification ready
- [ ] Scan codebase for untested utility modules
- [ ] Expand API client tests to 90%+ coverage (all methods, errors, edge cases)
- [ ] Create tests for all utility functions and type guards
- [ ] Create integration test directory
- [ ] Write component lifecycle integration tests
- [ ] Write error recovery integration tests
- [ ] Write relation creation integration tests
- [ ] Install @axe-core/playwright
- [ ] Create accessibility test suite
- [ ] Add coverage configuration to vitest.config.ts
- [ ] Add coverage scripts to package.json
- [ ] Create TESTING.md documentation
- [ ] Run coverage report and review gaps
- [ ] Verify coverage thresholds met
- [ ] Update CI to enforce coverage thresholds
- [ ] User sign-off
