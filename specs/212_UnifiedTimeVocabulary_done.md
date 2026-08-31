# 212 — One TIME Vocabulary

> **Status:** done
> **Depends on:** [210_RelocateEnterpriseArchitectureIntoDirection](210_RelocateEnterpriseArchitectureIntoDirection_done.md) — design: [docs/specs/enterprise-capability.md](../docs/specs/enterprise-capability.md)
> **Roadmap alignment:** SD1 / H1-1

---

## Problem Statement

TIME exists twice: `TimeGrade` (recorded assessments) and `TimeClassification` (computed suggestions) — two Go types for the same four business values, the last remnant of coverage smell C2. The suggestion also lives apart from the assessment it advises: an architect recording a TIME grade for a capability-application pair never sees what the fit data suggests for that exact pair, because the suggestion is only visible on a separate analysis tab.

That tab is on the enterprise architecture page, which spec 213 deletes. The suggestion never depended on the enterprise capability — it is computed per `(capability, component)` realisation from fit gaps — so it needs no rework, only a home that survives. Putting it beside the assessment is that home, and it is where the advice was always useful.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain / Enterprise Architect** | See the computed suggestion, with its confidence, while recording a TIME assessment |
| **Engineer / Product Manager** | Scan which realisations the fit data flags, to know where change is likely |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Suggestion and assessment speak one TIME language

  Scenario: Suggestion shown while assessing
    Given fit data yields a "Migrate" suggestion with medium confidence for a realisation
    When an architect opens the TIME assessment for that capability-application pair
    Then the suggestion and its confidence are shown beside the grade choices

  Scenario: The suggestion never fills in the answer
    Given a realisation with a "Migrate" suggestion
    When an architect opens the assessment without a recorded grade
    Then no grade is pre-selected

  Scenario: Recorded grade and suggestion can disagree
    Given a realisation suggested as "Migrate"
    When an architect records "Tolerate"
    Then the assessment reads "Tolerate" and still shows the "Migrate" suggestion

  Scenario: No suggestion without data
    Given a realisation lacking technical or functional fit data
    Then no suggested grade is shown, as before

  Scenario: Suggestions survive as a list
    When an architect opens the TIME assessments overview
    Then realisations carrying a suggestion show it, computed from the same inputs, thresholds and confidence rules as before
```

---

## Business Rules & Invariants

1. **One value object** — `TimeGrade` (Invest, Tolerate, Migrate, Eliminate) is the only TIME type; `TimeClassification` is deleted.
2. **Suggestion is advice** — a computed `TimeGrade` plus confidence; recording an assessment remains a free human choice, never defaulted from the suggestion.
3. **Unchanged computation** — gap inputs, the 1.5 threshold, and confidence rules are preserved exactly.
4. **One set of caches** — suggestion queries read Architecture Direction's `realization_cache` and `capability_node_cache` plus the relocated fit, importance and pillar caches; the duplicate `ea_realization_cache` and `domain_capability_metadata` are dropped.
5. **The suggestion is query-time context** — it is composed into assessment reads and never stored on the `TimeAssessment` aggregate.
6. **The collection is realisation-scoped** — `/time-assessments` returns one entry per realisation that carries a recorded grade *or* a suggested grade. An unassessed realisation has a null grade, no `assessedAt`, and no `delete` link. The pair-scoped read is unchanged: it still 404s when the pair has never been assessed.

---

## Acceptance Criteria

- [x] `TimeClassification` no longer exists; suggestion DTOs carry a `TimeGrade`
- [x] The pair-scoped TIME assessment read and its UI show the current suggestion and confidence
- [x] The TIME assessments overview shows each realisation's suggestion alongside its recorded grade
- [x] Suggestion results are value-identical to before for the same data
- [x] `ea_realization_cache` and `domain_capability_metadata`, their projectors and their backfill remnants are removed; the remaining caches serve both features
- [x] No surface for suggestions remains under `/enterprise-capabilities` or the enterprise architecture page

---

## Architecture

### Ownership

Architecture Direction, in-context.

### Domain Model

The suggestion calculator becomes an Architecture Direction domain service beside `TimeAssessment`, emitting `TimeGrade` + confidence. No aggregate changes.

### API Surface

`GET /time-suggestions` retires as a standalone analysis route; the pair-scoped time-assessment read and the `/time-assessments` collection compose the suggestion into their responses. The composition lives in a `TimeAssessmentView` over the assessment and suggestion read models, so neither read model gains a second responsibility. The `get_time_suggestions` agent tool retires; `list_time_assessments` carries the same per-realisation data and says so.

### Persistence

Drop `ea_realization_cache` and `domain_capability_metadata` with their projectors; re-point suggestion queries to `realization_cache` / `capability_node_cache` joins. The fit, importance and pillar caches stay.

### Frontend

The TIME assessment form and the assessments overview render the suggestion through the shared TIME grade component. The assessments overview is the Domain Board's capability drawer: its realising-applications list shows every direct realisation with its recorded grade or "unassessed", and the suggestion beside it. `useCapabilityAssessments` serves both from the one collection query, so `useTimeSuggestions` and the enterprise-architecture suggestion API retire with the route. The `TimeSuggestionsTab` on the enterprise architecture page is removed here rather than in 213, so the page's deletion carries no behaviour loss.

### Cross-Context Integration

None.

---

## Design Decisions

1. **Keep `TimeGrade` as the surviving name** — it is the recorded-judgement type with UI and constraints attached; the suggestion adapts to it. Alternative: a new neutral name for both (rejected: renames two surfaces instead of one, for aesthetics).
2. **Compose the suggestion into the assessment read, not the aggregate** — advice is query-time context, not state, and storing it would go stale whenever fit scores move.
3. **Fold the suggestions tab into the assessments overview rather than keeping a standalone analysis page** — the two lists cover the same realisations, and a separate page is what kept advice away from the decision. Alternative: a new standalone suggestions page (rejected: reproduces the problem the slice is fixing).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Cache consolidation with query rewrite | Regression risk in suggestion values | Acceptance criterion pins value-identical output; tests compare against fixtures from the current implementation |
| `GET /time-suggestions` retires | An agent tool and any external caller lose a route | The tool re-points to the assessments collection, which carries the same data per realisation; the route has no other consumer |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
