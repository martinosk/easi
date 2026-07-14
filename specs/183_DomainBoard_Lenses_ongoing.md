# 183 — Domain Board Lenses: Now / Journey / Target

> **Status:** ongoing
> **Depends on:** 180 (TIME assessments), 181 (realization roles), 182 (journeys)
> **Design doc:** [`docs/specs/capability-journeys.md`](../docs/specs/capability-journeys.md)

---

## Problem Statement

With assessments (180), roles (181), and journeys (182) recorded, the data for "where are we, what is changing, where are we going" exists — but the Domain Board still renders only the current state. This slice adds the mockup's three-lens presentation to the existing board (user decision 2026-07-12: extend the board, no new page): a **Now / Journey / Target** switcher where Journey shows every capability's change story with progress, moves visible in both affected domains, and Target projects the landscape as it will look when journeys land and standards hold.

This is a read-side, frontend slice: no new writes, no new aggregates. The bulk queries shipped in 180–182 feed alternative card renderers over the existing board view model.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise / Domain Architect** | Flip between today, the change portfolio, and the target picture in one place; spot idle Eliminate-apps and track in-flight work. |
| **Stakeholder / Engineer / PM** | Answer "is this capability settled, changing, or about to change — and what will it run on?" in five seconds. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Time lenses on the Domain Board

  Scenario: The board defaults to Now and offers three lenses
    When a user opens the Business Domains board
    Then the Now lens is active, rendering the board as today (plus TIME badges and role tints from 180/181)
    And a lens switcher offers Now, Journey, and Target with a one-line description of the active lens
    And the legend shows only entries relevant to the active lens

  Scenario: The lens is shareable
    Given a user selects the Journey lens
    Then the lens is reflected in the URL
    And opening that URL lands on the board with the Journey lens active

  Scenario: Journey lens — a capability with an active journey
    Given "Booking management" has an in-flight migration from "Seabook" to "Phoenix" at 60%
    When the Journey lens is active
    Then its card shows the from-chip, an arrow, the to-chip, the kind, an "in flight" pill, and a 60% progress bar

  Scenario: Journey lens — a capability with a completed journey
    Given "Track & trace" has a done consolidation onto "CargoFlow"
    Then its card shows the source label, the standard chip, a full progress bar, and a "done" pill

  Scenario: Journey lens — a capability with no journey
    Then its card shows its app chips and "no change planned"

  Scenario: Journey lens — derived sub-capability progress
    Given "Booking management" (with an active Seabook → Phoenix journey) has children with direct realisations
    When the user expands the sub-capability list on its card
    Then each child shows a derived status: done when it realises "Phoenix" at Full and not "Seabook";
         in flight when it realises both, or "Phoenix" only at Planned or Partial;
         not started when it realises only "Seabook"
    And each child row names the application(s) behind its status

  Scenario: Journey lens — a move renders in both domains
    Given "Invoicing" (Ferry freight) has a planned move to "Group functions" under "Accounts payable" as "Freight invoicing"
    When the Journey lens is active
    Then "Ferry freight" shows a ghost card "Invoicing — moving out" with a "to Group functions" trace link
    And "Group functions" shows an arriving card "Freight invoicing — arriving <period>" under "Accounts payable" with a "from Ferry freight" trace link
    And activating a trace link highlights both cards and scrolls the counterpart into view

  Scenario: Journey lens — portfolio summary
    When the Journey lens is active
    Then the toolbar shows counts of capabilities settled (done journey or no journey), in flight, and not started

  Scenario: Changes-only toggle
    Given the Journey or Target lens is active
    When the user enables "Highlight only what changed"
    Then capabilities and L1 groups without an active or planned change are dimmed
    And L1 groups containing changes are expanded
    And the toggle is unavailable in the Now lens

  Scenario: Target lens — projection per capability
    When the Target lens is active
    Then a capability with an active journey shows its target app as the standard chip (consolidations additionally tagged "consolidated")
    And a capability without a journey shows its standard-role chips
    And a capability with apps but no standard and no journey shows all chips
    And a capability with no apps shows "no standard defined"

  Scenario: Target lens — moves render only at the destination
    Given the planned move of "Invoicing"
    When the Target lens is active
    Then "Ferry freight" no longer shows the capability
    And "Group functions" shows "Freight invoicing" with its target app and "moved from Ferry freight"

  Scenario: Existing board behaviours survive every lens
    Then search filters capabilities and apps in all lenses
    And capability deep links open the drawer in all lenses
    And the drawer (with assessment, role, and journey sections) opens from any lens
    And the assign rail is available only in the Now lens

  Scenario: TIME badges appear only in the Now lens
    Then app chips carry grade badges in Now
    And Journey and Target chips render without grade badges
```

---

## Business Rules & Invariants

1. **Lenses are pure presentation** — no lens performs writes; all three render from read models. Switching lenses is instant client state.
2. **Journey status mapping** — no journey ⇒ steady ("no change planned", counts as settled); `planned` ⇒ not started; `in-flight` ⇒ in flight; most recent `done` journey ⇒ done (also settled). `abandoned` journeys render nothing on the board (drawer history only).
3. **Derived sub-capability status** (no stored state): for each descendant capability of a journey's capability holding direct realisations — realises the to-app at `Full` with no from-app ⇒ done; realises the to-app (any level) alongside a from-app, or the to-app only at `Planned`/`Partial` ⇒ in flight; realises only from-apps ⇒ not started; descendants realising neither are omitted from the breakdown.
4. **Moves render at both ends in the Journey lens** (ghost at source, arriving at destination) **and only at the destination in the Target lens**. Arriving cards nest under the target parent when set, else at the target domain's top level.
5. **Changes-only** exists in Journey and Target lenses only; "changed" means an active journey or a move arriving into the group.
6. **Summary counts** (Journey lens) follow rule 2's mapping over every capability card rendered on the board.
7. **Lens state lives in the URL** (query parameter) and composes with existing deep links; default is Now.
8. **TIME badges are Now-only**; role tints render in Now and on Target's standard chips.
9. **Lens-specific legends** — Now: role/level + TIME key with the "assessment, not commitment" disclaimer; Journey: status + moving/arriving keys; Target: status keys only (mockup).
10. **The assign rail and drag-assignment are Now-only affordances.**

---

## Acceptance Criteria

- [x] Lens switcher with Now default, per-lens description line, and per-lens legend; lens round-trips through the URL
- [x] Journey lens renders the four card states (active with progress bar, done, steady, move ghost/arriving) per the scenarios
- [x] Sub-capability breakdown derives statuses per rule 3 with no persisted journey-child state anywhere
- [x] Trace links highlight both ends of a move and scroll the counterpart into view
- [x] Summary counts and the changes-only toggle behave per rules 5–6, dimming and force-expanding as in the mockup
- [x] Target lens renders per scenario, including consolidated tags and moved-from cards, with moves absent at the source
- [x] Search, deep links, drawer, role visibility, and assign-rail behaviour verified in all lenses (rail Now-only)
- [x] All new styling uses design tokens (`--status-*`, `--skin-*`); zero hard-coded hex values (spec 179 rules)
- [x] Board data loads via the 180–182 bulk queries with no per-capability request fan-out beyond the existing per-domain pattern
- [x] Every BDD scenario has at least one corresponding test
- [x] Every modified file scores 10.0 per `easi-codehealth` (one exception: `useBoardJourneys.ts` — CodeScene MCP cannot emit a numeric score, but its review reports zero code smells)

---

## Architecture

### Ownership

Frontend slice in `features/business-domains`, composing read models owned by `architecturedirection` (assessments, roles, journeys via 180–182 bulk queries) and `capabilitymapping` (existing board data). No backend changes; if a 180–182 bulk query proves insufficient for a lens, the gap is fixed in that spec's surface, not with a new write model.

### Domain Model

None — presentation only. The derived sub-capability status (rule 3) is a pure view-model function over realisation + journey data, unit-tested exhaustively as the one piece of nontrivial logic.

### API Surface

No new endpoints. Board queries: existing realisations per domain + assessments/roles/journeys bulk queries keyed by the domain's capability ids.

### Persistence

None. Lens choice is URL state; no server or localStorage persistence beyond existing board behaviour.

### Frontend

`useDomainBoardPage`/`domainBoardViewModel` gain lens state and journey/move lookups; `BoardCapabilityCard` gains per-lens renderers (now / journey / target), progress bar, sub-capability breakdown expander, ghost/arriving variants, and trace interaction; toolbar gains the switcher, summary stats, and changes-only toggle. Chip variants extend `AppChip`. All chrome from tokens; interactions keyboard-accessible (cards and trace links focusable, Enter/Space activate — mockup behaviour).

### Cross-Context Integration

None beyond read-model queries.

---

## Design Decisions

1. **Extend the Domain Board rather than add a Landscape page** — user decision 2026-07-12; spec 179 built the board on this mockup's Now pattern, and one board keeps deep links, the rail, and role gating in one place.
2. **Sub-capability status derived from real realisations** — the mockup stores per-child statuses in its sample data, but deriving them from the descendants' realisations against the journey's from/to apps keeps a single source of truth and zero duplicate bookkeeping; the derivation reproduces every mockup child state.
3. **TIME badges Now-only** — mockup renders grades only in the Now lens; Journey/Target cards answer different questions, and grade noise there hides the story.
4. **Abandoned journeys invisible on the board** — an abandoned plan is audit history, not landscape narrative; showing it as "not started" (its capture status) would misreport intent.
5. **Lens in the URL** — the board's existing deep-link contract (spec 113 posture) extends naturally; a shared "look at the Journey view" link is a primary use.
6. **Assign rail Now-only** — assignment mutates current structure; offering it while viewing a projection invites edits against the wrong mental model.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Derived sub-capability status | Status quality depends on realisation upkeep (e.g. target app linked as `Planned` per child) | The rule is documented in the expander's empty state; keeping realisations honest is existing practice the derivation rewards |
| One board, three lenses | The board component grows | Per-lens renderers are separate components over one view model; codehealth gate enforces decomposition |
| Abandoned journeys hidden on board | A capability that just abandoned a plan looks "settled" | The drawer's journey history shows the abandonment one click away |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (component-integration tests over the lens provider; no backend changes, so no server integration tests)
- [x] API documentation updated (none — read-side frontend slice, no new endpoints)
- [ ] User sign-off
