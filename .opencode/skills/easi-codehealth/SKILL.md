---
name: easi-codehealth
description: MUST load when the user asks to check code health, run codehealth, refactor for codehealth, run a pre-commit codehealth check, or assess code quality of pending/branch changes. Enforces a per-file 10.0 verification — passing CodeScene quality gates alone is NOT sufficient.
compatibility: opencode
---

# EASI Code Health

## Iron Law

Every modified and untracked file in the pending change set must:
1. Score 10.0 on `mcp__codescene__code_health_score`, or be explicitly justified as inherent.
2. Pass a lead-developer review pass (Step 6) for design problems CodeScene cannot detect.

`quality_gates: passed` is necessary but not sufficient. A 10.0 score is necessary but not sufficient. Both are required.

## Procedure

Run in order. Do not skip steps.

### 1. Enumerate every pending-changes file
From `git status`: modified (`M`) and untracked (`??`). Test files (`*_test.go`, `*.test.ts`, `*.test.tsx`) count.

### 2. Score every file individually
Call `mcp__codescene__code_health_score` per file, in parallel. Build a file → score table.

### 3. Categorize sub-10.0 files
Each file under 10.0 is one of:

- **Fixable.** Refactor it. Patterns that reach 10.0:
  - Code Duplication in tests → table-driven tests
  - String Heavy / Primitive Obsession → typed struct
  - Complex Method / Bumpy Road → extract named helpers
  - Large Method (React) → extract custom hooks or sub-components
- **Inherent** — only these qualify:
  - Auto-generated: `backend/docs/docs.go`, `frontend/openapi.json`.
  - Pre-existing structural debt outside branch scope. Must be named in the report (file, score, smell, why deferred). If the branch introduced or worsened the smell, it is not inherent.

### 4. Refactor each fixable file
Apply the refactor. Re-run tests touching that file (`go test ./...` or `npm test -- --run <pattern>`) before re-scoring.

If a refactor worsens readability, cohesion, or domain fit (one-caller helpers created only to dodge a metric; splitting intentionally-parallel switches; renaming for the metric not for intent) — revert it. Find a better refactor or mark inherent with reasoning.

### 5. Lead-developer review pass
Re-read the full diff and the surrounding code. For every changed file, walk this checklist and act on every issue found:

- **Naming.** Identifiers reveal intent in domain terms. No `data`, `info`, `process`, `handle`, `manager`, `helper`, `util`, bare boolean `flag`. Function names match what they do.
- **Abstraction level.** Statements within one function sit at the same level.
- **Cohesion.** Each file/function/component does one thing.
- **Coupling.** No new cross-context imports. No reaching through objects (`a.b.c.d`). Data fetched at the right layer.
- **Domain modeling.** No primitive obsession (`string` that is an `OrderId`; `map[string]any` that is a typed payload). No anaemic structs that should hold behavior. No domain rules in handlers, no presentation in domain.
- **Semantic duplication.** Functions doing the same thing with different shapes. Parallel hierarchies. Copy-pasted-then-tweaked logic.
- **Dead or speculative code.** No unused exports, unreachable branches, unread parameters, one-caller "just in case" abstractions, flags for shipped features.
- **Error handling.** No swallowed errors. Wrap with context (`%w` in Go). Errors handled at the right layer. Typed errors over `panic`/`throw`. Query/error states surfaced in the UI.
- **Tests.** Names describe behavior, not implementation. Test through the public API. No over-mocking. Assertions do not lock in incidental detail. Edge cases covered (empty, nil, boundary, concurrency, auth-failure).
- **Comments.** No comments unless the user asked. Any `// what this does` comment is a refactor signal.
- **Architectural fit.** Respects bounded contexts and patterns from `easi-architecture-canvas`, `easi-go-backend-patterns`, `easi-frontend-patterns`, `easi-api-standards`.
- **Diff hygiene.** No reformat noise, debug prints, commented-out code, or TODOs without follow-up.

After fixing, re-run tests and re-score. Issues you choose not to fix go in the report with reasoning.

### 6. Re-score and re-run the safeguard
Re-score every refactored file.

### 7. Report
- Per-file score table for every file in the change set
- Quality gate verdict
- Inherent files: file, score, smell, reason
- Lead-developer review findings: per changed file, "no design issues" or list of issues + resolution (fixed / deferred-with-reason)
- Tests re-run

## Hard Gates

- Every modified + untracked file appears in the final score table.
- Every "inherent" claim names file, smell, and reason.
- Tests pass after refactor.
- Every changed file has been through the Step 6 review and the report records the result.

## Rationalization Prevention

| Excuse | Rule |
|--------|------|
| "Quality gates passed, we're done" | Score every file. |
| "It's only test code" | Test files are scored the same way. |
| "It was already that score" | "Pre-existing" is not "inherent". Fix it or justify deferral by name. |
| "Refactoring it would be a larger change" | Call it out explicitly with that reasoning. |
| "The score only dropped from 9.4 to 9.3" | Below 10.0 is below 10.0. |
| "The drift is microscopic" | If the branch made it worse, fix it. |
| "All files are at 10.0, so the review is done" | Run the Step 6 pass. |
| "CodeScene didn't flag it" | Step 6 covers what CodeScene cannot see. |
| "Extracting this helper got the score to 10.0" | One-caller helpers created only to dodge a metric: revert. |
| "Tests still pass after the refactor" | Verify the public contract; tests may be locking in implementation. |

## Inherent Allowlist

No further justification needed:
- `backend/docs/docs.go` — Swaggo-generated; regenerated via `make swagger`.
- `frontend/openapi.json` — generated from backend OpenAPI spec.

All other files require written justification.

## Output

Per-file score table, gate verdict, inherent list, and Step 6 review summary. No success claim without all four.
