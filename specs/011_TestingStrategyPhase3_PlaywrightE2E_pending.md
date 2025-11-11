# Phase 3: Add Playwright E2E Testing

## Description
Introduce true end-to-end testing using Playwright to cover canvas interactions, drag-and-drop functionality, and critical user workflows. This replaces the memory-leaking React Flow component tests removed in Phase 1 with superior browser-based tests.

## Problem Statement
After Phase 1, we lost test coverage for:
- React Flow canvas rendering and interactions
- Drag-and-drop component positioning
- Creating relations by connecting component handles
- Multi-view navigation and canvas state

These cannot be reliably tested in JSDOM because React Flow requires real browser rendering and proper canvas API. Playwright runs tests in real browser instances with proper cleanup.

## Requirements

### 1. Install and Configure Playwright
- Install `@playwright/test` and Chromium browser only
- Create `playwright.config.ts` with:
  - Test directory: `./e2e`
  - Base URL: `http://localhost:5173`
  - Web server config to start dev server automatically
  - Retry on CI, trace/video on failure
  - Single project (chromium only)
- Update `.gitignore` for playwright-report, test-results

### 2. Create E2E Test Structure
Create `frontend/e2e/` directory with:
- **`canvas-interactions.spec.ts`** - Canvas rendering, drag-drop, zoom, selection
- **`component-management.spec.ts`** - Create/edit/delete components via UI
- **`multi-view.spec.ts`** - View switching, separate canvas state per view
- **`fixtures/testData.ts`** - Shared test data

### 3. Add Test Data Attributes
Update components to include `data-testid` and `data-component-id` attributes:
- ComponentCanvas nodes need `data-component-id`
- Dialogs need `data-testid` for inputs/buttons
- Add `data-testid="canvas-loaded"` indicator for test synchronization

### 4. Write E2E Tests
Cover these critical flows:
- Components render on canvas
- Drag component updates position (persist after reload)
- Connect handles to create relation
- Select/deselect components
- Zoom controls work
- Create component via dialog (validation errors, success)
- Edit/delete components
- Switch between views maintains separate state

### 5. Update NPM Scripts
Add to `package.json`:
- `test:e2e` - Run Playwright tests
- `test:e2e:ui` - Interactive UI mode
- `test:e2e:debug` - Debug mode
- `test:all` - Run unit + E2E tests
- `test:ci` - CI mode with coverage

### 6. Configure CI Integration
Update CI workflow to:
- Install Playwright with `--with-deps chromium`
- Run E2E tests
- Upload playwright-report on failure

## Expected Outcomes
- ✅ Playwright configured and running
- ✅ E2E tests cover all canvas interactions removed in Phase 1
- ✅ Tests run in real browser (no JSDOM memory issues)
- ✅ 15+ test cases covering canvas, CRUD, multi-view
- ✅ CI pipeline includes E2E tests
- ✅ Test execution time < 3 minutes

## Coverage Comparison
**Phase 1 lost:** 11 test cases (canvas rendering/drag-drop)
**Phase 3 added:** 15+ E2E test cases with superior coverage

## Files to Create
- `frontend/playwright.config.ts`
- `frontend/e2e/canvas-interactions.spec.ts`
- `frontend/e2e/component-management.spec.ts`
- `frontend/e2e/multi-view.spec.ts`
- `frontend/e2e/fixtures/testData.ts`

## Files to Modify
- `frontend/package.json` - Add E2E scripts
- `frontend/.gitignore` - Ignore Playwright artifacts
- `frontend/src/components/ComponentCanvas.tsx` - Add data attributes
- `frontend/src/components/CreateComponentDialog.tsx` - Add data-testid
- `frontend/src/components/CreateRelationDialog.tsx` - Add data-testid
- `.github/workflows/test.yml` - Add E2E job (or create if missing)

## Dependencies
- Must complete Phase 1 (spec 009)
- Recommended: Phase 2 (spec 010) for cleaner organization

## Checklist
- [ ] Specification ready
- [ ] Install Playwright and Chromium
- [ ] Create playwright.config.ts
- [ ] Create e2e/ directory structure
- [ ] Add data-testid attributes to components
- [ ] Add canvas-loaded indicator
- [ ] Write canvas-interactions tests (6+ cases)
- [ ] Write component-management tests (4+ cases)
- [ ] Write multi-view tests (2+ cases)
- [ ] Create test fixtures
- [ ] Update package.json with E2E scripts
- [ ] Update .gitignore
- [ ] Run E2E tests locally to verify
- [ ] Configure CI pipeline
- [ ] Verify E2E tests pass in CI
- [ ] Verify all Phase 1 lost coverage restored
- [ ] Performance check: E2E complete in <3 min
- [ ] User sign-off
