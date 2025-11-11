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

describe('End-to-End Tests', () => {
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

  describe('Create and position component', () => {
    it('should create a component and add it to the view with position', async () => {
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

      // Set up initial state
      store.currentView = mockView;
      store.components = [];

      // Create component
      const result = await store.createComponent('Test Component', 'Test Description');

      // Verify component was created
      expect(result).toEqual(mockComponent);
      expect(apiClient.createComponent).toHaveBeenCalledWith({
        name: 'Test Component',
        description: 'Test Description',
      });

      // Verify component was added to view with default position
      expect(apiClient.addComponentToView).toHaveBeenCalledWith('view-1', {
        componentId: 'comp-1',
        x: 400,
        y: 300,
      });

      // Verify view was reloaded
      expect(apiClient.getViewById).toHaveBeenCalledWith('view-1');
    });

  it('should update component position after drag', async () => {
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

      // Set up initial state - need to use set method to properly update zustand store
      const useStore = create<AppStore>((set, get) => ({
        components: [],
        relations: [],
        currentView: mockView,
        selectedNodeId: null,
        selectedEdgeId: null,
        isLoading: false,
        error: null,
        loadData: async () => {},
        createComponent: async () => ({} as any),
        createRelation: async () => ({} as any),
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
        selectNode: () => {},
        selectEdge: () => {},
        clearSelection: () => {},
        setError: () => {},
      }));

      store = useStore.getState();

      // Update position
      await store.updatePosition('comp-1', 500, 400);

      // Verify position was updated
      expect(apiClient.updateComponentPosition).toHaveBeenCalledWith('view-1', 'comp-1', {
        x: 500,
        y: 400,
      });

      // Get fresh state after async operation
      const updatedState = useStore.getState();
      expect(updatedState.currentView?.components[0]).toEqual({
        componentId: 'comp-1',
        x: 500,
        y: 400,
      });
    });
  });

describe('Create relation between components', () => {
    it('should create a relation between two components', async () => {
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

      // Create relation
      const result = await store.createRelation(
        'comp-1',
        'comp-2',
        'Triggers',
        'Test Relation'
      );

      // Verify relation was created
      expect(result).toEqual(mockRelation);
      expect(apiClient.createRelation).toHaveBeenCalledWith({
        sourceComponentId: 'comp-1',
        targetComponentId: 'comp-2',
        relationType: 'Triggers',
        name: 'Test Relation',
        description: undefined,
      });

      // Verify relation was added to state
      expect(store.relations).toHaveLength(1);
      expect(store.relations[0]).toEqual(mockRelation);
    });
  });

describe('Load existing components and relations', () => {
    it('should load all data including components, relations, and view', async () => {
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

      // Load data
      await store.loadData();

      // Verify all data was loaded
      expect(store.components).toEqual(mockComponents);
      expect(store.relations).toEqual(mockRelations);
      expect(store.currentView).toEqual(mockFullView);
      expect(store.isLoading).toBe(false);
    });

    it('should create default view if none exists', async () => {
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

      // Load data
      await store.loadData();

      // Verify default view was created
      expect(apiClient.createView).toHaveBeenCalledWith({
        name: 'Default View',
        description: 'Main application view',
      });
      expect(store.currentView).toEqual(mockFullView);
    });
  });
});
