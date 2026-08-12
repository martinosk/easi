# 193 — Move Journey Target Among Realisations

> **Status:** done
> **Depends on:** 182 (Capability Journey)

---

## Problem Statement

When planning a `move` journey, the target application cannot be selected if it already realises the capability. Spec 182 rule 3 makes the capability's current realisations the implicit sources of a move, and rule 4 forbids the target being among the sources; the capture form implements both literally — it pins the (hidden) source list to *all* current realisations and filters those out of the target options. The result: every application already realising the capability is unselectable as the move target.

This inverts the most common move scenario. A move relocates the capability to another domain/parent under a new name — typically *staying on the application that already runs it* ("Invoicing" moves to "Group functions" as "Freight invoicing", still realised by SAP S/4). The apps most likely to be the move target are exactly the ones the form excludes.

The fix is in how a move's implicit sources are derived, not in the target-among-sources invariant itself: a move's implicit sources are the current realisations *excluding the chosen target*.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain / Enterprise Architect** | Plan a move journey whose target application already realises the capability — including the case where it is the sole realiser. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Move journey target selection

  Scenario: The move target may already realise the capability
    Given capability "Invoicing" is realised by "SAP S/4" and "Legacy Invoice"
    When the architect plans a move journey and opens the target application selector
    Then "SAP S/4" and "Legacy Invoice" are selectable
    And capturing the move targeting "SAP S/4" succeeds with "Legacy Invoice" as the sole implicit source

  Scenario: The sole realiser is the move target
    Given capability "Invoicing" is realised only by "SAP S/4"
    When the architect captures a move journey targeting "SAP S/4"
    Then the capture succeeds with no source applications

  Scenario: Switching the move target recomputes the implicit sources
    Given capability "Invoicing" is realised by "SAP S/4" and "Legacy Invoice"
    And the architect has selected "SAP S/4" as the move target
    When the architect changes the target to "Legacy Invoice"
    Then the captured journey lists "SAP S/4" as the implicit source

  Scenario: Non-move kinds keep the existing exclusion
    Given the architect is capturing a migration, consolidation, or carve-out
    When an application is selected as a source
    Then that application is not offered as a target
    And a capture whose target is among the sources is rejected
```

---

## Business Rules & Invariants

1. **Move target is unrestricted by realisation** — the target application selector for a `move` offers the full application catalog, including the capability's current realisers.
2. **Implicit sources of a move exclude the target** — a move's source applications are the capability's current realisations minus the chosen target application, recomputed whenever the target changes. (Amends 182 rule 3's derivation; a sole realiser chosen as target yields zero sources, which move's 0..n cardinality permits.)
3. **Target-among-sources stays invariant** — 182 rule 4 is unchanged for every kind, at every layer (form validation, backend aggregate). This spec changes what a move *sends* as sources, never what is *accepted*.

---

## Acceptance Criteria

- [x] For kind `move`, the target application selector offers every application in the catalog, including all current realisers
- [x] The captured move request carries `fromComponentIds` = current realisations minus the selected target; changing the target before submit recomputes the list
- [x] A move targeting the capability's sole realiser is captured with an empty source list
- [x] For `migration`, `consolidation`, and `carve-out`, selected sources remain excluded from the target options and target-among-sources captures remain rejected
- [x] Every BDD scenario has at least one corresponding test; the source-derivation rule has a unit test
- [x] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

Frontend-only. The journey capture form derives a move's implicit sources from the realisations prop minus the selected target, and stops filtering the move target options by the hidden source list. No API, domain model, persistence, or cross-context change: the backend aggregate's target-among-sources rejection is untouched and remains the backstop, and the request shape (`fromComponentIds`, `toComponentId`) is unchanged.

---

## Design Decisions

1. **Derive move sources as realisations-minus-target rather than exempting `move` from the invariant** — the target-among-sources rule stays intact in the form schema, the aggregate, and source-change operations, so no backend change, no event or upcaster work, and no existing journey is reinterpreted. Alternative rejected: relaxing the invariant for `move` at every layer — it touches event-sourced domain semantics to solve a UI affordance, and would let a journey record claim it moves *from* the very application it moves *to*.
2. **Full-catalog target options for a move** — a move's sources are implicit and the source field is hidden for that kind, so filtering the target list by an invisible list is inexplicable to the user. Non-move kinds keep the exclusion because there the sources are visible and user-chosen.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Sole realiser as target yields an empty source list | The journey's transition line shows a target with no sources | Accurate for a move — the story is the domain/parent/name change, not an app transition; the kind description already says realisations are implicit |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (none relevant — frontend-only change)
- [x] API documentation updated (no API change)
- [x] User sign-off
