import { describe, expect, it } from 'vitest';
import {
  createDynamicModeSlice,
  selectAnyDirty,
  selectDirtyForView,
  selectDynamicAdditions,
  selectDynamicDirty,
  selectDynamicPositionDeltas,
  selectDynamicRemovals,
  type DynamicModeActions,
  type DynamicModeState,
} from './dynamicModeSlice';

type Slice = DynamicModeState & DynamicModeActions;

function createStore() {
  let state: Slice;
  const setState = (
    partial: Partial<Slice> | ((s: Slice) => Partial<Slice>),
  ) => {
    const update = typeof partial === 'function' ? partial(state) : partial;
    state = { ...state, ...update };
  };
  const getState = () => state;
  state = createDynamicModeSlice(setState as never, getState as never, {} as never);
  return { getState };
}

function makeSeededStore(initial: Parameters<DynamicModeActions['enterDynamicMode']>[0]) {
  const store = createStore();
  store.getState().enterDynamicMode(initial);
  return store;
}

const singleEntityA: Parameters<DynamicModeActions['enterDynamicMode']>[0] = {
  entities: [{ id: 'A', type: 'component' }],
  positions: { A: { x: 0, y: 0 } },
};

describe('dynamicModeSlice', () => {
  it('starts with no active draft', () => {
    const { getState } = createStore();
    expect(getState().dynamicViewId).toBeNull();
    expect(getState().dynamicEntities).toEqual([]);
    expect(getState().dynamicOriginal).toBeNull();
  });

  it('enterDynamicMode captures original snapshot and seeds current draft', () => {
    const { getState } = createStore();
    const initial = {
      entities: [{ id: 'A', type: 'component' as const }, { id: 'B', type: 'component' as const }],
      positions: { A: { x: 10, y: 20 }, B: { x: 30, y: 40 } },
    };

    getState().enterDynamicMode(initial);

    expect(getState().dynamicOriginal).toEqual(initial);
    expect(getState().dynamicEntities).toEqual(initial.entities);
    expect(getState().dynamicPositions).toEqual(initial.positions);
  });

  it('exitDynamicMode clears the draft and snapshot', () => {
    const { getState } = createStore();
    getState().enterDynamicMode({
      entities: [{ id: 'A', type: 'component' }],
      positions: { A: { x: 0, y: 0 } },
    });

    getState().exitDynamicMode();

    expect(getState().dynamicViewId).toBeNull();
    expect(getState().dynamicOriginal).toBeNull();
    expect(getState().dynamicEntities).toEqual([]);
    expect(getState().dynamicPositions).toEqual({});
  });

  it('draftAddEntities appends new entities and merges positions', () => {
    const { getState } = createStore();
    getState().enterDynamicMode({
      entities: [{ id: 'A', type: 'component' }],
      positions: { A: { x: 0, y: 0 } },
    });

    getState().draftAddEntities(
      [{ id: 'B', type: 'component' }, { id: 'cap-1', type: 'capability' }],
      { B: { x: 100, y: 100 }, 'cap-1': { x: 200, y: 200 } },
    );

    expect(getState().dynamicEntities).toEqual([
      { id: 'A', type: 'component' },
      { id: 'B', type: 'component' },
      { id: 'cap-1', type: 'capability' },
    ]);
    expect(getState().dynamicPositions).toEqual({
      A: { x: 0, y: 0 },
      B: { x: 100, y: 100 },
      'cap-1': { x: 200, y: 200 },
    });
  });

  it('draftAddEntities does not duplicate already-included entities', () => {
    const { getState } = makeSeededStore(singleEntityA);

    getState().draftAddEntities([{ id: 'A', type: 'component' }, { id: 'B', type: 'component' }]);

    expect(getState().dynamicEntities).toEqual([
      { id: 'A', type: 'component' },
      { id: 'B', type: 'component' },
    ]);
  });

  it('draftRemoveEntities removes by id and drops their positions', () => {
    const { getState } = createStore();
    getState().enterDynamicMode({
      entities: [
        { id: 'A', type: 'component' },
        { id: 'B', type: 'component' },
        { id: 'C', type: 'component' },
      ],
      positions: { A: { x: 0, y: 0 }, B: { x: 1, y: 1 }, C: { x: 2, y: 2 } },
    });

    getState().draftRemoveEntities(['A', 'C']);

    expect(getState().dynamicEntities).toEqual([{ id: 'B', type: 'component' }]);
    expect(getState().dynamicPositions).toEqual({ B: { x: 1, y: 1 } });
  });

  it('draftSetPosition updates a single entity position', () => {
    const { getState } = createStore();
    getState().enterDynamicMode({
      entities: [{ id: 'A', type: 'component' }],
      positions: { A: { x: 0, y: 0 } },
    });

    getState().draftSetPosition('A', 50, 60);

    expect(getState().dynamicPositions.A).toEqual({ x: 50, y: 60 });
  });

  it('draftSetPositions bulk-updates positions', () => {
    const { getState } = createStore();
    getState().enterDynamicMode({
      entities: [{ id: 'A', type: 'component' }, { id: 'B', type: 'component' }],
      positions: { A: { x: 0, y: 0 }, B: { x: 0, y: 0 } },
    });

    getState().draftSetPositions({ A: { x: 10, y: 20 }, B: { x: 30, y: 40 } });

    expect(getState().dynamicPositions).toEqual({
      A: { x: 10, y: 20 },
      B: { x: 30, y: 40 },
    });
  });

  it('starts with all edge and entity-type filters enabled', () => {
    const { getState } = createStore();
    expect(getState().dynamicFilters.edges).toEqual({
      relation: true, realization: true, parentage: true, origin: true,
    });
    expect(getState().dynamicFilters.types).toEqual({
      component: true, capability: true, originEntity: true,
    });
  });

  it('draftSetEdgeFilter toggles a single edge type', () => {
    const { getState } = createStore();
    getState().draftSetEdgeFilter('realization', false);

    expect(getState().dynamicFilters.edges.realization).toBe(false);
    expect(getState().dynamicFilters.edges.relation).toBe(true);
  });

  it('draftSetTypeFilter toggles a single entity type', () => {
    const { getState } = createStore();
    getState().draftSetTypeFilter('originEntity', false);

    expect(getState().dynamicFilters.types.originEntity).toBe(false);
    expect(getState().dynamicFilters.types.component).toBe(true);
  });
});

describe('dynamicModeSlice — diff selectors', () => {
  it('selectDynamicAdditions returns entities present in current but not original', () => {
    const { getState } = makeSeededStore(singleEntityA);
    getState().draftAddEntities([{ id: 'B', type: 'component' }, { id: 'cap-1', type: 'capability' }]);

    expect(selectDynamicAdditions(getState())).toEqual([
      { id: 'B', type: 'component' },
      { id: 'cap-1', type: 'capability' },
    ]);
  });

  it('selectDynamicRemovals returns ids present in original but not current', () => {
    const { getState } = createStore();
    getState().enterDynamicMode({
      entities: [{ id: 'A', type: 'component' }, { id: 'B', type: 'component' }],
      positions: { A: { x: 0, y: 0 }, B: { x: 0, y: 0 } },
    });
    getState().draftRemoveEntities(['A']);

    expect(selectDynamicRemovals(getState())).toEqual([{ id: 'A', type: 'component' }]);
  });

  it('selectDynamicPositionDeltas returns only positions that differ from original', () => {
    const { getState } = createStore();
    getState().enterDynamicMode({
      entities: [{ id: 'A', type: 'component' }, { id: 'B', type: 'component' }],
      positions: { A: { x: 0, y: 0 }, B: { x: 5, y: 5 } },
    });
    getState().draftSetPosition('A', 100, 200);
    getState().draftSetPosition('B', 5, 5);

    expect(selectDynamicPositionDeltas(getState())).toEqual({ A: { x: 100, y: 200 } });
  });

  it('selectDynamicPositionDeltas includes positions for newly-added entities', () => {
    const { getState } = createStore();
    getState().enterDynamicMode({
      entities: [{ id: 'A', type: 'component' }],
      positions: { A: { x: 0, y: 0 } },
    });
    getState().draftAddEntities([{ id: 'B', type: 'component' }], { B: { x: 50, y: 50 } });

    expect(selectDynamicPositionDeltas(getState())).toEqual({ B: { x: 50, y: 50 } });
  });

  it('selectDynamicDirty is false right after enterDynamicMode', () => {
    const { getState } = createStore();
    getState().enterDynamicMode({
      entities: [{ id: 'A', type: 'component' }],
      positions: { A: { x: 0, y: 0 } },
    });

    expect(selectDynamicDirty(getState())).toBe(false);
  });

  it('selectDynamicDirty is true after any addition, removal, or position change', () => {
    const { getState } = createStore();
    const init = {
      entities: [{ id: 'A', type: 'component' as const }],
      positions: { A: { x: 0, y: 0 } },
    };

    getState().enterDynamicMode(init);
    getState().draftAddEntities([{ id: 'B', type: 'component' }]);
    expect(selectDynamicDirty(getState())).toBe(true);

    getState().exitDynamicMode();
    getState().enterDynamicMode(init);
    getState().draftRemoveEntities(['A']);
    expect(selectDynamicDirty(getState())).toBe(true);

    getState().exitDynamicMode();
    getState().enterDynamicMode(init);
    getState().draftSetPosition('A', 9, 9);
    expect(selectDynamicDirty(getState())).toBe(true);
  });
});

describe('dynamicModeSlice — per-view drafts', () => {
  const initA: Parameters<DynamicModeActions['enterDynamicMode']>[0] = {
    entities: [{ id: 'A', type: 'component' }],
    positions: { A: { x: 0, y: 0 } },
  };
  const initB: Parameters<DynamicModeActions['enterDynamicMode']>[0] = {
    entities: [{ id: 'X', type: 'component' }],
    positions: { X: { x: 100, y: 100 } },
  };

  it('enterDynamicMode populates draftsByView under the given viewId', () => {
    const { getState } = createStore();

    getState().enterDynamicMode(initA, 'view-a');

    expect(getState().draftsByView['view-a']).toEqual({
      original: initA,
      entities: initA.entities,
      positions: initA.positions,
      filters: getState().dynamicFilters,
      relations: [],
    });
  });

  it('draftAddEntities writes through to draftsByView under dynamicViewId', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');

    getState().draftAddEntities([{ id: 'B', type: 'component' }], { B: { x: 5, y: 5 } });

    expect(getState().draftsByView['view-a'].entities).toEqual([
      { id: 'A', type: 'component' },
      { id: 'B', type: 'component' },
    ]);
    expect(getState().draftsByView['view-a'].positions).toEqual({
      A: { x: 0, y: 0 },
      B: { x: 5, y: 5 },
    });
  });

  it('draftSetPosition, draftRemoveEntities, draftSetPositions all update draftsByView', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');

    getState().draftSetPosition('A', 99, 99);
    expect(getState().draftsByView['view-a'].positions.A).toEqual({ x: 99, y: 99 });

    getState().draftAddEntities([{ id: 'B', type: 'component' }], { B: { x: 1, y: 1 } });
    getState().draftSetPositions({ B: { x: 7, y: 7 } });
    expect(getState().draftsByView['view-a'].positions.B).toEqual({ x: 7, y: 7 });

    getState().draftRemoveEntities(['B']);
    expect(getState().draftsByView['view-a'].entities).toEqual([{ id: 'A', type: 'component' }]);
    expect(getState().draftsByView['view-a'].positions).toEqual({ A: { x: 99, y: 99 } });
  });

  it('draftSetEdgeFilter and draftSetTypeFilter mirror into draftsByView', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');

    getState().draftSetEdgeFilter('relation', false);
    getState().draftSetTypeFilter('capability', false);

    expect(getState().draftsByView['view-a'].filters.edges.relation).toBe(false);
    expect(getState().draftsByView['view-a'].filters.types.capability).toBe(false);
  });

  it('stashCurrentDraft copies active scalars into draftsByView', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');
    getState().draftAddEntities([{ id: 'B', type: 'component' }], { B: { x: 9, y: 9 } });

    getState().stashCurrentDraft('view-a');

    expect(getState().draftsByView['view-a'].entities).toContainEqual({ id: 'B', type: 'component' });
    expect(getState().draftsByView['view-a'].positions.B).toEqual({ x: 9, y: 9 });
  });

  it('hydrateDraftForView restores entities, positions, filters into scalars', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');
    getState().draftAddEntities([{ id: 'B', type: 'component' }], { B: { x: 50, y: 50 } });
    getState().draftSetEdgeFilter('relation', false);

    getState().enterDynamicMode(initB, 'view-b');
    expect(getState().dynamicEntities).toEqual(initB.entities);

    const hydrated = getState().hydrateDraftForView('view-a');

    expect(hydrated).toBe(true);
    expect(getState().dynamicViewId).toBe('view-a');
    expect(getState().dynamicEntities).toEqual([
      { id: 'A', type: 'component' },
      { id: 'B', type: 'component' },
    ]);
    expect(getState().dynamicPositions.B).toEqual({ x: 50, y: 50 });
    expect(getState().dynamicFilters.edges.relation).toBe(false);
  });

  it('hydrateDraftForView returns false when no entry exists', () => {
    const { getState } = createStore();

    expect(getState().hydrateDraftForView('view-x')).toBe(false);
    expect(getState().dynamicViewId).toBeNull();
  });

  it('discardDraftForView removes the entry; clears scalars when active', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');
    getState().draftAddEntities([{ id: 'B', type: 'component' }]);

    getState().discardDraftForView('view-a');

    expect(getState().draftsByView['view-a']).toBeUndefined();
    expect(getState().dynamicEntities).toEqual([]);
    expect(getState().dynamicViewId).toBeNull();
  });

  it('discardDraftForView leaves scalars alone when discarding a non-active view', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');
    getState().draftAddEntities([{ id: 'B', type: 'component' }]);
    getState().enterDynamicMode(initB, 'view-b');

    getState().discardDraftForView('view-a');

    expect(getState().draftsByView['view-a']).toBeUndefined();
    expect(getState().dynamicViewId).toBe('view-b');
    expect(getState().dynamicEntities).toEqual(initB.entities);
  });

  it('exitDynamicMode clears scalars but does not clear draftsByView', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');
    getState().draftAddEntities([{ id: 'B', type: 'component' }]);

    getState().exitDynamicMode();

    expect(getState().dynamicViewId).toBeNull();
    expect(getState().draftsByView['view-a']).toBeDefined();
  });

  it('selectDirtyForView returns true for stashed dirty drafts even after switching away', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');
    getState().draftAddEntities([{ id: 'B', type: 'component' }]);
    getState().enterDynamicMode(initB, 'view-b');

    expect(selectDirtyForView(getState(), 'view-a')).toBe(true);
    expect(selectDirtyForView(getState(), 'view-b')).toBe(false);
  });

  it('selectDirtyForView reads active scalars when viewId === dynamicViewId', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');

    expect(selectDirtyForView(getState(), 'view-a')).toBe(false);

    getState().draftSetPosition('A', 50, 50);

    expect(selectDirtyForView(getState(), 'view-a')).toBe(true);
  });

  it('selectAnyDirty is true if any stashed draft is dirty', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');
    getState().draftAddEntities([{ id: 'B', type: 'component' }]);
    getState().enterDynamicMode(initB, 'view-b');

    expect(selectAnyDirty(getState())).toBe(true);
  });

  it('selectAnyDirty is false when no view is dirty', () => {
    const { getState } = createStore();
    getState().enterDynamicMode(initA, 'view-a');
    getState().enterDynamicMode(initB, 'view-b');

    expect(selectAnyDirty(getState())).toBe(false);
  });
});
