import type { StateCreator } from 'zustand';
import type { ViewId, ViewportState } from '../types/storeTypes';
import {
  loadViewportStatesFromStorage,
  saveViewportStatesToStorage,
} from '../utils/storageHelpers';

export interface ViewportSliceState {
  viewportStates: Record<ViewId, ViewportState>;
}

export interface ViewportActions {
  saveViewportState: (viewId: ViewId, viewport: ViewportState) => void;
  getViewportState: (viewId: ViewId) => ViewportState | undefined;
}

export const createViewportSlice: StateCreator<
  ViewportSliceState & ViewportActions,
  [],
  [],
  ViewportSliceState & ViewportActions
> = (set, get) => ({
  viewportStates: loadViewportStatesFromStorage(),

  saveViewportState: (viewId: ViewId, viewport: ViewportState) => {
    const { viewportStates } = get();
    const newStates = { ...viewportStates, [viewId]: viewport };
    set({ viewportStates: newStates });
    saveViewportStatesToStorage(newStates);
  },

  getViewportState: (viewId: ViewId) => {
    return get().viewportStates[viewId];
  },
});
