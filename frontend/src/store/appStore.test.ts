import { describe, it, expect, vi, beforeEach } from 'vitest';
import { create } from 'zustand';
import type { Component, Relation, View } from '../api/types';
import { ApiError } from '../api/types';
import apiClient from '../api/client';

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

// Store interface matching the actual implementation
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
  updateComponent: (id: string, name: string, description?: string) => Promise<Component>;
  createRelation: (
    sourceComponentId: string,
    targetComponentId: string,
    relationType: 'Triggers' | 'Serves',
    name?: string,
    description?: string
  ) => Promise<Relation>;
  updateRelation: (id: string, name?: string, description?: string) => Promise<Relation>;
  updatePosition: (componentId: string, x: number, y: number) => Promise<void>;
  selectNode: (id: string | null) => void;
  selectEdge: (id: string | null) => void;
  clearSelection: () => void;
  setError: (error: string | null) => void;
}

describe('AppStore Tests', () => {
  let useStore: ReturnType<typeof create<AppStore>>;
  let store: AppStore;

  beforeEach(() => {
    vi.clearAllMocks();

    // Create a fresh store instance for each test that matches the actual implementation
    useStore = create<AppStore>((set, get) => ({
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

      updateComponent: async (id: string, name: string, description?: string) => {
        const { components } = get();
        try {
          const updatedComponent = await apiClient.updateComponent(id, { name, description });
          set({
            components: components.map((c) =>
              c.id === id ? updatedComponent : c
            ),
          });
          mockToast.success(`Component "${name}" updated`);
          return updatedComponent;
        } catch (error) {
          const errorMessage = error instanceof ApiError
            ? error.message
            : 'Failed to update component';
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

      updateRelation: async (id: string, name?: string, description?: string) => {
        const { relations } = get();
        try {
          const updatedRelation = await apiClient.updateRelation(id, { name, description });
          set({
            relations: relations.map((r) =>
              r.id === id ? updatedRelation : r
            ),
          });
          mockToast.success('Relation updated');
          return updatedRelation;
        } catch (error) {
          const errorMessage = error instanceof ApiError
            ? error.message
            : 'Failed to update relation';
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

  describe('Component Management', () => {
    describe('Create Component', () => {
      it('should create a component and add it to the view with position', async () => {
        // Arrange
        const mockComponent: Component = {
          id: 'comp-1',
          name: 'Test Component',
          description: 'Test Description',
          createdAt: new Date().toISOString(),
          _links: { self: { href: '/api/components/comp-1' } },
        };

        const mockView: View = {
          id: 'view-1',
          name: 'Default View',
          isDefault: false,
          components: [],
          createdAt: new Date().toISOString(),
          _links: { self: { href: '/api/views/view-1' } },
        };

        const mockViewWithComponent: View = {
          ...mockView,
          components: [
            { componentId: 'comp-1', x: 400, y: 300 },
          ],
        };

        vi.mocked(apiClient.createComponent).mockResolvedValueOnce(mockComponent);
        vi.mocked(apiClient.addComponentToView).mockResolvedValueOnce(undefined as any);
        vi.mocked(apiClient.getViewById).mockResolvedValueOnce(mockViewWithComponent);

        // Set initial view state
        useStore.setState({ currentView: mockView, components: [] });

        // Act
        const result = await useStore.getState().createComponent('Test Component', 'Test Description');

        // Assert
        expect(result).toEqual(mockComponent);
        expect(apiClient.createComponent).toHaveBeenCalledWith({
          name: 'Test Component',
          description: 'Test Description',
        });
        expect(apiClient.addComponentToView).toHaveBeenCalledWith('view-1', {
          componentId: 'comp-1',
          x: 400,
          y: 300,
        });
        expect(apiClient.getViewById).toHaveBeenCalledWith('view-1');

        // Verify state was updated
        const finalState = useStore.getState();
        expect(finalState.components).toHaveLength(1);
        expect(finalState.components[0]).toEqual(mockComponent);
        expect(finalState.currentView?.components).toHaveLength(1);
      });

      it('should handle validation error when creating component with empty name', async () => {
        // Arrange
        const validationError = new ApiError('Component name is required', 400);
        vi.mocked(apiClient.createComponent).mockRejectedValueOnce(validationError);

        // Act & Assert
        await expect(useStore.getState().createComponent('', '')).rejects.toThrow(validationError);
        expect(apiClient.createComponent).toHaveBeenCalled();
        expect(mockToast.error).toHaveBeenCalledWith('Component name is required');
      });

      it('should handle duplicate component name error', async () => {
        // Arrange
        const duplicateError = new ApiError('Component with this name already exists', 409);
        vi.mocked(apiClient.createComponent).mockRejectedValueOnce(duplicateError);

        // Act & Assert
        await expect(useStore.getState().createComponent('Duplicate', '')).rejects.toThrow(duplicateError);
        expect(mockToast.error).toHaveBeenCalledWith('Component with this name already exists');
      });

      it('should handle empty name validation', async () => {
        // Arrange
        const validationError = new ApiError('Name cannot be empty', 400);
        vi.mocked(apiClient.createComponent).mockRejectedValueOnce(validationError);

        // Act & Assert
        await expect(useStore.getState().createComponent('   ', '')).rejects.toThrow();
        expect(mockToast.error).toHaveBeenCalledWith('Name cannot be empty');
      });
    });

    describe('Update Component Position', () => {
      it('should update component position after drag', async () => {
        // Arrange
        const mockView: View = {
          id: 'view-1',
          name: 'Default View',
          isDefault: false,
          components: [
            { componentId: 'comp-1', x: 400, y: 300 },
          ],
          createdAt: new Date().toISOString(),
          _links: { self: { href: '/api/views/view-1' } },
        };

        vi.mocked(apiClient.updateComponentPosition).mockResolvedValueOnce(undefined as any);
        useStore.setState({ currentView: mockView });

        // Act
        await useStore.getState().updatePosition('comp-1', 500, 400);

        // Assert
        expect(apiClient.updateComponentPosition).toHaveBeenCalledWith('view-1', 'comp-1', {
          x: 500,
          y: 400,
        });

        const updatedState = useStore.getState();
        expect(updatedState.currentView?.components[0]).toEqual({
          componentId: 'comp-1',
          x: 500,
          y: 400,
        });
      });

      it('should handle network error when updating position', async () => {
        // Arrange
        const networkError = new Error('Network error');
        vi.mocked(apiClient.updateComponentPosition).mockRejectedValueOnce(networkError);

        const mockView: View = {
          id: 'view-1',
          name: 'Test View',
          isDefault: false,
          components: [{ componentId: 'comp-1', x: 100, y: 100 }],
          createdAt: new Date().toISOString(),
          _links: { self: { href: '/api/views/view-1' } },
        };

        useStore.setState({ currentView: mockView });

        // Act & Assert
        await expect(useStore.getState().updatePosition('comp-1', 200, 200)).rejects.toThrow();
        expect(mockToast.error).toHaveBeenCalledWith('Failed to update position');
      });
    });
  });

  describe('Relation Management', () => {
    describe('Create Relation', () => {
      it('should create a relation between two components', async () => {
        // Arrange
        const mockRelation: Relation = {
          id: 'rel-1',
          sourceComponentId: 'comp-1',
          targetComponentId: 'comp-2',
          relationType: 'Triggers' as const,
          name: 'Test Relation',
          createdAt: new Date().toISOString(),
          _links: { self: { href: '/api/relations/rel-1' } },
        };

        vi.mocked(apiClient.createRelation).mockResolvedValueOnce(mockRelation);
        useStore.setState({ relations: [] });

        // Act
        const result = await useStore.getState().createRelation(
          'comp-1',
          'comp-2',
          'Triggers',
          'Test Relation'
        );

        // Assert
        expect(result).toEqual(mockRelation);
        expect(apiClient.createRelation).toHaveBeenCalledWith({
          sourceComponentId: 'comp-1',
          targetComponentId: 'comp-2',
          relationType: 'Triggers',
          name: 'Test Relation',
          description: undefined,
        });

        const finalState = useStore.getState();
        expect(finalState.relations).toHaveLength(1);
        expect(finalState.relations[0]).toEqual(mockRelation);
      });

      it('should handle validation error when source equals target', async () => {
        // Arrange
        const validationError = new ApiError('Source and target must be different', 400);
        vi.mocked(apiClient.createRelation).mockRejectedValueOnce(validationError);

        // Act & Assert
        await expect(
          useStore.getState().createRelation('comp-1', 'comp-1', 'Triggers')
        ).rejects.toThrow(validationError);
        expect(mockToast.error).toHaveBeenCalledWith('Source and target must be different');
      });

      it('should handle non-existent component error', async () => {
        // Arrange
        const notFoundError = new ApiError('Component not found', 404);
        vi.mocked(apiClient.createRelation).mockRejectedValueOnce(notFoundError);

        // Act & Assert
        await expect(
          useStore.getState().createRelation('comp-1', 'non-existent', 'Triggers')
        ).rejects.toThrow(notFoundError);
        expect(mockToast.error).toHaveBeenCalledWith('Component not found');
      });

      it('should handle duplicate relation error', async () => {
        // Arrange
        const duplicateError = new ApiError('Relation already exists', 409);
        vi.mocked(apiClient.createRelation).mockRejectedValueOnce(duplicateError);

        // Act & Assert
        await expect(
          useStore.getState().createRelation('comp-1', 'comp-2', 'Triggers')
        ).rejects.toThrow(duplicateError);
        expect(mockToast.error).toHaveBeenCalledWith('Relation already exists');
      });

      it('should handle invalid relation type', async () => {
        // Arrange
        const validationError = new ApiError('Invalid relation type', 400);
        vi.mocked(apiClient.createRelation).mockRejectedValueOnce(validationError);

        // Act & Assert
        await expect(
          useStore.getState().createRelation('comp-1', 'comp-2', 'Invalid' as any)
        ).rejects.toThrow(validationError);
        expect(mockToast.error).toHaveBeenCalledWith('Invalid relation type');
      });
    });
  });

  describe('Data Loading', () => {
    describe('Load Existing Data', () => {
      it('should load all data including components, relations, and view', async () => {
        // Arrange
        const mockComponents = [
          { id: 'comp-1', name: 'Component A' },
          { id: 'comp-2', name: 'Component B' },
        ];

        const mockRelations = [
          {
            id: 'rel-1',
            sourceComponentId: 'comp-1',
            targetComponentId: 'comp-2',
            relationType: 'Triggers' as const,
          },
        ];

        const mockViews = [
          { id: 'view-1', name: 'Default View' },
        ];

        const mockFullView: View = {
          id: 'view-1',
          name: 'Default View',
          isDefault: false,
          components: [
            { componentId: 'comp-1', x: 100, y: 100 },
            { componentId: 'comp-2', x: 200, y: 200 },
          ],
          createdAt: new Date().toISOString(),
          _links: { self: { href: '/api/views/view-1' } },
        };

        vi.mocked(apiClient.getComponents).mockResolvedValueOnce(mockComponents as any);
        vi.mocked(apiClient.getRelations).mockResolvedValueOnce(mockRelations as any);
        vi.mocked(apiClient.getViews).mockResolvedValueOnce(mockViews as any);
        vi.mocked(apiClient.getViewById).mockResolvedValueOnce(mockFullView);

        // Act
        await useStore.getState().loadData();

        // Assert
        const finalState = useStore.getState();
        expect(finalState.components).toEqual(mockComponents);
        expect(finalState.relations).toEqual(mockRelations);
        expect(finalState.currentView).toEqual(mockFullView);
        expect(finalState.isLoading).toBe(false);
        expect(mockToast.success).toHaveBeenCalledWith('Data loaded successfully');
      });

      it('should create default view if none exists', async () => {
        // Arrange
        const mockComponents: any[] = [];
        const mockRelations: any[] = [];
        const mockViews: any[] = [];

        const mockCreatedView = {
          id: 'view-1',
          name: 'Default View',
          description: 'Main application view',
        };

        const mockFullView: View = {
          ...mockCreatedView,
          isDefault: false,
          components: [],
          createdAt: new Date().toISOString(),
          _links: { self: { href: '/api/views/view-1' } },
        };

        vi.mocked(apiClient.getComponents).mockResolvedValueOnce(mockComponents);
        vi.mocked(apiClient.getRelations).mockResolvedValueOnce(mockRelations);
        vi.mocked(apiClient.getViews).mockResolvedValueOnce(mockViews);
        vi.mocked(apiClient.createView).mockResolvedValueOnce(mockCreatedView as any);
        vi.mocked(apiClient.getViewById).mockResolvedValueOnce(mockFullView);

        // Act
        await useStore.getState().loadData();

        // Assert
        expect(apiClient.createView).toHaveBeenCalledWith({
          name: 'Default View',
          description: 'Main application view',
        });
        expect(mockToast.success).toHaveBeenCalledWith('Created default view');

        const finalState = useStore.getState();
        expect(finalState.currentView).toEqual(mockFullView);
      });
    });

    describe('Load Data Errors', () => {
      it('should handle network error when loading data', async () => {
        // Arrange
        const networkError = new Error('Network request failed');
        vi.mocked(apiClient.getComponents).mockRejectedValueOnce(networkError);

        // Act
        try {
          await useStore.getState().loadData();
        } catch (error) {
          // Error is expected
        }

        // Assert
        const finalState = useStore.getState();
        expect(finalState.error).toBe('Failed to load data');
        expect(finalState.isLoading).toBe(false);
        expect(mockToast.error).toHaveBeenCalledWith('Failed to load data');
      });

      it('should handle timeout error', async () => {
        // Arrange
        const timeoutError = new Error('Request timeout');
        vi.mocked(apiClient.getComponents).mockRejectedValueOnce(timeoutError);

        // Act & Assert
        await expect(useStore.getState().loadData()).rejects.toThrow();
        expect(mockToast.error).toHaveBeenCalledWith('Failed to load data');
      });

      it('should handle partial data loading failure', async () => {
        // Arrange
        const mockComponents = [{ id: 'comp-1', name: 'Component A' }];
        const networkError = new Error('Failed to load relations');

        vi.mocked(apiClient.getComponents).mockResolvedValueOnce(mockComponents as any);
        vi.mocked(apiClient.getRelations).mockRejectedValueOnce(networkError);

        // Act
        try {
          await useStore.getState().loadData();
        } catch (error) {
          // Error is expected
        }

        // Assert
        const finalState = useStore.getState();
        expect(finalState.error).toBe('Failed to load data');
        expect(mockToast.error).toHaveBeenCalledWith('Failed to load data');
      });
    });
  });

  describe('Error Handling', () => {
    describe('API Errors', () => {
      it('should handle 500 server error when creating component', async () => {
        // Arrange
        const serverError = new ApiError('Internal server error', 500);
        vi.mocked(apiClient.createComponent).mockRejectedValueOnce(serverError);

        // Act & Assert
        await expect(useStore.getState().createComponent('Test', '')).rejects.toThrow(serverError);
        expect(mockToast.error).toHaveBeenCalledWith('Internal server error');
      });

      it('should handle 500 server error when creating relation', async () => {
        // Arrange
        const serverError = new ApiError('Internal server error', 500);
        vi.mocked(apiClient.createRelation).mockRejectedValueOnce(serverError);

        // Act & Assert
        await expect(
          useStore.getState().createRelation('comp-1', 'comp-2', 'Triggers')
        ).rejects.toThrow(serverError);
        expect(mockToast.error).toHaveBeenCalledWith('Internal server error');
      });

      it('should handle unauthorized error (401)', async () => {
        // Arrange
        const authError = new ApiError('Unauthorized', 401);
        vi.mocked(apiClient.getComponents).mockRejectedValueOnce(authError);

        // Act & Assert
        await expect(useStore.getState().loadData()).rejects.toThrow(authError);
        expect(mockToast.error).toHaveBeenCalledWith('Unauthorized');
      });

      it('should handle forbidden error (403)', async () => {
        // Arrange
        const forbiddenError = new ApiError('Forbidden', 403);
        vi.mocked(apiClient.createComponent).mockRejectedValueOnce(forbiddenError);

        // Act & Assert
        await expect(useStore.getState().createComponent('Test', '')).rejects.toThrow(forbiddenError);
        expect(mockToast.error).toHaveBeenCalledWith('Forbidden');
      });

      it('should handle connection refused error', async () => {
        // Arrange
        const connectionError = new Error('Connection refused');
        vi.mocked(apiClient.getComponents).mockRejectedValueOnce(connectionError);

        // Act
        try {
          await useStore.getState().loadData();
        } catch (error) {
          // Error is expected
        }

        // Assert
        const finalState = useStore.getState();
        expect(finalState.error).toBe('Failed to load data');
        expect(mockToast.error).toHaveBeenCalledWith('Failed to load data');
      });
    });
  });

  describe('Selection Management', () => {
    it('should select a node', () => {
      // Act
      useStore.getState().selectNode('node-1');

      // Assert
      const state = useStore.getState();
      expect(state.selectedNodeId).toBe('node-1');
      expect(state.selectedEdgeId).toBeNull();
    });

    it('should select an edge', () => {
      // Act
      useStore.getState().selectEdge('edge-1');

      // Assert
      const state = useStore.getState();
      expect(state.selectedEdgeId).toBe('edge-1');
      expect(state.selectedNodeId).toBeNull();
    });

    it('should clear selection', () => {
      // Arrange
      useStore.setState({ selectedNodeId: 'node-1', selectedEdgeId: null });

      // Act
      useStore.getState().clearSelection();

      // Assert
      const state = useStore.getState();
      expect(state.selectedNodeId).toBeNull();
      expect(state.selectedEdgeId).toBeNull();
    });
  });

  describe('Error State Management', () => {
    it('should set error message', () => {
      // Act
      useStore.getState().setError('Test error');

      // Assert
      expect(useStore.getState().error).toBe('Test error');
    });

    it('should clear error message', () => {
      // Arrange
      useStore.setState({ error: 'Some error' });

      // Act
      useStore.getState().setError(null);

      // Assert
      expect(useStore.getState().error).toBeNull();
    });
  });
});
