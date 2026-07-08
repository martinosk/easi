# 174 — Standardise Capability List-Views

> **Status:** done
> **Depends on:** _none_

---

## Problem Statement

Capabilities are presented as a list/tree in three different surfaces, each built with
its own controls and layout even though they solve the same problem:

- **Architecture Canvas "Explorer"** (`CapabilitiesSection`) — custom DOM + global CSS,
  rich features (multi-select, maturity colour dot, custom-colour swatch, context menu,
  selection, "not in view" dimming), search over name + description.
- **Business Domains "Explorer"** (`CapabilityExplorer`) — Mantine `Paper`/`Stack`,
  L1-only drag, no expand/collapse, **no search at all**.
- **Value Streams "Capabilities"** (`CapabilitySidebar`) — custom DOM + global CSS,
  expandable tree, drag, search over name, "Mapped" badge.

The inconsistency is confusing to users and costly to maintain (three code paths, three
duplicate tree builders — `treeUtils.buildCapabilityTree`, the `useCapabilityTree` hook,
and `CapabilityExplorer`'s inline copy). We standardise all three on the **Value Streams** list-view
pattern, extracted into one shared component, and enhance search so the matched text is
**emphasised (bold)** within each result rather than only filtering the list.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | A single, predictable capability list across Canvas, Business Domains, and Value Streams, with search that clearly shows *where* the match is. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Standardised capability list-view

  Scenario: Consistent list across all three surfaces
    Given the capability list is shown on Canvas, Business Domains, or Value Streams
    Then it uses the same expandable tree layout, level labels, and search box

  Scenario: Search highlights the matched text
    Given a capability list with a search box
    When I type a query that matches part of a capability name
    Then matching capabilities remain visible with their ancestors
    And the matched substring is rendered in bold within the name
    And capabilities matching only by description are not shown

  Scenario: Business Domains gains search
    Given the Business Domains Explorer
    When I type into its new search box
    Then the capability list filters to matches with the substring bolded

  Scenario: Canvas retains all existing capabilities
    Given the Architecture Canvas Explorer
    Then multi-select, context menu, maturity dot, custom-colour swatch,
      selection highlight, and "not in view" dimming all continue to work

  Scenario: Drag rules are preserved per surface
    Given the Business Domains Explorer
    Then only L1 capabilities are draggable onto a domain
    And on Value Streams any unmapped capability is draggable
```

---

## Business Rules & Invariants

1. **Single component** — all three surfaces render through one shared `CapabilityTree`.
2. **Bold match** — when a search query is active, the matched substring of a capability
   name is visually emphasised (bold) wherever the list is shown.
3. **Name-only search** — search matches capability names only, on all three surfaces.
   Canvas's existing description matching is deliberately removed (Decision 5).
4. **No Canvas feature loss** — every existing Canvas Explorer behaviour is preserved,
   with the single deliberate exception of description matching in search (rule 3).
5. **L1-only domain drag** — Business Domains keeps L1-only drag; L2–L4 are expandable
   context, not drop sources.
6. **Per-surface badges** — "Mapped" (Value Streams) and "Assigned" (Business Domains)
   badges are preserved.
7. **Mantine-native** — the shared component uses Mantine v8 primitives and a scoped CSS
   module; no global bespoke `.cap-tree-*` / `.capability-tree-*` / `.tree-search`
   classes, no inline static styles, no hard-coded design tokens in `.tsx`.

---

## Acceptance Criteria

- [x] One shared `CapabilityTree` component renders the list on all three surfaces.
- [x] Search bolds the matched substring on all three surfaces.
- [x] Search matches capability names only on all three surfaces (Canvas description
      matching removed).
- [x] `TreeSearchInput` lives in `components/shared/` with scoped styles; the global
      `.tree-search` styles are gone and all navigation sections use the new import.
- [x] Business Domains Explorer has a working search box.
- [x] Canvas multi-select, context menu, maturity dot, custom-colour swatch, selection,
      and "not in view" dimming continue to work.
- [x] Business Domains drag remains L1-only; Value Streams drag remains unmapped-only.
- [x] Existing `data-testid`s are preserved; existing tests pass (two sanctioned test
      updates: Canvas description-search test inverted per rule 3; retired
      `.capability-tree-item` class selectors replaced with the new row test ids).
- [x] New/updated unit tests cover render, expand/collapse, filtering, bold highlight,
      and drag predicates.
- [x] Lint is clean and Code Health is 10.0 on every changed file; full test suite
      passes except one pre-existing timeout flake in `CapabilityUIConsistency`
      ("should render all capabilities in tree regardless of view presence"), which
      fails identically at the pre-branch HEAD.

---

## Architecture

### Ownership

Frontend-only change. No bounded context, API, or persistence impact. Touches three
frontend feature areas: `navigation` (Canvas Explorer), `business-domains`, and
`value-streams`, plus a new shared component under `features/capabilities`.

### Frontend

- New `features/capabilities/components/CapabilityTree/` (component + CSS module + index)
  owns: search input with Mantine `Highlight` bold matching, name-only tree filtering,
  auto-expand-on-search, expand/collapse, indentation, loading/empty states. It renders
  no section header — each surface keeps its own chrome (Canvas keeps `TreeSection` with
  count and add button, matching its sibling sections; Value Streams keeps its panel
  header), so Canvas stays visually consistent with the other navigation sections.
- `TreeSearchInput` moves from `features/navigation/components/shared/` to
  `components/shared/` with a scoped CSS module; the global `.tree-search` styles are
  removed from `navigation.css` and the five navigation section consumers migrate to
  the new import.
- Extension points via props: `getRowProps` (draggable, handlers, selected, dimmed,
  title), `renderRight` (per-surface badge/indicator slot), controlled or internal
  expansion, `onVisibleItemsChange` (for Canvas multi-select range).
- Consolidate the three tree builders onto one `useCapabilityTree` hook /
  `CapabilityTreeNode` type. The hook keeps `orphanedL1Ids` and gains an option for
  Canvas's orphan handling: Canvas roots **any** capability whose parent is absent
  (all levels), while Business Domains and Value Streams root only L1s and drop
  non-L1 orphans — both behaviours are preserved.
- Each surface keeps its own drag serialization, selection, and badge logic in the slots.
- Implementation notes: rows are `role="treeitem"` with keyboard activation (Enter/Space)
  and expose `data-draggable` (pinned by Business Domains tests). Canvas rows carry new
  `capability-tree-item-<id>` test ids replacing the retired `.capability-tree-item`
  class selectors in tests. The standardised default expansion is fully collapsed —
  only L1 roots are visible until expanded (per user decision during review); search
  still auto-expands to reveal matches. This replaces Business Domains'
  always-expanded rendering and Value Streams' roots-expanded default. Canvas is
  unaffected: its expansion is controlled by persisted per-view state.

---

## Design Decisions

1. **Shared component with slots, not a hard fork** — keeps one layout/search code path
   while letting each surface keep distinct drag/select/badge behaviour. Alternative:
   copy the Value Streams component into each surface (rejected — re-introduces drift).
2. **Mantine `Highlight` for bold matching** — already a dependency, handles substring
   marking. Must be styled bold-only (transparent background), not the default yellow
   `Mark`. Alternative: hand-rolled string splitting (rejected — reinvents Mantine).
3. **Mantine-native rebuild rather than lifting the bespoke `.cap-tree-*` CSS** — aligns
   with the post-spec-168 single-vocabulary rule. Alternative: promote the global CSS to
   a shared stylesheet (rejected — violates `easi-frontend-styling`).
4. **Canvas keeps every feature (reskin only)** and **Business Domains keeps L1-only
   drag** — per explicit user decision during planning.
5. **Name-only search on all surfaces** — Canvas's description matching is dropped (per
   user decision during spec review): a description-only match has no substring to bold
   in the name and reads as a false positive. Alternative: keep description matching
   with un-bolded rows (rejected — inconsistent and confusing).
6. **Surfaces keep their own section chrome** — on Canvas the Capabilities section must
   stay visually identical to its sibling `TreeSection` sections (Applications, Vendors,
   Internal Teams, Acquired Entities); a shared header would converge across surfaces
   while diverging within Canvas. Alternative: header/count/add props on the shared
   component (rejected).
7. **Promote `TreeSearchInput` to `components/shared/`** with a scoped CSS module.
   Alternative: reuse it in place (rejected — global `.tree-search` class and a
   capabilities→navigation import violate rule 7).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Generic shared component with slots | More props/indirection than three bespoke lists | Structured slots + clear prop names; one well-tested component |
| Mantine-native rebuild | Larger diff than lifting existing CSS | Preserve `data-testid`s; cover with unit + e2e tests |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (frontend-only; all three surfaces
      verified in the browser via mock-API mode)
- [x] API documentation updated (no API changes — frontend-only)
- [x] User sign-off
