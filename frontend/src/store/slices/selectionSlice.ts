import type { StateCreator } from 'zustand';
import type { ComponentId, RelationId } from '../types/storeTypes';

export interface SelectionState {
  selectedNodeId: ComponentId | null;
  selectedEdgeId: RelationId | null;
}

export interface SelectionActions {
  selectNode: (id: ComponentId | null) => void;
  selectEdge: (id: RelationId | null) => void;
  clearSelection: () => void;
}

export const createSelectionSlice: StateCreator<
  SelectionState & SelectionActions,
  [],
  [],
  SelectionState & SelectionActions
> = (set) => ({
  selectedNodeId: null,
  selectedEdgeId: null,

  selectNode: (id: ComponentId | null) => {
    set({ selectedNodeId: id, selectedEdgeId: null });
  },

  selectEdge: (id: RelationId | null) => {
    set({ selectedEdgeId: id, selectedNodeId: null });
  },

  clearSelection: () => {
    set({ selectedNodeId: null, selectedEdgeId: null });
  },
});
