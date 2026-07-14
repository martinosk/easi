# 184 — Landscape Signals

> **Status:** pending
> **Depends on:** 180 (TIME assessments), 181 (realization roles), 182 (journeys), 183 (board toolbar surface)
> **Design doc:** [`docs/specs/capability-journeys.md`](../docs/specs/capability-journeys.md)

---

## Problem Statement

Assessments, roles, and journeys are maintained by different people at different times, so they drift apart: an app assessed Eliminate with no journey away from it; a journey migrating off an app someone assessed Invest; a standard app graded Tolerate. Today nobody sees these disagreements until a review meeting stumbles on one.

**Signals** are computed cross-checks over the three layers — *questions for governance, not errors* (mockup). They are never persisted, acknowledged, or dismissed: fixing the underlying data is the only way a signal disappears. The board toolbar shows a live count; a drawer lists each signal with a type badge and a click-through to the capability concerned.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | A standing list of places where recorded judgements and plans contradict each other, to feed governance conversations. |
| **Domain Architect** | See which of their capabilities carry unanswered questions (Eliminate with no plan, stale intent). |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Landscape signals

  Scenario: Gap — Eliminate with no journey
    Given a realisation of "Freight pricing" by "Seabook" is assessed Eliminate
    And "Freight pricing" has no active journey
    Then a "gap" signal reads that "Freight pricing" is realised by "Seabook", assessed Eliminate, but no journey is planned

  Scenario: Priority — the assessment supports starting a planned journey
    Given "Transport planning" has a planned journey whose from-apps include "LogiPlan"
    And the "LogiPlan" realisation is assessed Eliminate
    Then a "priority" signal reads that the assessment supports starting the planned journey

  Scenario: Contradiction — journey moves away from an Invest app
    Given "Booking management" has an active journey whose from-apps include "Phoenix"
    And the "Phoenix" realisation of "Booking management" is assessed Invest
    Then a "contradiction" signal reads that the journey moves the capability away from an Invest app, so assessment or journey needs updating

  Scenario: Tension — standard app assessed Tolerate or Eliminate
    Given "Salesforce" holds the standard role for "CRM"
    And its realisation is assessed Tolerate
    Then a "tension" signal reads that the standard for "CRM" is assessed Tolerate

  Scenario: Carve-out candidate — mixed Invest/Eliminate profile
    Given "CargoFlow" is assessed Invest on at least one capability
    And assessed Eliminate on "Warehouse management", which has no active journey
    Then a "carve-out" signal names the mixed profile and asks whether a carve-out is warranted

  Scenario: Signals count and drawer
    Given the landscape produces five signals
    When a user opens the Business Domains board
    Then the toolbar shows a Signals button with count 5
    And opening it lists each signal with its type badge, text, and location (domain · capability)
    And the drawer states that signals are questions for governance, not errors

  Scenario: Click-through
    When the user activates a signal row
    Then the capability's drawer opens (its board card scrolled into view)

  Scenario: Signals are live, not stored
    Given a "gap" signal for "Freight pricing"
    When an architect captures a journey on "Freight pricing"
    Then the next signals fetch no longer contains that signal
    And no signal state was written anywhere

  Scenario: Empty state
    Given a landscape with no disagreements
    Then the Signals button shows 0 and the drawer reads "No signals."
```

---

## Business Rules & Invariants

Signal predicates — evaluated per current assessment (180), role (181), and journey (182) state. "Active journey" means status `planned` or `in-flight`.

1. **Gap** — a realisation assessed `Eliminate` on a capability with no active journey.
2. **Priority** — a realisation assessed `Eliminate` whose capability has a `planned` journey listing that app among its from-apps.
3. **Contradiction** — a realisation assessed `Invest` whose capability has an active journey listing that app among its from-apps.
4. **Tension** — a realisation holding the `standard` role while assessed `Tolerate` or `Eliminate`.
5. **Carve-out candidate** — an application assessed `Invest` on ≥ 1 capability and `Eliminate` on ≥ 1 other capability that has no active journey; one signal per such application, anchored to one of its Eliminate capabilities.
6. **Computed at query time, never persisted** — no table, no events, no acknowledge/dismiss state.
7. **Deterministic** — the same data yields the same signals in a stable order (by domain, capability, type).
8. **Duplicates collapse** — a (type, capability, application) triple yields at most one signal.
9. **Read permission equals board read permission**; signals expose no data the caller could not already read.
10. **Rules 1–3 consider only *active* journeys.** A done journey does not suppress a gap: if an Eliminate-assessed app is still realising the capability after its journey completed, the question stands.

---

## Acceptance Criteria

- [ ] Each predicate 1–5 fires exactly per its rule, covered by unit tests including boundary cases (done journey ≠ suppression per rule 10; from-app membership; mixed-profile counting)
- [ ] Signals endpoint returns type, text, location, capability reference (and application where applicable) in deterministic order
- [ ] Board toolbar shows the Signals button with live count; drawer renders type-badged rows and the governance framing text; empty state reads "No signals."
- [ ] Activating a signal opens the referenced capability's drawer and scrolls its card into view
- [ ] Changing the underlying data changes the next fetch's signals; no signal state exists in any store
- [ ] Every BDD scenario has at least one corresponding test
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

`architecturedirection` — every input (assessments, roles, journeys) is local, and capability/domain/application names come from its existing denormalised caches. The engine is a query-time application service over read models, the `TimeSuggestionReadModel` composition precedent.

### Domain Model

None. The predicates are pure functions over read-model rows — implemented as a unit-tested domain service with no aggregate, no events.

### API Surface

One read endpoint returning the signal list (tenant-wide; optional domain filter), HATEOAS links to the referenced capability resources. Shape per `easi-api-standards`.

### Persistence

None.

### Frontend

Toolbar Signals button with count (board, all lenses) opening a drawer of signal rows (type badge, text, where); row activation reuses the board's existing deep-link/scroll/drawer machinery. Query follows `easi-frontend-data`; the count invalidates alongside assessment/role/journey mutations via their `mutationEffects.ts`.

### Cross-Context Integration

None — reads local read models only.

---

## Design Decisions

1. **Computed, never persisted** — mockup framing ("questions for governance, not errors"). An acknowledge/dismiss workflow would turn honest questions into ticket bookkeeping and let disagreements be hidden without being resolved.
2. **Active-journey semantics (rule 10) deviate from the mockup's sample logic** — the mockup checks for the *presence* of a transition object; in the live system terminal journeys persist forever, so presence-based suppression would silence every gap after any historical journey. Anchoring rules 1–3 to *active* journeys preserves the mockup's intent on its own sample data while staying correct over history.
3. **Fixed rule set in code** — five rules, unit-tested. Configurable rule authoring rejected: no user need, large surface.
4. **Signals live in `architecturedirection`** — the alternative (frontend-computed, mockup-style) rejected: the count belongs on the toolbar before the board fans out per-domain data, and server-side rules stay consistent for any future consumer (assistant tools, one-pagers).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Never persisted | No trail of "we saw and discussed this signal" | Resolving actions (journeys, re-assessments, role changes) are themselves evented and auditable |
| Fixed rule set | New disagreement patterns need code changes | Each rule is a small pure function; adding one is a unit-tested one-file change |
| Query-time computation | Cost grows with landscape size | Inputs are indexed read-model rows; same posture as TIME suggestions, which holds at production scale |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
