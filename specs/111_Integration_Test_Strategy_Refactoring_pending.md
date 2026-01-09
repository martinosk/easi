# Integration Test Strategy Refactoring

## Description
Refactor backend integration tests to reduce maintenance burden, improve test value, and properly test the CQRS/ES architecture. Current tests bypass the domain layer with raw SQL inserts, creating tight coupling to database schema and providing low value relative to maintenance cost.

## Problem Statement
Current integration tests for read models:
- Insert test data via raw SQL, bypassing commands, aggregates, events, and projectors
- Duplicate schema knowledge across 15+ helper methods per test file
- Break whenever database schema changes, requiring updates in multiple places
- Test SQL syntax correctness rather than business behavior
- Create illusion of coverage while testing very little domain logic

## Goals
1. Reduce schema coupling in tests
2. Test actual system behavior (command → event → projector → read model)
3. Separate unit-testable logic from integration concerns
4. Establish sustainable patterns for future test development

## Phase 1: Extract and Unit Test Read Model Logic

### Scope
Identify and extract pure functions from read models that can be unit tested without database.

### Requirements
- Extract categorization, calculation, and transformation logic from read models
- Create unit tests for extracted functions (no DB dependency)
- Target: `categorizeResults`, `categorizeGap`, summary calculations, DTO mapping logic

### Affected Areas
- `strategic_fit_analysis_read_model.go` - categorization logic
- `enterprise_capability_link_read_model.go` - conflict detection logic
- Similar patterns in other read models

### Success Criteria
- Pure logic functions have dedicated unit tests
- Unit tests run in < 1 second total
- No database setup required for logic tests

## Phase 2: Create Command-Based Test Fixtures

### Scope
Replace raw SQL test helpers with fixtures that use the command bus.

### Requirements
- Create `internal/testing/fixtures/` package
- Implement fixture builders that dispatch real commands
- Fixtures should handle cleanup via event sourcing (delete commands or test tenant isolation)
- One fixture per bounded context

### Structure
```
internal/testing/fixtures/
├── capability_fixtures.go
├── business_domain_fixtures.go
├── application_fixtures.go
└── metamodel_fixtures.go
```

### Success Criteria
- Fixtures use command bus, not raw SQL
- Single place to update when domain changes
- Tests exercise full event → projector → read model flow

## Phase 3: Migrate Existing Integration Tests

### Scope
Incrementally migrate existing integration tests to use new fixtures.

### Requirements
- Migrate tests opportunistically (when they break or are touched)
- Delete redundant tests that duplicate unit test coverage
- Keep one "smoke test" per read model verifying basic query execution
- Remove tests that only verify data permutations (cover in unit tests)

### Migration Priority
1. Tests that break frequently due to schema changes
2. Tests with most duplicated setup code
3. New tests (write using fixtures from start)

### Target State Per Read Model
- 1-2 integration tests: "query executes and returns expected structure"
- 0 integration tests for logic permutations (covered by unit tests)
- All tests use command-based fixtures

## Phase 4: Establish Testing Guidelines

### Scope
Document patterns and guidelines for future test development.

### Requirements
- Document when to use unit vs integration tests
- Provide examples of good vs bad integration test patterns
- Add to CLAUDE.md or create dedicated testing guide

### Guidelines Summary
- **Unit test**: Domain logic, value objects, categorization, calculations
- **Integration test**: SQL query execution, HTTP handler responses, full command→read flow
- **Never**: Raw SQL inserts to set up test data in integration tests

## Files Affected

### Phase 1
- New unit test files for read model logic
- Possible extraction of pure functions to separate files

### Phase 2
- New: `internal/testing/fixtures/*.go`

### Phase 3
- `*_integration_test.go` files (migrate incrementally)
- Delete redundant SQL helper methods

### Phase 4
- `CLAUDE.md` or new `docs/testing-guidelines.md`

## Out of Scope
- Frontend test changes
- E2E/Playwright test changes
- Changing the read model implementations themselves
- Changing the CQRS/ES architecture

## Dependencies
- None (can start immediately)

## Risks
- Phase 2 requires commands to exist for all test scenarios (may need to add commands)
- Some read models may query data not created by commands in current scope

## Checklist
- [ ] Specification approved
- [ ] Phase 1: Extract pure functions from read models
- [ ] Phase 1: Unit tests for extracted logic
- [ ] Phase 2: Create fixtures package structure
- [ ] Phase 2: Implement capability fixtures
- [ ] Phase 2: Implement business domain fixtures
- [ ] Phase 2: Implement remaining fixtures as needed
- [ ] Phase 3: Migrate first integration test file as proof of concept
- [ ] Phase 3: Document migration pattern
- [ ] Phase 3: Migrate remaining tests (ongoing)
- [ ] Phase 4: Document testing guidelines
- [ ] User sign-off
