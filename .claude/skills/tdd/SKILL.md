---
name: tdd
description: Test-Driven Development workflow. Use this skill for ALL implementation tasks — features, bug fixes, new domain models, new endpoints, new UI components. TDD is the default development approach.
argument-hint: <description of what to implement>
---

Implement `$ARGUMENTS` using strict RED-GREEN-REFACTOR TDD.

No production code without a failing test first.

## RED-GREEN-REFACTOR Cycle

### RED: Write a Failing Test First

- Write a test that describes the desired **behavior**, not the implementation
- Run the test — confirm it **fails for the right reason** (not a syntax error or missing import)
- The failing test defines the next increment of work

### GREEN: Minimum Code to Pass

- Write **only** enough production code to make the failing test pass
- Resist adding functionality not demanded by a test
- Run the test — confirm it passes

### REFACTOR: Clean Up (Only If Valuable)

- Assess after every GREEN
- All tests must still pass after refactoring

### Repeat

Continue the cycle until the full behavior described in `$ARGUMENTS` is implemented.

## Project Test Conventions

Follow conventions in [docs/backend/testing.md](/docs/backend/testing.md) (Go) and [docs/frontend/README.md](/docs/frontend/README.md) (TypeScript).

### What to Test

| Layer | Test Focus | Example |
|-------|------------|---------|
| **Value objects** | Validation, equality, immutability | `NewEmail("")` returns `ErrEmailEmpty` |
| **Aggregates** | Command handling, state transitions, event emission | Creating a component emits `ComponentCreated` |
| **Projectors** | Event → read model mapping | `ComponentCreated` inserts row with correct fields |
| **API handlers** | HTTP status codes, response shape, HATEOAS links | `POST /components` returns `201` with `Location` header |
| **React hooks** | Data fetching, state management, error handling | `useComponents()` returns loading then data |
| **React components** | User interactions, conditional rendering | Click "Delete" shows confirmation dialog |
| **Utility functions** | Pure input/output, edge cases | `filterByCreator([], [userId], map)` returns `[]` |

### What NOT to Test

- Type definitions, interfaces, DTOs, imports
- Framework wiring (router setup, DI container config)
- Code that simply delegates to another tested function
- Implementation details (internal state, private methods)

## Steps

### 1. Understand the requirement

Parse `$ARGUMENTS`. If the requirement is unclear, ask the user before writing any code.

Identify:
- What **behavior** needs to exist (not what code to write)
- Which **layer(s)** are involved (domain, application, infrastructure, frontend)
- What **existing code** to read first (related aggregates, handlers, components)

### 2. Plan the test sequence

Break the requirement into small, incremental behavior steps. Each step = one RED-GREEN cycle.

Order from inside out:
1. **Domain model** first (value objects → aggregates → domain services)
2. **Application layer** next (command handlers, projectors, read models)
3. **Infrastructure** last (API handlers, repositories)
4. **Frontend** after backend is solid (hooks → components → pages)

Present the sequence to the user as a brief numbered list before starting.

### 3. Execute RED-GREEN-REFACTOR cycles

For each behavior step:

1. **RED**: Write the failing test
2. Run the test — show the failure
3. **GREEN**: Write minimum production code
4. Run the test — show it passes
5. **REFACTOR**: Assess and improve
6. Run the test again after refactoring

Show progress: `[3/8] GREEN | NewEmail rejects empty string`

### 4. Verify the full build

After all cycles are complete:

- **Backend**: `go build ./...` and `go test ./...`
- **Frontend**: `npm run build` and `npm test -- --run`
- Fix any issues before reporting success

### 5. Report

```
## TDD Implementation Complete

**Implemented**: <what was built>
**Cycles**: <N> RED-GREEN-REFACTOR cycles

### Tests Written
| # | Test | Layer | Status |
|---|------|-------|--------|
| 1 | description | domain/app/infra/frontend | PASS |

### Production Code Created/Modified
- <file>: <what was added/changed>

### Checklist
- [ ] Every production code line has a failing test that demanded it
- [ ] All tests pass
- [ ] Build passes (backend and/or frontend)
- [ ] No speculative code ("just in case" logic without tests)
```

## Anti-Patterns to Avoid

- Writing production code without a failing test first
- Testing implementation details (internal state, call counts on internal methods)
- Over-mocking — prefer real objects, use fakes/mocks only at system boundaries
- Tests that only assert "no error" without checking the actual result
- Identity test values: `0` for `+/-`, `1` for `*`, empty string, `true/true`
- Speculative code not demanded by any test
- Rewriting an entire file wholesale — use incremental edits
