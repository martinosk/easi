import type { StateCreator } from 'zustand';
import type { View } from '../../api/types';
import type { ComponentId, Position, EdgeType } from '../types/storeTypes';
import apiClient from '../../api/client';
import { handleApiCall, optimisticUpdate } from '../utils/apiHelpers';

export interface LayoutActions {
  updatePosition: (componentId: ComponentId, position: Position) => Promise<void>;
  setEdgeType: (edgeType: EdgeType) => Promise<void>;
  setColorScheme: (colorScheme: string) => Promise<void>;
}

type StoreWithDependencies = {
  currentView: View | null;
};

type ViewPropertyKey = 'edgeType' | 'colorScheme';

async function updateViewProperty<K extends ViewPropertyKey>(
  currentView: View,
  propertyKey: K,
  newValue: string,
  set: (partial: { currentView: View }) => void,
  apiCall: () => Promise<unknown>,
  successMessage: string,
  errorMessage: string
): Promise<void> {
  const previousValue = currentView[propertyKey];

  set({
    currentView: {
      ...currentView,
      [propertyKey]: newValue,
    },
  });

  await optimisticUpdate({
    apiCall,
    onSuccess: () => {},
    onError: () => set({
      currentView: {
        ...currentView,
        [propertyKey]: previousValue,
      },
    }),
    successMessage,
    errorMessage,
  });
}

export const createLayoutSlice: StateCreator<
  StoreWithDependencies & LayoutActions,
  [],
  [],
  LayoutActions
> = (set, get) => ({
  updatePosition: async (componentId: ComponentId, position: Position) => {
    const { currentView } = get();

    if (!currentView) {
      return;
    }

    await handleApiCall(
      () => apiClient.updateComponentPosition(currentView.id, componentId, position),
      'Failed to update position'
    );

    const updatedComponents = currentView.components.map((vc) =>
      vc.componentId === componentId ? { ...vc, ...position } : vc
    );

    set({
      currentView: {
        ...currentView,
        components: updatedComponents,
      },
    });
  },

  setEdgeType: async (edgeType: EdgeType) => {
    const { currentView } = get();
    if (!currentView) return;

    await updateViewProperty(
      currentView,
      'edgeType',
      edgeType,
      set,
      () => apiClient.updateViewEdgeType(currentView.id, { edgeType }),
      'Edge type updated',
      'Failed to update edge type'
    );
  },

  setColorScheme: async (colorScheme: string) => {
    const { currentView } = get();
    if (!currentView) return;

    await updateViewProperty(
      currentView,
      'colorScheme',
      colorScheme,
      set,
      () => apiClient.updateViewColorScheme(currentView.id, { colorScheme }),
      'Color scheme updated',
      'Failed to update color scheme'
    );
  },
});
