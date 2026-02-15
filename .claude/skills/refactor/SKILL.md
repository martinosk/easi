---
name: refactor
description: Analyze and refactor code for improved health, domain alignment, and architectural compliance. Supports single-file, multi-file, and cross-context architectural refactoring.
argument-hint: <file-or-directory | --scope=architecture>
---

Refactor code in the EASI codebase — from single-file code health improvements to multi-file architectural changes.

**Usage**: `/refactor <file-or-directory>` or `/refactor --scope=architecture`

**Arguments**:
- `file(s)`: Files to analyze and refactor (single-file mode)
- `directory or package`: Directory to analyze for structural patterns (multi-file mode)
- `--scope=architecture`: Cross-context architectural analysis (architecture mode)
- No argument: prompt the user

## Before You Start

Remind the user: **commit or stash any uncommitted work before refactoring.** Refactoring should always start from a clean working tree so changes can be reviewed and reverted cleanly.

## Steps

### 1. Identify target and mode

Parse `$ARGUMENTS` to determine the refactoring mode:

| Argument | Mode | Scope |
|----------|------|-------|
| Single file (`.go`, `.ts`, `.tsx`) | **Single-file** | Code health + domain alignment for one file |
| Directory or Go package path | **Multi-file** | Structural patterns across files in the directory |
| `--scope=architecture` | **Architecture** | Cross-context coupling, ACL compliance, published language gaps |
| No argument | Ask the user | — |

### 2. Analyze

Run analysis appropriate to the mode. Use [codescene-refactoring.md](codescene-refactoring.md) as references throughout.

#### Single-file mode

1. **CodeScene code health**: Run `mcp__codescene__code_health_score` and `mcp__codescene__code_health_review` on the file.
2. **Read the file** to understand the full context.
3. **Domain analysis**: Check for:
   - Logic in the wrong layer (business logic in projectors/handlers, infrastructure in domain)
   - Primitive obsession (string/int where a value object should exist)
   - Aggregate doing too much (mixed responsibilities)
   - Missing or misplaced validation
   - Anemic domain models and complex read models

#### Multi-file mode

1. **Glob** for all source files in the directory.
2. **CodeScene code health** on each file — note any below 10.0.
3. **Structural analysis**: Check for:
   - Layer violations (domain importing infrastructure, handler containing business logic)
   - Duplicated patterns across files that indicate a missing abstraction
   - Inconsistent naming or organization vs project conventions
   - Anemic domain models and complex read models
   - Read model / projector alignment (does each projector serve exactly one read model?)

#### Architecture mode

1. **Cross-context coupling**: Grep for imports between bounded contexts that bypass published language packages.
2. **SQL join analysis**: Look for SQL queries that join tables owned by different contexts.
3. **Published language gaps**: Check if events and shared types are properly exposed via `publishedlanguage` packages.
4. **ACL compliance**: Verify anti-corruption layers exist where contexts communicate.
5. **Cache projector coverage**: Check if cross-context data needs are served by local cache projectors rather than direct queries.

### 3. Classify findings

Classify each finding into a priority:

| Priority | Criteria | Action |
|----------|----------|--------|
| **Critical** | Breaks architectural boundaries, cross-context coupling, data integrity risk | Must fix immediately |
| **High** | Code health < 4.0, logic in wrong layer, missing domain concept | Must fix |
| **Normal** | Code health 4.0–8.9, minor pattern violations, style inconsistencies | Fix — iterate until resolved |
| **Polish** | Code health 9.0–9.9, minor duplication or string-heavy args in tests | Fix — iterate until 10.0 using table-driven tests or shared helpers |
| **Skip** | Inherent complexity (event-sourced type switches, domain event constructors with many args), already 10.0 | Leave as-is, document reason |

**Goal: All files must reach 10.0 Code Health.** Only skip findings that are inherent to the domain model or architectural pattern (see codescene-refactoring.md for examples). For everything else, iterate until resolved.

### 4. Execute refactoring

Follow these rules:

**General:**
- Preserve all existing behavior — refactoring only, no functional changes
- Respect project code style (no added comments, no over-engineering)

**CodeScene-driven:**
- Re-check with `mcp__codescene__code_health_score` after changes — iterate up to 3 times

**Domain/Architectural:**
- When moving code between files, ensure all references are updated
- When creating new files (migrations, projectors), follow existing naming conventions
- When modifying events or published language, check all subscribers

**"DRY = Knowledge, Not Code":**
- Only extract shared abstractions when the duplicated code represents the **same business concept**
- Structurally similar code that serves different purposes should remain separate

### 6. Verify

- **Backend changes**: Run `go build ./...` and `go test ./...` (or scoped to affected packages)
- **Frontend changes**: Run `npm run build` and `npm test -- --run`
- **Both**: Run both verification suites
- If verification fails, fix the issue before continuing

### 7. Report

```
## Refactoring Complete: <target>

### Changes Applied
| # | File | Change | Before | After |
|---|------|--------|--------|-------|
| 1 | path | description | score/state | score/state |

### Code Health Summary (if applicable)
Before: <original-score(s)>
After:  <final-score(s)>

### Remaining Items
- <any findings that were skipped or couldn't be addressed, with reasons>

### Checklist
- [ ] All changes are behavior-preserving
- [ ] Build passes
- [ ] Tests pass
- [ ] No new cross-context coupling introduced
```

## When NOT to Refactor

Stop and discuss with the user if:
- The refactoring would require **changing public API contracts** (breaking change)
- The "smell" is **inherent to the pattern** (event-sourced type switches, domain event constructors) — see Skip priority

## Rules

- Never exclude test files from analysis unless explicitly asked
