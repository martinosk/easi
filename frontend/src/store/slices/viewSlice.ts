import type { StateCreator } from 'zustand';
import type { ViewId } from '../types/storeTypes';

export interface ViewState {
  currentViewId: ViewId | null;
  isInitialized: boolean;
  openViewIds: ViewId[];
}

export interface ViewActions {
  setCurrentViewId: (viewId: ViewId | null) => void;
  setInitialized: (initialized: boolean) => void;
  openView: (viewId: ViewId) => void;
  closeView: (viewId: ViewId) => void;
  setOpenViewIds: (viewIds: ViewId[]) => void;
}

export const createViewSlice: StateCreator<ViewState & ViewActions, [], [], ViewState & ViewActions> = (set, get) => ({
  currentViewId: null,
  isInitialized: false,
  openViewIds: [],

  setCurrentViewId: (viewId: ViewId | null) => {
    if (get().currentViewId !== viewId) {
      set({ currentViewId: viewId });
    }
  },

  setInitialized: (initialized: boolean) => {
    set({ isInitialized: initialized });
  },

  openView: (viewId: ViewId) => {
    set((s) => (s.openViewIds.includes(viewId) ? {} : { openViewIds: [...s.openViewIds, viewId] }));
  },

  closeView: (viewId: ViewId) => {
    set((s) => ({ openViewIds: s.openViewIds.filter((id) => id !== viewId) }));
  },

  setOpenViewIds: (viewIds: ViewId[]) => {
    set({ openViewIds: [...viewIds] });
  },
});
