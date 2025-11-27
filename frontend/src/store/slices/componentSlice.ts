import type { StateCreator } from 'zustand';
import type { Component, View, Relation } from '../../api/types';
import type { ComponentId, ComponentData, Position, ViewId } from '../types/storeTypes';
import apiClient from '../../api/client';
import { handleApiCall } from '../utils/apiHelpers';
import toast from 'react-hot-toast';
import { ApiError } from '../../api/types';

export interface ComponentState {
  components: Component[];
}

export interface ComponentActions {
  createComponent: (data: ComponentData) => Promise<Component>;
  updateComponent: (id: ComponentId, data: ComponentData) => Promise<Component>;
  deleteComponent: (id: ComponentId) => Promise<void>;
  updateComponentColor: (viewId: ViewId, componentId: ComponentId, color: string) => Promise<void>;
  clearComponentColor: (viewId: ViewId, componentId: ComponentId) => Promise<void>;
}

type StoreWithDependencies = ComponentState & {
  currentView: View | null;
  relations: Relation[];
};

export const createComponentSlice: StateCreator<
  StoreWithDependencies & ComponentActions,
  [],
  [],
  ComponentState & ComponentActions
> = (set, get) => ({
  components: [],

  createComponent: async (data: ComponentData) => {
    const { components, currentView } = get();

    const newComponent = await handleApiCall(
      () => apiClient.createComponent(data),
      'Failed to create component'
    );

    set({ components: [...components, newComponent] });

    if (currentView) {
      const defaultPosition: Position = { x: 400, y: 300 };

      await apiClient.addComponentToView(currentView.id, {
        componentId: newComponent.id,
        ...defaultPosition,
      });

      const updatedView = await apiClient.getViewById(currentView.id);
      set({ currentView: updatedView });
    }

    toast.success(`Component "${data.name}" created`);
    return newComponent;
  },

  updateComponent: async (id: ComponentId, data: ComponentData) => {
    const { components } = get();

    const updatedComponent = await handleApiCall(
      () => apiClient.updateComponent(id, data),
      'Failed to update component'
    );

    set({
      components: components.map((c) =>
        c.id === id ? updatedComponent : c
      ),
    });

    toast.success(`Component "${data.name}" updated`);
    return updatedComponent;
  },

  deleteComponent: async (id: ComponentId) => {
    const { components, relations, currentView } = get();

    try {
      await apiClient.deleteComponent(id);

      set({
        components: components.filter((c) => c.id !== id),
        relations: relations.filter(
          (r) => r.sourceComponentId !== id && r.targetComponentId !== id
        ),
      });

      if (currentView) {
        const updatedView = await apiClient.getViewById(currentView.id);
        set({ currentView: updatedView });
      }

      toast.success('Component deleted from model');
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to delete component';

      toast.error(errorMessage);
      throw error;
    }
  },

  updateComponentColor: async (viewId: ViewId, componentId: ComponentId, color: string) => {
    const { currentView } = get();
    if (!currentView) return;

    const previousComponents = currentView.components;
    const updatedComponents = currentView.components.map((vc) =>
      vc.componentId === componentId ? { ...vc, customColor: color } : vc
    );

    set({
      currentView: {
        ...currentView,
        components: updatedComponents,
      },
    });

    try {
      await apiClient.updateComponentColor(viewId, componentId, color);
    } catch (error) {
      set({
        currentView: {
          ...currentView,
          components: previousComponents,
        },
      });
      throw error;
    }
  },

  clearComponentColor: async (viewId: ViewId, componentId: ComponentId) => {
    const { currentView } = get();
    if (!currentView) return;

    const previousComponents = currentView.components;
    const updatedComponents = currentView.components.map((vc) =>
      vc.componentId === componentId ? { ...vc, customColor: undefined } : vc
    );

    set({
      currentView: {
        ...currentView,
        components: updatedComponents,
      },
    });

    try {
      await apiClient.clearComponentColor(viewId, componentId);
    } catch (error) {
      set({
        currentView: {
          ...currentView,
          components: previousComponents,
        },
      });
      throw error;
    }
  },
});
