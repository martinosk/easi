# 168 — Frontend UI Vocabulary Overhaul

> **Status:** pending
> **Depends on:** —
> **Conceptual basis:** This session's audit of `frontend/src/`, plus the descriptive section of `.opencode/skills/easi-frontend-patterns/SKILL.md` ("UI Styling — pick a vocabulary").

---

## TL;DR

**Standardise on Mantine v8.** The codebase already has a Mantine theme (`src/theme/mantine.ts`) whose tokens are a 1:1 mirror of the CSS variables in `src/index.css`. Mantine wins in 33 source files; the `.btn` + ad-hoc-feature-class system wins in 45. The third "vocabulary" — bare HTML wrapped in 176 `style={{}}` blocks across 46 files — is not a vocabulary, it is the absence of one, and it ships browser-default UI exactly like the `DirectionPanel` regression that triggered this spec. Migrate detail panels, list views, page layouts, and shared dialogs to Mantine in vertical slices. Retire the `.btn` system and the duplicated `.dialog` rules. Keep `src/index.css` only for the variable block and a small set of app-shell / canvas rules that genuinely belong outside the design system.

---

## Problem Statement

EASI ships three incompatible UI vocabularies and a contributor cannot tell which one applies to a given surface without grepping the neighbours. The cracks are now user-visible: the recent `DirectionPanel` shipped with `<button>` elements that rendered as browser defaults next to polished `.btn`-styled siblings, and was only caught after release. The user reaction — *why are there so many different ways to do this* — is correct. The status quo is a slow, silent quality regression.

The audit below counts the actual fragmentation, identifies the seams that ship the worst inconsistencies, and proposes a single end-state with a migration plan that does not require freezing feature work.

---

## Audit Summary

### Inventory

- **Total source `.tsx` files** (excluding tests/examples): **~212**.
- **Files importing `@mantine/core`** (excluding tests): **33** — 16% of source `.tsx`.
- **Files with `className="btn ..."`**: **45** — 21% of source `.tsx`.
- **Files with at least one `style={{}}` block**: **46** — 22% of source `.tsx`, totalling **176 inline-style occurrences**.
- **Files using both Mantine and `.btn` *in the same file*: **0**.** The mix is at the surface level, not the file level.
- **`src/index.css`**: **2,865 lines, ~404 class selectors**. Includes one full `.dialog` block at line 895 and a *second* full `.dialog` block at line 1714 (genuine duplication, not an override).
- **`src/App.css`**: 42 lines of Vite scaffolding (`#root`, `.logo`, etc.) — dead, can be deleted today.
- **Feature CSS files**: **24** sibling `.css` files under `src/features/**` and `src/components/**`, ranging from 30 lines (`HelpTooltip.css`) to 434 lines (`EnterpriseArchPage.css`).
- **`@dnd-kit/*`**: in `package.json`, **0 imports** anywhere in `src/`. Dead dependency; remove during cleanup.

### The two real vocabularies

| Vocabulary | Files | Where it lives | Form library |
|---|---:|---|---|
| **Mantine v8** (`@mantine/core`, `@mantine/hooks`) | 33 | `settings/` (pure), all `Create*Dialog` / `Edit*Dialog` / `Add*Dialog` / `Delete*Dialog` across `capabilities/`, `components/`, `origin-entities/`, `relations/`, `canvas/` toolbar, `business-domains/VisualizationArea.tsx`, plus `shared/HelpTooltip` and `shared/MaturitySlider` | `react-hook-form` + `@hookform/resolvers/zod` |
| **`.btn` + feature-scoped `.css` + ad-hoc kebab classes** | 45 | `enterprise-architecture/` (pure source), `invitations/`, `users/`, `importing/`, `edit-grants/`, `business-domains/` detail panels, `origin-entities/` detail panels, `relations/` detail/action panels, `auth/LoginPage`, `releases/`, plus `shared/ConfirmationDialog` and `shared/ErrorScreen` | Raw `<form>` + manual `useState` |

The split is consistent enough to be a *latent design rule*: **dialogs and modals use Mantine + react-hook-form; detail panels, lists, and pages use `.btn` + feature CSS + raw forms.** No one wrote that down. New contributors find it by accident, or — as with `DirectionPanel` — fail to find it and ship bare HTML.

### Per-feature breakdown

Where a feature contains both vocabularies, the split is dialog-vs-panel:

| Feature | Mantine files | `.btn` files | Pattern |
|---|---:|---:|---|
| `settings/` | 7 | 0 | Mantine-pure |
| `capabilities/` | 7 | 1 | Mantine dialogs, `.btn` detail panel |
| `origin-entities/` | 6 | 4 | Mantine create/edit dialogs, `.btn` detail panels |
| `components/` | 4 | 2 | Mantine dialogs, `.btn` detail panels |
| `relations/` | 3 | 2 | Mantine dialogs, `.btn` detail/action panels |
| `canvas/` | 4 | 0 | Mantine-pure |
| `business-domains/` | 1 | 7 | One Mantine canvas-like surface (`VisualizationArea`); rest `.btn` |
| `enterprise-architecture/` | 0 (source) | 8 | `.btn`-pure |
| `invitations/`, `users/`, `importing/`, `edit-grants/`, `auth/`, `releases/` | 0 | 12 | `.btn`-pure |
| `architecture-direction/` | 0 | 1 | `.btn` (after the post-release hot-fix) |

Files with the most accumulated inline styles (which is where the *third* vocabulary — "bare HTML + `style={{}}`" — actually shows up):

| File | `style={{}}` blocks |
|---|---:|
| `architecture-direction/components/CaptureDirectionForm.tsx` | **25** |
| `enterprise-architecture/components/DomainCapabilityPanel.tsx` | 18 |
| `enterprise-architecture/components/EnterpriseCapabilityCard.tsx` | 13 |
| `business-domains/components/CapabilityExplorer.tsx` | 11 |
| `business-domains/components/{ReassignConfirmDialog,DetailsSidebar}.tsx` | 7 each |
| `components/layout/DockviewLayout.tsx` | 7 |

`CaptureDirectionForm.tsx` is the canonical bad example: 25 inline-style blocks, a fieldset with hard-coded `#E5E7EB`, hard-coded greys (`#6B7280`, `#B45309`, `#B91C1C`) that *exist as `--color-gray-500` / `--color-warning` / `--color-error`-ish variables five files away*, raw `<button>` / `<select>` / `<input>` / `<fieldset>` / `<legend>` everywhere. It is what `DirectionPanel.tsx` looked like before the hot-fix; it still ships today and is the next regression waiting to be reported.

### Dialog fragmentation specifically

This is the worst inconsistency, because dialogs are the loudest UI element. Three distinct dialog vocabularies are live:

1. **Mantine `Modal`** — 18 files (all the `Create*Dialog` / `Edit*Dialog` / `Delete*Dialog` / `Add*Dialog` listed above). Backed by `MantineProvider` in `src/main.tsx`; tests need `MantineTestWrapper`.
2. **`.dialog` CSS class system** — 14 files including `shared/ConfirmationDialog`, `InviteUserModal`, `CreateEnterpriseCapabilityModal`, the four `ImportingDialog` step components, `CreateViewDialog`, `ChangeRoleModal`, etc. The class is defined **twice in `index.css`** (lines 895 and 1714) with different rules.
3. **Bare HTML with inline-style overlay** — `CaptureDirectionForm` and the dialog *inside* `ReassignConfirmDialog`.

The user-visible consequence: confirmation dialogs across the app have at least two visibly different chromes, and a contributor adding a new dialog cannot answer "which one do I use" without grepping siblings.

### Git history — what does the codebase imply about intent?

- **Pre-Dec 2025**: `.btn` + feature `.css` + ad-hoc kebab classes was the only vocabulary. `src/index.css` grew to its current 2,865 lines.
- **2025-12-20** (`f9744b27` "Introduce Dockview UI framework with React Query"): Mantine added to `package.json`, `src/theme/mantine.ts` created with tokens mirroring the existing CSS variables. The commit message says "Add Mantine theming infrastructure" — i.e. the intent was an incremental migration, not a parallel system.
- **2026-01-08** (`e654dfb8` "FE dependencies"): `react-hook-form` and `@hookform/resolvers` added — paired exclusively with Mantine from the moment they landed.
- **Post-Jan 2026**: every new create/edit/delete dialog has shipped Mantine + react-hook-form. Every new detail panel has shipped `.btn` + raw form. **No one drew the boundary; everyone copied the nearest sibling.**

There is no half-finished migration to point at. There is a half-articulated migration — Mantine was adopted with intent but the boundary stopped at "dialogs that I happened to be writing this week."

### Cost of the status quo (concrete)

- **Contributor cost**: every new component requires reading 2–3 siblings to learn which vocabulary applies. This session's `DirectionPanel` regression is the proof: the author did not perform that read, and shipped bare HTML.
- **Token / design-system cost**: the same six grey shades, six spacing values, five radii, four shadows, and four font sizes are defined twice — once as CSS custom properties in `src/index.css` and again in `src/theme/mantine.ts`. Changing the brand grey is a two-file change today. Hard-coded hexes in inline styles (`#E5E7EB`, `#6B7280`, `#B45309`, `#B91C1C`, `#000000`, `#1e40af`, `#3b82f6`) drift past both.
- **CSS bloat**: `src/index.css` contains many feature-specific rules (`.audit-*`, `.capability-*`, `.user-menu-*`, `.error-boundary-*`, two full `.dialog` blocks) that should live next to the components or — preferably — not exist at all. The file ships to every page on first load.
- **Test churn**: Mantine components need `MantineTestWrapper`; `.btn` components do not. Tests in mixed features (`origin-entities`, `capabilities`) have to know which wrapper their component needs. Three places already get this wrong and silently render without `MantineProvider` (warnings in test logs, not failures).
- **Dead code already shipping**: `@dnd-kit/*` (three packages) is in `package.json`, imported nowhere. `App.css` is unused Vite scaffolding. Two `.dialog` definitions in `index.css`. `react-colorful` is one file.

---

## User Personas

| Persona | Needs |
|---|---|
| **Frontend contributor (human or agent)** | Open a new feature folder, write a button, and have it visually match the rest of the app without first investigating the local convention. |
| **Existing user** | Confirmation dialogs, list rows, action buttons, and form inputs look and behave the same wherever they appear in the app. No "this surface looks unfinished" moments. |
| **Designer / EM reviewing UX** | One design system to reason about. Brand-token changes propagate everywhere without per-feature touch-ups. |
| **Reviewer / PR author** | An automated linter rejects bare `<button>`, bare `<input>`, and inline-style soup at PR time, so the `DirectionPanel` regression cannot recur. |

---

## Recommended End-State

**Mantine v8 is the single UI vocabulary** for all interactive primitives (buttons, inputs, selects, checkboxes, radios, modals, popovers, tooltips, badges, tabs, accordions, alerts, sliders), all form composition (`react-hook-form` + `zod` via `@hookform/resolvers`), and all layout primitives (`Stack`, `Group`, `Box`, `SimpleGrid`, `Container`, `Paper`, `Card`).

**`src/index.css` is reduced** to:

1. The CSS variable block (`:root { --color-gray-* / --spacing-* / --font-size-* / --shadow-* / --radius-* }`), kept only because non-Mantine surfaces (canvas nodes, page background, scroll containers) still consume them.
2. App-shell rules that are genuinely outside the Mantine surface: `body`, `.app-container`, `.app-header*`, `.toolbar*`, `.main-content`, `.canvas-section`, `.canvas-container`, `.dockview-theme-light*`. These compose Mantine surfaces but are not Mantine components themselves.
3. ReactFlow / `@xyflow/react` node skins (`.component-node*`, `.capability-node*`, `.origin-entity-node*`, `.react-flow__handle*`). The canvas is a graphics surface, not a form; Mantine is the wrong fit there.

Everything else in `index.css` — `.btn*`, `.dialog*` (both blocks), `.form-input*`, `.form-group*`, `.detail-panel*`, `.empty-state`, all `.audit-*`, all the `.user-menu-*`, all `.error-boundary-*`, `.login-*` — is deleted in tandem with the migrations.

**Feature-scoped `.css` files** survive only where they style genuinely component-specific composition (e.g. `ChatPanel.css` for the chat scroll behaviour, `MaturityAnalysisTab.css` for the analysis layout). Each surviving file consumes the CSS variables; none invent design tokens. The bar to keep one is: "this cannot be expressed with Mantine layout primitives + a single `style={...}` on one element". Most cannot meet that bar.

**Tokens unify** by deleting `src/theme/mantine.ts`'s hard-coded values and reading from the CSS variables via `var(--…)`. The Mantine theme becomes a thin adapter; `:root` is the single source of truth. This direction (CSS-vars as source, Mantine theme as consumer) is chosen because the canvas / app-shell rules cannot read from a JS theme object, while Mantine can read from CSS variables natively.

**`MantineProvider` is the default for tests**: the existing `MantineTestWrapper` becomes the default wrapper inside `renderWithProviders` so every test renders inside `MantineProvider`. No per-test decision about which wrapper to use.

**Dead code is removed in the same pass**: `@dnd-kit/*` packages, `App.css`, the duplicate `.dialog` block, and any unbacked class names found during migration.

---

## Style System Invariants

These hold across the codebase once the overhaul completes. Each is mechanically checkable.

1. **No bare interactive HTML in source `.tsx`.** No `<button>`, `<input>`, `<select>`, `<textarea>`, `<form>`, `<dialog>`, `<fieldset>`, `<legend>` outside `src/components/canvas/**` and `src/test/**`. The lint rule rejects them. Mantine's `Button`, `TextInput`, `Select`, `Textarea`, `Modal`, etc., are the only allowed renderings.
2. **No `style={{}}` prop in source `.tsx`** except (a) ReactFlow node positioning, (b) dynamic values that genuinely cannot live in CSS (computed pixel offsets, runtime-colour from data). Static styling goes to Mantine props (`p`, `gap`, `c`, `fz`) or a feature-scoped `.module.css`.
3. **No hard-coded hex colours, pixel sizes, or rem values in `.tsx` files.** All design tokens come from CSS variables (`var(--color-gray-500)`) or Mantine theme references (`c="gray.5"`).
4. **No `className="btn*"`, `className="dialog*"`, `className="form-input*"`, `className="form-group*"` anywhere in `src/`** once the migration completes. The rules backing them are removed from `index.css` in the same slice that removes the last usage.
5. **All feature-scoped `.css` files use CSS variables, never literal colours/spacings.** Enforced by a stylelint rule.
6. **Tests render inside `MantineProvider` by default.** `renderWithProviders` wraps every render; no per-test decision.
7. **The Mantine theme reads from CSS variables, not from JS literals.** Brand colour changes are a one-line edit in `:root`.

---

## Acceptance Criteria

- [x] `src/theme/mantine.ts` references CSS variables (`var(--…)`) for colours, spacing, radii, shadows; no duplicated hex / rem literals.
- [x] `src/App.css` deleted.
- [x] `@dnd-kit/core`, `@dnd-kit/sortable`, `@dnd-kit/utilities` removed from `package.json` and from `vite.config.ts` manualChunks.
- [x] `src/components/shared/ConfirmationDialog` is a thin Mantine wrapper used by every confirm-y action across the app; no other confirmation primitives survive.
- [x] An ADR-style note in the frontend skill reflects the single-vocabulary world. (The skill was split into `easi-frontend-styling` and `easi-frontend-data`; old `easi-frontend-patterns` removed.)
- [ ] `src/index.css` ≤ 600 lines and contains only: `:root` token block, body/app-shell rules, ReactFlow / canvas node skins, dockview overrides. No `.btn*`, no `.dialog*`, no `.form-input*`, no `.form-group*`, no `.detail-panel*`, no per-feature class blocks.
- [ ] Zero source `.tsx` files (excluding `src/components/canvas/**` and tests) contain `<button>`, `<input>`, `<select>`, `<textarea>`, `<form>`, `<fieldset>`, `<legend>` as raw elements. A lint rule enforces this and is part of `npm run lint`.
- [ ] Zero `className="btn"` / `className="btn-*"` usages remain in `src/`. The lint rule rejects them.
- [ ] Zero `className="dialog"` / `className="dialog-*"` usages remain in `src/`. The `.dialog` rules in `index.css` are deleted.
- [ ] All `Create*Dialog`, `Edit*Dialog`, `Delete*Dialog`, `Add*Dialog`, `Confirmation*Dialog`, `*Modal` components in `src/features/**` and `src/components/shared/**` use Mantine `Modal`.
- [ ] `renderWithProviders` wraps every render in `MantineProvider` by default; per-component `MantineTestWrapper` imports are removed.
- [ ] Each migration slice ships with an E2E screenshot diff against the pre-migration screenshot showing the surface is visually equivalent or intentionally improved.
- [ ] CodeScene `pre_commit_code_health_safeguard` passes on every modified file in every slice. (Passing for Slices 0–2; re-verify per slice.)

---

## Migration Plan (Vertical Slices)

Each slice is independently shippable, independently reviewable, and does not require freezing feature work on unrelated surfaces. Slices are ordered by *user-visible payoff first, lowest-risk first within that*. Total estimated effort is ~3–4 weeks of one engineer, but no slice individually is more than a few days.

### Slice 0 — Plumbing (no UI change)

- Make `src/theme/mantine.ts` read tokens from CSS variables via `var(--…)`. Single source of truth: `:root` in `index.css`.
- Move `MantineProvider` wrapping into `renderWithProviders`; delete per-test `MantineTestWrapper` imports.
- Delete `src/App.css`. Delete `@dnd-kit/*` from `package.json`.
- Add lint rules (forbid bare interactive HTML, forbid `className="btn*"`, forbid hex literals in `.tsx`) gated to *warning* until each subsequent slice silences them by region.
- Remove the second `.dialog` block at `src/index.css:1714` (the orphan); audit which file relied on which definition, fix the loser.

### Slice 1 — Shared Confirmation Dialog

- Rewrite `src/components/shared/ConfirmationDialog.tsx` as a Mantine `Modal` + `Button` composition. Same props, same call sites.
- Verify every call site (`grep -r "ConfirmationDialog" frontend/src/`) renders correctly. This is the single highest-leverage slice: 14 dialog surfaces inherit the new chrome at once.
- Delete the `.dialog*` rules in `index.css` *not* used by other surfaces; flag the ones still in use for their owning slice.

### Slice 2 — `architecture-direction` (the regression that started this)

- Rewrite `DirectionPanel.tsx`, `DirectionStatusBadge.tsx`, `CaptureDirectionForm.tsx` in Mantine.
- Delete `DirectionPanel.css`.
- Promote `CaptureDirectionForm` to react-hook-form + zod for consistency with the rest of the Mantine-side forms (`CreateCapabilityDialog` is the reference).
- Tests: the existing `CaptureDirectionForm.test.tsx` and `DirectionPanel.test.tsx` continue to pass under the new render (now Mantine-wrapped by default per Slice 0).
- This slice doubles as a worked example for the rest.

### Slice 3 — `enterprise-architecture` detail panels and page chrome

- `EnterpriseCapabilityDetailPanel`, `MaturityGapDetailPanel`, `MaturityAnalysisTab`, `StrategicFitTab`, `TimeSuggestionsTab`, `EnterpriseArchHeader`, `DomainCapabilityDockPanel`, `EnterpriseCapabilitiesTable`, `EnterpriseCapabilityCard`, `EnterpriseCapabilitiesEmptyState`, `CreateEnterpriseCapabilityModal`.
- Largest single slice (8+ files, 434 lines of `EnterpriseArchPage.css`). May split into 3a (panels) and 3b (page chrome) if a single PR gets unwieldy.
- After this slice, `src/features/enterprise-architecture/pages/EnterpriseArchPage.css` should be ≤ 100 lines (canvas-adjacent layout only).

### Slice 4 — `business-domains` detail panels

- Migrated to Mantine: `DetailsSidebar`, `DomainsSidebar`, `DomainForm` (now `react-hook-form` + zod via `lib/schemas/businessDomain.ts`), `DomainDialogs` (Mantine `Modal`, no more native `<dialog>` ref), `DomainCard`, `DomainList`, `CapabilityExplorer`, `CapabilityExplorerSidebar`, `StrategicImportanceSection`, `PageLoadingStates`, `DepthSelector` (Mantine `SegmentedControl`), `ShowApplicationsToggle` (Mantine `Switch`), `ApplicationChip` + `ApplicationChipList`, `VisualizationArea`, `NestedCapabilityGrid`, `grid/CapabilityItem`, `DockviewBusinessDomainsLayout`. Feature CSS replaced by four small `*.module.css` files for the visualization tile, drop-zone, chip, and explorer surfaces; `visualization.css` deleted.
- Deleted as dead code (no consumers reachable from `BusinessDomainsRouter`): `DomainDetailPage`, `CapabilityAssociationManager`, `CapabilitySelectorModal`, `CapabilityTagList`, `ReassignConfirmDialog`, `DomainFilter` (the navigation feature owns the live `DomainFilter`), `CapabilityDetailPanel`, `ViewModeToggle`. `useDomainDialogManager` was simplified — no more `dialogRef` or imperative `showModal`/`close` effect.
- `DockviewToolbar.tsx` is kept on the global `.toolbar*` app-shell classes (per Decision 6 / Slice 8 sweep); not migrated here.

### Slice 5 — `origin-entities` and `relations` detail panels

- Migrated to Mantine: `AcquiredEntityDetails`, `InternalTeamDetails`, `VendorDetails`, `OriginRelationshipDetails`, `OriginEntityDetailsPanel` (loading/error states), `RelationDetails`, `RealizationDetailsContent`, `RealizationActions`, `RealizationLevelBadge`, `OriginBadge`, `InheritedRealizationInfo`. Shared chrome (`OriginEntityActions` + `OriginEntityRelationshipsList`) extracted to `origin-entities/components/OriginEntityPanelChrome.tsx` so the three origin-entity panels stop duplicating the action bar and applications list. Dialogs in both features were already Mantine and untouched.
- `DetailField` shared primitive kept on its existing `.detail-*` classes for now — used by `components/`, `capabilities/`, and the migrated panels. Slice 6 / Slice 8 retires those rules. `.btn*`, `.detail-panel/header/title/content/actions/loading/error/info/date/id`, `.realization-list/item/name/level`, `.level-badge`, `.origin-badge/direct/inherited`, `.relation-type-*`, `.reference-link/icon`, `.origin-relationship-type` are now unused in these two features.

### Slice 6 — `components` and `capabilities` remaining detail surfaces

- Migrated to Mantine: `ComponentDetails`, `ComponentFitScores`, `ComponentOriginsSection`, `CapabilityDetails`, `RealizationFitContext`, and the shared `DetailField` primitive (the latter retired the `.detail-field/-label/-value` rules, which is what Slice 5 had deferred). Realization rows render as Mantine `Group` + `Badge`; fit-score rows as `Paper` with Mantine `Button` + `Textarea` + `ColorSwatch`; origins render as `Stack` + `Divider` + `Text`; capability badges (level, maturity, tags) all use Mantine `Badge` with theme colors. Feature CSS retired: `ComponentFitScores.css`, `RealizationFitContext.css`. `index.css` lost ~240 lines: the entire `.detail-*` block, `.realization-*` block, `.tag-list/-badge`, `.expert-list/-item/-contact`, `.level-badge`, `.maturity-badge`, `.badge-genesis/-custom-build/-product/-commodity/-default`, `.type-link`, `.origin-direct/-inherited`.
- `index.css` is now 2,642 lines (down from 2,865). Slice 8 sweeps what remains.

### Slice 7 — `invitations`, `users`, `importing`, `edit-grants`, `releases`, `auth`

- Largely list views, tables, and the login page. Migrate to Mantine `Table`, `Card`, `Paper`, `TextInput`, `PasswordInput`, `Button`.
- `ImportDialog` and its four step components consolidate onto the Mantine `Stepper`.
- `LoginPage` rewritten in Mantine; `.login-*` rules deleted.

### Slice 8 — Sweep and enforce

- Delete every remaining `.btn*`, `.dialog*`, `.form-input*`, `.form-group*`, `.detail-panel*`, `.empty-state`, `.audit-*`, `.user-menu-*`, `.error-boundary-*`, `.login-*` rule from `index.css`.
- Promote the lint rules from *warning* to *error*. CI now rejects bare interactive HTML, `btn*` / `dialog*` class names, and hex literals in `.tsx`.
- Verify `index.css` ≤ 600 lines.
- Update `easi-frontend-patterns/SKILL.md`: replace the "two vocabularies" section with the single-vocabulary rules and link to this spec.

---

## What We Are NOT Doing

- **Not replacing Mantine.** Mantine v8 is the chosen target. Tailwind, Chakra, Radix, MUI, Ant Design, headless-only systems, and CSS-in-JS libraries are out of scope. The investment in `src/theme/mantine.ts`, `MantineTestWrapper`, and the 33 existing Mantine surfaces is preserved and extended.
- **Not migrating canvas (`src/features/canvas/**` and `src/components/canvas/**`) node skins to Mantine.** ReactFlow nodes are graphics, not form controls; Mantine is the wrong fit. The `.component-node`, `.capability-node`, `.origin-entity-node`, `.react-flow__handle` rules stay in `index.css`. The canvas *toolbar* (`AutoLayoutButton`, `DynamicModeToolbar`, etc.) is already Mantine and stays Mantine.
- **Not touching dockview's own DOM (`.dv-*` classes).** Those are dockview-owned; we only theme via `.dockview-theme-light` overrides.
- **Not introducing CSS Modules, styled-components, emotion, or any new styling abstraction.** Mantine's built-in style props + a small number of `*.module.css` files for genuinely composed feature layouts are the only mechanisms.
- **Not changing the brand palette, spacing scale, or typography during this overhaul.** Tokens are unified (CSS variables become the single source) but their *values* do not change. A separate design-pass spec can change values later; this spec is a refactor, not a redesign.
- **Not migrating the `chat/` feature's `ChatPanel.css` in this overhaul.** It styles a scroll/streaming behaviour that is genuinely component-specific. It will consume the CSS variables (Slice 0 ensures the variables remain available) and is otherwise left alone unless it accumulates inline styles later.
- **Not adding a new component library on top of Mantine** (no internal `<EasiButton>` wrapper around Mantine's `Button`). Mantine components are used directly. Wrappers exist only where they encode behaviour, not appearance — e.g. a `HateoasButton` that hides itself when the corresponding `_link` is absent (and renders as a Mantine `Button` otherwise).
- **Not rewriting the existing Mantine surfaces.** They are the target. Slices migrate *to* their conventions.

---

## Architecture

### Token layering

```text
src/index.css :root  →  CSS variables (single source)
        │
        ├──▶  src/theme/mantine.ts  reads via var(--…)  →  MantineProvider theme
        │
        ├──▶  src/index.css canvas / app-shell rules  →  consume var(--…) directly
        │
        └──▶  remaining feature .module.css files  →  consume var(--…) directly
```

The Mantine theme becomes a thin pass-through. The CSS-variable block is owned, audited, and changed in one place.

### Confirmation dialog as a load-bearing primitive

The current `shared/ConfirmationDialog` is the choke point that 14 features route through. Rewriting it in Mantine in Slice 1 lifts every consumer at once without per-call-site changes. This is the highest-leverage move in the plan and the reason Slice 1 sits before any per-feature migration.

### Lint enforcement

- **ESLint** — `react/forbid-elements` configured to forbid bare interactive HTML in `src/**` excluding `src/components/canvas/**` and `src/test/**`. A custom rule forbids `className="btn"` / `className="dialog"` / etc. once each slice silences its usages.
- **Stylelint** — added to the lint pipeline to forbid hard-coded hex / rem / px in feature `.css` files (must use `var(--…)`).
- Rules ship in Slice 0 as *warning*; promoted to *error* in Slice 8 once the last usage is gone.

### Test strategy

- `renderWithProviders` is the single render helper. Wraps in `QueryClientProvider`, `MantineProvider`, `BrowserRouter`. No per-component wrapper decision.
- Each slice ships a Playwright screenshot of the migrated surface, compared against a pre-migration baseline. A surface that comes out visually identical or visually improved passes; one that comes out visually regressed (chrome lost, alignment broken) blocks the slice.
- Existing unit and component tests are *not* rewritten; they pass under the new render. If a test was secretly relying on a `.btn` class for selection, it switches to `role="button"` or `getByRole`.

---

## Design Decisions

1. **Mantine wins, not the `.btn` system.** Three reasons. **(a)** Mantine is already the chosen direction for the loudest UI surface (dialogs and forms) and is paired with the chosen form stack (react-hook-form + zod). Reversing to `.btn` means rewriting that pairing, with no equivalent in the `.btn` world. **(b)** Mantine ships accessibility, keyboard handling, focus traps, ARIA, and disabled-state semantics out of the box; the `.btn` system has none of these and adding them by hand to every CSS variant costs more than migrating the consumers. **(c)** Mantine is mature, actively maintained, and TypeScript-first; the `.btn` system is bespoke and only kept alive by inertia. The 45 `.btn`-using files are mostly thin (a panel header, a few buttons, an empty state) — the migration is wide but not deep.

2. **CSS variables as the single token source, Mantine theme as a consumer.** The alternative — make the Mantine theme JS object the source and have CSS read from generated CSS-vars — is technically possible (Mantine emits CSS vars) but it leaves the canvas rules in an awkward position (they would need to consume Mantine's generated names rather than our own). Keeping `:root` as the source lets every layer — Mantine, canvas, dockview overrides, any remaining feature CSS — consume the same names. Brand changes are a one-file edit either way; this layering is easier to reason about.

3. **`shared/ConfirmationDialog` is migrated before per-feature surfaces.** Highest leverage, lowest risk (one file, ~80 lines), 14 consumer surfaces inherit the new chrome free. If anything breaks it surfaces immediately under one screenshot diff rather than 14.

4. **Slices are vertical (one feature at a time), not horizontal (all buttons app-wide first, then all inputs, …).** Horizontal slices leave the app in an inconsistent state for the entire migration window; vertical slices leave each feature fully consistent at the end of its slice. A user who lands on a migrated feature sees a coherent surface; a user who lands on a not-yet-migrated feature sees the old surface unchanged. Vertical slices are also independently revertible.

5. **Lint rules ship as warnings first, promoted to errors in Slice 8.** Hard-enforcing on day one would block every PR until every slice ships, which defeats the "no freeze on feature work" requirement. Warnings give reviewers a visible signal without blocking; the final promotion to error happens when the last forbidden usage is gone.

6. **Canvas node skins stay as bespoke CSS.** ReactFlow nodes are graphics surfaces with strong layout assumptions; wrapping them in Mantine `Paper` introduces unwanted padding, focus rings, and shadow defaults that fight ReactFlow's own. The boundary "canvas is bespoke; everything else is Mantine" is clear and easy to lint.

7. **No internal wrapper layer (no `<EasiButton>` around `<Button>`).** Wrappers around design-system components create a second design system that drifts. The codebase already has one drifting design system; the spec is not to add a third. Mantine components are used directly; wrappers exist only where they encode HATEOAS-style behaviour (e.g. a button that hides on absent `_links`).

8. **`@dnd-kit/*` removed, not "left for later."** It is in the dependency tree, in the lockfile, in every install, and imported by zero source files. Removing it during the overhaul costs nothing and prevents a future contributor from picking it up unaware that the codebase has not adopted it.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Migrate to Mantine (vs to `.btn`) | 45 `.btn`-using files need rewriting; each gets a `MantineProvider`-wrapped test render | Slice 0 makes `MantineProvider` the default in tests so per-file test rework is one import deletion per file, not a rewrite. The 45 files are mostly small. |
| CSS variables as the single source of truth | Mantine theme cannot use TS type-safety on token names | Tokens are stable (already named, already consumed in two places); the loss is mild and the gain — one-file brand edit — is large. |
| Vertical slices over horizontal | The codebase carries two vocabularies for the duration of the migration | This is the status quo today, so the trade-off is "stay where we are, but with a known end-date" — strict improvement. Lint warnings make the remaining usages visible. |
| Canvas stays bespoke | The "one vocabulary" rule has an explicit exception | The exception is small, scoped to `src/**/canvas/**`, and lintable. Calling it out in the spec means future contributors don't accidentally try to "fix" it. |
| Confirmation dialog rewritten before any feature slice | A bug in the rewrite affects 14 surfaces simultaneously | Slice 1 is one file, ~80 lines, with screenshot diffs against the existing surface. The blast radius is the same as a CSS change to the existing `.dialog` rule — which already happens. |
| Lint rules promoted to error only in Slice 8 | A new feature shipped during Slice 4 could regress and the warning would be ignored | Lint warnings flagged in CI; PR reviewers gate on them. The cost of a missed warning is one extra round in the migration; the cost of error-from-day-one is freezing every PR. |
| No `<EasiButton>` wrapper layer | A future Mantine API change touches every call site rather than one wrapper | Mantine v8 → v9 is one major version; touching call sites with codemods is cheap. Wrappers create their own drift, which is the *originating problem* this spec exists to solve. |

---

## Decision Log — Alternatives Considered and Rejected

**Alt A — Standardise on the `.btn` + feature-CSS system; remove Mantine.**
*Why it loses.* Mantine is the chosen direction for every new dialog and form (33 files, all the create/edit/delete dialogs). Reversing means rewriting all of those *plus* finding a replacement for `react-hook-form` + `zod` integration (no equivalent exists in the `.btn` world). Also reverses accessibility, focus management, and ARIA defaults to "build it yourself per component." Net: more work, lower quality endpoint.

**Alt B — Keep both vocabularies but draw a hard, lint-enforced boundary.** (e.g. "dialogs are Mantine; panels are `.btn`. Lint enforces it.")
*Why it loses.* This codifies the current confusion rather than removing it. Contributors still have to learn "which surface am I on" before writing a button. The `.btn` system still has no accessibility primitives. `src/index.css` still ships 2,865 lines to every page. `src/theme/mantine.ts` and `:root` still duplicate every token. The user's underlying complaint — *why are there so many different ways to do this* — is reaffirmed, not addressed.

**Alt C — Migrate to Tailwind / Chakra / Radix / MUI.**
*Why it loses.* Throws away the existing Mantine investment (33 surfaces, theme, test wrapper, react-hook-form pairing) without a corresponding gain. Mantine v8 is current, mature, TypeScript-first, and ships everything we need. The spec's goal is to reduce vocabularies from three to one, not to swap one of them out for a fourth.

**Alt D — Build an internal `<Easi*>` wrapper layer on top of Mantine.**
*Why it loses.* See Decision 7. The codebase's originating problem is a drifting bespoke layer on top of the platform. Adding another bespoke layer on top of Mantine recreates the problem with a different label. Mantine components are stable enough and well-typed enough to be used directly.

**Alt E — Big-bang migration (one PR, one weekend, every file).**
*Why it loses.* Requires freezing feature work, blocks the team, ships a single PR no reviewer can meaningfully read, and has no incremental fallback if something regresses. Vertical slices ship value at every step.

**Alt F — Do nothing; let new code "naturally" trend Mantine.**
*Why it loses.* This is the de facto policy today, and it produced the `DirectionPanel` regression. Without an enforced lint rule, "new code trends Mantine" is unverifiable; contributors copy the nearest sibling, which is as often `.btn` as Mantine. The accumulated `.btn` codebase is not going to migrate itself.

---

## Checklist

- [x] Specification ready
- [x] Slice 0 — Plumbing
- [x] Slice 1 — Shared ConfirmationDialog
- [x] Slice 2 — `architecture-direction`
- [x] Slice 3 — `enterprise-architecture`
- [x] Slice 4 — `business-domains`
- [x] Slice 5 — `origin-entities` and `relations`
- [x] Slice 6 — `components` and `capabilities`
- [ ] Slice 7 — `invitations`, `users`, `importing`, `edit-grants`, `releases`, `auth`
- [ ] Slice 8 — Sweep + lint enforcement
- [ ] User sign-off after Slice 8
