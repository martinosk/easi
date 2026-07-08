# 174 — Standardise Capability List-Views

> **Status:** ongoing
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

The inconsistency is confusing to users and costly to maintain (three code paths, two
duplicate tree builders). We standardise all three on the **Value Streams** list-view
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
3. **No Canvas feature loss** — every existing Canvas Explorer behaviour is preserved.
4. **L1-only domain drag** — Business Domains keeps L1-only drag; L2–L4 are expandable
   context, not drop sources.
5. **Per-surface badges** — "Mapped" (Value Streams) and "Assigned" (Business Domains)
   badges are preserved.
6. **Mantine-native** — the shared component uses Mantine v8 primitives and a scoped CSS
   module; no global bespoke `.cap-tree-*` / `.capability-tree-*` classes, no inline
   static styles, no hard-coded design tokens in `.tsx`.

---

## Acceptance Criteria

- [ ] One shared `CapabilityTree` component renders the list on all three surfaces.
- [ ] Search bolds the matched substring on all three surfaces.
- [ ] Business Domains Explorer has a working search box.
- [ ] Canvas multi-select, context menu, maturity dot, custom-colour swatch, selection,
      and "not in view" dimming continue to work.
- [ ] Business Domains drag remains L1-only; Value Streams drag remains unmapped-only.
- [ ] Existing `data-testid`s are preserved; existing tests pass.
- [ ] New/updated unit tests cover render, expand/collapse, filtering, bold highlight,
      and drag predicates.
- [ ] `npm run lint` and `npm run test` pass; per-file Code Health is 10.0.

---

## Architecture

### Ownership

Frontend-only change. No bounded context, API, or persistence impact. Touches three
frontend feature areas: `navigation` (Canvas Explorer), `business-domains`, and
`value-streams`, plus a new shared component under `features/capabilities`.

### Frontend

- New `features/capabilities/components/CapabilityTree/` (component + CSS module + index)
  owns: header, search input (reusing `TreeSearchInput`) with Mantine `Highlight` bold
  matching, tree filtering, auto-expand-on-search, expand/collapse, indentation,
  loading/empty states.
- Extension points via props: `getRowProps` (draggable, handlers, selected, dimmed,
  title), `renderRight` (per-surface badge/indicator slot), controlled or internal
  expansion, `searchFields`, `onVisibleItemsChange` (for Canvas multi-select range).
- Consolidate the two tree builders onto one `useCapabilityTree` hook / `CapabilityTreeNode`
  type; preserve Canvas's orphan-L1-as-root behaviour via a hook option.
- Each surface keeps its own drag serialization, selection, and badge logic in the slots.

---

## Design Decisions

1. **Shared component with slots, not a hard fork** — keeps one layout/search code path
   while letting each surface keep distinct drag/select/badge behaviour. Alternative:
   copy the Value Streams component into each surface (rejected — re-introduces drift).
2. **Mantine `Highlight` for bold matching** — already a dependency, handles substring
   marking. Alternative: hand-rolled string splitting (rejected — reinvents Mantine).
3. **Mantine-native rebuild rather than lifting the bespoke `.cap-tree-*` CSS** — aligns
   with the post-spec-168 single-vocabulary rule. Alternative: promote the global CSS to
   a shared stylesheet (rejected — violates `easi-frontend-styling`).
4. **Canvas keeps every feature (reskin only)** and **Business Domains keeps L1-only
   drag** — per explicit user decision during planning.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Generic shared component with slots | More props/indirection than three bespoke lists | Structured slots + clear prop names; one well-tested component |
| Mantine-native rebuild | Larger diff than lifting existing CSS | Preserve `data-testid`s; cover with unit + e2e tests |

---

## Checklist

- [x] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
