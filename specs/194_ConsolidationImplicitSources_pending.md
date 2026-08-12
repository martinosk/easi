# 194 — Consolidation Implicit Sources

> **Status:** pending
> **Depends on:** 182 (Capability Journey), 193 (Move Journey Target Among Realisations)

---

## Problem Statement

A consolidation journey's source applications are hand-picked, but "consolidation" means the capability's fragmented realisation collapses onto one application — all of it. If some realisers are left out, the change is a partial migration, not a consolidation; spec 182 already provides `migration` for exactly that. Hand-picking invites capture errors (a forgotten realiser silently narrows the recorded story) and blurs the line between the two kinds.

It also carries a sibling of the 193 bug: consolidating *onto* one of the existing realisers — the normal case ("three tracking tools merged onto one of them") — only works if the architect manually leaves the target out of the source list, and spec 182's "consolidation ≥ 2 from-apps" rule wrongly rejects the two-realiser case (A and B merging onto A yields a single from-app).

This slice makes consolidation sources implicit, the same derivation 193 gives moves: all current realisations, minus the target when it is one of them. Migration remains the kind for moving a chosen subset.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain / Enterprise Architect** | Capture a consolidation without hand-picking sources, including consolidating onto one of the current realisers — trusting that the journey records every realiser as a source. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Consolidation journey with implicit sources

  Scenario: Consolidating onto an existing realiser
    Given capability "Shipment tracking" is realised by "TrackIt", "CargoEye" and "Phoenix"
    When the architect captures a consolidation journey targeting "Phoenix"
    Then the capture succeeds with "TrackIt" and "CargoEye" as the implicit sources
    And no source application field was offered during capture

  Scenario: Consolidating onto a new application
    Given capability "Shipment tracking" is realised by "TrackIt" and "CargoEye"
    When the architect captures a consolidation journey targeting the catalog application "Phoenix"
    Then the capture succeeds with "TrackIt" and "CargoEye" as the implicit sources

  Scenario: Two realisers merging onto one of them
    Given capability "Shipment tracking" is realised by "TrackIt" and "Phoenix"
    When the architect captures a consolidation journey targeting "Phoenix"
    Then the capture succeeds with "TrackIt" as the sole implicit source

  Scenario: Switching the consolidation target recomputes the implicit sources
    Given capability "Shipment tracking" is realised by "TrackIt" and "Phoenix"
    And the architect has selected "Phoenix" as the consolidation target
    When the architect changes the target to "TrackIt"
    Then the captured journey lists "Phoenix" as the implicit source

  Scenario: A consolidation needs something to consolidate
    Given capability "Shipment tracking" is realised by at most one application
    When the architect attempts to capture a consolidation journey
    Then the capture is blocked with an explanation that consolidation requires at least two current realisers

  Scenario: Sources of an active consolidation stay editable
    Given an active consolidation journey with sources "TrackIt" and "CargoEye"
    When the architect changes the sources to "TrackIt" alone
    Then the change is accepted
    And changing the sources to none is rejected

  Scenario: Migration and carve-out keep explicit sources
    Given the architect is capturing a migration or carve-out
    Then the source application field is offered and its 182 cardinality rules apply unchanged
```

---

## Business Rules & Invariants

1. **Consolidation sources are implicit at capture** — the capability's current realisations minus the chosen target, recomputed whenever the target changes. No source field is shown for the kind. (Amends 182 rule 3 for `consolidation`; mirrors 193 rule 2.)
2. **Consolidation requires ≥ 2 current realisers** — with fewer than two realisations there is nothing to consolidate; capture is blocked at the form with a clear explanation. A subset-merge among many realisers is a `migration`.
3. **Source cardinality for consolidation becomes ≥ 1** — at every validating layer. The merge involves the sources plus the target, so one source already means two applications; 182's "≥ 2 from-apps" wrongly rejects two realisers merging onto one of them.
4. **Consolidation target is unrestricted by realisation** — the target selector offers the full application catalog, including current realisers. (Mirrors 193 rule 1.)
5. **Sources of an active consolidation remain editable** — the implicit derivation is a capture-time convenience, not a lifetime constraint; the existing change-sources operation applies unchanged, under the ≥ 1 floor.
6. **Target-among-sources stays invariant** — 182 rule 4 unchanged for every kind, at every layer.

---

## Acceptance Criteria

- [ ] For kind `consolidation`, the capture form shows no source field, offers the full catalog as targets, and submits `fromComponentIds` = current realisations minus the selected target, recomputed on target change
- [ ] A consolidation capture on a capability with fewer than two current realisations is blocked in the form with an explanatory message
- [ ] A consolidation with exactly one source application is accepted end-to-end (form validation, API, aggregate); zero sources remains rejected
- [ ] Changing sources on an active consolidation journey accepts one source and rejects zero
- [ ] `migration` and `carve-out` capture behavior, including source field and cardinality rules, is unchanged
- [ ] API documentation reflects the relaxed consolidation cardinality
- [ ] Every BDD scenario has at least one corresponding test; rules 1–3 have unit tests
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

Frontend: the journey capture form treats `consolidation` like `move` for source derivation (hidden source field, realisations-minus-target, shared with the 193 mechanism) and adds the ≥ 2-realisers gate; the form schema's consolidation minimum drops to one source. The change-sources form needs no structural change — its cardinality floor follows the shared schema helpers.

Backend: the journey-kind value object's source-count validation for consolidation drops from two to one. No other domain, API-shape, persistence, or cross-context change; the backend stays permissive about *which* sources a consolidation lists (as it is for move) — the all-realisers semantics is a capture derivation, not an aggregate invariant.

---

## Design Decisions

1. **Implicit derivation lives in the form; the backend stays permissive** — verifying "sources = current realisations minus target" in the aggregate would need a cross-context realisation lookup at capture time, and realisations drift while a journey is active, making the check stale the moment it passes. Mirrors 193 decision 1 and the move posture. Alternative rejected: handler-side verification against the realisation read model — adds coupling for no invariant the domain can keep true.
2. **Cardinality floor ≥ 1, not "≥ 2 realisers", in the backend** — the ≥ 2-realisers precondition is about capability state the `architecturedirection` context does not own; the structural residue the aggregate *can* enforce is that a merge names at least one source. The user-facing precondition is enforced where the realisation list is at hand: the capture form. Alternative rejected: keeping ≥ 2 sources — wrongly rejects the two-realiser merge.
3. **Sources stay editable after capture** — the landscape drifts while a journey is in flight, and 182's change-sources operation with its events already records such adjustments. Alternative rejected: locking or continuously re-deriving sources — re-derivation would rewrite captured history and contradict 182 rule 6's discrete-event posture.
4. **Migration is the kind for partial merges** — architects consolidating a subset of realisers onto one app record a migration; keeping consolidation total is what makes the two kinds distinct. No new kind introduced.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Backend accepts any ≥ 1 source list for consolidation | An API client can record a "consolidation" that omits realisers | The form is the only shipped capture surface and always derives the full list; the journey remains a plan, not a model mutation (182 rule 10) |
| Consolidation blocked under two realisers | An architect anticipating a second realiser cannot pre-capture the consolidation | Realisations can be added first at `Planned` level (182 decision 3 posture), after which capture proceeds |
| Structural validation no longer distinguishes consolidation from migration | The kinds differ only in source derivation and story told | That distinction is the point: total vs. partial; the kind label carries the meaning, as it already does for move |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
