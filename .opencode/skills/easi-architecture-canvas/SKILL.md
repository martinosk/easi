---
name: easi-architecture-canvas
description: MUST load when working on EASI's Architecture Canvas — any code under `frontend/src/features/canvas/`, `frontend/src/features/views/`, or `frontend/src/store/slices/dynamicModeSlice.ts`. Load when adding canvas nodes/edges, changing what appears on a view, building handle interactions, drag-drop, dynamic-view-mode drafts, save / discard flows, or anything that touches canvas state ownership.
compatibility: opencode
---

# EASI Architecture Canvas

The Architecture Canvas is EASI's primary modeling surface — a React Flow canvas that renders Application Components, Capabilities, Origin Entities and the relationships between them. Editing happens in **Dynamic View Mode**, a per-view draft layer that lets the user stage which entities to show and where to place them before committing.

This skill defines the **state-ownership rules** that keep the canvas correct under that model. Get these wrong and you produce silent data loss bugs (orphan entities, dropped relations, stale edges).

---

## Iron Rule: Model is Immediate, Only Views Are Draftable

Two kinds of state coexist on the canvas. Conflate them and the system breaks.

| Kind | Examples | Persistence |
|------|----------|-------------|
| **Model** — facts about the world | Components, Capabilities, Origin Entities, parent-child links, realizations, origin links, component-to-component relations, attribute changes | Persist to the backend **at the moment** the user commits the action. No exceptions, no draft. |
| **View** — what the user is looking at *right now* | Which entities a particular view contains, their (x, y) positions on that view, per-view edge / type filters | May be queued in `dynamicModeSlice` and committed on Save (or discarded) |

The mental model that drives this: **the model is shared, persistent truth about the architecture; a view is one user's current angle on that truth**. Discarding a view-draft can never roll back a model change, because the model change was already a public fact at the moment the user clicked Submit.

---

## What Belongs in `dynamicModeSlice`

Only view-membership and per-view positioning state. The slice should expose:

- `dynamicEntities: EntityRef[]` — which entities the draft view shows
- `dynamicPositions: Record<string, Position>` — where each sits on this view
- `dynamicFilters: DynamicFilters` — per-view edge / type toggles

…and the corresponding draft actions (`draftAddEntities`, `draftRemoveEntities`, `draftSetPosition`, `draftSetEdgeFilter`, `draftSetTypeFilter`, `resetDraft`, `enterDynamicMode`, `discardDraftForView`).

**Do not add** anything like:

- `dynamicRelations` / `draftAddRelation` — relations are model state
- `dynamicCapabilities` / `dynamicComponents` with unsaved attributes — attribute edits are model state
- Any "draft parent change" or "pending realization" structure — also model state

If you find yourself reaching for one of these, the design is wrong: move the operation behind an immediate backend mutation hook (`useChangeCapabilityParent`, `useCreateRelation`, etc.) and only record view-placement in the draft.

---

## Edges Are Derived, Never Drafted

Every edge on the canvas — parent, realization, origin, component-relation — is **computed from current model state** at render time by `useCanvasEdges`:

```typescript
return [
  ...createRelationEdges(relationsBetweenCanvasComponents(relations, componentIdsOnCanvas), ctx),
  ...createParentEdges(projection.capabilities, capabilities, ctx),
  ...createRealizationEdges(capabilityRealizations, projection.capabilities, projection.components, ctx),
  ...createOriginRelationshipEdges(originRelationships, originEntityNodeIds, componentIdsOnCanvas, ctx),
];
```

The rule each creator follows is identical: **if the relation exists in the model and both endpoints are visible on the current view, draw the edge**. There is no per-view "show this edge" toggle and no draft-edge concept. As a consequence:

- Hiding an edge in dynamic mode means hiding (or removing from view) one of its endpoints.
- A "draft relation" makes no sense — you cannot draw an edge between an entity that exists and a relation that does not yet, and you cannot persist the relation later because there is nowhere for it to land.
- A relation's existence in the model is sufficient and necessary for it to be rendered.

---

## Worked Example: Handle-Click "Create Related Entity"

User clicks the bottom handle on an L2 capability, picks "Capability (child of)", fills the Create Capability dialog, hits Submit. The orchestrator (`useCreateRelatedEntity.handleEntityCreated`) handles every mode the same way for the model steps and only branches for the view step:

```typescript
// 1. Entity creation already happened inside the dialog (POST /api/v1/capabilities)
//    — that is a model fact, committed before this hook is even called.

// 2. Dispatch the relation immediately. ALWAYS. Regardless of dynamic mode.
//    PATCH /api/v1/capabilities/{newId}/parent  ← model state
await runRegularModePersist(spec, current, dispatchRelation);

if (dynamicViewId) {
  // 3a. View placement goes into the draft — that is the only thing that can be deferred.
  draftAddEntities(
    [{ id: entityId, type: targetTypeToEntityType[current.entry.targetType] }],
    { [entityId]: targetPosition },
  );
  setPending(null);
  return;
}

// 3b. Regular mode — view placement is also immediate.
await safeAddToView(addToView, current.entry.targetType, entityId, targetPosition);
setPending(null);
```

Behavioral guarantees this design produces:

- The new L3 capability and its parent link are visible to every other user on every other view immediately, not after this user happens to click Save.
- Discarding the dynamic-view draft removes the new L3 from *this view* but leaves the entity and its parent link intact in the model — which is correct, because the user committed to creating them.
- The parent edge appears on the canvas as soon as the read model catches up, because both endpoints are now visible on the view and the parent link exists in `useCapabilities()`.

---

## Anti-Pattern: Relation Routed Through the Draft

This shape is what produced the spec-165 orphan-capability bug and is the exact thing this skill exists to prevent:

```typescript
// WRONG — relation queued in the draft store
if (dynamicViewId) {
  draftAddEntities([...], {...});
  draftAddRelation(spec);   // ← model state in a view-scoped draft
  setPending(null);
  return;
}
```

Symptom: the entity gets created in the backend (the dialog already POSTed it), the entity appears on the canvas (the draft places it on the view), but **the relation never reaches the backend**. The save flow only persists view-membership; a draft-stored relation has nowhere to land. Result: a permanent orphan — e.g. an L3 capability with no parent — and no error to the user.

If you see `draftAddX(spec)` for anything that isn't `Entities` (view membership) or `Position`, you are looking at this bug.

---

## Reviewer Checklist

When reviewing canvas / view / dynamic-mode code, fail the change if any of these are true:

- [ ] A field on `dynamicModeSlice.DynamicModeState` represents anything other than view membership, view positions, or per-view filters.
- [ ] A `draftAdd*` action writes data that the save flow does not persist (`saveDraft` in `features/canvas/utils/saveDraft.ts` is the source of truth for what the draft can contain — anything outside its `additions` / `removals` / `positionDeltas` is wrong).
- [ ] A canvas-mutation orchestrator (handle-click create, drag-connect, context-menu create-and-link, etc.) takes a different code path for `dynamicViewId !== null` that defers a model mutation. The dynamic-mode branch may only differ in whether view placement is drafted or immediate.
- [ ] An edge type is "stored" anywhere instead of derived in `useCanvasEdges`.
- [ ] Discarding a draft would undo any change that was already visible to other users / other views.

---

## Reference Files

| Concern | File |
|---------|------|
| Draft store (view state only) | `frontend/src/store/slices/dynamicModeSlice.ts` |
| Save flow (proves what the draft can hold) | `frontend/src/features/canvas/utils/saveDraft.ts` |
| Edge derivation from model | `frontend/src/features/canvas/hooks/useCanvasEdges.ts` |
| Canonical create-related-entity orchestrator | `frontend/src/features/canvas/hooks/useCreateRelatedEntity.ts` |
| Canvas edge creators | `frontend/src/features/canvas/utils/edgeCreators.ts` |
