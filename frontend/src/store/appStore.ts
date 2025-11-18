import { create } from 'zustand';
import apiClient from '../api/client';
import type { Component, Relation, View } from '../api/types';
import { ApiError } from '../api/types';
import toast from 'react-hot-toast';

export interface ViewportState {
  x: number;
  y: number;
  zoom: number;
}

interface AppStore {
  // Data
  components: Component[];
  relations: Relation[];
  views: View[];
  currentView: View | null;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;

  // Canvas state per view
  viewportStates: Record<string, ViewportState>;

  // Loading states
  isLoading: boolean;
  error: string | null;

  // Actions
  loadData: () => Promise<void>;
  loadViews: () => Promise<void>;
  switchView: (viewId: string) => Promise<void>;
  saveViewportState: (viewId: string, viewport: ViewportState) => void;
  getViewportState: (viewId: string) => ViewportState | undefined;
  createComponent: (name: string, description?: string) => Promise<Component>;
  updateComponent: (id: string, name: string, description?: string) => Promise<Component>;
  createRelation: (
    sourceComponentId: string,
    targetComponentId: string,
    relationType: 'Triggers' | 'Serves',
    name?: string,
    description?: string
  ) => Promise<Relation>;
  updateRelation: (id: string, name?: string, description?: string) => Promise<Relation>;
  deleteComponent: (id: string) => Promise<void>;
  deleteRelation: (id: string) => Promise<void>;
  removeComponentFromView: (componentId: string) => Promise<void>;
  updatePosition: (componentId: string, x: number, y: number) => Promise<void>;
  setEdgeType: (edgeType: string) => Promise<void>;
  setLayoutDirection: (direction: string) => Promise<void>;
  applyAutoLayout: () => Promise<void>;
  selectNode: (id: string | null) => void;
  selectEdge: (id: string | null) => void;
  clearSelection: () => void;
  setError: (error: string | null) => void;
}

// Flag to prevent concurrent loadData calls (prevents race condition in StrictMode)
let isLoadingData = false;

// Load viewport states from localStorage
const loadViewportStatesFromStorage = (): Record<string, ViewportState> => {
  try {
    const stored = localStorage.getItem('viewportStates');
    return stored ? JSON.parse(stored) : {};
  } catch {
    return {};
  }
};

// Save viewport states to localStorage
const saveViewportStatesToStorage = (states: Record<string, ViewportState>) => {
  try {
    localStorage.setItem('viewportStates', JSON.stringify(states));
  } catch (error) {
    console.error('Failed to save viewport states:', error);
  }
};

export const useAppStore = create<AppStore>((set, get) => ({
  // Initial state
  components: [],
  relations: [],
  views: [],
  currentView: null,
  selectedNodeId: null,
  selectedEdgeId: null,
  viewportStates: loadViewportStatesFromStorage(),
  isLoading: false,
  error: null,

  // Load all data
  loadData: async () => {
    // Prevent concurrent calls (fixes StrictMode double-call issue)
    if (isLoadingData) {
      return;
    }

    isLoadingData = true;
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
        views.push(currentView);
        toast.success('Created default view');
      } else {
        // Use default view or first view
        currentView = views.find(v => v.isDefault) || views[0];
      }

      // Load full view with components
      const fullView = await apiClient.getViewById(currentView.id);
      set({ currentView: fullView, views, isLoading: false });

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

  // Load views
  loadViews: async () => {
    try {
      const views = await apiClient.getViews();
      set({ views });
    } catch (error) {
      console.error('Failed to load views:', error);
    }
  },

  // Switch to a different view
  switchView: async (viewId: string) => {
    const { currentView } = get();

    // Don't switch if already on this view
    if (currentView?.id === viewId) {
      return;
    }

    try {
      const fullView = await apiClient.getViewById(viewId);
      set({ currentView: fullView });
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to switch view';
      toast.error(errorMessage);
      throw error;
    }
  },

  // Save viewport state for a view
  saveViewportState: (viewId: string, viewport: ViewportState) => {
    const { viewportStates } = get();
    const newStates = { ...viewportStates, [viewId]: viewport };
    set({ viewportStates: newStates });
    saveViewportStatesToStorage(newStates);
  },

  // Get viewport state for a view
  getViewportState: (viewId: string) => {
    return get().viewportStates[viewId];
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

  // Update a component
  updateComponent: async (id: string, name: string, description?: string) => {
    const { components } = get();

    try {
      const updatedComponent = await apiClient.updateComponent(id, { name, description });

      set({
        components: components.map((c) =>
          c.id === id ? updatedComponent : c
        ),
      });

      toast.success(`Component "${name}" updated`);
      return updatedComponent;
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to update component';

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

  // Update a relation
  updateRelation: async (id: string, name?: string, description?: string) => {
    const { relations } = get();

    try {
      const updatedRelation = await apiClient.updateRelation(id, { name, description });

      set({
        relations: relations.map((r) =>
          r.id === id ? updatedRelation : r
        ),
      });

      toast.success('Relation updated');
      return updatedRelation;
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to update relation';

      toast.error(errorMessage);
      throw error;
    }
  },

  // Delete a component from the model
  deleteComponent: async (id: string) => {
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

  // Delete a relation from the model
  deleteRelation: async (id: string) => {
    const { relations } = get();

    try {
      await apiClient.deleteRelation(id);

      set({
        relations: relations.filter((r) => r.id !== id),
      });

      toast.success('Relation deleted from model');
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to delete relation';

      toast.error(errorMessage);
      throw error;
    }
  },

  // Remove a component from the current view only
  removeComponentFromView: async (componentId: string) => {
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

  setEdgeType: async (edgeType: string) => {
    const { currentView } = get();

    if (!currentView) {
      return;
    }

    const previousEdgeType = currentView.edgeType;

    set({
      currentView: {
        ...currentView,
        edgeType,
      },
    });

    try {
      await apiClient.updateViewEdgeType(currentView.id, { edgeType });
      toast.success('Edge type updated');
    } catch (error) {
      set({
        currentView: {
          ...currentView,
          edgeType: previousEdgeType,
        },
      });

      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to update edge type';

      toast.error(errorMessage);
      throw error;
    }
  },

  setLayoutDirection: async (layoutDirection: string) => {
    const { currentView } = get();

    if (!currentView) {
      return;
    }

    const previousLayoutDirection = currentView.layoutDirection;

    set({
      currentView: {
        ...currentView,
        layoutDirection,
      },
    });

    try {
      await apiClient.updateViewLayoutDirection(currentView.id, { layoutDirection });
      toast.success('Layout direction updated');
    } catch (error) {
      set({
        currentView: {
          ...currentView,
          layoutDirection: previousLayoutDirection,
        },
      });

      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to update layout direction';

      toast.error(errorMessage);
      throw error;
    }
  },

  applyAutoLayout: async () => {
    const { currentView, components, relations } = get();

    if (!currentView) {
      return;
    }

    try {
      const { calculateDagreLayout } = await import('../utils/layout');

      const nodes = components
        .filter((component) =>
          currentView.components.some((vc) => vc.componentId === component.id)
        )
        .map((component) => {
          const viewComponent = currentView.components.find(
            (vc) => vc.componentId === component.id
          );

          const position = viewComponent
            ? { x: viewComponent.x, y: viewComponent.y }
            : { x: 400, y: 300 };

          return {
            id: component.id,
            type: 'component',
            position,
            data: {
              label: component.name,
              description: component.description,
            },
          };
        });

      const edges = relations
        .filter((relation) => {
          const sourceInView = currentView.components.some(
            (vc) => vc.componentId === relation.sourceComponentId
          );
          const targetInView = currentView.components.some(
            (vc) => vc.componentId === relation.targetComponentId
          );
          return sourceInView && targetInView;
        })
        .map((relation) => ({
          id: relation.id,
          source: relation.sourceComponentId,
          target: relation.targetComponentId,
        }));

      const layoutedNodes = calculateDagreLayout(nodes, edges, {
        direction: (currentView.layoutDirection as 'TB' | 'LR' | 'BT' | 'RL') || 'TB',
      });

      const positions = layoutedNodes.map((node) => ({
        componentId: node.id,
        x: node.position.x,
        y: node.position.y,
      }));

      await apiClient.updateMultiplePositions(currentView.id, { positions });

      const updatedView = await apiClient.getViewById(currentView.id);
      set({ currentView: updatedView });

      toast.success('Layout applied');
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to apply layout';

      toast.error(errorMessage);
      throw error;
    }
  },
}));
