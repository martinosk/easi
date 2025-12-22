import type { StateCreator } from 'zustand';
import type { CapabilityId } from '../types/storeTypes';

export interface CanvasCapabilityState {
  selectedCapabilityId: CapabilityId | null;
}

export interface CanvasCapabilityActions {
  selectCapability: (capabilityId: CapabilityId | null) => void;
}

export const createCanvasCapabilitySlice: StateCreator<
  CanvasCapabilityState & CanvasCapabilityActions,
  [],
  [],
  CanvasCapabilityState & CanvasCapabilityActions
> = (set) => ({
  selectedCapabilityId: null,

  selectCapability: (capabilityId: CapabilityId | null) => {
    set({ selectedCapabilityId: capabilityId });
  },
});
