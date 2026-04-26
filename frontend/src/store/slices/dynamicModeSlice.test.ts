import { describe, expect, it } from 'vitest';
import {
  createDynamicModeSlice,
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
  it('starts with dynamic mode disabled and no draft', () => {
    const { getState } = createStore();
    expect(getState().dynamicEnabled).toBe(false);
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

    expect(getState().dynamicEnabled).toBe(true);
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

    expect(getState().dynamicEnabled).toBe(false);
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
