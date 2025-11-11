import { describe, it, expect, vi, beforeEach } from 'vitest';
import apiClient from '../api/client';
import { create } from 'zustand';
import type { Component, Relation, View } from '../api/types';
import { ApiError } from '../api/types';

// Mock the API client
vi.mock('../api/client');

// Mock react-hot-toast
const mockToast = {
  success: vi.fn(),
  error: vi.fn(),
};
vi.mock('react-hot-toast', () => ({
  default: mockToast,
}));

// Import the store type
interface AppStore {
  components: Component[];
  relations: Relation[];
  currentView: View | null;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  isLoading: boolean;
  error: string | null;
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

describe('Error Handling Tests', () => {
  let store: AppStore;

  beforeEach(() => {
    vi.clearAllMocks();

    // Create a fresh store instance for each test
    const useStore = create<AppStore>((set, get) => ({
      components: [],
      relations: [],
      currentView: null,
      selectedNodeId: null,
      selectedEdgeId: null,
      isLoading: false,
      error: null,

      loadData: async () => {
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
            mockToast.success('Created default view');
          } else {
            currentView = views[0];
          }

          const fullView = await apiClient.getViewById(currentView.id);
          set({ currentView: fullView, isLoading: false });
          mockToast.success('Data loaded successfully');
        } catch (error) {
          const errorMessage = error instanceof ApiError
            ? error.message
            : 'Failed to load data';
          set({ error: errorMessage, isLoading: false });
          mockToast.error(errorMessage);
          throw error;
        }
      },

      createComponent: async (name: string, description?: string) => {
        const { currentView, components } = get();
        try {
          const newComponent = await apiClient.createComponent({ name, description });
          set({ components: [...components, newComponent] });

          if (currentView) {
            const centerX = 400;
            const centerY = 300;
            await apiClient.addComponentToView(currentView.id, {
              componentId: newComponent.id,
              x: centerX,
              y: centerY,
            });
            const updatedView = await apiClient.getViewById(currentView.id);
            set({ currentView: updatedView });
          }

          mockToast.success(`Component "${name}" created`);
          return newComponent;
        } catch (error) {
          const errorMessage = error instanceof ApiError
            ? error.message
            : 'Failed to create component';
          mockToast.error(errorMessage);
          throw error;
        }
      },

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
          mockToast.success('Relation created');
          return newRelation;
        } catch (error) {
          const errorMessage = error instanceof ApiError
            ? error.message
            : 'Failed to create relation';
          mockToast.error(errorMessage);
          throw error;
        }
      },

      updatePosition: async (componentId: string, x: number, y: number) => {
        const { currentView } = get();
        if (!currentView) return;

        try {
          await apiClient.updateComponentPosition(currentView.id, componentId, { x, y });
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
          mockToast.error(errorMessage);
          throw error;
        }
      },

      selectNode: (id: string | null) => {
        set({ selectedNodeId: id, selectedEdgeId: null });
      },

      selectEdge: (id: string | null) => {
        set({ selectedEdgeId: id, selectedNodeId: null });
      },

      clearSelection: () => {
        set({ selectedNodeId: null, selectedEdgeId: null });
      },

      setError: (error: string | null) => {
        set({ error });
      },
    }));

    store = useStore.getState();
  });

  describe('Invalid component creation', () => {
    it('should handle validation error when creating component', async () => {
      const validationError = new ApiError('Component name is required', 400);
      vi.mocked(apiClient.createComponent).mockRejectedValueOnce(validationError);

      await expect(store.createComponent('', '')).rejects.toThrow(validationError);
      expect(apiClient.createComponent).toHaveBeenCalled();
    });

    it('should handle duplicate component name error', async () => {
      const duplicateError = new ApiError('Component with this name already exists', 409);
      vi.mocked(apiClient.createComponent).mockRejectedValueOnce(duplicateError);

      await expect(store.createComponent('Duplicate', '')).rejects.toThrow(duplicateError);
    });

    it('should handle empty name validation', async () => {
      const validationError = new ApiError('Name cannot be empty', 400);
      vi.mocked(apiClient.createComponent).mockRejectedValueOnce(validationError);

      await expect(store.createComponent('   ', '')).rejects.toThrow();
    });
  });

  describe('Invalid relation creation', () => {
    it('should handle validation error when source equals target', async () => {
      const validationError = new ApiError('Source and target must be different', 400);
      vi.mocked(apiClient.createRelation).mockRejectedValueOnce(validationError);

      

      await expect(
        store.createRelation('comp-1', 'comp-1', 'Triggers')
      ).rejects.toThrow(validationError);
    });

    it('should handle non-existent component error', async () => {
      const notFoundError = new ApiError('Component not found', 404);
      vi.mocked(apiClient.createRelation).mockRejectedValueOnce(notFoundError);

      

      await expect(
        store.createRelation('comp-1', 'non-existent', 'Triggers')
      ).rejects.toThrow(notFoundError);
    });

    it('should handle duplicate relation error', async () => {
      const duplicateError = new ApiError('Relation already exists', 409);
      vi.mocked(apiClient.createRelation).mockRejectedValueOnce(duplicateError);

      

      await expect(
        store.createRelation('comp-1', 'comp-2', 'Triggers')
      ).rejects.toThrow(duplicateError);
    });

    it('should handle invalid relation type', async () => {
      const validationError = new ApiError('Invalid relation type', 400);
      vi.mocked(apiClient.createRelation).mockRejectedValueOnce(validationError);

      

      await expect(
        store.createRelation('comp-1', 'comp-2', 'Invalid' as any)
      ).rejects.toThrow(validationError);
    });
  });

  describe('Network failure scenarios', () => {
    it('should handle network error when loading data', async () => {
      const networkError = new Error('Network request failed');
      vi.mocked(apiClient.getComponents).mockRejectedValueOnce(networkError);

      try {
        await store.loadData();
      } catch (error) {
        // Error is expected
      }

      // Check state after error handling completes
      expect(store.error).toBe('Failed to load data');
      expect(store.isLoading).toBe(false);
    });

    it('should handle timeout error', async () => {
      const timeoutError = new Error('Request timeout');
      vi.mocked(apiClient.getComponents).mockRejectedValueOnce(timeoutError);

      

      await expect(store.loadData()).rejects.toThrow();
    });

    it('should handle 500 server error when creating component', async () => {
      const serverError = new ApiError('Internal server error', 500);
      vi.mocked(apiClient.createComponent).mockRejectedValueOnce(serverError);

      

      await expect(store.createComponent('Test', '')).rejects.toThrow(serverError);
    });

    it('should handle 500 server error when creating relation', async () => {
      const serverError = new ApiError('Internal server error', 500);
      vi.mocked(apiClient.createRelation).mockRejectedValueOnce(serverError);

      

      await expect(
        store.createRelation('comp-1', 'comp-2', 'Triggers')
      ).rejects.toThrow(serverError);
    });

    it('should handle network error when updating position', async () => {
      const networkError = new Error('Network error');
      vi.mocked(apiClient.updateComponentPosition).mockRejectedValueOnce(networkError);


      store.currentView = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [{ componentId: 'comp-1', x: 100, y: 100 }],
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      await expect(store.updatePosition('comp-1', 200, 200)).rejects.toThrow();
    });

    it('should handle unauthorized error (401)', async () => {
      const authError = new ApiError('Unauthorized', 401);
      vi.mocked(apiClient.getComponents).mockRejectedValueOnce(authError);

      

      await expect(store.loadData()).rejects.toThrow(authError);
    });

    it('should handle forbidden error (403)', async () => {
      const forbiddenError = new ApiError('Forbidden', 403);
      vi.mocked(apiClient.createComponent).mockRejectedValueOnce(forbiddenError);

      

      await expect(store.createComponent('Test', '')).rejects.toThrow(forbiddenError);
    });

    it('should handle connection refused error', async () => {
      const connectionError = new Error('Connection refused');
      vi.mocked(apiClient.getComponents).mockRejectedValueOnce(connectionError);

      try {
        await store.loadData();
      } catch (error) {
        // Error is expected
      }

      expect(store.error).toBe('Failed to load data');
    });

    it('should handle partial data loading failure', async () => {
      const mockComponents = [{ id: 'comp-1', name: 'Component A' }];
      const networkError = new Error('Failed to load relations');

      vi.mocked(apiClient.getComponents).mockResolvedValueOnce(mockComponents as any);
      vi.mocked(apiClient.getRelations).mockRejectedValueOnce(networkError);

      try {
        await store.loadData();
      } catch (error) {
        // Error is expected
      }

      expect(store.error).toBe('Failed to load data');
    });
  });
});
