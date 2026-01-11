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

## Phase 1: Unit Test Existing Pure Functions

### Scope
Add unit tests for already-extracted pure functions that lack test coverage.

### Analysis
Review found that `categorizeGap` logic is already extracted to `valueobjects/gap_category.go` but has no unit tests. The `categorizeResults` method in read models delegates to this value object, so no additional extraction is needed.

### Requirements
- Add unit tests for `valueobjects/gap_category.go` covering all categorization scenarios
- Test boundary conditions: liability threshold (gap>=2), concern threshold (gap>=1), high importance (>=4)
- Verify existing value object test patterns and follow conventions

### Affected Files
- New: `valueobjects/gap_category_test.go`

### Success Criteria
- `CategorizeGap` function has comprehensive unit tests
- Unit tests run in < 1 second
- All threshold combinations covered (liability, concern, aligned)

## Phase 2: Create Command-Based Test Fixtures

### Scope
Replace raw SQL test helpers with fixtures that use the command bus.

### Requirements
- Create `internal/testing/fixtures/` package
- Implement fixture builders that dispatch real commands via handlers
- Fixtures set tenant context and provide necessary infrastructure (event store, command bus)
- Cleanup via t.Cleanup() with delete commands or transaction rollback
- One fixture per bounded context

### Structure
```
internal/testing/fixtures/
├── base.go              # Shared test infrastructure (DB, tenant context, command bus setup)
├── capability_fixtures.go
├── business_domain_fixtures.go
├── application_fixtures.go
└── metamodel_fixtures.go
```

### Implementation Approach
1. Create `TestContext` struct that holds DB, tenant context, and command dispatcher
2. Fixture methods accept `TestContext` and return created entity IDs
3. Each fixture method dispatches appropriate commands (CreateCapability, AssignToBusinessDomain, etc.)
4. Use existing handler infrastructure rather than bypassing to raw commands

### Success Criteria
- Fixtures use command bus, not raw SQL
- Single place to update when domain changes
- Tests exercise full event → projector → read model flow

## Phase 3: Migrate Existing Integration Tests

### Scope
Incrementally migrate existing integration tests to use new fixtures.

### Test Type Classification
1. **Handler tests** (`*_handlers_*_test.go`) - Test HTTP handler command→response flow
   - SHOULD use command-based fixtures for all data setup
   - Tests should exercise the full command→event→projector→response cycle
   - Example: `capability_handlers_integration_test.go`

2. **Read model tests** (`*_read_model_*_test.go`) - Test SQL query correctness
   - MAY use raw SQL for projection table setup when testing specific query edge cases
   - Focus on "does the SQL query return correct results for this data state"
   - Example: `strategic_fit_analysis_handlers_integration_test.go` (tests read model queries)

### Requirements
- Handler tests: Replace all raw SQL with fixtures
- Read model tests: Use fixtures for aggregate tables, raw SQL acceptable for projection tables when testing specific query scenarios
- Delete redundant tests that duplicate unit test coverage
- Keep one "smoke test" per read model verifying basic query execution

### Migration Priority
1. Handler tests (highest value from migration)
2. Read model tests with duplicate setup code
3. New tests (write using fixtures from start)

### Target State Per Read Model
- 1-2 integration tests: "query executes and returns expected structure"
- 0 integration tests for logic permutations (covered by unit tests)
- Handler tests use command-based fixtures

## Phase 4: Establish Testing Guidelines

### Scope
Document patterns and guidelines for future test development in CLAUDE.md.

### Requirements
- Add "Backend Integration Testing" section to CLAUDE.md
- Document when to use unit vs integration tests
- Provide clear anti-patterns to avoid

### Guidelines Summary (for CLAUDE.md)
```markdown
## Backend Integration Testing

### Test Categories
- **Unit test**: Domain logic, value objects, aggregates, pure functions
- **Integration test**: SQL query execution, HTTP handler responses, full command→read flow

### Integration Test Rules
- ALWAYS use command-based fixtures from `internal/testing/fixtures/`
- NEVER insert test data via raw SQL - this bypasses domain logic
- Keep integration tests focused on "query executes correctly" rather than data permutations
- Data permutation tests belong in unit tests for the underlying logic

### Anti-patterns
- Raw SQL INSERT statements in test setup (couples tests to schema)
- Testing categorization logic via integration tests (should be unit tested)
- Multiple helper methods per test file that duplicate INSERT statements
```

## Files Affected

### Phase 1
- New: `internal/capabilitymapping/domain/valueobjects/gap_category_test.go`

### Phase 2
- New: `internal/testing/fixtures/base.go`
- New: `internal/testing/fixtures/capability_fixtures.go`
- New: `internal/testing/fixtures/business_domain_fixtures.go`
- New: `internal/testing/fixtures/application_fixtures.go`
- New: `internal/testing/fixtures/metamodel_fixtures.go`

### Phase 3
- `*_integration_test.go` files (migrate incrementally)
- Delete redundant SQL helper methods as tests are migrated

### Phase 4
- `CLAUDE.md` - add Backend Integration Testing section

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
- [x] Specification reviewed and validated
- [x] Phase 1: Add unit tests for gap_category.go
- [x] Phase 2: Create fixtures package with base.go
- [x] Phase 2: Implement capability fixtures
- [x] Phase 2: Implement business domain fixtures
- [x] Phase 2: Implement application fixtures
- [x] Phase 2: Implement strategic analysis fixtures
- [x] Phase 3: Create proof-of-concept fixture tests (all passing)
- [x] Phase 3: Document migration guidelines (handler vs read model tests)
- [x] Phase 4: Add testing guidelines to CLAUDE.md
- [ ] User sign-off

## Notes on Phase 3 Migration
Existing integration tests remain functional. Migration guidelines added to spec:
- **Handler tests** should migrate to fixtures when touched
- **Read model tests** may keep raw SQL for projection tables
- New tests should use fixtures from the start
