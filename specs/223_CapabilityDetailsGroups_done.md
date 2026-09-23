# 223 — Capability Details: Collapsible, Re-arrangeable Groups

> **Status:** done
> **Depends on:** 220 (capability details edited in place), 211 (maturity journeys)
> **Roadmap alignment:** outside roadmap: presentation only; the capability panel that 220 unified is regrouped, no domain or contract change. Respects SD6's principle that every affordance is a HATEOAS link.

---

## Problem Statement

Spec 220 gave capabilities one detail panel on every surface, but that panel is a flat list of fourteen sections. On the Business Domains drawer the journey block, the strategic importance block, the identity fields, the ownership fields, the tags, the experts and the realising applications run into each other with no visual structure, and a user looking for one thing scrolls past everything else.

This spec arranges the panel into named groups. Each group collapses, and the user can move groups up and down so the things they look at first sit at the top. The arrangement is remembered and applies on every surface that shows the panel, so the Architecture Canvas details pane and the Business Domains drawer share one layout.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architecture maintainer** | Find the field to fix without scrolling past unrelated sections; keep the groups they use most at the top |
| **Domain owner** | Open a capability from the domain board and see transition and fitness first, with the identity and ownership fields folded away |
| **Read-only viewer** | The same grouped panel, foldable, with no edit controls |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Capability details arranged in collapsible, re-arrangeable groups

  Scenario: Groups and their contents
    Given capability "Order Management" has every field populated and one realising application
    When I open its details on the Business Domains drawer
    Then the capability name heads the panel above every group
    And I see these groups in this default order, each with a header I can click:
      | Group                  | Contents                                                              |
      | Description            | description, level, status                                            |
      | Transition             | plan journey action, current journeys with their plan summary         |
      | Fitness                | strategic importance, maturity                                        |
      | Metadata               | ownership model, primary owner, EA owner, tags, experts, created      |
      | Realising applications | the realising applications list                                       |
    And the one-pager action and the history section follow the groups

  Scenario: The canvas shows the same groups
    Given "Order Management" is placed on the current view
    When I open its details on the Architecture Canvas
    Then I see the Description, Fitness, Metadata and Realising applications groups in the same order as the drawer
    And an "In this view" group, holding the custom colour control and "Remove from view", follows Realising applications
    And the Transition group and the strategic importance content are absent, because they belong to the domain board

  Scenario: Collapse and expand a group
    Given the Metadata group is expanded
    When I click its header
    Then its contents are hidden and the header stays
    When I click the header again
    Then its contents are shown

  Scenario: Move a group
    Given the default order
    When I activate "Move Fitness up"
    Then Fitness sits above Transition
    And "Move Description up" and "Move <last group> down" are disabled

  Scenario: The arrangement is remembered and shared across surfaces
    Given I moved Fitness above Transition and collapsed Metadata on the Business Domains drawer
    When I open any capability on the Architecture Canvas
    Then Fitness sits above Description's successor in the same relative order
    And Metadata is collapsed
    When I reload the page
    Then the same arrangement applies

  Scenario: An unknown stored arrangement falls back to the default
    Given the stored arrangement names a group that no longer exists or omits a group
    When I open capability details
    Then unknown names are ignored and omitted groups appear at their default position, expanded

  Scenario: Editing still happens in place
    Given the capability resource carries an "edit" link
    When I rename the capability in the heading and edit the description inside Description
    Then both edits behave exactly as in 220
```

---

## Business Rules & Invariants

1. **Five fixed groups plus the canvas group** — the panel has exactly these groups: Description, Transition, Fitness, Metadata, Realising applications, and on the canvas only, In this view. No host adds a group of its own.
2. **Every 220 section lands in one group, the name heads the panel** — the name renders above the groups on every surface and never collapses or moves; description, level and status in Description; journeys in Transition; strategic importance and maturity in Fitness; ownership model, primary owner, EA owner, tags, experts and created in Metadata; the realising applications list in its own group; custom colour and remove-from-view in In this view. The one-pager action and the history section stay below the groups, ungrouped.
3. **Host slots are named and placed** — the Business Domains drawer supplies the Transition content and the strategic importance content; the canvas supplies the In this view content. A group whose content the host does not supply is not rendered, and the empty Transition group never renders on the canvas.
4. **Every group collapses** — a group header toggles its content; collapsing hides content without unmounting edits already in progress.
5. **Every group moves** — each group header carries move-up and move-down controls; the first group's move-up and the last group's move-down are disabled. Moving changes the order of every rendered group on every surface.
6. **One remembered arrangement** — order and collapsed state are stored once per browser under one key and read by every surface. Default: the order in rule 1, all groups expanded.
7. **Robust to change** — a stored order is reconciled against the known group ids: unknown ids are dropped, missing ids are appended in default order, and unknown collapsed ids are ignored.
8. **Field affordances unchanged** — every field keeps the in-place control, link gating and commit behaviour that 220 settled. Group headers are structural controls and are not gated by links.
9. **Titles come from the group** — a section inside a group does not repeat the group title; the Transition heading the journey section used to render moves to the group header.

---

## Acceptance Criteria

- [x] A shared grouped-details primitive renders named groups as collapsible items with move-up and move-down controls; a test covers toggle, reorder and disabled edges.
- [x] A layout hook persists order and collapsed ids under one storage key, reconciles stored order against known ids, and falls back to the default; a test covers each rule.
- [x] The capability panel renders the name above the groups and the five groups in the default order with the contents in rule 2; the panel test asserts the grouping.
- [x] The Business Domains drawer supplies the Transition and strategic importance content as separate slots; the drawer test asserts the journey content sits in the Transition group and strategic importance in Fitness.
- [x] The canvas details pane renders the same groups plus In this view after Realising applications, and no Transition group; the canvas test asserts this.
- [x] Collapsing a group and moving a group on one surface is reflected on the other surface after remount (panel test, and mock-mode Playwright `e2e/mock/capability-details-groups.spec.ts` across the drawer and the canvas).
- [x] The one-pager action and history section render below the groups on every surface.
- [x] All existing 220 in-place editing tests still pass unchanged in behaviour.

---

## Architecture

### Ownership

Frontend only, in the capabilities feature and a shared presentation primitive. Business Domains keeps hosting the panel and supplying its domain-context slots.

### Domain Model

No change.

### API Surface

No change.

### Persistence

Browser-local only: the group order and collapsed ids, keyed once per browser, not per surface and not per capability.

### Frontend

- A shared primitive takes an ordered list of groups (id, title, content) and a layout (order, collapsed, toggle, move) and renders a Mantine accordion with one item per rendered group. Move controls sit beside, not inside, the accordion control.
- A shared layout hook owns the storage key, the default order and the reconciliation rule.
- The capability panel builds its groups from its own sections and the host slots. The panel component's slot props become `transition`, `strategicImportance` and `viewMembership`.
- The Business Domains drawer passes the journey section as `transition` and the strategic importance section as `strategicImportance`. The canvas host resolves whether the capability is on the current view with actionable links and passes `viewMembership` only then.
- The journey section drops its own Transition heading; the view membership section drops its own In this view heading.

### Cross-Context Integration

None.

---

## Design Decisions

1. **Move controls instead of drag-and-drop** — up and down buttons on the group header are keyboard-accessible, need no new dependency and are enough for five groups. Alternatives considered: _drag-and-drop_ (rejected: adds a library or hand-rolled drag handling for a five-item list, and is hard to make accessible).
2. **One arrangement for every surface** — the user asked for the same layout on the drawer and the canvas; a single storage key gives that for free and avoids two diverging layouts. Alternative: _per-surface storage_ (rejected: contradicts the ask).
3. **Groups absent rather than empty** — a host that has no content for Transition or In this view supplies nothing, and the group is not rendered, so the canvas never shows an empty Transition header. Alternative: _render every group everywhere with an empty state_ (rejected: noise, and the domain-only content is a 220 rule).
4. **Ungrouped one-pager action and history** — both are actions on the whole record rather than fields, and history already collapses on its own. Alternative: _a History group_ (rejected: double collapse).
5. **Local storage, not the user profile** — the arrangement is a per-browser convenience like the domain board's view mode; promoting it to a server-side preference is not warranted by five groups.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Browser-local arrangement | Does not follow the user to another device | Default order is sensible; one click per move restores it |
| Move buttons | Slower than drag for large reorders | Five groups; at most a handful of clicks |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (mock-mode Playwright scenario; no backend change)
- [x] API documentation updated (no API change)
- [x] User sign-off
