import type { StateCreator } from 'zustand';
import type { ViewId } from '../types/storeTypes';

export interface ViewState {
  currentViewId: ViewId | null;
  isInitialized: boolean;
}

export interface ViewActions {
  setCurrentViewId: (viewId: ViewId | null) => void;
  setInitialized: (initialized: boolean) => void;
}

export const createViewSlice: StateCreator<
  ViewState & ViewActions,
  [],
  [],
  ViewState & ViewActions
> = (set, get) => ({
  currentViewId: null,
  isInitialized: false,

  setCurrentViewId: (viewId: ViewId | null) => {
    if (get().currentViewId !== viewId) {
      set({ currentViewId: viewId });
    }
  },

  setInitialized: (initialized: boolean) => {
    set({ isInitialized: initialized });
  },
});
