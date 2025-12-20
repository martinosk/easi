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
- [x] Create renderWithProviders function that wraps:
  - MantineProvider with test theme
  - Router context (MemoryRouter)
  - QueryClientProvider for React Query
- [x] Support partial options injection for testing specific scenarios
- [x] Export from central test utilities location (src/test/helpers/index.ts)

### Phase 2: Mock Factories
- [x] Create API client mock factory with sensible defaults
- [x] Create feature-specific API mock factories (components, relations, views, capabilities, etc.)
- [x] Create common entity factories (component, capability, relation, etc.)
- [x] Ensure mocks are typed correctly

### Phase 3: Test Setup Consolidation
- [x] Global mocks already centralized in setup.ts
- [x] Entity builders with sensible defaults created
- [x] Mock factories exportable from central location

### Phase 4: Migration
- [x] New utilities available for existing tests (gradual adoption)
- [x] All test utilities exported from src/test/helpers/index.ts
- [x] Testing patterns established via entityBuilders and mockApiClient

## Incremental Delivery
1. First: Create unified render utility (new tests can use it immediately)
2. Second: Create mock factories
3. Third: Consolidate setup
4. Fourth: Migrate existing tests (can be done gradually)

## Checklist
- [x] Specification ready
- [x] Render utility created (renderWithProviders)
- [x] Mock factories created (all feature APIs + entities)
- [x] Setup consolidated (central exports from test/helpers)
- [x] Existing tests can use new utilities (gradual adoption)
- [x] Testing patterns established
- [x] User sign-off
