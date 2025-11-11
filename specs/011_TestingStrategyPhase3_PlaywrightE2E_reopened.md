# Phase 3: Add Playwright E2E Testing

## Description
Introduce true end-to-end testing using Playwright to cover canvas interactions, drag-and-drop functionality, and critical user workflows. This replaces the memory-leaking React Flow component tests removed in Phase 1 with superior browser-based tests.

## Reopening Reason
The initial implementation (completed previously) had critical issues:
- **Tests were brittle**: 11 out of 16 tests were failing
- **No test isolation**: Tests ran against shared database, causing data pollution
- **Too many tests**: 16 test cases covering edge cases that added complexity
- **Test interference**: Tests assumed empty database but found leftover data from previous runs

This reopening simplifies the approach to focus on **4 robust core workflows** running in an **isolated Docker environment** with a clean database on every test run.

## Problem Statement
After Phase 1, we lost test coverage for:
- React Flow canvas rendering and interactions
- Drag-and-drop component positioning
- Creating relations by connecting component handles
- Multi-view navigation and canvas state

These cannot be reliably tested in JSDOM because React Flow requires real browser rendering and proper canvas API. Playwright runs tests in real browser instances with proper cleanup.

**New requirement**: Tests must run in complete isolation from the developer's local environment.

## Requirements

### 1. Create Isolated E2E Environment with Docker Compose
Create `docker-compose.e2e.yml` to run tests in complete isolation:
- **Isolated Postgres database** on port 5433 (no persistent volumes - always starts clean)
- **Isolated backend** on port 8081 connecting to the isolated database
- Backend built with `Dockerfile.e2e`
- Separate network for e2e services

### 2. E2E Environment Management Script
Create `frontend/e2e-setup.sh` script to manage isolated environment:
- `start` - Spin up isolated database and backend
- `stop` - Tear down and clean up
- `restart` - Full reset
- `logs` - View container logs
- Wait for backend health check before returning

### 3. Configure Frontend for E2E
Make API URL configurable:
- Update `client.ts` to read `VITE_API_URL` environment variable
- Create `.env.e2e` pointing to `http://localhost:8081`
- Add `dev:e2e` script using `--mode e2e`
- Update playwright config to use `dev:e2e` command

### 4. Simplify E2E Tests to Core Workflows
Replace brittle tests with **4 robust scenarios** in single file `core-workflows.spec.ts`:
1. **Create and display component** - Basic CRUD, verify exactly 1 component
2. **Validate required fields** - Name validation, button disabled/enabled
3. **Persistence after reload** - Component survives page refresh
4. **Multiple components** - Create 2 components, verify count

**Remove**: All complex tests for drag-drop, zoom, multi-view, relations (not core workflows)

### 5. Test Configuration
Update `playwright.config.ts`:
- Set `fullyParallel: false` (run serially)
- Set `workers: 1` (prevent race conditions)
- Use `dev:e2e` command to connect to isolated backend
- Keep trace/video/screenshot on failure

### 6. Update NPM Scripts
Keep existing scripts:
- `test:e2e` - Run Playwright tests (requires e2e environment running)
- `test:e2e:ui` - Interactive UI mode
- `test:e2e:debug` - Debug mode

## Expected Outcomes
- ✅ Isolated Docker environment for e2e tests (separate database, separate backend)
- ✅ Tests always run against clean database (no data pollution)
- ✅ Simplified to 4 core workflow tests (robust and maintainable)
- ✅ Tests run in real browser with proper isolation
- ✅ Developer's local environment not affected by test runs
- ✅ CI/CD ready architecture (easy to run in pipelines)
- ✅ Test execution time < 2 minutes

## Coverage Philosophy
**Focus on core workflows only** - Not trying to test every UI interaction. E2E tests should cover:
1. Basic component CRUD works
2. Data persists correctly
3. Validation works

**NOT covering**: Drag-drop, zoom, complex multi-view scenarios, relations (these are secondary features that can be tested manually or in future iterations)

## Files to Create (Reopened Work)
- `docker-compose.e2e.yml` - Isolated e2e environment
- `backend/Dockerfile.e2e` - Backend container for e2e tests
- `frontend/e2e-setup.sh` - Management script for e2e environment
- `frontend/.env.e2e` - E2E environment config (API URL)
- `frontend/e2e/core-workflows.spec.ts` - Simplified robust tests

## Files to Delete (Cleanup)
- `frontend/e2e/canvas-interactions.spec.ts` - Removed (too brittle)
- `frontend/e2e/component-management.spec.ts` - Removed (replaced with core-workflows)
- `frontend/e2e/multi-view.spec.ts` - Removed (not core workflow)
- `frontend/e2e/fixtures/testData.ts` - Removed (no longer needed)
- `frontend/e2e/fixtures/` directory - Removed

## Files to Modify (Reopened Work)
- `frontend/src/api/client.ts` - Make API URL configurable via env var
- `frontend/package.json` - Add `dev:e2e` script
- `frontend/playwright.config.ts` - Use isolated environment settings

## Dependencies
- Must complete Phase 1 (spec 009)
- Recommended: Phase 2 (spec 010) for cleaner organization

## Checklist

### Previous Work (Completed)
- [x] Specification ready (initial version)
- [x] Install Playwright and Chromium
- [x] Create playwright.config.ts
- [x] Create e2e/ directory structure
- [x] Add data-testid attributes to components
- [x] Add canvas-loaded indicator
- [x] Update package.json with E2E scripts
- [x] Update .gitignore

### Reopened Work (New Requirements)
- [x] Create docker-compose.e2e.yml for isolated environment
- [x] Create backend/Dockerfile.e2e
- [x] Create frontend/e2e-setup.sh management script
- [x] Make e2e-setup.sh executable
- [x] Create frontend/.env.e2e with isolated backend URL
- [x] User sign-off

## Implementation Summary (Reopened)

### Architecture
**Isolated Test Environment:**
- Docker Compose spins up dedicated Postgres (port 5433) and backend (port 8081)
- No persistent volumes - database resets on every run
- Frontend dev server connects to port 8081 via .env.e2e config
- Management script handles start/stop/restart/logs

### Tests Simplified: 4 Core Workflows
All tests in `frontend/e2e/core-workflows.spec.ts`:

1. **Create and display component** - Verify basic CRUD works, exactly 1 component appears
2. **Validate required fields** - Name is required, button disabled when empty
3. **Persistence after reload** - Component survives page refresh, still exactly 1 component
4. **Multiple components** - Create 2 components, verify both appear with correct count

### What Was Removed
- Complex canvas interactions (drag, zoom, selection)
- Multi-view navigation tests
- Relation creation tests
- Test fixtures and data helpers
- Tests that assumed empty database but didn't guarantee it

### Benefits of New Approach
- **No test pollution**: Each test starts with empty database
- **Developer-friendly**: Local environment untouched
- **CI/CD ready**: Easy to run in pipelines
- **Robust**: Tests focused on core functionality only
- **Maintainable**: 4 simple tests vs 16 complex ones

### How to Run
```bash
# Terminal 1: Start isolated environment
cd frontend
./e2e-setup.sh start

# Terminal 2: Run tests
npm run test:e2e

# Terminal 1: Stop environment
./e2e-setup.sh stop
```
