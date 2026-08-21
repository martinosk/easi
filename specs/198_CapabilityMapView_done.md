# 198 — Capability Map View (Reintroduced)

> **Status:** ongoing
> **Depends on:** 179_UI_Redesign_DesignSystem (done), 195_SubCapabilityJourneys (done)

---

## Problem Statement

Spec 179 rebuilt the Business Domains page as an all-domains board and removed the nested capability map that spec 058 had introduced. The board is well received, but users miss the map: a spatial, nested rendering of a single domain's capability hierarchy that shows L1 blocks containing their L2–L4 children at a glance. The board's collapsible list groups answer "what is in this domain?", while the map answered "what does this domain look like?" — a shape-of-the-landscape view architects used for orientation and stakeholder conversations.

This spec reintroduces the capability map as a supplement to the board, not a replacement. The board remains the default and keeps all its behavior. The map is a second, read-only way of viewing one domain at a time, reachable via a view toggle on the same page.

The original map's drag-to-arrange positions were persisted through the ViewLayouts bounded context, which spec 192 retired. The reintroduced map therefore uses deterministic auto-layout and needs no position store and no backend changes.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect** | See one domain's full capability hierarchy spatially, drill into any capability, present the domain landscape to others |
| **Stakeholder (read-only)** | Orient themselves in an unfamiliar domain without expanding list groups one by one |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Capability map view on the Business Domains page

  Scenario: Switching to the map view
    Given I am on the Business Domains page showing the board
    When I switch the view toggle from "Board" to "Map"
    Then the board is replaced by a nested capability map for one selected domain
    And the domain's L1 capabilities render as blocks containing their child capabilities

  Scenario: View choice persists
    Given I have selected the "Map" view
    When I reload the Business Domains page
    Then the map view is shown again

  Scenario: Switching domains within the map
    Given I am in the map view showing domain "Sales"
    When I select domain "Logistics" in the domain selector
    Then the map renders the capability hierarchy of "Logistics"

  Scenario: Controlling visible depth
    Given I am in the map view with the depth selector at "L1–L2"
    When I change the depth selector to "L1–L4"
    Then L3 and L4 capabilities become visible nested inside their parents
    And my depth choice persists across page reloads

  Scenario: Drilling into a capability
    Given I am in the map view
    When I click a capability cell
    Then the existing capability drawer opens for that capability
    And it shows the same content it shows when opened from the board

  Scenario: The map is a Now-only view
    Given the board is showing the Journey lens
    When I switch to the map view
    Then the map renders with Now coloring and no lens switcher
    And switching back to the board restores the Journey lens

  Scenario: Showing realising applications on the map
    Given I am in the map view with "Show apps" off
    When I switch "Show apps" on
    Then each capability cell lists its realising applications as the same chips the board uses
    And my choice persists across page reloads

  Scenario: Opening an application from the map
    Given the map shows application chips
    When I click an application chip
    Then the application drawer opens for that application
    And the capability drawer does not open

  Scenario: Searching in the map
    Given I am in the map view with capabilities visible
    When I type a query into the search field
    Then cells whose capability name or code does not match are visually de-emphasized
    And matching cells remain fully visible

  Scenario: Domain with no assigned capabilities
    Given the selected domain has no L1 capabilities assigned
    When I view it in the map
    Then an empty state explains the domain has no capabilities and points to the board's assign flow

  Scenario: The board is unaffected
    Given the map view exists
    When I switch back to "Board"
    Then the board renders and behaves exactly as before this spec
```

---

## Business Rules & Invariants

1. **Board stays default** — a user who has never chosen a view sees the board.
2. **Map is read-only** — the map offers no assignment, no drag-and-drop, no create/edit/delete actions; all mutation flows remain on the board. Drill-in via the capability drawer is the only interaction besides navigation controls.
3. **Deterministic auto-layout** — cells at each nesting level are ordered alphabetically by capability name (capabilities carry no code field in the frontend model); the same data always renders the same map. No positions are persisted anywhere.
4. **One domain at a time** — the map renders exactly one domain; a selector switches between domains.
5. **Depth bounds** — the depth selector offers L1 only, L1–L2, L1–L3, and L1–L4; default is L1–L2.
6. **Shared drill-in** — clicking a capability opens the same capability drawer the board uses, with identical content.
7. **Now-only map** — the map always renders the Now view (maturity coloring and its legend); the Journey/Target lenses, their controls, and the changes-only toggle exist only on the board. Switching views never mutates the board's lens state; the map ignores it and the board restores it.
8. **Client-side persistence only** — view choice, selected domain, depth, and the apps toggle persist per user in the browser; nothing is stored server-side.
9. **Apps toggle** — the map toolbar offers a "Show apps" switch, off by default; when on, each cell shows its realising applications as the same chips board cards use, and clicking a chip opens the application drawer (drill-in, consistent with Rule 2).

---

## Acceptance Criteria

- [x] A Board/Map toggle is visible on the Business Domains page; default is Board (Rule 1)
- [x] Map view renders the selected domain's capabilities as nested blocks, ordered by name (Rule 3)
- [x] A domain selector switches the mapped domain without leaving the map (Rule 4)
- [x] Depth selector offers the four levels, defaults to L1–L2, and takes effect immediately (Rule 5)
- [x] View choice, selected domain, and depth survive a page reload (Rule 8)
- [x] Clicking a map cell opens the existing capability drawer for that capability (Rule 6)
- [x] The map shows no lens switcher and renders Now coloring regardless of the board's lens; the board's lens survives a round-trip through the map (Rule 7)
- [x] A persisted "Show apps" switch adds the board's application chips to cells; chip click opens the application drawer (Rule 9)
- [x] Search de-emphasizes non-matching cells in the map
- [x] Empty domain shows an explanatory empty state
- [x] The map exposes no mutation affordances (Rule 2)
- [x] All board tests still pass; board behavior is unchanged
- [x] No backend, API, or database changes are introduced

---

## Architecture

### Ownership

Frontend-only change inside the `business-domains` feature (`frontend/src/features/business-domains/`). No bounded context on the backend is touched.

### Domain Model

None. No new aggregates, events, or read models.

### API Surface

None. The map reads the data already fetched for the board (domains, per-domain capability trees, realizations, journeys, lens indexes). No new endpoints and no changes to existing ones.

### Persistence

Browser-local only: view mode, selected domain id, and depth level per user. A stored domain id that no longer exists falls back to the first domain.

### Frontend

- The Business Domains page gains a view-mode state (board | map) surfaced as a segmented toggle in the toolbar area; the toolbar's lens switcher and search remain visible and functional in both views.
- A new map component subtree renders one `DomainBoardViewModel`'s L1 groups as a nested grid: L1 blocks containing recursively nested child cells down to the selected depth. It consumes the existing page hook's data — no new data hooks or queries.
- Depth selector component with localStorage persistence (pattern mirrors the pre-179 `usePersistedDepth`).
- Cell click delegates to the existing `openCapabilityById` flow so the capability drawer, application drawer, and navigation behavior are reused unchanged.
- Lens coloring reuses the existing `BoardLensProvider`/lens modules; map cells resolve their color through the same lens functions as board cards.
- Board-only affordances (assign rail, create-domain, context menus, drag handlers) are not rendered in map view.

### Cross-Context Integration

None.

---

## Design Decisions

1. **View toggle on the existing page, not a separate route** — the map is a different projection of the same data the page already loads; a toggle keeps lens, search, and drawers shared. Alternatives considered: separate nav entry (rejected: duplicates page state and data loading), per-domain "Map" button like pre-179 (rejected: hides the view behind each card and complicates returning to the board).
2. **Auto-layout instead of drag-to-arrange** — the position store the old map relied on (ViewLayouts) was retired by spec 192. Deterministic ordering by capability code makes the map shareable and consistent for all users with zero backend scope. Alternatives considered: per-user localStorage positions (rejected: layouts diverge between users, defeating the map's presentation purpose), new backend position store (rejected: large scope; would warrant its own spec if demand appears).
3. **Read-only first slice** — all mutations stay on the board, which already has tested flows for assignment, editing, and deletion. Duplicating them in the map doubles surface area without new user value. Alternative considered: full board parity (rejected for scope).
4. **Default depth L1–L2** — L1-only duplicates what the collapsed board already shows; L1–L4 is visually dense for large domains. L1–L2 gives the "shape of the domain" the map exists for.
5. **Reuse `DomainBoardViewModel` as the map's data source** — the board's view model already contains the full capability tree, journey annotations, and lens inputs per domain. Rendering the map from it guarantees board/map consistency and avoids new queries.

---

6. **Now-only map (revised after user feedback)** — an earlier revision mirrored the board's Journey/Target lenses onto map cells via status pills and border colors. Users found it confusing: the map cannot carry the lenses' full semantics (ghost cards, arriving moves, summaries), so the lens appeared broken, and Target hid moving L1s in a view meant to show domain contents. The lenses now live exclusively on the board; the map always renders Now. Alternative considered: full lens parity on the map (rejected: duplicates the board's change-planning surface without its semantics).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Auto-layout only | Users who arranged the old map lose custom layouts | Deterministic ordering is predictable; a position store can be specced later if demand is real |
| One domain at a time | No all-domains map | The board covers the all-domains overview; the map is for depth on one domain |
| Read-only map | Users must switch to board to assign/edit | Toggle is one click and view state persists |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (component + page level; verified in-browser via Playwright)
- [x] API documentation updated (no API changes)
- [x] User sign-off
