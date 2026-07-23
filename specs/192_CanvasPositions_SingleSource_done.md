# 192 — Canvas Element Positions: Single Source of Truth

> **Status:** done
> **Depends on:** 060_ViewLayouts_Context (partially reversed by this spec)

---

## Problem Statement

Canvas element positions are split across two stores. Spec 060 introduced the ViewLayouts bounded context and moved positions into `viewlayouts.element_positions`, but Slice 10 ("Cleanup") was deliberately deferred — *"verify in production first"* — and never completed. The old store, `architectureviews.view_element_positions`, stayed live and writable.

Because both stores remained active, the authoritative one silently changed twice as unrelated work landed on top:

| Period | Component/capability positions written to |
|--------|-------------------------------------------|
| until 2025-11-30 | `architectureviews.view_element_positions` |
| 2025-11-30 → 2026-04-26 | `viewlayouts.element_positions` (spec 060 cutover) |
| 2026-04-26 onward | `architectureviews.view_element_positions` (spec 164 always-on draft mode routed saves through the views API) |

Neither store is authoritative for the whole estate. Which one holds a view's current positions depends on when that view was last edited. Origin entity positions were never migrated and have always lived in ArchitectureViews.

The production symptom is that views whose layout container is empty render every component and capability stacked at the canvas default origin, because the canvas read positions exclusively from ViewLayouts — a store that has received no writes since April 2026. Read-only views show this directly; editable views were masked because draft mode reseeded positions from the view.

This spec makes ArchitectureViews the single owner of canvas element positions and reconciles the split data so no view loses its layout.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect (editor)** | Positions they arrange on a canvas persist and are exactly what they see on reload |
| **Stakeholder (read-only)** | Opening a shared view shows the architecture as its author arranged it, not a pile of overlapping nodes |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Canvas element positions survive from a single store

  Scenario: Read-only viewer opens a view that has no layout container
    Given an architecture view whose elements have positions in ArchitectureViews
    And no layout container exists for that view
    And the user has no edit permission on the view
    When the user opens the view
    Then each component and capability renders at its stored position
    And no two elements are stacked at the canvas default origin

  Scenario: View last arranged while ViewLayouts was authoritative
    Given a view whose element positions were last written to ViewLayouts
    And ArchitectureViews holds older positions for the same elements
    When reconciliation has run
    And the user opens the view
    Then each element renders at the position last chosen by a user
    And no element reverts to its older position

  Scenario: View last arranged while ArchitectureViews was authoritative
    Given a view whose element positions were last written to ArchitectureViews
    And ViewLayouts holds older positions for the same elements
    When reconciliation has run
    And the user opens the view
    Then each element renders at the position last chosen by a user

  Scenario: Editor rearranges and reloads
    Given an editable architecture view
    When the user drags a component to a new position and saves
    And the user reloads the view
    Then the component renders at the new position

  Scenario: Editor applies auto-layout and reloads
    Given an editable architecture view
    When the user applies auto-layout and saves
    And the user reloads the view
    Then every element renders at its auto-layout position

  Scenario: Element with no recorded position
    Given a view containing an element that has no position in either store
    When the user opens the view
    Then the element renders at the canvas default origin
```

---

## Business Rules & Invariants

1. **Single owner** — ArchitectureViews is the sole store of canvas element positions. No canvas read or write path may consult ViewLayouts for a position.
2. **Newest wins** — where both stores hold a position for the same element in the same view, reconciliation keeps the one with the later `updated_at`.
3. **No position is lost** — reconciliation never reduces the set of positioned elements for a view; an element positioned in either store before reconciliation is positioned after it.
4. **Reconciliation is idempotent** — running it more than once produces the same result and never overwrites a position written after it ran.
5. **Tenant isolation holds** — reconciliation operates strictly within a tenant; positions never cross tenants.
6. **Origin entities are unaffected** — their positions already live in ArchitectureViews and must not be rewritten.
7. **Default origin is a last resort** — the canvas default position applies only to elements with no stored position in ArchitectureViews after reconciliation.

---

## Acceptance Criteria

- [x] For every architecture view, `architectureviews.view_element_positions` holds the most recently written position for each of its components and capabilities
- [x] No element that had a position in either store before reconciliation is without one afterwards
- [x] Re-running reconciliation changes no rows
- [x] Reconciliation touches no rows belonging to another tenant
- [x] Origin entity position rows are byte-identical before and after reconciliation
- [x] The canvas renders component, capability and origin entity positions from the view payload alone
- [x] No canvas code path issues a request to `/api/v1/layouts/**`
- [x] The ViewLayouts tables, schema, API and Go context are removed
- [x] A read-only view with no layout container renders its elements at their stored positions
- [x] Dragging and saving, then reloading, shows the saved positions
- [x] Applying auto-layout and saving, then reloading, shows the auto-layout positions
- [x] Unit tests cover the view-payload position source, missing-position default, and draft precedence

---

## Architecture

### Ownership

ArchitectureViews owns canvas element positions outright after this change, and the ViewLayouts bounded context is retired in the same release. The Architecture Canvas stops depending on it entirely, the Business Domain grid stopped persisting positions with spec 179, and no other consumer exists — so its Go context (API, event handlers, repositories, domain model), its tables and its schema are removed.

### Domain Model

No new aggregates, value objects or domain events. This is a consolidation of an existing read model onto one owner. The ArchitectureViews view aggregate already carries element positions and already exposes them on its read model.

### API Surface

No new or changed endpoints. The canvas already receives element positions on the view resource; it simply stops issuing a second request to the layouts resource. Permission model is unchanged — position writes continue to go through the existing view element endpoints and their existing permission checks.

### Persistence

A one-time reconciliation migration (131) folds ViewLayouts positions into ArchitectureViews for `architecture-canvas` containers. Selection is newest-wins on each store's per-element `updated_at`; rows present in only one store are carried across, with the element type derived from the owning read model. The migration must be safe to re-run and must respect tenant boundaries.

A follow-up migration (132) drops `viewlayouts.element_positions`, `viewlayouts.layout_containers` and the `viewlayouts` schema. Migration ordering guarantees reconciliation completes before the drop.

### Frontend

The Architecture Canvas derives every node position from the view payload. This applies identically whether or not a draft is active, so entering and leaving draft mode never shifts a node. Canvas position writes — drag, auto-layout, drop — all target the view.

### Cross-Context Integration

None added. The ViewLayouts subscriptions to `ApplicationComponentDeleted`, `CapabilityDeleted`, `BusinessDomainDeleted` and `ViewDeleted` are removed together with the context; the equivalent cleanup for `view_element_positions` already lives in ArchitectureViews.

---

## Design Decisions

1. **ArchitectureViews wins over ViewLayouts as the owner** — every write path has targeted ArchitectureViews since April 2026, origin entity positions never left it, and it is the store the view resource already serves to the client. Consolidating there requires no API change and no new client request. Alternative considered: complete spec 060 forward and make ViewLayouts authoritative (rejected — it would require repointing every write path, keeping both stores in sync through rollout, and the separation it was built for no longer has a second consumer).

2. **Reconcile by newest-wins on `updated_at` rather than picking a store per view** — the authoritative store changed mid-life, so a per-view rule based on dates would misclassify any view edited across a boundary. Per-element timestamps are already recorded in both tables and give the correct answer without guessing. Alternative considered: treat ViewLayouts as authoritative wherever a container has elements (rejected — it would discard every drag-save made since April 2026).

3. **Reconcile the data before changing the read path, and ship them together** — changing the canvas to read from ArchitectureViews while the middle-window cohort still has its current positions only in ViewLayouts would visibly revert those views. The migration and the frontend change are a single deployable unit. Alternative considered: have the client merge both stores at read time (rejected — it makes every canvas load pay for a second request and leaves the split permanently).

4. **Retire the ViewLayouts context in the same release** — originally this spec left the context in place and proposed retirement separately, but with the canvas repointed and the Business Domain grid persisting nothing since spec 179, no consumer references the context or its tables. Leaving a live, writable-but-unread store is exactly the failure mode spec 060's outcome documents, so the context code, tables and schema go now (decided 2026-07-23). The `business-domain-grid` rows are dead data since 2026-07-11 and are dropped with the tables.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Reconcile rather than dual-read | A view edited between the migration running and the new frontend reaching the user could write to the old path | Ship migration and frontend in the same release; reconciliation is idempotent and can be re-run |
| Drop the ViewLayouts tables | The pre-reconciliation position history is gone; migrations are forward-only | Reconciliation (131) preserves every newest position before the drop (132); migration ordering guarantees the sequence |
| Newest-wins on `updated_at` | Clock skew between writes could pick the wrong row in a near-tie | Windows are months apart in practice; ties are visually indistinguishable |
| ArchitectureViews owns presentation data | Reverses spec 060's domain/presentation split | The split's only other consumer is gone; revisit if a second layout consumer appears |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant — none; the reconciliation is a one-shot data migration, validated at deploy time
- [x] API documentation updated — layouts endpoints removed from the generated docs; `PATCH /views/{id}/layout` route annotation corrected
- [x] User sign-off (2026-07-23)
