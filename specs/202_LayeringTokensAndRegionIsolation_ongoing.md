# 202 — Layering Tokens and Region Isolation

> **Status:** ongoing
> **Depends on:** 201 (the bug this generalises), 179 (canvas workspace shell)

---

## Problem Statement

Spec 201 shipped a bug where the header overflow menu was painted behind the explorer tree on the Architecture Canvas. The cause was not the menu: a feature stylesheet gave the tree `z-index: 1000` and the shell pane containing it created no stacking context, so a feature-local number competed at document level with Mantine's portal layer (popover 300).

Frontend z-index today is eleven unrelated literals (`1, 2, 3, 5, 10, 50, 100, 200, 1000, 10000`) with no relationship to each other or to Mantine's own scale (`app 100, modal 200, popover 300, overlay 400`). Nothing prevents the next feature from repeating spec 201's failure with a different number.

This spec makes stacking a property of the design system: a named layer scale in the token file, aligned with Mantine; shell regions that own their stacking contexts; and a lint gate so a raw z-index literal cannot enter the codebase again.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect / Stakeholder (any user)** | Menus, tooltips and dialogs always render on top of page content, on every page and at every window width |
| **Frontend developer** | One vocabulary for "how high does this sit", enforced by tooling, instead of guessing a number |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Overlay layering is governed by the design system

  Scenario: Feature content can never rise above a portaled overlay
    Given a shell region (explorer, canvas, details, main content) contains an element with the highest in-region layer
    When a Mantine Menu, Popover or Tooltip is opened from the header
    Then the overlay is painted above that element

  Scenario: Layer tokens track Mantine's scale
    Given the token file defines shell, panel and popover layers
    Then each equals the corresponding Mantine default z-index
    And the in-region layers are all lower than the shell layer

  Scenario: A raw z-index literal fails lint
    Given a stylesheet or component outside the token file sets "z-index: 1000" or "zIndex: 5"
    When "npm run lint" runs
    Then it fails naming the file and line
    And a value written as "var(--layer-…)" passes
```

---

## Business Rules & Invariants

1. **Single layer scale** — every z-index in `frontend/src` (CSS, CSS modules, TSX) is expressed as `var(--layer-*)`; the only numeric definitions live in `src/theme/tokens.css`.
2. **Two tiers** — in-region layers (`base`, `raised`, `floating`, `pinned`) order elements *inside* a region; shell layers (`shell`, `panel`, `popover`) order regions and portaled overlays. Every in-region value is lower than every shell value.
3. **Mantine alignment** — `--layer-shell`, `--layer-panel`, `--layer-popover` equal Mantine's `app`, `modal`, `popover` defaults. Nothing in EASI defines a value above `popover` except through Mantine itself.
4. **Regions isolate** — every shell layout region (`explorerPane`, `canvasPane`, `detailsPane`, the main content area) creates its own stacking context (`isolation: isolate`). Feature code positions relative to its region, never to the viewport.
5. **Lint is the gate** — `npm run lint` fails on any z-index literal outside `tokens.css`. Test files and `node_modules` are excluded.

---

## Acceptance Criteria

- [x] `tokens.css` defines `--layer-base/raised/floating/pinned` and `--layer-shell/panel/popover`; a unit test asserts rule 2 and rule 3 against `getDefaultZIndex` from `@mantine/core` — `src/theme/layers.test.ts`.
- [x] Every existing z-index site in `frontend/src` uses a `--layer-*` token; `grep -rn "z-index\|zIndex" src` outside `tokens.css` shows only `var(--layer-` values — 14 sites migrated.
- [x] `explorerPane`, `canvasPane`, `detailsPane` and the main content region carry `isolation: isolate` — main region is the new `<main data-testid="main-region">` in `App.tsx`.
- [x] A layering check runs as part of `npm run lint`; a unit test proves it flags numeric `z-index:` / `zIndex:` and accepts `var(--layer-*)` — `scripts/check-layering.ts`, `scripts/layering.test.ts`.
- [x] `easi-frontend-styling` skill documents the layer scale and the "regions isolate, features never position to the viewport" rule.
- [x] Spec 203's overlay smoke test passes on all pages.

---

## Architecture

### Ownership

Frontend only — `frontend/src/theme/tokens.css`, `frontend/src/components/layout/`, a lint script under `frontend/scripts/`, and the affected feature stylesheets. No backend change.

### Frontend

- Layer tokens are CSS custom properties; Mantine consumes its own `--mantine-z-index-*` variables, EASI code consumes `--layer-*`. Equality between the two scales is enforced by a test rather than by referencing Mantine variables from the token file, so the token file stays self-describing.
- The layering check is a small Node script with a pure core (`findLayeringViolations(files)`) that is unit-tested, and a CLI entry that scans `src/**/*.{css,ts,tsx}` and exits non-zero on violations. It is chained into `npm run lint` after Biome, because Biome cannot express a value-level CSS rule.
- Region isolation lives in the shell's CSS modules (`CanvasWorkspace.module.css`, the main content container), not in feature code.

---

## Design Decisions

1. **Numeric tokens + equality test, not `var(--mantine-z-index-*)` in tokens.css** — the token file remains readable as a scale on its own; the test fails loudly if a Mantine upgrade changes its defaults. Alternative: reference Mantine variables directly (rejected: the scale becomes invisible in the file and silently drifts with upgrades).
2. **Custom lint script instead of a Biome rule** — Biome 2.5 has no value-level CSS rule and no custom-rule API. Alternative: Stylelint (rejected: a second lint toolchain for one rule).
3. **Four in-region layers only** — `base/raised/floating/pinned` cover every current use (0/1/2/3, 5, 10, 50). Alternative: keep arbitrary small integers (rejected: they are the same problem at a smaller scale).
4. **`isolation: isolate` over `position: relative; z-index: 0`** — expresses intent, has no side effect on containing blocks.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Regions isolate | A feature can no longer intentionally overlay a sibling region | Overlays that must escape a region go through Mantine portals — the only sanctioned way |
| Lint gate | Adds a step to `npm run lint` | Script runs in well under a second; failure message names file and line |

---

## Checklist

- [x] Specification ready — approved by user directive ("implement all 4", 2026-08-26)
- [x] Implementation done
- [x] Unit tests implemented and passing — `layers.test.ts`, `layering.test.ts`
- [x] Integration tests implemented if relevant — spec 203 mock-mode overlay smoke test
- [x] API documentation updated — no API change
- [ ] User sign-off
