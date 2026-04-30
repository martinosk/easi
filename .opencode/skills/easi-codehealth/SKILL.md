---
name: easi-codehealth
description: MUST load when the user asks to check code health, run codehealth, refactor for codehealth, run a pre-commit codehealth check, or assess code quality of pending/branch changes. Enforces a per-file 10.0 verification — passing CodeScene quality gates alone is NOT sufficient.
compatibility: opencode
---

# EASI Code Health

## Overview

Codehealth prompts in EASI are checked at the per-file level, not just the gate level. CodeScene's `pre_commit_code_health_safeguard` returns `quality_gates: passed` whenever no file *crosses a threshold* on this branch — but a file already at 9.4 that drifts to 9.3 still passes the gate. The project standard (recorded in user memory: "ALWAYS pursue 10.0 Code Health") is that every modified or added file in the pending change set must score 10.0, except for files where the remaining smell is genuinely inherent.

This skill exists because the LLM failure mode is to stop at "quality_gates: passed" and report success while sub-10.0 files quietly remain. That's a regression in disguise.

## Iron Law

**`quality_gates: passed` is necessary but not sufficient.** Every modified and untracked file in the pending change set must be individually scored. Files below 10.0 must be either refactored to 10.0 or explicitly justified as inherent — never silently accepted.

## Procedure

Run these steps in order. Do not skip steps even when an earlier step looks clean.

### 1. Run the pre-commit safeguard
Call `mcp__codescene__pre_commit_code_health_safeguard` for the repo. Note any `degraded` verdicts and any introduced findings. This is the gate signal — useful, but not the verdict.

### 2. Enumerate every pending-changes file
From `git status`, collect both modified (`M`) and untracked (`??`) files in scope. Test files (`*_test.go`, `*.test.ts`, `*.test.tsx`) count — score them too.

### 3. Score every file individually
Call `mcp__codescene__code_health_score` per file. Run these in parallel — they are independent. Build a table of file → score.

### 4. Categorize sub-10.0 files
For each file under 10.0, decide one of:

- **Fixable** — the smell can be addressed within the spirit of this branch. Refactor it. Common patterns that reliably reach 10.0:
  - **Code Duplication in tests** → table-driven tests
  - **String Heavy / Primitive Obsession** → bundle related primitives into a typed struct
  - **Complex Method / Bumpy Road** → extract named helpers or sub-functions
  - **Large Method (React)** → extract custom hooks or sub-components
  - See user memory `codescene-refactoring.md` for the full pattern list (and known anti-patterns — splitting a structurally-similar switch into sub-methods *increases* duplication).
- **Inherent** — only these qualify:
  - **Auto-generated files**: `backend/docs/docs.go`, `frontend/openapi.json`. Score is bounded by the generator and the API surface. Note this and move on.
  - **Pre-existing structural debt outside the branch's scope**: the smell existed before this branch started, the branch touched the file only incidentally, and fixing it would require structural changes far beyond the spec. Must be called out by name in the report (file, score, smell, why deferred), never silently skipped. If the branch *introduced* or *worsened* a smell, it does not qualify as inherent.

### 5. Refactor each fixable file
Apply the refactor. Then re-run any tests touching that file (`go test ./...` for backend changes, the relevant `npm test -- --run <pattern>` for frontend) before re-scoring. A 10.0 score with broken tests is worthless.

### 6. Re-score and re-run the safeguard
After all fixes, re-score every file you refactored, and re-run `pre_commit_code_health_safeguard`. Both must agree.

### 7. Report
Final report must include:
- Final per-file score table for every file in the change set
- Quality gate verdict
- Any inherent files explicitly listed with reason
- Tests that were re-run

Never report success off `quality_gates: passed` alone. Never report "all files at 10.0" without having scored every file in the change set.

## Hard Gates

- **Hard gate before reporting success**: every modified + untracked file appears in your final score table. If you cannot point at a score for every file, you have not finished.
- **Hard gate before claiming inherent**: name the file, the smell category, and *why* it cannot be addressed in this branch. "Pre-existing" alone is not enough — pre-existing smells are still fixable; they just may be out of scope.
- **Hard gate before refactor commits**: tests pass after the refactor.

## Rationalization Prevention

| Excuse | Reality |
|--------|---------|
| "Quality gates passed, we're done" | Gates measure threshold-crossing, not 10.0. Score every file. |
| "It's only test code" | User memory: "ALWAYS check code health on test files too." Test code is scored the same way. |
| "It was already that score before the branch" | "Pre-existing" doesn't auto-qualify as inherent. Explain why fixing it is out of scope, or fix it. |
| "Refactoring it would be a larger change" | That may be true — and it must be called out explicitly with that reasoning, not silently dropped. |
| "The score only dropped from 9.4 to 9.3" | Both are below 10.0. The branch standard is 10.0. |
| "The drift is microscopic (0.3%)" | Direction matters. If the branch made it worse, that is a regression even if the gate passed. |

## What Counts as Inherent (Explicit Allowlist)

Only these qualify without further justification:
- `backend/docs/docs.go` — Swaggo-generated; size grows with API surface; regenerated via `make swagger`.
- `frontend/openapi.json` — generated from backend OpenAPI spec.

Any other file requires a written justification in the report.

## Output

A per-file score table for every file in the pending change set, gate verdict, and an explicit inherent-files list. No success claim without all three.
