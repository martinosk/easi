---
name: mutation-test
description: Run AI-powered mutation testing on source code to find gaps in test coverage. Use when verifying test quality, checking if tests catch real bugs, or improving test suites.
---

Apply semantically meaningful mutations that represent realistic developer bugs — not random syntactic changes.

**Usage**: `/mutation-test <file-or-directory>` or `/mutation-test --branch`

## Steps

### 1. Identify the target

Parse `$ARGUMENTS`:

- **File or directory**: Find source files (`.go`, `.ts`, `.tsx`), skip test files and generated code.
- **`--branch`**: Use `git diff main...HEAD --name-only` to find changed source files, then scope mutations only to changed functions/lines.
- **No argument**: Ask the user.

### 2. Read source and tests

- Read the target source file(s) and their corresponding test file(s).
- If no tests exist for a target, report it as untested and skip it.

### 3. Scan for red flags, then plan mutations

**First, scan the existing tests for common weaknesses:**

- Tests that only assert "no error" / "doesn't throw" without checking return values
- Tests using identity values: `0` for `+/-`, `1` for `*/`, empty string, `true/true` or `false/false` for logical operators
- Tests that check only one side of a condition (never test the boundary)
- Tests that verify a function was called but not with what arguments
- Tests with no assertions at all

Note any red flags found — these indicate where mutations are most likely to survive.

**Then plan 5–15 mutations per file.** Use [mutation-strategies.md](mutation-strategies.md) as a reference.

Prioritize:
- Code paths where red flags were found
- Business logic over boilerplate
- Branching, error handling, edge cases
- Domain boundaries (validation, state transitions)

Skip:
- Type definitions, interfaces, DTOs, imports, generated code
- Code with no tests at all (report as untested, don't mutate)
- **Equivalent mutants** — mutations that produce identical behavior (e.g., `+= 0` → `-= 0`, mutations in dead code). See the equivalent mutants section in mutation-strategies.md.

For each mutation, record: **Location** (file:function:line), **Category**, **Description**, **Expected catching test**.

Present the plan as a numbered table and wait for the user before proceeding.

### 4. Execute mutations one at a time

For each planned mutation:

1. **Apply** a single, minimal change using Edit.
2. **Run tests**: `go test ./<package>/...` (Go) or `npm test -- --run <pattern>` (TypeScript).
3. **Record**: tests fail → **killed**; tests pass → **survived**; error/timeout → **inconclusive**.
4. **Revert** immediately using Edit.
5. **Verify revert** by re-reading the changed line.

**Rules:**
- ONE mutation at a time. Never stack.
- ALWAYS revert before the next mutation. If revert fails, stop and alert the user.
- Never modify test files.
- Mutations must be syntactically valid — no compile errors.
- Show a running tally: `[3/10] Killed | removed nil check in ValidateOrder`

### 5. Generate the report

```
## Mutation Testing Report

**Target**: <file-path>
**Mutations**: <total> | Killed: <n> | Survived: <n> | Inconclusive: <n>
**Mutation Score**: <killed / (killed + survived)>% — <interpretation>

Score interpretation: 90%+ excellent | 70–89% good, some gaps | 50–69% significant gaps | <50% tests miss behavior

### Red Flags Found in Test Suite

- <List weaknesses found in Step 3, if any>

### Surviving Mutants

| # | Location | Mutation | Category | How to Fix |
|---|----------|----------|----------|------------|
| 1 | file:line | description | category | specific test to add, with example values |

### Equivalent Mutants (Skipped)

| # | Location | Mutation | Why Equivalent |
|---|----------|----------|----------------|
| 1 | file:line | description | explanation |

### Killed Mutants

| # | Location | Mutation | Caught By |
|---|----------|----------|-----------|
| 1 | file:line | description | test name |

### How to Strengthen Tests

For each surviving mutant, provide a concrete fix pattern:
- **Boundary not tested**: Add a test at the exact boundary value (e.g., test `age=18` not just `age=25`)
- **Identity values**: Replace test inputs — use `(10, 3)` not `(10, 1)` for multiplication
- **Only one branch tested**: Add the complementary case (test `(true, false)` not just `(true, true)`)
- **No side-effect verification**: Assert on what was called and with what arguments
- **Missing error path test**: Add a test that triggers the error condition
```

### 6. Final verification

Run the relevant test suite one final time to confirm source code is fully restored.
