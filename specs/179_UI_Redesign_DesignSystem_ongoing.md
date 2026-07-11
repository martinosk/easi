# 179 — UI Redesign: Design System, White-Label Skins, Domain Board

> **Status:** ongoing
> **Depends on:** 168 (Mantine vocabulary), 174 (shared CapabilityTree)

---

## Problem Statement

The frontend reads as generated-default UI: a blue→purple gradient header, Tailwind
stock palettes, system fonts, gradient canvas nodes, and raw Dockview chrome. Nothing
about it says "EASI". Beyond aesthetics, the Business Domains page buries the most
valuable view in the product — *which applications realise which capabilities* — behind
a four-panel dock layout that requires selecting one domain at a time.

The mockup `mockups/capability-journey-mockup.html` (its "Now" lens) demonstrates the
target UX: a landscape board of all domains at once, collapsible L1 groups, compact L2
capability cards with realising-application chips, and a detail drawer. Its *visual
identity* is tenant-specific and must NOT be copied; its *interaction
patterns* are the blueprint.

## Goals

1. **A real EASI identity** — typography, palette, and component language chosen for
   this product (an enterprise-architecture instrument), not framework defaults.
2. **White-label skinnability** — chrome colours are a swappable token set so each
   tenant can carry its own brand without code changes.
3. **Domain Board** — rebuild the Business Domains page on the mockup's "Now" pattern.
4. **Coherence** — every page speaks the same design language; all gradients and
   stock-palette styling are eliminated.

---

## Design Direction — "Instrument"

EASI is a survey instrument for application landscapes. The design thesis:

> **Chrome is quiet and achromatic. Colour is reserved for meaning.**

The UI itself uses ink, paper, and hairlines. Saturated colour appears only where it
encodes model semantics (status, realisation, maturity, TIME) — which makes the
landscape readable at a glance and makes tenant skinning trivial (the tenant hue slots
into the small set of brand surfaces without fighting decoration).

### Typography (self-hosted via @fontsource)

| Role | Face | Usage |
|---|---|---|
| Display | **Schibsted Grotesk** | page titles, panel/domain headers, wordmark, nav |
| Body | **Inter** (variable) | all UI text; base size 14px |
| Data mono | **Spline Sans Mono** | application codes, IDs, levels, counts, chips |

The mono face is a first-class citizen: every application chip, level tag (L1…L4), and
count renders in it. That recurring "notation" texture is the signature of the design.

### Token architecture (two layers + skin layer)

All tokens live in `src/theme/tokens.css` (`:root`). `src/theme/mantine.ts` continues
to consume them via `var(--…)` — no literal values in TS.

**Layer 1 — primitives.** A custom cool-graphite neutral ramp replaces Tailwind gray:

```
--n-50 #F7F9FA  --n-100 #EFF2F4  --n-200 #E2E7EB  --n-300 #CBD4DA  --n-400 #9FACB6
--n-500 #71818D --n-600 #55636E  --n-700 #3D4A54  --n-800 #28323B  --n-900 #161E25
```

**Layer 2 — semantic roles** (what components consume):

```
--bg #F2F5F7   --surface #FFFFFF   --well #F7F9FA
--line #E2E7EB --line-strong #C1CCD4
--ink #1B242C  --muted #5C6B77     --faint #8D9BA7

status.positive  #0C7A55 / bg #E1F3EB / line #99D4BC   (standard, done, Full)
status.progress  #B45E06 / bg #FCF0DC / line #EACB8F   (in-flight, Partial, Migrate)
status.neutral   #5C6B77 / bg #EEF2F5 / line #C9D4DC   (idle, Tolerate)
status.future    #5F4FC7 / bg #EDEAFC / line #C0B8EE   (planned, incoming)
status.danger    #B3362F / bg #FBEAE8 / line #EBB4AF   (errors, Eliminate)
```

TIME mapping: I → positive, T → neutral, M → progress, E → danger.

**Layer 3 — skin (the white-label axis).** A skin overrides ONLY:

```
--skin-brand          appbar / brand surface
--skin-on-brand       text+icons on brand surface
--skin-on-brand-muted secondary text on brand surface
--skin-accent-0..9    the Mantine primary ramp (buttons, links, active states)
--skin-focus          focus ring colour
```

Skins are `[data-skin="…"]` blocks in `src/theme/skins.css`, applied by
`applySkin(name)` (`src/theme/skin.ts`) setting `document.documentElement.dataset.skin`,
persisted in localStorage. Shipped skins:

- **`easi`** (default) — graphite: brand `#232B33`, accent ramp = the neutral ramp
  (near-black primary buttons), focus `#3457D5`.
- **`harbor`** (demo) — deep maritime blue `#1F3A5F`: proves a navy-branded tenant
  (e.g. a shipping line) is a 10-line CSS block, without shipping any real tenant's hue.
- **`evergreen`** (demo) — deep spruce `#1E3D33`.

`SessionTenant` already reaches the client; a later spec can map tenant → skin
server-side. This spec ships the mechanism plus a skin picker under Settings →
Appearance (admin/dev affordance).

### Shape & depth

- Radii: xs 4 / sm 6 / md 8 / lg 12 / xl 16; controls default `sm`, cards `lg`.
- Borders carry structure; shadows only on overlays (menus, modals, drawers).
- No gradients anywhere. `linear-gradient` count in `src/` must reach zero.

---

## UX Changes

### App shell

- Replace the gradient header with a flat `--skin-brand` appbar: wordmark "easi" in
  Schibsted Grotesk, quiet nav items (on-brand-muted → on-brand when active, active
  indicated by a 2px underline rule), user menu, tenant name visible.
- Kill the legacy unused `MainLayout` component + styles.
- Dockview is removed from the product entirely (user decision during review — the
  re-skinned dock chrome still read as clutter). The canvas workspace is a fixed
  three-pane layout (`CanvasWorkspace`): Explorer 280px | canvas | Details 350px,
  hairline separators, quiet uppercase pane headers, per-pane visibility toggles
  persisted to localStorage `easi-canvas-panels`. The `dockview` dependency is gone.

### Business Domains → Domain Board (mockup "Now" lens)

- **Board of all domains at once**: responsive card grid (`minmax(440px, 1fr)`), one
  card per business domain, replacing the pick-one-domain dock layout.
- Inside a domain card: **collapsible L1 groups** (header: name, `L1 · n sub` tag,
  distinct-app count, amber-highlighted when > 3 apps). Expanded groups list **L2+
  capability cards**: name, level tag, maturity dot, realising-application chips
  (mono; tinted by `realizationLevel` — Full → positive, Partial → progress, Planned →
  future/dashed; dimmed when `origin: 'Inherited'`), `n apps` flag when multi-realised,
  italic "no realising application mapped" empty state. (TIME / Standard-Application
  badges are per-Enterprise-Capability data in another context — deliberately out of
  scope here; a later spec can join them in.)
- **Toolbar**: search (filters capabilities/apps, auto-expands matches), maturity
  legend, depth is no longer a global toggle — the hierarchy is progressive disclosure.
- **Detail drawer** (right, 440px): replaces the details dock panel. Breadcrumb
  (domain · L1), capability name, realising applications with realisation level +
  origin + notes, strategic importance section, maturity, actions (existing HATEOAS
  gated actions preserved).
- **Assignment UX preserved**: capability explorer becomes a toggleable right rail
  ("Assign capabilities" mode) using the shared `CapabilityTree`; L1 drag onto a domain
  card assigns, exactly as today (same hooks/invariants, `associateCapability`).
- Deep links `/business-domains/:domainId` and `?capability=` keep working (scroll to
  + expand + open drawer).
- Roles: `stakeholder` never sees the assign rail (as today, where explorer is hidden).

### Canvas

- Component nodes: flat `--surface` cards, hairline border, small colour-coded accent
  edge by node kind — no gradient fill, no translate-on-hover.
- Capability nodes: graphite ramp by level (backend colour-scheme override stays).
- Edges/handles/minimap/controls re-tinted to tokens.

### Everything else

- **Login** rebuilt: no gradient wash — quiet `--bg` page, left-aligned wordmark, a
  plain bordered card. First impression = instrument, not SaaS template.
- **Icons**: adopt `@tabler/icons-react` as the single icon language; every emoji
  glyph used as an icon (⛔ 🔒 👁️ 📦 🏢 🏪 👥 📚 ★ ✎ 🗑) is replaced.
- **Settings feature**: its CSS references undefined variables (`--blue-500`,
  `--red-200`, … — Tailwind-style names that were never defined). These are bugs, not
  styling: rewire to real tokens.
- **Radial context menu** keeps its interaction, loses glassmorphism (`backdrop-filter`,
  bouncy cubic-bezier) for flat surface + hairline + shadow-lg.
- Value Streams, Enterprise Architecture, Users, Invitations, Settings, One-Pagers,
  Chat: re-skinned to tokens/typography. No layout rebuilds beyond what the audit
  flags as slop; consistency is the goal.

---

## Business Rules & Invariants

1. Zero `linear-gradient`/`radial-gradient` in `frontend/src`.
2. All colour/spacing/type values flow from `tokens.css`; no new hard-coded hex/px in
   `.tsx` (per `easi-frontend-styling`).
3. Skins may only change layer-3 variables; semantic status colours are NOT skinnable
   (meaning stays stable across tenants).
4. No DFDS identity in the shipped product: mockup hues/fonts are not copied.
5. All existing behaviours preserved: HATEOAS action gating, L1-only domain drag,
   deep links, role visibility, canvas draft mode, `data-testid`s.
6. Mockup interaction patterns adopted; its DOM/CSS is NOT ported — everything is
   rebuilt Mantine-native + CSS modules.

## Acceptance Criteria

- [x] tokens.css + skins.css + skin.ts + fonts shipped; Mantine theme consumes tokens.
- [x] Skin picker in Settings; `easi`, `harbor`, `evergreen` selectable; persists.
- [x] New appbar; gradient header gone; legacy MainLayout deleted.
- [x] Business Domains renders the Domain Board with search, collapsible L1 groups,
      app chips (incl. L1 own realizations), drawer, assignment rail; deep links work;
      board scrolls independently.
- [x] Canvas nodes flat; Dockview removed entirely (fixed three-pane workspace).
- [x] `grep -r "linear-gradient" frontend/src` → 0 hits.
- [x] Build green (tsc + vite), 1673 unit tests passing, lint at main's baseline
      (single pre-existing logo.svg parse error). Code Health pass: see checklist.

## Design Decisions

1. **Achromatic chrome, colour = meaning** — makes tenant skinning structurally cheap
   and the landscape legible. Alternative (brand-accent-everywhere): rejected — it's
   what every generated UI does, and it fights the status palette.
2. **Skins as CSS variable sets** — zero-JS theming, no rebuild per tenant.
   Alternative (Mantine theme objects per tenant): rejected — duplicates the token
   source of truth established by 168.
3. **Dockview dropped everywhere** — originally the plan kept a re-skinned dock on the
   canvas; user review overruled it ("get completely rid of it"). The fixed three-pane
   workspace keeps the panel toggles and per-user visibility persistence, loses drag/
   tab/resize chrome. Simpler code, calmer chrome, one less dependency.
4. **Board rebuilt Mantine-native** — the mockup is vanilla HTML/CSS; porting it
   verbatim would violate the single-vocabulary rule.

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing (1674 tests, 176 files)
- [x] Integration verified (live browser pass against seeded backend + screenshot review;
      Playwright e2e specs compile and target unchanged selectors, full runtime pass pending)
- [x] API documentation updated (no API changes — frontend-only)
- [ ] User sign-off
