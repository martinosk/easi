import type { StateCreator } from 'zustand';
import type { CapabilityId, ViewId } from '../types/storeTypes';
import type { View } from '../../api/types';
import apiClient from '../../api/client';

export interface CanvasCapability {
  capabilityId: CapabilityId;
  x: number;
  y: number;
}

export interface CanvasCapabilityState {
  canvasCapabilities: CanvasCapability[];
  selectedCapabilityId: CapabilityId | null;
}

type StoreWithView = CanvasCapabilityState & {
  currentView: View | null;
};

export interface CanvasCapabilityActions {
  addCapabilityToCanvas: (capabilityId: CapabilityId, x: number, y: number) => Promise<void>;
  removeCapabilityFromCanvas: (capabilityId: CapabilityId) => Promise<void>;
  updateCapabilityPosition: (capabilityId: CapabilityId, x: number, y: number) => void;
  selectCapability: (capabilityId: CapabilityId | null) => void;
  clearCanvasCapabilities: () => void;
  syncCanvasCapabilitiesFromView: (view: View) => void;
  updateCapabilityColor: (viewId: ViewId, capabilityId: CapabilityId, color: string) => Promise<void>;
}

export const createCanvasCapabilitySlice: StateCreator<
  StoreWithView & CanvasCapabilityActions,
  [],
  [],
  CanvasCapabilityState & CanvasCapabilityActions
> = (set, get) => ({
  canvasCapabilities: [],
  selectedCapabilityId: null,

  addCapabilityToCanvas: async (capabilityId: CapabilityId, x: number, y: number) => {
    const { canvasCapabilities, currentView } = get();
    const exists = canvasCapabilities.some((c) => c.capabilityId === capabilityId);
    if (exists || !currentView) return;

    set({
      canvasCapabilities: [...canvasCapabilities, { capabilityId, x, y }],
    });

    try {
      await apiClient.addCapabilityToView(currentView.id, { capabilityId, x, y });
    } catch (error) {
      set({
        canvasCapabilities: canvasCapabilities.filter((c) => c.capabilityId !== capabilityId),
      });
      throw error;
    }
  },

  removeCapabilityFromCanvas: async (capabilityId: CapabilityId) => {
    const { canvasCapabilities, selectedCapabilityId, currentView } = get();
    if (!currentView) return;

    const removed = canvasCapabilities.find((c) => c.capabilityId === capabilityId);
    set({
      canvasCapabilities: canvasCapabilities.filter((c) => c.capabilityId !== capabilityId),
      selectedCapabilityId: selectedCapabilityId === capabilityId ? null : selectedCapabilityId,
    });

    try {
      await apiClient.removeCapabilityFromView(currentView.id, capabilityId);
    } catch (error) {
      if (removed) {
        set({
          canvasCapabilities: [...get().canvasCapabilities, removed],
        });
      }
      throw error;
    }
  },

  updateCapabilityPosition: (capabilityId: CapabilityId, x: number, y: number) => {
    const { canvasCapabilities, currentView } = get();
    set({
      canvasCapabilities: canvasCapabilities.map((c) =>
        c.capabilityId === capabilityId ? { ...c, x, y } : c
      ),
    });

    if (currentView) {
      apiClient.updateCapabilityPositionInView(currentView.id, capabilityId, x, y).catch(() => {
      });
    }
  },

  selectCapability: (capabilityId: CapabilityId | null) => {
    set({ selectedCapabilityId: capabilityId });
  },

  clearCanvasCapabilities: () => {
    set({ canvasCapabilities: [], selectedCapabilityId: null });
  },

  syncCanvasCapabilitiesFromView: (view: View) => {
    const capabilities: CanvasCapability[] = (view.capabilities || []).map((vc) => ({
      capabilityId: vc.capabilityId,
      x: vc.x,
      y: vc.y,
    }));
    set({ canvasCapabilities: capabilities, selectedCapabilityId: null });
  },

  updateCapabilityColor: async (viewId: ViewId, capabilityId: CapabilityId, color: string) => {
    const { currentView } = get();
    if (!currentView) return;

    const previousCapabilities = currentView.capabilities;
    const updatedCapabilities = currentView.capabilities.map((vc) =>
      vc.capabilityId === capabilityId ? { ...vc, customColor: color } : vc
    );

    set({
      currentView: {
        ...currentView,
        capabilities: updatedCapabilities,
      },
    });

    try {
      await apiClient.updateCapabilityColor(viewId, capabilityId, color);
    } catch (error) {
      set({
        currentView: {
          ...currentView,
          capabilities: previousCapabilities,
        },
      });
      throw error;
    }
  },
});
