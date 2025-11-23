import type { StateCreator } from 'zustand';
import type { View, Component, Relation } from '../../api/types';
import type { ViewId, ComponentId } from '../types/storeTypes';
import apiClient from '../../api/client';
import toast from 'react-hot-toast';
import { ApiError } from '../../api/types';

export interface ViewState {
  views: View[];
  currentView: View | null;
  isLoading: boolean;
  error: string | null;
}

export interface ViewActions {
  loadData: () => Promise<void>;
  loadViews: () => Promise<void>;
  switchView: (viewId: ViewId) => Promise<void>;
  removeComponentFromView: (componentId: ComponentId) => Promise<void>;
  setError: (error: string | null) => void;
}

type StoreWithDependencies = ViewState & {
  components: Component[];
  relations: Relation[];
  syncCanvasCapabilitiesFromView: (view: View) => void;
};

let isLoadingData = false;

export const createViewSlice: StateCreator<
  StoreWithDependencies & ViewActions,
  [],
  [],
  ViewState & ViewActions
> = (set, get) => ({
  views: [],
  currentView: null,
  isLoading: false,
  error: null,

  loadData: async () => {
    if (isLoadingData) {
      return;
    }

    isLoadingData = true;
    set({ isLoading: true, error: null });

    try {
      const [components, relations] = await Promise.all([
        apiClient.getComponents(),
        apiClient.getRelations(),
      ]);

      set({ components, relations });

      const views = await apiClient.getViews();
      let currentView: View;

      if (views.length === 0) {
        currentView = await apiClient.createView({
          name: 'Default View',
          description: 'Main application view',
        });
        views.push(currentView);
        toast.success('Created default view');
      } else {
        currentView = views.find(v => v.isDefault) || views[0];
      }

      const fullView = await apiClient.getViewById(currentView.id);
      set({ currentView: fullView, views, isLoading: false });
      get().syncCanvasCapabilitiesFromView(fullView);

      toast.success('Data loaded successfully');
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to load data';

      set({ error: errorMessage, isLoading: false });
      toast.error(errorMessage);
      throw error;
    } finally {
      isLoadingData = false;
    }
  },

  loadViews: async () => {
    try {
      const views = await apiClient.getViews();
      set({ views });
    } catch (error) {
      console.error('Failed to load views:', error);
    }
  },

  switchView: async (viewId: ViewId) => {
    const { currentView } = get();

    if (currentView?.id === viewId) {
      return;
    }

    try {
      const fullView = await apiClient.getViewById(viewId);
      set({ currentView: fullView });
      get().syncCanvasCapabilitiesFromView(fullView);
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to switch view';
      toast.error(errorMessage);
      throw error;
    }
  },

  removeComponentFromView: async (componentId: ComponentId) => {
    const { currentView } = get();

    if (!currentView) {
      return;
    }

    try {
      await apiClient.removeComponentFromView(currentView.id, componentId);

      const updatedComponents = currentView.components.filter(
        (vc) => vc.componentId !== componentId
      );

      set({
        currentView: {
          ...currentView,
          components: updatedComponents,
        },
      });

      toast.success('Component removed from view');
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to remove component from view';

      toast.error(errorMessage);
      throw error;
    }
  },

  setError: (error: string | null) => {
    set({ error });
  },
});
