---
name: easi-boyscouting
description: MUST load after completing a feature, bug fix, or any behavior-changing implementation, before declaring the work done. A deliberate cleanup pass over the affected area — files you modified AND files you read while doing the work — so the code is measurably better than you found it. Distinct from `easi-codehealth`: codehealth scores the change set; boyscouting decides what additional structural cleanups belong in this branch.
compatibility: opencode
---

# EASI Boyscouting

## Iron Law

**The branch ends in a better state than it started — across the whole affected area, not just the lines the task forced you to touch.**

"Affected area" = files modified by the task + files read while doing the task + the nearest structural neighbors (same package, same feature folder, same aggregate, the call sites that exercise what you changed).

If you walked through a smell to do your task, the smell is in scope. If you'd have to go looking for it, it isn't.

## When to Run

After GREEN (TDD) and before `easi-codehealth`'s per-file scoring pass. The ordering matters:

1. **TDD GREEN** — tests pass, feature works.
2. **Boyscouting (this skill)** — decide what extra cleanup belongs in this branch, do it, re-run tests.
3. **`easi-codehealth`** — verify every modified + untracked file scores 10.0 or is justified inherent.
4. **PR.**

Doing boyscouting *after* codehealth is wrong: any cleanup you then perform creates new files codehealth never scored.

## Procedure

### 1. List the affected area
From the work just completed, write down:
- Files modified (`git status` — `M` and `??`).
- Files read while reasoning about the task (the imports you opened, the call sites you traced, the tests you consulted).
- Their nearest neighbors: same `internal/<context>/` package, same `frontend/src/features/<feature>/` folder, the aggregate the change lives on.

Anything outside that list is not in scope for this branch.

### 2. Walk the area with these prompts
For each file in the area, ask:

- **Did this file's design make the task harder than it should have been?** Did you copy-paste because the right helper didn't exist? Did you have to thread the same value through five layers? Did you fight a name that didn't mean what it said? Those are signals — the structure is wrong, and you now have the context to fix it.
- **Did the new code make existing code redundant?** Old special cases the new path subsumes. Dead branches. Comments that now describe behavior the code no longer has (delete them; per [[feedback_no_comments]] do not update stale comments — remove them).
- **Did the new code make a duplication visible that wasn't before?** Two functions that now obviously do the same thing. Parallel switches over the same enum. A primitive that has become an identity (string that is really an `OrderId`, `map[string]any` that is a typed payload).
- **Did you read a test that locks in implementation rather than behavior?** Asserts on call counts where it should assert on outcomes. Mocks where real code would have worked. Names that describe the test's mechanics, not the behavior under test.
- **Did you find unused exports, unread parameters, one-caller "just in case" abstractions, or flags for shipped features?** Delete them. Trust git history.
- **Did you spot a violation of an existing skill in the area?** [[easi-frontend-styling]] (bare HTML / `style={{}}` where Mantine belongs), [[easi-api-standards]] (missing HATEOAS link, wrong DTO shape), [[easi-frontend-data]] (missing cache invalidation, duplicated permission logic), [[easi-domain-driven-design]] (domain rule leaking into a handler, anaemic struct). If you walked through one, it's in scope.

### 3. Sort findings into three buckets

| Bucket | Definition | Action |
|--------|------------|--------|
| **In-branch fix** | Small, contained, structurally tied to what you just changed. Touches files in the affected area. Refactor + cleanup, no behavior change. | Do it now. |
| **Spec-worthy** | Cross-cutting, changes behavior, or large enough that reviewers would have to context-switch to understand it. | Stop. Write a spec per [[easi-spec-driven-development]]. Do not bundle it into this branch. |
| **Out of area** | Real smell, but in code you didn't touch and didn't read. | Leave it. Boyscouting is not a license to audit the repo. |

If a finding sits between "in-branch fix" and "spec-worthy" — bias toward spec-worthy. A bloated PR loses review quality faster than a missed cleanup loses code quality.

### 4. Apply the in-branch fixes
- Each fix is its own logical change.
- Re-run the affected tests after each fix (`go test ./internal/<context>/...` or `npm test -- --run <pattern>`).
- If a refactor breaks a test, the test was probably locking in implementation — fix the test or revert the refactor. Do not weaken a real assertion to make a refactor pass.
- Per [[easi-test-driven-development]] refactor step: behavior does not change. If you find you're changing behavior, stop — that's a separate task, write a test first.

### 5. Re-walk and stop
Re-walk the bucket list one more time. The goal is not "exhaustively perfect" — the goal is "measurably better, and you can name the improvements." When the remaining findings are all out-of-area or spec-worthy, you are done.

Hand off to `easi-codehealth`.

## Decision Heuristics

| Situation | Boyscout? |
|-----------|-----------|
| Smell in a file you modified | Yes. |
| Smell in a file you read while reasoning about the task | Yes. |
| Smell in the same package as a modified file | Yes, if cheap and contained. |
| Smell in a sibling feature folder you didn't open | No. |
| Stale comment in code you walked through | Yes — delete it. |
| Dead export proven unused by a `git grep` | Yes — delete it. |
| Dead-looking export, but you're not sure | Leave it. The cost of guessing wrong is higher than the cost of leaving it. |
| Renaming for clarity, one file, internal symbol | Yes. |
| Renaming an exported domain term across the codebase | Spec-worthy. Not this branch. |
| Extract a one-caller helper to dodge a CodeScene smell | No — see [[easi-codehealth]] Step 4: revert one-caller helpers created only to chase a metric. |
| Pulling a duplicated pattern into a shared abstraction | Only if there are 3+ real callers AND the abstraction sits inside a single bounded context. |
| Architectural restructuring (move package, split aggregate, change context boundary) | Always spec-worthy. Never boyscout. |
| Frontend visual polish unrelated to the change | No. The diff stays focused. |

## Rationalization Prevention

| Excuse | Rule |
|--------|------|
| "The task is done, codehealth passed, we're good." | Boyscouting runs *before* codehealth. If you skipped it, the branch left value on the floor. |
| "Cleaning that up would inflate the diff." | Then it's spec-worthy, not skip-worthy. Write the spec. |
| "I didn't introduce that smell — it was already there." | Irrelevant. If you walked through it, you own it for this pass. |
| "Refactoring it would break unrelated tests." | Those tests are locking in implementation. Either fix them (in scope) or spec the refactor (out of scope). Don't pretend the smell isn't there. |
| "It's just a comment — leaving it is harmless." | Stale comments mislead readers. Delete, don't update (see [[feedback_no_comments]]). |
| "I might break something I don't understand." | Then it's not in your affected area. Leave it. |
| "This file is at 9.8 already, close enough." | Per [[codescene-refactoring]] and `easi-codehealth`: below 10.0 is below 10.0. If it's in the affected area, fix it. |
| "The duplication is only two call sites — Rule of Three says wait." | Correct. Leave it. Note it for the next time you touch the area. |
| "I should also clean up this other module while I'm here." | Out of area. Stop. Open a separate branch or a spec. |
| "The change feels small enough to skip the spec." | If you're rationalizing past the "spec-worthy" bucket, write the spec. Reviewers' confusion costs more than your saved minute. |

## Boundary with Other Skills

| Skill | Relationship |
|-------|--------------|
| [[easi-test-driven-development]] | Boyscouting is an expanded REFACTOR step — but bounded to the affected area, not the file. |
| [[easi-codehealth]] | Codehealth verifies; boyscouting cleans. Run boyscouting first so codehealth has the final state to score. |
| [[easi-spec-driven-development]] | Any finding too large for an in-branch fix becomes a spec. Boyscouting must not subvert the spec gate. |
| [[easi-domain-driven-design]] | Use as the lens for "is this design wrong?" — primitive obsession, anaemic structs, leaked domain rules. |
| [[easi-architecture-canvas]], [[easi-frontend-data]], [[easi-frontend-styling]], [[easi-go-backend-patterns]], [[easi-api-standards]] | Use as the per-area conformance checklists when walking files in those areas. |

## Hard Gates

- Affected area is written down before the walk.
- Every finding is sorted into in-branch / spec-worthy / out-of-area, with a one-line reason.
- All in-branch fixes have associated test runs.
- No behavior change in any boyscout fix (behavior change = restart from RED).
- Hand-off to `easi-codehealth` happens only after the walk is complete.

## Output

A short boyscout report appended to the implementation summary:

```
Affected area: <list of files / packages / feature folders>
Findings:
  - <file:area>: <smell> -> in-branch fix (done)
  - <file:area>: <smell> -> spec-worthy (spec NNN drafted / not yet drafted)
  - <file:area>: <smell> -> out of area (left)
Tests re-run: <commands>
Handing off to easi-codehealth.
```

If the report has zero findings, that is a valid outcome — but say so explicitly. Silent skipping is not.
