# 170 — Direction Map: Physical Movement Canvas

> **Status:** pending
> **Depends on:** [166 — Logical Capability Rename](166_LogicalCapability_Rename_pending.md), [167 — Direction on a Logical Capability](167_Direction_Aggregate_Capture_pending.md)
> **Conceptual basis:** [`mockups/architecture-direction-model.md`](../mockups/architecture-direction-model.md), [`mockups/architecture-direction.html`](../mockups/architecture-direction.html)

---

## Problem Statement

After spec 167 ships, an architect can answer "is there a Direction on this Logical Capability?" by opening the Logical's detail surface. That's the lookup case. The question this slice solves is the *survey* case: "what does the proposed re-shape of the business architecture look like, all together, on one surface?"

Today the architecture group stitches that picture together by hand — slides with arrows between domains, whiteboard sketches, screenshots from the mockup. Stakeholders can't independently form the same picture, and the picture goes stale the moment a Direction advances.

This slice introduces the **Direction Map** — a single visual surface where every active Direction is rendered: source domain(s) on one side, target domain(s) on the other, the physical movement between them visible at a glance. Filter by status (Draft / Proposed / Agreed) and the surface tightens to "what the group is actively debating" or "what the group has actually agreed."

After this slice ships, an architect or stakeholder can open one tab and see the proposed re-shape of physical reality — and a domain owner can see what's *moving in* and *out* of their domain in five seconds.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect (in a working session)** | A working surface for the group's strategic conversation — the canvas everyone looks at when discussing what to consolidate or decompose next. |
| **Domain Owner** | A scannable picture of what's moving in and out of their domain so they can prepare for the conversations that involve them. |
| **Product Manager / Engineer** | A high-level orientation — "what's the overall shape of where the architecture group is taking us" — without needing to read every individual Direction. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Direction Map

  Scenario: Survey active Directions on one canvas
    Given the system has Directions in various statuses
    When I open the Direction Map
    Then I see every active Direction visualised as physical movement between domains
    And consolidate, decompose, and stay Directions are visually distinguishable

  Scenario: Filter by status to focus the picture
    Given the Direction Map is showing all active Directions
    When I filter to a specific status
    Then only Directions matching that status remain visible
    And the picture re-renders within a second

  Scenario: Click into a Direction to inspect or edit
    Given I am viewing the Direction Map
    When I select a Direction
    Then the Direction's detail surfaces in context (drawer, panel, or detail page — chosen by the implementation)
    And edits made there propagate back to the map without a page reload

  Scenario: A domain owner orients on their domain
    Given I am a domain owner viewing the Direction Map
    When I focus on my domain
    Then I can see every Direction that has my domain as a source or target
    And I can tell which way the movement is going

  Scenario: A Direction's source becomes stale
    Given a Direction references a deleted physical capability
    When I view the Direction Map
    Then the Direction renders with the stale-reference indicator carried from spec 167

  Scenario: A reader sees the map but cannot edit
    Given I have read-only access
    When I open the Direction Map
    Then I see the same picture as an architect
    And the click-into surface is read-only

  Scenario: Stay-type Directions are visible but visually quiet
    Given a Logical Capability has a stay-type Direction (group has explicitly evaluated and decided no change)
    When I view the Direction Map
    Then the stay Direction is visible as a deliberate decision
    And it does not visually compete with movement-type Directions for attention
```

---

## Business Rules & Invariants

1. **The Direction Map is read-side only.** It does not own state. Every entity it renders is owned by another aggregate (Direction, Logical Capability, Business Domain). The map is a projection.
2. **Active Directions only.** `Rejected` Directions do not appear on the map by default; if needed for retrospective viewing, that's a separate filter (decided at implementation time).
3. **Visual distinguishability is normative for the three Direction types.** A user must be able to tell at a glance whether each rendered Direction is `consolidate`, `decompose`, or `stay`. The exact visual encoding (colour, shape, stripe, etc.) is settled during implementation; the validated mockup at `mockups/architecture-direction.html` is the starting point.
4. **Status filtering is a primary affordance.** A user must be able to narrow the canvas to one status (`draft`, `proposed`, `agreed`) or any subset, and the picture must update in under a second on production data.
5. **Domain is the spatial primitive.** The map's layout is organised around business domains; movements render *between* domains. Other layouts (per-status swim lanes, per-horizon timelines) are out of this slice — they are different views and would land as separate specs.
6. **Click-through reuses spec 167's editor.** The map does not introduce a parallel editing surface for Directions. Selecting a Direction surfaces the same editor used on the Logical Capability detail page.
7. **The map respects the user's read authorisation.** A user sees only the Directions whose underlying Logical Capabilities they can read. Directions on out-of-scope Logicals are not rendered.
8. **Stale references render but do not break the map.** A Direction with a deleted source capability appears with the same stale indicator surfaced in spec 167; the map renders the rest of the picture without interruption.
9. **Stay-type Directions appear but are visually subordinate.** They are real decisions and stakeholders need to see them, but they should not compete for attention with movement Directions, which is the picture's primary job. The visual treatment is settled during implementation.
10. **Performance budget: the canvas renders the active-Direction set in under two seconds at production scale** (~1300 physical capabilities, expected dozens to low-hundreds of active Directions). If profiling during implementation reveals this is unrealistic, the slice's scope contracts to a default-filtered view (e.g. agreed-only or per-domain) before the full unfiltered view is offered.

---

## Acceptance Criteria

- [ ] Opening the Direction Map shows every active Direction visualised as movement between business domains
- [ ] The three Direction types (consolidate, decompose, stay) are visually distinguishable at a glance
- [ ] A status filter narrows the visible Directions in under a second on production-scale data
- [ ] A reader sees the same picture as an architect; click-through to a Direction surfaces the spec 167 editor in read-only mode
- [ ] An architect can edit a Direction from the map's click-through surface and see the change reflected on the map without a page reload
- [ ] Stale references render without breaking the rest of the picture
- [ ] Stay-type Directions appear, visually subordinate to movement Directions
- [ ] A domain-focused view (the user can tell what's moving in / out of any one domain) is reachable from the default surface
- [ ] The unfiltered map renders within two seconds at production scale; if implementation finds this unattainable, a default filter (agreed-only or current-user's-domain) is documented and shipped instead
- [ ] All BDD scenarios above have at least one corresponding test (canvas rendering may use snapshot tests; click-through and filtering are exercised end-to-end)
- [ ] CodeScene `pre_commit_code_health_safeguard` passes on every modified file

---

## Architecture

### Ownership

The Direction Map is a read-side feature of `architecturedirection`, with read references into `enterprisearchitecture` (Logical Capabilities) and `capabilitymapping` (physical capabilities and business domains).

### Domain Model

No new aggregates. The map is a projection that joins:
- Active Directions (status ≠ rejected, from `architecturedirection`)
- Their parent Logical Capability (from `enterprisearchitecture`)
- Their source physical capabilities and the domains those belong to (from `capabilitymapping`)
- Their target placements (where applicable)

The projection is rebuildable from the source contexts' event streams.

### API Surface

A read endpoint that returns the data backing the map — sufficient information to render every Direction, its sources, its targets, and its status. The exact shape (single rich response vs. multiple endpoints, server-side filtering vs. client-side) is settled during implementation. The contract obligation: a single request returns everything needed for the default view.

The map endpoint MUST honour the caller's read authorisation; the response excludes Directions on out-of-scope Logical Capabilities.

### Persistence

A new read model in `architecturedirection` populated by projections over Direction events plus subscribed events from the upstream contexts. The read model is rebuildable from event-store replay.

### Frontend

A new Direction Map surface in the architecture-direction area of the UI. EASI's existing canvas / dockview infrastructure is the natural reuse target; whether to use it or to build a purpose-specific surface is settled at implementation time, with the constraint that the layout must remain legible at production scale and the click-through must reuse the spec 167 Direction editor.

### Cross-Context Integration

The map projection subscribes to the published-language events of `architecturedirection`, `enterprisearchitecture`, and `capabilitymapping`. The subscription pattern follows EASI's established cross-context projection model.

---

## Design Decisions

1. **Map is a projection, not a new aggregate.** The map is a view over several other aggregates; making it a first-class aggregate would introduce a synchronisation surface that does not pay back. Read-side composition is the established pattern.

2. **Status filter as the primary affordance.** Of all the ways to narrow the picture (status, domain, horizon, type), status is the one most aligned with the "what is the group debating right now / what have we agreed" daily question. Other filters can land as follow-ups if needed.

3. **Click-through reuses spec 167's editor.** Two editors for the same aggregate is a recipe for drift; one editor reachable from multiple surfaces stays consistent.

4. **Domain as spatial primitive, not horizon.** The map answers "where is the architecture moving" — a domain question. Horizon (Now / Next / Later) is a different question answered by spec 171's Target Architecture view. Mixing the two on one canvas was tried in the mockup iterations and judged messy; this slice commits to the cleaner separation.

5. **Stay-type Directions visible but subordinate.** They are real decisions but they don't *move* anything. Treating them with the same visual weight as a consolidation arrow miscommunicates emphasis. Settled visually during implementation; the mockup's "stays" badge is the starting point.

6. **Performance budget is explicit.** Production scale is non-trivial; the slice ships a usable map either at full scale unfiltered or with a sensible default filter, whichever the profiling warrants. The slice does not ship if neither is possible — it would fail the five-second test.

7. **Decomposition entry point is NOT in this slice.** The model says decomposition starts from a single physical capability, not from the map. The map renders existing decompose-type Directions; creating a new one starts elsewhere (a per-capability flow, settled in a future spec).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Map as projection | Recomputation cost on every relevant upstream event | Standard EASI projection pattern; cost is bounded |
| Domain as spatial primitive | The map cannot show horizon at the same time | Spec 171 is the horizon view; the two views are complementary |
| Click-through reuses 167's editor | Map's rendering and 167's edit form must stay in sync as 167 evolves | Acceptable: any change to the editor surfaces consistently in both surfaces |
| Stays visually subordinate | Architects who emphasise stays may want them more prominent | Visual encoding can be tuned post-deploy without re-spec'ing |
| Performance budget enforces a default filter if needed | The unfiltered view may not ship in this slice | Acceptable: a useful filtered view is better than an unusable unfiltered one |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
