import { create } from 'zustand';
import apiClient from '../api/client';
import type { Component, Relation, View } from '../api/types';
import { ApiError } from '../api/types';
import toast from 'react-hot-toast';

interface AppStore {
  // Data
  components: Component[];
  relations: Relation[];
  currentView: View | null;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;

  // Loading states
  isLoading: boolean;
  error: string | null;

  // Actions
  loadData: () => Promise<void>;
  createComponent: (name: string, description?: string) => Promise<Component>;
  createRelation: (
    sourceComponentId: string,
    targetComponentId: string,
    relationType: 'Triggers' | 'Serves',
    name?: string,
    description?: string
  ) => Promise<Relation>;
  updatePosition: (componentId: string, x: number, y: number) => Promise<void>;
  selectNode: (id: string | null) => void;
  selectEdge: (id: string | null) => void;
  clearSelection: () => void;
  setError: (error: string | null) => void;
}

export const useAppStore = create<AppStore>((set, get) => ({
  // Initial state
  components: [],
  relations: [],
  currentView: null,
  selectedNodeId: null,
  selectedEdgeId: null,
  isLoading: false,
  error: null,

  // Load all data
  loadData: async () => {
    set({ isLoading: true, error: null });

    try {
      // Load components and relations in parallel
      const [components, relations] = await Promise.all([
        apiClient.getComponents(),
        apiClient.getRelations(),
      ]);

      set({ components, relations });

      // Load or create default view
      const views = await apiClient.getViews();
      let currentView: View;

      if (views.length === 0) {
        // Create default view
        currentView = await apiClient.createView({
          name: 'Default View',
          description: 'Main application view',
        });
        toast.success('Created default view');
      } else {
        // Use first view
        currentView = views[0];
      }

      // Load full view with components
      const fullView = await apiClient.getViewById(currentView.id);
      set({ currentView: fullView, isLoading: false });

      toast.success('Data loaded successfully');
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to load data';

      set({ error: errorMessage, isLoading: false });
      toast.error(errorMessage);
      throw error;
    }
  },

  // Create a new component
  createComponent: async (name: string, description?: string) => {
    const { currentView, components } = get();

    try {
      // Create component
      const newComponent = await apiClient.createComponent({ name, description });

      set({ components: [...components, newComponent] });

      // Add to current view if exists
      if (currentView) {
        // Calculate center position (default canvas center)
        const centerX = 400;
        const centerY = 300;

        await apiClient.addComponentToView(currentView.id, {
          componentId: newComponent.id,
          x: centerX,
          y: centerY,
        });

        // Reload view to get updated components list
        const updatedView = await apiClient.getViewById(currentView.id);
        set({ currentView: updatedView });
      }

      toast.success(`Component "${name}" created`);
      return newComponent;
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to create component';

      toast.error(errorMessage);
      throw error;
    }
  },

  // Create a new relation
  createRelation: async (
    sourceComponentId: string,
    targetComponentId: string,
    relationType: 'Triggers' | 'Serves',
    name?: string,
    description?: string
  ) => {
    const { relations } = get();

    try {
      const newRelation = await apiClient.createRelation({
        sourceComponentId,
        targetComponentId,
        relationType,
        name,
        description,
      });

      set({ relations: [...relations, newRelation] });
      toast.success('Relation created');
      return newRelation;
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to create relation';

      toast.error(errorMessage);
      throw error;
    }
  },

  // Update component position
  updatePosition: async (componentId: string, x: number, y: number) => {
    const { currentView } = get();

    if (!currentView) {
      return;
    }

    try {
      await apiClient.updateComponentPosition(currentView.id, componentId, { x, y });

      // Update local state
      const updatedComponents = currentView.components.map((vc) =>
        vc.componentId === componentId ? { ...vc, x, y } : vc
      );

      set({
        currentView: {
          ...currentView,
          components: updatedComponents,
        },
      });
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to update position';

      toast.error(errorMessage);
      throw error;
    }
  },

  // Select a node
  selectNode: (id: string | null) => {
    set({ selectedNodeId: id, selectedEdgeId: null });
  },

  // Select an edge
  selectEdge: (id: string | null) => {
    set({ selectedEdgeId: id, selectedNodeId: null });
  },

  // Clear selection
  clearSelection: () => {
    set({ selectedNodeId: null, selectedEdgeId: null });
  },

  // Set error
  setError: (error: string | null) => {
    set({ error });
  },
}));
