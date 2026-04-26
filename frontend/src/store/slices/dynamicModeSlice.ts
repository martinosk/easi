import type { StateCreator } from 'zustand';
import type { DynamicFilters, EdgeType, EntityRef, EntityType } from '../../features/canvas/utils/dynamicMode';

export type Position = { x: number; y: number };

export interface DynamicModeSnapshot {
  entities: EntityRef[];
  positions: Record<string, Position>;
}

export interface DraftEntry {
  original: DynamicModeSnapshot;
  entities: EntityRef[];
  positions: Record<string, Position>;
  filters: DynamicFilters;
}

export interface DynamicModeState {
  dynamicOriginal: DynamicModeSnapshot | null;
  dynamicViewId: string | null;
  dynamicEntities: EntityRef[];
  dynamicPositions: Record<string, Position>;
  dynamicFilters: DynamicFilters;
  draftsByView: Record<string, DraftEntry>;
}

export interface DynamicModeActions {
  enterDynamicMode: (initial: DynamicModeSnapshot, viewId?: string | null) => void;
  exitDynamicMode: () => void;
  resetDraft: () => void;
  draftAddEntities: (entities: EntityRef[], positions?: Record<string, Position>) => void;
  draftRemoveEntities: (ids: string[]) => void;
  draftSetPosition: (id: string, x: number, y: number) => void;
  draftSetPositions: (updates: Record<string, Position>) => void;
  draftSetEdgeFilter: (edge: EdgeType, enabled: boolean) => void;
  draftSetTypeFilter: (type: EntityType, enabled: boolean) => void;
  stashCurrentDraft: (viewId: string) => void;
  hydrateDraftForView: (viewId: string) => boolean;
  discardDraftForView: (viewId: string) => void;
}

const defaultFilters: DynamicFilters = {
  edges: { relation: true, realization: true, parentage: true, origin: true },
  types: { component: true, capability: true, originEntity: true },
};

type SliceState = DynamicModeState & DynamicModeActions;

function snapshotEntry(s: DynamicModeState): DraftEntry | null {
  if (!s.dynamicOriginal) return null;
  return {
    original: { entities: [...s.dynamicOriginal.entities], positions: { ...s.dynamicOriginal.positions } },
    entities: [...s.dynamicEntities],
    positions: { ...s.dynamicPositions },
    filters: cloneFilters(s.dynamicFilters),
  };
}

function cloneFilters(f: DynamicFilters): DynamicFilters {
  return { edges: { ...f.edges }, types: { ...f.types } };
}

function withMirror(s: DynamicModeState, patch: Partial<DynamicModeState>): Partial<DynamicModeState> {
  const next = { ...s, ...patch };
  const id = next.dynamicViewId;
  if (!id) return patch;
  const entry = snapshotEntry(next);
  if (!entry) return patch;
  return { ...patch, draftsByView: { ...next.draftsByView, [id]: entry } };
}

export const createDynamicModeSlice: StateCreator<SliceState, [], [], SliceState> = (set) => ({
  dynamicOriginal: null,
  dynamicViewId: null,
  dynamicEntities: [],
  dynamicPositions: {},
  dynamicFilters: defaultFilters,
  draftsByView: {},

  enterDynamicMode: (initial, viewId = null) => {
    set((s) => {
      const original = { entities: [...initial.entities], positions: { ...initial.positions } };
      const base: Partial<DynamicModeState> = {
        dynamicOriginal: original,
        dynamicViewId: viewId,
        dynamicEntities: [...initial.entities],
        dynamicPositions: { ...initial.positions },
      };
      if (!viewId) return base;
      const entry: DraftEntry = {
        original,
        entities: [...initial.entities],
        positions: { ...initial.positions },
        filters: cloneFilters(s.dynamicFilters),
      };
      return { ...base, draftsByView: { ...s.draftsByView, [viewId]: entry } };
    });
  },
  exitDynamicMode: () => {
    set({
      dynamicOriginal: null,
      dynamicViewId: null,
      dynamicEntities: [],
      dynamicPositions: {},
      dynamicFilters: defaultFilters,
    });
  },
  resetDraft: () => {
    set((s) => {
      if (!s.dynamicOriginal) return {};
      const patch: Partial<DynamicModeState> = {
        dynamicEntities: [...s.dynamicOriginal.entities],
        dynamicPositions: { ...s.dynamicOriginal.positions },
      };
      return withMirror(s, patch);
    });
  },

  draftAddEntities: (entities, positions) => {
    set((s) => {
      const existing = new Set(s.dynamicEntities.map((e) => e.id));
      const additions = entities.filter((e) => !existing.has(e.id));
      const patch: Partial<DynamicModeState> = {
        dynamicEntities: [...s.dynamicEntities, ...additions],
        dynamicPositions: positions ? { ...s.dynamicPositions, ...positions } : s.dynamicPositions,
      };
      return withMirror(s, patch);
    });
  },

  draftRemoveEntities: (ids) => {
    const drop = new Set(ids);
    set((s) => {
      const nextPositions: Record<string, Position> = {};
      for (const [id, pos] of Object.entries(s.dynamicPositions)) {
        if (!drop.has(id)) nextPositions[id] = pos;
      }
      const patch: Partial<DynamicModeState> = {
        dynamicEntities: s.dynamicEntities.filter((e) => !drop.has(e.id)),
        dynamicPositions: nextPositions,
      };
      return withMirror(s, patch);
    });
  },

  draftSetPosition: (id, x, y) => {
    set((s) => withMirror(s, { dynamicPositions: { ...s.dynamicPositions, [id]: { x, y } } }));
  },

  draftSetPositions: (updates) => {
    set((s) => withMirror(s, { dynamicPositions: { ...s.dynamicPositions, ...updates } }));
  },

  draftSetEdgeFilter: (edge, enabled) => {
    set((s) =>
      withMirror(s, {
        dynamicFilters: { ...s.dynamicFilters, edges: { ...s.dynamicFilters.edges, [edge]: enabled } },
      }),
    );
  },

  draftSetTypeFilter: (type, enabled) => {
    set((s) =>
      withMirror(s, {
        dynamicFilters: { ...s.dynamicFilters, types: { ...s.dynamicFilters.types, [type]: enabled } },
      }),
    );
  },

  stashCurrentDraft: (viewId) => {
    set((s) => {
      const entry = snapshotEntry(s);
      if (!entry) return {};
      return { draftsByView: { ...s.draftsByView, [viewId]: entry } };
    });
  },

  hydrateDraftForView: (viewId) => {
    let hydrated = false;
    set((s) => {
      const entry = s.draftsByView[viewId];
      if (!entry) return {};
      hydrated = true;
      return {
        dynamicViewId: viewId,
        dynamicOriginal: { entities: [...entry.original.entities], positions: { ...entry.original.positions } },
        dynamicEntities: [...entry.entities],
        dynamicPositions: { ...entry.positions },
        dynamicFilters: cloneFilters(entry.filters),
      };
    });
    return hydrated;
  },

  discardDraftForView: (viewId) => {
    set((s) => {
      const { [viewId]: _, ...rest } = s.draftsByView;
      const isActive = s.dynamicViewId === viewId;
      if (!isActive) return { draftsByView: rest };
      return {
        draftsByView: rest,
        dynamicOriginal: null,
        dynamicViewId: null,
        dynamicEntities: [],
        dynamicPositions: {},
        dynamicFilters: defaultFilters,
      };
    });
  },
});

export type DynamicDiffState = Pick<DynamicModeState, 'dynamicOriginal' | 'dynamicEntities' | 'dynamicPositions'>;

export function selectDynamicAdditions(state: DynamicDiffState): EntityRef[] {
  if (!state.dynamicOriginal) return [];
  const originalIds = new Set(state.dynamicOriginal.entities.map((e) => e.id));
  return state.dynamicEntities.filter((e) => !originalIds.has(e.id));
}

export function selectDynamicRemovals(state: DynamicDiffState): EntityRef[] {
  if (!state.dynamicOriginal) return [];
  const currentIds = new Set(state.dynamicEntities.map((e) => e.id));
  return state.dynamicOriginal.entities.filter((e) => !currentIds.has(e.id));
}

function isPositionChanged(orig: Position | undefined, cur: Position): boolean {
  return !orig || orig.x !== cur.x || orig.y !== cur.y;
}

export function selectDynamicPositionDeltas(state: DynamicDiffState): Record<string, Position> {
  if (!state.dynamicOriginal) return {};
  const original = state.dynamicOriginal.positions;
  const out: Record<string, Position> = {};
  for (const e of state.dynamicEntities) {
    const cur = state.dynamicPositions[e.id];
    if (cur && isPositionChanged(original[e.id], cur)) out[e.id] = cur;
  }
  return out;
}

export function selectDynamicDirty(state: DynamicDiffState): boolean {
  if (!state.dynamicOriginal) return false;
  if (selectDynamicAdditions(state).length > 0) return true;
  if (selectDynamicRemovals(state).length > 0) return true;
  if (Object.keys(selectDynamicPositionDeltas(state)).length > 0) return true;
  return false;
}

function entryAsDiffState(entry: DraftEntry): DynamicDiffState {
  return {
    dynamicOriginal: entry.original,
    dynamicEntities: entry.entities,
    dynamicPositions: entry.positions,
  };
}

export function selectDirtyForView(state: DynamicModeState, viewId: string): boolean {
  if (state.dynamicViewId === viewId) return selectDynamicDirty(state);
  const entry = state.draftsByView[viewId];
  if (!entry) return false;
  return selectDynamicDirty(entryAsDiffState(entry));
}

function dirtyOutsideDraftMap(state: DynamicModeState): boolean {
  if (!state.dynamicViewId) return false;
  if (state.draftsByView[state.dynamicViewId]) return false;
  return selectDynamicDirty(state);
}

export function selectAnyDirty(state: DynamicModeState): boolean {
  const stashedDirty = Object.keys(state.draftsByView).some((id) => selectDirtyForView(state, id));
  return stashedDirty || dirtyOutsideDraftMap(state);
}
