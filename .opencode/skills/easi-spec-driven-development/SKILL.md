---
name: easi-spec-driven-development
description: "Spec-driven development workflow for EASI. Use when starting new features or behavior-changing work. Defines how to find, read, create, and maintain specs in /specs."
compatibility: opencode
---

# EASI Spec-Driven Development

Non-trivial features and behavior changes in EASI are anchored to a spec in `/specs`. The spec is the hand-off artifact from planning into implementation — it must be complete and approved before any code is written.

The spec maps directly to the orchestrator's three-phase workflow:

| Orchestrator Phase | What happens in EASI |
|--------------------|----------------------|
| **Phase 1 — Research** | Explore the codebase, identify affected contexts, find similar patterns, understand scope |
| **Phase 2 — Plan** | Write the spec (collaboration loop), reach consistency gate, check `Specification ready` |
| **Phase 3 — Implement** | Execute against each acceptance criterion, TDD, inline reviews, rename to `_done` |

The `Specification ready` checkbox is the **Phase 2 human gate**. Nothing is implemented until a human has reviewed the spec and that box is checked.

---

## When to Use This Skill

| Scenario | Use spec workflow? |
|----------|-------------------|
| New feature or user-facing behavior | Yes |
| Behavior change to an existing feature | Yes |
| Bug fix that reveals missing or wrong behavior | Yes — spec the correct behavior first |
| Bug fix with a clear, unambiguous reproduction step | No |
| Refactor with no behavior change | No |
| Typo, copy change, or trivial config tweak | No |
| Spike or investigation | No |

---

## Spec Lifecycle

Specs move through states reflected in the filename suffix:

| Suffix | Meaning |
|--------|---------|
| `_pending.md` | Spec written, not yet in active development |
| `_ongoing.md` | In active development |
| `_done.md` | Fully implemented and signed off |
| `_superseded.md` | Replaced by a later spec — leave in place, do not delete |

Always rename the file when status changes.

---

## Naming Convention

```
{NNN}_{ShortDescription}_{status}.md
```

- Three-digit sequential number
- CamelCase short description
- Status suffix

Examples: `126_EditGrants_AccessDelegation_done.md`, `132_ValueStreams_ongoing.md`

---

## Required Checklist

Every spec must contain:

```markdown
## Checklist
- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
```

Check each item only when genuinely complete. Do not batch-check at the end.

---

## Phase 1 — Research

Before writing a single line of the spec, understand the system:

1. Find related specs in `/specs` — read any `_ongoing` or `_done` specs that touch the same bounded context
2. Identify the affected bounded context(s) in `backend/internal/` and `frontend/src/features/`
3. Locate similar patterns — find an analogous aggregate, handler, or component to use as reference
4. Assess cross-context impact — which events or read models would be affected?
5. Produce a short research summary (can be inline notes or a `memory/` file for larger features)

**Phase 1 human gate:** The human reviews research findings before spec writing begins. For large or architecturally significant features, produce a design document at `docs/specs/{feature-name}.md` covering problem statement, proposed approach, alternatives, and key decisions. Human approves the design doc before Phase 2 starts.

---

## Phase 2 — Writing the Spec (Collaboration Loop)

Write the spec before touching any code. Use `specs/001_SpecTemplate_pending.md` as the starting point.

### Artifact Order

Produce the spec sections in this order — each informs the next:

1. **Problem Statement** — Why this change? What user/system pain does it address?
2. **User-Facing Behavior (BDD Scenarios)** — Observable behavior as Gherkin scenarios. These become acceptance tests.
3. **Business Rules & Invariants** — Numbered rules. Each is a potential test case.
4. **Acceptance Criteria** — Measurable pass/fail conditions derived from the scenarios and rules.
5. **Architecture** — Aggregate design, domain events, API endpoints, HATEOAS, migrations, frontend, cross-context integration.
6. **Design Decisions** — Numbered decisions with rationale and alternatives rejected.

### Collaboration Loop

1. **Draft** — Human writes the first version (or agent drafts from research findings)
2. **Critique** — Agent categorizes issues: gaps, ambiguities, conflicts, scope violations. Reference the specific text for each issue.
3. **Decide** — Human accepts/rejects/modifies. Document rejected suggestions with a one-line reason.
4. **Refine** — Agent updates the spec, preserving the human's language where possible
5. **Repeat** — Max 2 iterations. If unresolved after 2, escalate — do not keep iterating.

### Scope Check

A spec is too broad if:
- More than 3 components are affected
- Multiple independent behaviors are described
- The change cannot be deployed and validated independently

When scope is too broad, propose a split into vertical slices. Each slice gets its own numbered spec. Human approves the split before continuing.

A spec must never contain "future work", "out of scope", or "nice to have" sections. If something is worth doing, it warrants its own numbered spec. If it's not worth a spec, it doesn't belong in writing at all.

### Consistency Gate (before checking `Specification ready`)

- [ ] Intent is unambiguous — two developers would interpret it the same way
- [ ] Every behavior in the problem statement has at least one BDD scenario
- [ ] Every BDD scenario has at least one acceptance criterion
- [ ] Architecture is constrained to what the intent requires — no over-engineering
- [ ] Concepts are named consistently throughout all sections
- [ ] No section contradicts another
- [ ] Scope covers one vertical slice

This gate is a **hard stop**. Do not check `Specification ready` until all items pass.

**Phase 2 human gate:** Human reviews the full spec. Once they approve, check `- [x] Specification ready` and rename the file to `_pending.md` (or directly to `_ongoing.md` if implementation starts immediately). This is the primary review artifact — a correct spec prevents entire sessions of wrong code.

---

## Phase 3 — Implementation

### Before implementing

1. Find the spec in `/specs`
2. Confirm `- [x] Specification ready` is checked — if not, stop and ask the human
3. Read the full spec — business rules, invariants, cross-context integration, design decisions
4. Check all declared dependencies are `_done`
5. Rename the file to `_ongoing`

### While implementing

6. Work through each acceptance criterion in order
7. Follow TDD: write a failing test first, make it pass, then refactor (RED → GREEN → REFACTOR)
8. After each discrete unit of work, run the inline review checkpoint:
   - Changed Go code → `easi-go-backend-patterns`, `easi-backend-testing`
   - Changed API handlers → `easi-api-standards`
   - Changed frontend TypeScript/React → `easi-frontend-patterns`
   - Changed database migrations → `easi-database-migrations`
   - All changes → `structure-review` as baseline
9. Check off each checklist item as it is genuinely completed — do not batch at the end
10. Update the spec when you discover gaps or new invariants during implementation

### When done

11. Verify every checklist item is genuinely checked
12. Run final `/code-review --changed` on all modified files
13. Rename the file to `_done`
14. Update any specs that reference this one

---

## Key Rules

1. **No implementation without a spec** — for covered scenarios (see When to Use)
2. **No implementation when `Specification ready` is unchecked** — this is the Phase 2 gate
3. **Rename the file at each status change**: `pending` → `ongoing` → `done`
4. **BDD scenarios are contracts** — every scenario must have a corresponding test; no scenario exists without a test
5. **Treat business rule invariants as test cases** — each numbered rule should drive at least one unit test
6. **Acceptance criteria are the definition of done** — implementation is not complete until every criterion passes
7. **Update the spec during implementation** — discoveries belong in the spec, not just in code
8. **No future or out-of-scope content** — create a new numbered spec instead
9. **Categorize all critiques** — gap / ambiguity / conflict / scope violation, with reference to specific text
10. **Architecture detail is mandatory** when introducing: a new bounded context, new API endpoints, cross-context events, DB migrations, permission changes, or new frontend routes

---

## Spec Structure Reference

See `specs/001_SpecTemplate_pending.md` for the canonical template.

**Required for all specs:** Problem Statement, User-Facing Behavior (BDD), Acceptance Criteria, Checklist

**Required when the change introduces new domain concepts or cross-context integration:** Architecture section (Ownership, Domain Model, API Surface, Persistence, Frontend, Cross-Context Integration)

**Always include:** Design Decisions (numbered, with rationale), Trade-offs
