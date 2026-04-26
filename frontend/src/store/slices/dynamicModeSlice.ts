import type { StateCreator } from 'zustand';
import type {
  DynamicFilters,
  EdgeType,
  EntityRef,
  EntityType,
} from '../../features/canvas/utils/dynamicMode';

export type Position = { x: number; y: number };

export interface DynamicModeSnapshot {
  entities: EntityRef[];
  positions: Record<string, Position>;
}

export interface DynamicModeState {
  dynamicEnabled: boolean;
  dynamicOriginal: DynamicModeSnapshot | null;
  dynamicEntities: EntityRef[];
  dynamicPositions: Record<string, Position>;
  dynamicFilters: DynamicFilters;
}

export interface DynamicModeActions {
  enterDynamicMode: (initial: DynamicModeSnapshot) => void;
  exitDynamicMode: () => void;
  draftAddEntities: (entities: EntityRef[], positions?: Record<string, Position>) => void;
  draftRemoveEntities: (ids: string[]) => void;
  draftSetPosition: (id: string, x: number, y: number) => void;
  draftSetPositions: (updates: Record<string, Position>) => void;
  draftSetEdgeFilter: (edge: EdgeType, enabled: boolean) => void;
  draftSetTypeFilter: (type: EntityType, enabled: boolean) => void;
}

const defaultFilters: DynamicFilters = {
  edges: { relation: true, realization: true, parentage: true, origin: true },
  types: { component: true, capability: true, originEntity: true },
};

export const createDynamicModeSlice: StateCreator<
  DynamicModeState & DynamicModeActions,
  [],
  [],
  DynamicModeState & DynamicModeActions
> = (set) => ({
  dynamicEnabled: false,
  dynamicOriginal: null,
  dynamicEntities: [],
  dynamicPositions: {},
  dynamicFilters: defaultFilters,

  enterDynamicMode: (initial) => {
    set({
      dynamicEnabled: true,
      dynamicOriginal: { entities: [...initial.entities], positions: { ...initial.positions } },
      dynamicEntities: [...initial.entities],
      dynamicPositions: { ...initial.positions },
    });
  },
  exitDynamicMode: () => {
    set({
      dynamicEnabled: false,
      dynamicOriginal: null,
      dynamicEntities: [],
      dynamicPositions: {},
      dynamicFilters: defaultFilters,
    });
  },

  draftAddEntities: (entities, positions) => {
    set((s) => {
      const existing = new Set(s.dynamicEntities.map((e) => e.id));
      const additions = entities.filter((e) => !existing.has(e.id));
      return {
        dynamicEntities: [...s.dynamicEntities, ...additions],
        dynamicPositions: positions ? { ...s.dynamicPositions, ...positions } : s.dynamicPositions,
      };
    });
  },

  draftRemoveEntities: (ids) => {
    const drop = new Set(ids);
    set((s) => {
      const nextPositions: Record<string, Position> = {};
      for (const [id, pos] of Object.entries(s.dynamicPositions)) {
        if (!drop.has(id)) nextPositions[id] = pos;
      }
      return {
        dynamicEntities: s.dynamicEntities.filter((e) => !drop.has(e.id)),
        dynamicPositions: nextPositions,
      };
    });
  },

  draftSetPosition: (id, x, y) => {
    set((s) => ({ dynamicPositions: { ...s.dynamicPositions, [id]: { x, y } } }));
  },

  draftSetPositions: (updates) => {
    set((s) => ({ dynamicPositions: { ...s.dynamicPositions, ...updates } }));
  },

  draftSetEdgeFilter: (edge, enabled) => {
    set((s) => ({
      dynamicFilters: { ...s.dynamicFilters, edges: { ...s.dynamicFilters.edges, [edge]: enabled } },
    }));
  },

  draftSetTypeFilter: (type, enabled) => {
    set((s) => ({
      dynamicFilters: { ...s.dynamicFilters, types: { ...s.dynamicFilters.types, [type]: enabled } },
    }));
  },
});

export function selectDynamicAdditions(state: DynamicModeState): EntityRef[] {
  if (!state.dynamicOriginal) return [];
  const originalIds = new Set(state.dynamicOriginal.entities.map((e) => e.id));
  return state.dynamicEntities.filter((e) => !originalIds.has(e.id));
}

export function selectDynamicRemovals(state: DynamicModeState): EntityRef[] {
  if (!state.dynamicOriginal) return [];
  const currentIds = new Set(state.dynamicEntities.map((e) => e.id));
  return state.dynamicOriginal.entities.filter((e) => !currentIds.has(e.id));
}

function isPositionChanged(orig: Position | undefined, cur: Position): boolean {
  return !orig || orig.x !== cur.x || orig.y !== cur.y;
}

export function selectDynamicPositionDeltas(state: DynamicModeState): Record<string, Position> {
  if (!state.dynamicOriginal) return {};
  const original = state.dynamicOriginal.positions;
  const out: Record<string, Position> = {};
  for (const e of state.dynamicEntities) {
    const cur = state.dynamicPositions[e.id];
    if (cur && isPositionChanged(original[e.id], cur)) out[e.id] = cur;
  }
  return out;
}

export function selectDynamicDirty(state: DynamicModeState): boolean {
  if (!state.dynamicOriginal) return false;
  if (selectDynamicAdditions(state).length > 0) return true;
  if (selectDynamicRemovals(state).length > 0) return true;
  if (Object.keys(selectDynamicPositionDeltas(state)).length > 0) return true;
  return false;
}
