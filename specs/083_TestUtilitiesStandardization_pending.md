# Test Utilities Standardization

## Description
Create a unified test utility infrastructure to reduce boilerplate, improve consistency, and make tests easier to write and maintain.

## Current State
- Test setup is scattered across individual test files
- `mantineTestWrapper.tsx` exists but usage is inconsistent
- Mock patterns vary between test files
- Store mocking is repeated in each test file
- No standard way to render components with all required providers

## Target State
- Single render utility that wraps components with all required providers
- Standardized mock factories for common dependencies
- Consistent patterns across all test files
- Reduced boilerplate in individual tests

## Requirements

### Phase 1: Unified Render Utility
- [ ] Create renderWithProviders function that wraps:
  - MantineProvider with test theme
  - Router context (when React Router migration complete)
  - Any other required context providers
- [ ] Support partial store state injection for testing specific scenarios
- [ ] Export from central test utilities location

### Phase 2: Mock Factories
- [ ] Create API client mock factory with sensible defaults
- [ ] Create store mock factory for Zustand state
- [ ] Create common entity factories (component, capability, relation, etc.)
- [ ] Ensure mocks are typed correctly

### Phase 3: Test Setup Consolidation
- [ ] Move all global mocks to central setup file
- [ ] Document mock override patterns for specific tests
- [ ] Remove duplicated mock code from individual test files

### Phase 4: Migration
- [ ] Update existing tests to use new utilities
- [ ] Remove deprecated test helpers
- [ ] Document testing patterns for new tests

## Incremental Delivery
1. First: Create unified render utility (new tests can use it immediately)
2. Second: Create mock factories
3. Third: Consolidate setup
4. Fourth: Migrate existing tests (can be done gradually)

## Checklist
- [ ] Specification ready
- [ ] Render utility created
- [ ] Mock factories created
- [ ] Setup consolidated
- [ ] Existing tests migrated
- [ ] Testing patterns documented
- [ ] User sign-off
