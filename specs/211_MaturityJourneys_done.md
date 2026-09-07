# 211 — Maturity Journeys

> **Status:** done
> **Depends on:** [210_RelocateEnterpriseArchitectureIntoDirection](210_RelocateEnterpriseArchitectureIntoDirection_done.md), 182 (journeys), 196 (milestone order) — design: [docs/specs/enterprise-capability.md](../docs/specs/enterprise-capability.md)
> **Roadmap alignment:** SD2 / H1-1

---

## Problem Statement

Maturity ambition lives as a number on an enterprise capability, and the gap analysis derived from it is the only reason the enterprise capability still earns a surface. The number is inert: nobody owns it, nothing schedules it, and it describes a destination with no route. Meanwhile a capability's maturity rises through work — replacing a system, yes, but just as often through process change, clearer ownership, better data quality or new skills — and EASI already has the artifact for planned work on a capability, with a period, a progress state and ordered steps: the journey.

Every journey kind today is application-shaped (from-applications → to-application), which is why maturity never fitted. But `move` already proves a kind can define its own target instead. This slice adds the maturity journey, so a capability's maturity ambition becomes a plan someone owns rather than a field someone once filled in.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain / Enterprise Architect** | Plan a capability's maturity uplift — the level to reach, by when, and the steps to get there — without pretending it is an application migration |
| **Engineer / Product Manager** | See what a capability's maturity plan involves and how far it has got |
| **Stakeholder** | Read which capabilities are being deliberately matured, to what level, and by when |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Maturity as a journey

  Scenario: Planning a maturity journey
    Given capability "Hazardous Booking" sits at maturity 30 (Custom Build)
    When an architect plans a maturity journey targeting 65 (Product) by Q4 2027
    Then the journey is planned against that capability with target maturity 65
    And no from-applications or to-application were required

  Scenario: The uplift need not be technical
    Given a planned maturity journey
    When an architect adds the milestones "Name a single owner", "Agree the data contract" and "Retire the manual reconciliation"
    Then all three are recorded as milestones in the order given

  Scenario: Maturity target must be above today's maturity
    Given capability "Hazardous Booking" sits at maturity 30
    When an architect plans a maturity journey targeting 25
    Then the plan is rejected, because a maturity journey must raise maturity

  Scenario: Application fields are refused on a maturity journey
    When a maturity journey is planned with a to-application
    Then the plan is rejected

  Scenario: A maturity journey runs the normal lifecycle
    Given a planned maturity journey
    Then it can be started, progressed, completed and abandoned exactly as other journeys

  Scenario: The gap is visible while the journey is open
    Given an in-flight maturity journey targeting 65 on a capability at maturity 30
    When a user opens the capability
    Then the journey shows the current maturity, the target, and the remaining gap

  Scenario: The gap closes as the capability matures
    Given the journey above
    When the capability's maturity is updated to 65
    Then the remaining gap reads zero

  Scenario: One maturity journey at a time
    Given a capability with an active maturity journey
    When an architect plans a second maturity journey for it
    Then the plan is rejected

  Scenario: Maturity journeys compose across the hierarchy like any other
    Given an L2 capability has an active maturity journey
    When a user opens its L1 ancestor
    Then the maturity journey appears among the sub-capability journeys
```

---

## Business Rules & Invariants

1. **`maturity` is a journey kind** — alongside migration, consolidation, carve-out and move, on the existing `CapabilityJourney` aggregate. No new aggregate.
2. **A maturity journey carries a target maturity (0–99) and no applications** — zero from-applications, no to-application; supplying either is rejected, mirroring how `move` refuses application fields.
3. **The target must exceed the capability's current maturity at planning time** — a journey that does not raise maturity is not a maturity journey. The check is made once, when the journey is planned.
4. **Only maturity journeys carry a target maturity** — supplying one on any other kind is rejected, mirroring `ErrJourneyMoveFieldsOnNonMove`.
5. **The lifecycle is unchanged** — planned → in-flight → done / abandoned, with progress, target period, note and milestones exactly as other kinds.
6. **One active maturity journey per capability** — a capability may run a maturity journey alongside an application journey, but not two maturity journeys.
7. **The gap is derived, never stored** — remaining gap is target maturity minus the capability's current maturity, computed at read time from the capability cache, so it tracks maturity changes without the journey being touched.
8. **Milestones carry the non-technical work** — a maturity journey imposes no structure on what a milestone may say.

---

## Acceptance Criteria

- [x] A maturity journey can be planned, started, progressed, completed and abandoned against a domain capability, carrying a target maturity and no applications
- [x] Planning is rejected when the target does not exceed current maturity, when application fields are supplied, or when an active maturity journey already exists for that capability
- [x] A target maturity supplied on a non-maturity kind is rejected
- [x] The journey read model exposes target maturity, current maturity and the derived remaining gap
- [x] Maturity journeys appear in the journey list, the timeline view and the sub-capability journey composition alongside other kinds
- [x] The capability drawer's journey section renders a maturity journey with its target, gap and milestones

---

## Architecture

### Ownership

Architecture Direction, entirely in-context.

### Domain Model

`JourneyKind` gains `maturity`, with `ValidateSourceCount` requiring zero from-applications. `CapabilityJourney` gains a `targetMaturity` value object, validated on `PlanCapabilityJourney` beside the existing move-field validation: present exactly for the maturity kind, above the capability's current maturity, within 0–99. `JourneyPlanned` gains a `targetMaturity` field; no other event changes. `TargetMaturity` moves out of the retiring enterprise-capability vocabulary and becomes the journey's own value object, keeping its 0–99 range and its Genesis / Custom Build / Product / Commodity sections.

### API Surface

`POST /capabilities/{id}/journey` accepts `kind: "maturity"` with `targetMaturity`; the journey DTO gains a `maturity` object carrying `targetMaturity`, `currentMaturity` and `maturityGap`. `GET /capabilities/{id}/journey` returns `journeys` (0–2 active journeys, one per track) instead of a single `journey`. No new routes.

### Persistence

`capability_journeys` gains a nullable `target_maturity` column, and the single-active partial unique index is re-scoped to `(tenant_id, capability_id, kind = 'maturity')` so the two tracks are independent. Current maturity is read from the existing `capability_node_cache`; the gap is computed in the query, so no backfill is needed — pre-existing journeys have no target and no gap.

### Frontend

`CaptureJourneyForm` offers the maturity kind and swaps its application pickers for a maturity target field; the journey panel renders target, current and gap for every active journey. Journey list, timeline and sub-capability composition pick the kind up without change.

### Cross-Context Integration

None. Current maturity already arrives in Architecture Direction through `capability_node_cache`, fed by published Capability Mapping events.

---

## Design Decisions

1. **A journey kind, not a new aggregate** — the lifecycle, milestones, progress, period and hierarchy composition are all wanted verbatim, and the `move` kind already establishes that a kind may replace the application fields with its own target. Alternative — a separate `MaturityPlan` aggregate (rejected: it would duplicate the entire journey lifecycle and force every journey surface to merge two sources).
2. **Target maturity lives on the journey, not on the capability** — the ambition and the plan to reach it are one thing; splitting them re-creates the inert number this slice exists to remove. Alternative — target maturity as capability metadata (rejected in the design doc, decision 1: catalog-wide gap analysis is better served by TIME suggestions and fit gaps, which are computed from evidence rather than typed in).
3. **Gap is derived at read time** — storing it would go stale the moment a capability's maturity changes, and the capability cache is already local to the context.
4. **The target must raise maturity** — enforcing it at planning time keeps the concept honest without freezing the journey if the capability later regresses; a regression makes the gap grow, which is information, not an error.

5. **A journey occupies a *track*, and one-active-per-capability becomes one-active-per-track** (decided during implementation) — rule 6 lets a maturity journey run beside an application journey, which amends spec 182 rule 1. `JourneyKind.TrackKinds()` names the kinds sharing a track (`maturity` alone; the four application kinds together); the handler's active-journey lookup and the partial unique index are both scoped by it. Alternative — a stored `track` column — rejected: the track is derivable from the kind, and storing it would need a projector write and a backfill for no gain.

6. **`GET /capabilities/{id}/journey` returns `journeys`, not `journey`** (decided during implementation) — with two tracks the endpoint can no longer name one journey. Alternatives rejected: a second `maturity-journey` route (fragments the concept D5 exists to unify) and a kind-specific second field on the envelope (special-cases one kind in the wire shape). The drawer renders each returned journey with the existing journey card.

7. **Two capture rels, not one** (decided during implementation) — `x-capture` is emitted when the application track is free and `x-capture-maturity` when the maturity track is, so the client derives the offered kinds from links rather than re-deriving the rule. `x-change-sources` is withheld from maturity journeys, which have no sources to change.

8. **The gap is a nested `maturity` object, mirroring `move`** (decided during implementation) — the DTO already models kind-specific fields as an optional nested object; a third set of flat nullable columns would read as unconditional fields that are usually null. Section names (Genesis / Custom Build / Product / Commodity) are deliberately not returned: the maturity scale is MetaModel-owned and Architecture Direction should not publish a second copy of it.

9. **Current maturity is resolved by the handler and validated by the aggregate** (decided during implementation) — `PlanCapabilityJourney` takes `CurrentMaturity` as a fact so rule 3 stays a domain invariant rather than a handler check; the handler reads it from `capability_node_cache` only for the maturity kind.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Maturity ambition exists only where a journey exists | No catalog-wide "where should we invest" list from target maturity alone | Investment candidates come from TIME suggestions and strategic fit, which are computed from evidence rather than declared; spec 212 puts the suggestion where it is acted on |
| One more journey kind on a shared form | `CaptureJourneyForm` grows a second kind-conditional branch after `move` | The branch is the same shape as move's; if a third appears, the form splits per kind |
| Target maturity is fixed once planned | Re-aiming a journey means abandoning and re-planning | Matches how the other kinds treat their targets; editing the target is a later slice if the need appears |
| Two active journeys per capability (rule 6) | The Domain Board card still shows one journey, so a maturity journey can be hidden behind an application one | The board card is the five-second answer about the application landscape; `selectBoardJourney` deterministically prefers the application journey, and the drawer, history and bulk query all show both |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
