# 204 — Retire Global Stylesheets and Hand-Rolled Overlays

> **Status:** ongoing
> **Depends on:** 202 (layer tokens), 168 (Mantine as the single UI vocabulary)

---

## Problem Statement

Twelve unscoped stylesheets survive from before spec 168/179 (`navigation.css`, `canvas.css`, `views.css`, `ChatPanel.css`, `ContextMenu.css`, `StageFlowDiagram.css`, `ValueStreamDetailPage.css`, four settings stylesheets, `HelpTooltip.css` — ~2,400 lines). Global class selectors are where cross-cutting rules hide: the rule that caused spec 201's bug sat in `navigation.css` for two redesigns because nothing marked it as unused and nothing scoped it to the component it styled.

Four overlays are hand-rolled instead of Mantine: `ContextMenu` (radial + linear, `z-index: 10000`, own click-outside/Escape handling), `ColorPicker` (`z-index: 1000`, own click-outside), `FilterPopover` (own click-outside, positioned by a global class), and `ChatPanel` (`position: fixed`, `z-index: 200`, own Escape handling). Each re-implements what Mantine's portal layer already provides, and each carries its own z-index.

This spec finishes the spec 168 consolidation: every stylesheet becomes a CSS module owned by the component it styles, and every overlay renders through Mantine's portal layer.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Any user** | Identical look and behaviour of tree, canvas, chat, context menus, colour picker, filters and settings pages |
| **Frontend developer** | Styles co-located with components; unused selectors visible; overlays that cannot be out-stacked by page content |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Overlays render through Mantine's portal layer

  Scenario: Context menu opens on top of everything and closes like a Mantine menu
    Given a node, edge, tree item or board card with a context menu
    When the user right-clicks it
    Then the menu (radial or linear) appears at the pointer position
    And it closes on click outside and on Escape
    And it is the topmost element at its centre point

  Scenario: Colour picker is a popover
    Given the details panel of a node with colour edit access
    When the user clicks the colour swatch
    Then a popover with the hex picker opens below the swatch
    And clicking outside commits the chosen colour and closes the popover

  Scenario: Tree filter is a popover
    When the user clicks the filter icon in the explorer
    Then the filter panel opens as a popover anchored to the icon
    And clicking outside closes it

  Scenario: Chat panel is a drawer
    When the user opens the assistant
    Then the chat panel slides in from the right without blocking the page behind it
    And Escape closes it

Feature: Every stylesheet is a CSS module

  Scenario: No global selectors remain
    When the frontend builds
    Then no file under src imports a non-module .css file except index.css, tokens.css, skins.css and third-party styles

  Scenario: Existing behaviour is preserved
    When the unit test suite and both Playwright projects run
    Then they pass with the same assertions as before the migration
```

---

## Business Rules & Invariants

1. **CSS modules only** — the only non-module stylesheets under `frontend/src` are `index.css`, `theme/tokens.css`, `theme/skins.css`. Third-party stylesheets (`@xyflow/react`, Mantine) stay as imports.
2. **Overlays are Mantine** — `Popover` for the colour picker and the tree filter, `Drawer` (non-modal: no overlay, no scroll lock, no focus trap) for the chat panel, `Menu` for the linear context menu. The radial context menu keeps its custom geometry but renders through Mantine `Portal`, uses `useClickOutside` from `@mantine/hooks`, and sits on `--layer-popover`.
3. **No hand-rolled dismissal** — no `document.addEventListener('mousedown' | 'keydown')` for click-outside or Escape inside overlay components; Mantine owns dismissal.
4. **Public API unchanged** — `ContextMenu`, `ColorPicker`, `FilterPopover`, `ChatPanel` keep their props; the 30+ consumers of `ContextMenu` do not change.
5. **Selectors reachable by tests are `data-testid`** — e2e helpers and unit tests never query hashed module classes; `.tree-item` and similar become `data-testid` hooks.
6. **Cross-component styling goes through props, not `:global`** — `CanvasWorkspace.module.css` stops reaching into `.navigation-tree`; the tree exposes what the shell needs (a `variant`/class prop or its own module rules).

---

## Acceptance Criteria

- [ ] `find src -name '*.css' ! -name '*.module.css'` lists only `index.css`, `theme/tokens.css`, `theme/skins.css`.
- [ ] `grep -rn "addEventListener('mousedown'\|addEventListener('keydown'" src/components/shared src/features/chat src/features/navigation/components/FilterPopover*` returns nothing.
- [ ] `ColorPicker`, `FilterPopover` render Mantine `Popover`; `ChatPanel` renders Mantine `Drawer`; `LinearContextMenu` renders Mantine `Menu`; `RadialContextMenu` renders inside Mantine `Portal`.
- [ ] Unit tests for each converted overlay cover: opens, closes on outside click, closes on Escape, and (for the colour picker) commits on close.
- [ ] `e2e/helpers.ts` and all unit tests locate tree items via `data-testid`.
- [ ] `CanvasWorkspace.module.css` contains no `:global(` selector.
- [ ] `npm run test`, `npm run lint`, `npm run build`, and `npx playwright test --project=mock` pass.
- [ ] Spec 203's overlay smoke test is extended with one context-menu and one colour-picker case.

---

## Architecture

### Ownership

Frontend only. Bounded contexts touched: navigation (explorer tree), canvas, views, chat, value-streams, settings, shared components. No behaviour or API change.

### Frontend

- Each `X.css` becomes `X.module.css` next to the component that imports it; consumers switch from string class names to `classes.name`. Where a stylesheet served several components (e.g. `navigation.css` → sections, tree items, dialogs), it is split so each module is imported by the components that use it, not by a barrel `index.ts`.
- Overlay components keep their props. Positioning for context menus uses a zero-size anchor at the pointer coordinates as the Mantine target, so `x`/`y` stays the contract.
- The chat `Drawer` uses `withOverlay={false}`, `lockScroll={false}`, `trapFocus={false}`, `closeOnClickOutside={false}` — the page stays interactive, Escape closes, the slide-in comes from Mantine transitions.

---

## Design Decisions

1. **Radial menu stays custom** — Mantine has no radial menu and the geometry is domain UX. It gains the portal layer, Mantine dismissal and the layer token; that is the whole failure class. Alternative: replace radial with linear everywhere (rejected: a UX change outside this spec's intent).
2. **Split per consumer, not one module per old file** — a 300-line module imported by twenty components reproduces the global-stylesheet problem with a hash. Alternative: rename `.css` → `.module.css` mechanically (rejected: cohesion is the point).
3. **`data-testid` over class selectors in tests** — hashed module classes are implementation detail; test hooks are contract.
4. **Drawer for chat rather than keeping `position: fixed`** — puts the panel in Mantine's modal layer (200) by construction and removes the hand-rolled Escape handler.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Large mechanical diff | Review load | One commit per feature area; no behaviour change per commit; tests unchanged in intent |
| Splitting stylesheets | Some duplicated small rules across modules | Shared visual primitives move to Mantine props or a shared module, not back to a global file |

---

## Checklist

- [x] Specification ready — approved by user directive ("implement all 4", 2026-08-26)
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated — no API change
- [ ] User sign-off
