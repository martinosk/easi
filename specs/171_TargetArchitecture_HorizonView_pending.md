# 171 — Target Architecture by Horizon

> **Status:** pending
> **Depends on:** [166 — Logical Capability Rename](166_LogicalCapability_Rename_pending.md), [167 — Direction on a Logical Capability](167_Direction_Aggregate_Capture_pending.md), [169 — Standard Application Designation](169_StandardAppDesignation_pending.md), [170 — Direction Map](170_DirectionMap_Canvas_pending.md)
> **Conceptual basis:** [`mockups/architecture-direction-model.md`](../mockups/architecture-direction-model.md), [`mockups/architecture-direction.html`](../mockups/architecture-direction.html)

---

## Problem Statement

The Direction Map (spec 170) shows movement: what's proposed to move, where, and at what status. The Application Portfolio (spec 169) shows the standard-app picture. Neither answers the question a domain owner or product manager actually asks during an investment decision: *"What does my domain look like at the target — and what does it look like a step or two before we get there?"*

This slice introduces the **Target Architecture by Horizon** view — a synthesis surface that composes Directions, Standard App designations, and the underlying physical capability map into one picture, scrubbable across three horizons: **Now** (current state), **Next** (after agreed Directions intended for the near term land), **Later** (the full target after every agreed Direction realises).

Each capability on the surface is classified relative to the chosen horizon: *native* to the domain (always was here), *inbound* (arriving from another domain via a consolidate Direction), *decomposed-in* (arriving from a decompose Direction), or *transitional* (still here but leaving in a later horizon). Where a Standard App designation is agreed for a capability's logical, the application is shown inline.

After this slice ships, anyone at DFDS can pick a horizon and see what their domain — or the whole organisation — looks like at that point. The path from current to target becomes legible, without dates and without a project plan.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain Owner** | See what their domain looks like Now / Next / Later; spot capabilities arriving, leaving, or changing application; plan accordingly. |
| **Product Manager / Engineer** | Answer "where will my capability live a year from now, and what app will it use?" without asking an architect. |
| **Enterprise Architect** | Use the synthesis surface to validate that the agreed Directions and designations actually compose into a coherent target — gaps and conflicts surface here. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Target Architecture by Horizon

  Scenario: Pick a horizon and see the implied landscape
    Given the system has agreed Directions and Standard App designations across multiple horizons
    When I open the Target Architecture surface
    And I select a horizon
    Then the surface renders the per-domain landscape implied by every agreed Direction at or before that horizon
    And switching between horizons recomputes the landscape in under a second

  Scenario: Capabilities are classified relative to the chosen horizon
    Given a horizon is selected
    Then each capability on the surface is classified as one of: native, inbound, decomposed-in, transitional
    And the classifications are visually distinguishable

  Scenario: Standard applications appear inline where agreed
    Given a Logical Capability has an agreed Standard App designation
    When the Logical's physical capabilities are rendered at a horizon at or after the designation's horizon
    Then the agreed application appears inline on each rendered capability

  Scenario: A domain owner orients on their domain
    Given I am a domain owner viewing Target Architecture
    When I focus on my domain
    Then I see the capabilities that will be in my domain at the chosen horizon
    And I see which are arriving (inbound or decomposed-in) and which are leaving (transitional)

  Scenario: The view is read-only
    Given I am viewing Target Architecture
    Then the surface is purely informational
    And edits to Directions, designations, or capabilities happen on their respective surfaces, not here

  Scenario: Draft and proposed Directions do not appear
    Given the system has Directions in draft and proposed status
    When I view Target Architecture
    Then those Directions do not influence the rendered landscape at any horizon
    And only agreed Directions contribute

  Scenario: A capability with no Direction stays where it is at every horizon
    Given a Logical Capability has no agreed Direction
    Then its physical capabilities appear native in their existing domains at every horizon
```

---

## Business Rules & Invariants

1. **Target Architecture is read-only.** It composes data from other contexts; it owns no aggregates.
2. **The horizon enum has three values:** `now`, `next`, `later`. There are no dates. The horizon is a categorical commitment about ordering, not about timing.
3. **Only `agreed` Directions contribute to the projected landscape.** `Draft` and `proposed` Directions, however informative, do not shape what the surface renders. The surface is the answer to "what has the group actually decided," not "what is the group considering."
4. **A Direction's effect manifests at and after its declared horizon.** A Direction with horizon `next` does not change the `now` view; it does change the `next` and `later` views.
5. **Capability classification is computed per-horizon per-capability:** `native` (in this domain at this horizon by default), `inbound` (arriving here from a consolidate Direction realised at or before this horizon), `decomposed-in` (arriving here from a decompose Direction), `transitional` (still rendered here but leaving at a later horizon by an agreed Direction).
6. **Standard application is inline only for `agreed` designations whose horizon is at or before the chosen horizon.** Proposed and draft designations don't show an inline application.
7. **The view honours the caller's read authorisation.** Capabilities and domains the caller cannot read are not rendered.
8. **Performance budget: switching horizon recomputes in under one second** at production scale. If profiling shows this unattainable, the implementation may pre-materialise per-horizon snapshots (settled during implementation).
9. **Stale references render visibly.** If a Direction or a designation references a deleted entity, the affected cell renders with the stale indicator carried from prior specs.
10. **No editing affordances appear on this surface.** Click-through to a capability, Direction, or designation routes to its respective edit surface (specs 167, 169, or the existing capability detail page).

---

## Acceptance Criteria

- [ ] The Target Architecture surface renders the per-domain landscape for a selected horizon (`now` / `next` / `later`)
- [ ] Switching horizon recomputes the landscape in under one second on production-scale data, or a documented per-horizon pre-materialisation strategy is in place
- [ ] Each capability on the surface is classified as one of `native` / `inbound` / `decomposed-in` / `transitional` and the classifications are visually distinct
- [ ] Where a Logical Capability has an agreed Standard App designation at or before the chosen horizon, the application appears inline on the Logical's rendered physicals
- [ ] Draft and proposed Directions do not influence the rendered landscape at any horizon
- [ ] A non-architect product manager can answer "where will my capability live a year from now and what app will it use?" by reading the surface; no training required
- [ ] The surface is read-only; click-through navigates to the relevant edit surface (capability detail, Direction editor, designation editor)
- [ ] Stale references render with the stale indicator carried from prior specs
- [ ] The view honours read authorisation; out-of-scope domains and capabilities are not rendered
- [ ] All BDD scenarios above have at least one corresponding test
- [ ] CodeScene `pre_commit_code_health_safeguard` passes on every modified file

---

## Architecture

### Ownership

A read-side feature in `architecturedirection`, composing data from `enterprisearchitecture` (Logical Capabilities and their mappings to physicals), `capabilitymapping` (physical capabilities and domains), and `architecturedirection` itself (Directions and Standard App designations).

### Domain Model

No new aggregates. The slice introduces a read model that materialises, per horizon, the projected landscape: for each (domain, horizon) pair, the set of physical capabilities present and their classifications, with optional inline application.

The projection logic is the load-bearing piece. Given the rule set in Business Rules, the implementation can either compute on demand from the underlying read models or pre-materialise three snapshots (one per horizon). The choice is a performance trade-off settled during implementation.

### API Surface

A read endpoint per horizon (or one endpoint with a horizon parameter) returns the data needed to render the surface. The contract obligation: one request per horizon returns the full per-domain landscape including classification and inline applications. The surface honours read authorisation.

### Persistence

A read model in `architecturedirection`, populated by projections over the relevant published-language events from all three contexts. Whether per-horizon snapshots are pre-materialised or computed on demand is a profiling-driven choice during implementation.

### Frontend

A new Target Architecture surface in the architecture-direction area of the UI. Layout: domain-organised, with a horizon scrubber prominent. Each rendered capability shows its name, classification (visually), and inline application when applicable. Click-through routes to the appropriate edit surface in the existing UI (a capability's detail page, a Direction's editor, a designation's editor).

The view is the synthesis surface — its job is *answering*, not editing. Layout choices that keep the answer scannable in five seconds take precedence over layout choices that maximise data density.

### Cross-Context Integration

The projection subscribes to:
- `architecturedirection`: Direction events and StandardAppDesignation events.
- `enterprisearchitecture`: Logical Capability events and LogicalCapabilityMapping events.
- `capabilitymapping`: physical capability events and BusinessDomain events.

Subscription patterns follow the established EASI cross-context projection model.

---

## Design Decisions

1. **Three horizons, no dates.** The conceptual model commits to this and the user has explicitly rejected date-based planning. The surface is informed by what *has been agreed* and *which horizon the agreement targets*, not by a calendar. Alternative considered: surface dates from the Directions for clients who want them — rejected; would re-introduce project-management semantics the model explicitly excludes.

2. **Only `agreed` Directions contribute.** The surface is the architecture group's collective answer. Including drafts would mean the surface changes whenever an architect saves a sketch, which is the opposite of "stable, dependable, alignment-decision-ready." Alternative considered: a toggle to include proposed Directions — rejected for this slice; if useful later, lands as a separate spec.

3. **Read-only synthesis.** Editing on this surface would invite contradictions (e.g., a user editing a draft Direction from a Now view that doesn't show drafts). Routing all edits back to the canonical edit surfaces keeps the model consistent.

4. **Capability classification is computed, not stored.** The classifications (`native` / `inbound` / `decomposed-in` / `transitional`) are derivable from Directions and the existing physical capability map. Storing them would create a synchronisation problem; computing them is the cleaner pattern.

5. **Inline application only for `agreed` designations.** Proposed designations are uncertain by definition; surfacing them in a synthesis view that the user expects to be answer-grade misrepresents certainty. They live on the Application Portfolio surface where their proposal-status is visible.

6. **Performance can ship pre-materialised if needed.** Computing the landscape from raw events on every navigation may be too slow at production scale. The slice budgets for this contingency: per-horizon snapshots are an acceptable implementation choice if profiling demands it.

7. **Layout settled at implementation time.** The mockup at `mockups/architecture-direction.html` shows one validated layout (per-domain cards with capability rows). It is the starting point. Whether to use it as-is or evolve is settled during implementation, against the constraint that a non-architect can answer the daily-alignment question in five seconds without training.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Three horizons, no dates | Some users will want quarter-level resolution | Acceptable per the model's principles; if quarter-level is genuinely needed, a separate spec adds it later |
| Only agreed Directions | Drafts and proposals don't affect the synthesis view | Drafts are visible on their respective surfaces; the synthesis view is the answer surface |
| Read-only | Click-through cost per edit | One click is cheaper than the alternative (contradictory editing surfaces) |
| Classifications computed | Recomputation cost on every relevant event | Standard projection pattern; pre-materialisation is the escape valve |
| Inline app only for agreed designations | Proposed apps are invisible here | The Application Portfolio surface (spec 169) is where proposals live |
| Layout settled in implementation | Surface design isn't pre-committed | Validated mockup is the starting point; iteration is bounded |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
