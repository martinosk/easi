---
name: easi-spec-driven-development
description: "Spec-driven development workflow for EASI. Use when starting new features or behavior-changing work. Defines how to find, read, create, and maintain specs in /specs."
compatibility: opencode
---

# EASI Spec-Driven Development

Non-trivial features and behavior changes in EASI are anchored to a spec in `/specs`. The spec is the source of truth for what gets built and whether it's done.

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

For covered scenarios, if a spec already exists: find it, check it's ready, then implement. If no spec exists: write one before touching code.

**Hard gate:** If `- [ ] Specification ready` is unchecked, stop. Ask the user to review the spec before writing any implementation code.

## Spec Lifecycle

Specs move through states reflected in the filename suffix:

| Suffix | Meaning |
|--------|---------|
| `_pending.md` | Not yet implemented |
| `_ongoing.md` | In active development |
| `_done.md` | Fully implemented and signed off |
| `_superseded.md` | Replaced by a later spec — leave in place |

Always rename the file when status changes.

## Naming Convention

```
{NNN}_{ShortDescription}_{status}.md
```

- Three-digit sequential number
- CamelCase short description
- Status suffix

Examples: `126_EditGrants_AccessDelegation_done.md`, `132_ValueStreams_ongoing.md`

## Required Checklist

Every spec must contain:

```markdown
## Checklist
- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] User sign-off
```

Check each item only when genuinely complete.

## Creating a New Spec

Write the spec before touching any code. Use a collaboration loop to ensure it's complete before marking `Specification ready`.

### Collaboration Loop

1. **Draft** — Human writes the first version
2. **Critique** — Agent categorizes issues: gaps, ambiguities, conflicts, scope violations. Reference the specific text for each issue.
3. **Decide** — Human accepts/rejects/modifies. Document rejected suggestions with a one-line reason.
4. **Refine** — Agent updates the spec, preserving the human's language where possible
5. **Repeat** — Max 2 iterations. If unresolved after 2, escalate to the human — do not keep iterating.

### Scope Check

A spec is too broad if:
- More than 3 components are affected
- Multiple independent behaviors are described
- The change can't be deployed and validated independently

When scope is too broad, propose a split into vertical slices. Each slice gets its own numbered spec. Human approves the split before continuing.

### Consistency Gate (before checking `Specification ready`)

- Intent is unambiguous — two developers would interpret it the same way
- Every behavior has at least one acceptance criterion or BDD scenario
- Concepts are named consistently throughout
- No section contradicts another
- Scope covers one vertical slice

## Spec Structure

**Required for all specs:** Description/Problem Statement, Checklist

**Recommended for feature specs:** User Personas, Core Concepts, Business Rules & Invariants (numbered), User Stories/Acceptance Criteria, Bounded Context Considerations, Design Decisions (numbered, with rationale), Architecture, Trade-offs

**Include full architecture detail** when introducing: a new bounded context or aggregate, new API endpoints, cross-context event subscriptions, DB migrations, permission changes, or new frontend routes.

## The Workflow

### Before implementing

1. Find the spec in `/specs`
2. Check `- [x] Specification ready` — if unchecked, stop and ask the user
3. Read the full spec — business rules, invariants, cross-context integration, design decisions
4. Check declared dependencies are `_done`

### While implementing

5. Rename to `_ongoing`
6. Implement against each acceptance criterion
7. Check off checklist items as completed — don't batch at the end
8. Update the spec when you discover gaps or new invariants

### When done

9. Verify every checklist item is genuinely checked
10. Rename to `_done`
11. Update any specs that reference this one

## Key Rules

1. No implementation without a spec — for covered scenarios (see When to Use)
2. No implementation when `Specification ready` is unchecked
3. Rename the file at each status change: `pending` → `ongoing` → `done`
4. Treat business rule invariants as test cases
5. Acceptance criteria are the definition of done
6. Update the spec during implementation — discoveries belong in the spec
7. Don't skip the cross-context integration section
8. Categorize all critiques (gap / ambiguity / conflict / scope violation) and reference specific text
