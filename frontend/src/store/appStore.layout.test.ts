import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useAppStore } from './appStore';
import apiClient from '../api/client';
import { ApiError } from '../api/types';
import type { View } from '../api/types';

vi.mock('../api/client');

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock('../utils/layout', () => ({
  calculateDagreLayout: vi.fn((nodes, edges, options) => {
    return nodes.map((node, index) => ({
      ...node,
      position: {
        x: index * 200,
        y: index * 150,
      },
    }));
  }),
}));

const mockToast = await import('react-hot-toast').then(m => m.default);

describe('AppStore Layout Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAppStore.setState({
      currentView: null,
      components: [],
      relations: [],
    });
  });

  describe('setEdgeType', () => {
    it('should update edge type successfully', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({ currentView: mockView });
      vi.mocked(apiClient.updateViewEdgeType).mockResolvedValueOnce(undefined as any);

      await useAppStore.getState().setEdgeType('step');

      expect(useAppStore.getState().currentView?.edgeType).toBe('step');
      expect(apiClient.updateViewEdgeType).toHaveBeenCalledWith('view-1', { edgeType: 'step' });
      expect(mockToast.success).toHaveBeenCalledWith('Edge type updated');
    });

    it('should rollback on API error', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({ currentView: mockView });
      const apiError = new ApiError('Failed to update', 500);
      vi.mocked(apiClient.updateViewEdgeType).mockRejectedValueOnce(apiError);

      await expect(useAppStore.getState().setEdgeType('step')).rejects.toThrow(apiError);

      expect(useAppStore.getState().currentView?.edgeType).toBe('default');
      expect(mockToast.error).toHaveBeenCalledWith('Failed to update');
    });

    it('should handle generic error', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({ currentView: mockView });
      const genericError = new Error('Network error');
      vi.mocked(apiClient.updateViewEdgeType).mockRejectedValueOnce(genericError);

      await expect(useAppStore.getState().setEdgeType('step')).rejects.toThrow(genericError);

      expect(useAppStore.getState().currentView?.edgeType).toBe('default');
      expect(mockToast.error).toHaveBeenCalledWith('Failed to update edge type');
    });

    it('should do nothing if no current view', async () => {
      useAppStore.setState({ currentView: null });

      await useAppStore.getState().setEdgeType('step');

      expect(apiClient.updateViewEdgeType).not.toHaveBeenCalled();
      expect(mockToast.success).not.toHaveBeenCalled();
    });

    it('should handle all valid edge types', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      const edgeTypes = ['default', 'step', 'smoothstep', 'straight'];

      for (const edgeType of edgeTypes) {
        useAppStore.setState({ currentView: { ...mockView, edgeType: 'default' } });
        vi.mocked(apiClient.updateViewEdgeType).mockResolvedValueOnce(undefined as any);

        await useAppStore.getState().setEdgeType(edgeType);

        expect(useAppStore.getState().currentView?.edgeType).toBe(edgeType);
      }
    });
  });

  describe('setLayoutDirection', () => {
    it('should update layout direction successfully', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({ currentView: mockView });
      vi.mocked(apiClient.updateViewLayoutDirection).mockResolvedValueOnce(undefined as any);

      await useAppStore.getState().setLayoutDirection('LR');

      expect(useAppStore.getState().currentView?.layoutDirection).toBe('LR');
      expect(apiClient.updateViewLayoutDirection).toHaveBeenCalledWith('view-1', { layoutDirection: 'LR' });
      expect(mockToast.success).toHaveBeenCalledWith('Layout direction updated');
    });

    it('should rollback on API error', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({ currentView: mockView });
      const apiError = new ApiError('Failed to update', 500);
      vi.mocked(apiClient.updateViewLayoutDirection).mockRejectedValueOnce(apiError);

      await expect(useAppStore.getState().setLayoutDirection('LR')).rejects.toThrow(apiError);

      expect(useAppStore.getState().currentView?.layoutDirection).toBe('TB');
      expect(mockToast.error).toHaveBeenCalledWith('Failed to update');
    });

    it('should handle generic error', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({ currentView: mockView });
      const genericError = new Error('Network error');
      vi.mocked(apiClient.updateViewLayoutDirection).mockRejectedValueOnce(genericError);

      await expect(useAppStore.getState().setLayoutDirection('LR')).rejects.toThrow(genericError);

      expect(useAppStore.getState().currentView?.layoutDirection).toBe('TB');
      expect(mockToast.error).toHaveBeenCalledWith('Failed to update layout direction');
    });

    it('should do nothing if no current view', async () => {
      useAppStore.setState({ currentView: null });

      await useAppStore.getState().setLayoutDirection('LR');

      expect(apiClient.updateViewLayoutDirection).not.toHaveBeenCalled();
      expect(mockToast.success).not.toHaveBeenCalled();
    });

    it('should handle all valid directions', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      const directions = ['TB', 'LR', 'BT', 'RL'];

      for (const direction of directions) {
        useAppStore.setState({ currentView: { ...mockView, layoutDirection: 'TB' } });
        vi.mocked(apiClient.updateViewLayoutDirection).mockResolvedValueOnce(undefined as any);

        await useAppStore.getState().setLayoutDirection(direction);

        expect(useAppStore.getState().currentView?.layoutDirection).toBe(direction);
      }
    });
  });

  describe('applyAutoLayout', () => {
    it('should apply layout and update positions', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [
          { componentId: 'comp-1', x: 0, y: 0 },
          { componentId: 'comp-2', x: 0, y: 0 },
        ],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      const mockComponents = [
        { id: 'comp-1', name: 'Component 1', description: 'Desc 1' },
        { id: 'comp-2', name: 'Component 2', description: 'Desc 2' },
      ];

      const mockRelations = [
        {
          id: 'rel-1',
          sourceComponentId: 'comp-1',
          targetComponentId: 'comp-2',
          relationType: 'Triggers' as const,
        },
      ];

      const mockUpdatedView: View = {
        ...mockView,
        components: [
          { componentId: 'comp-1', x: 0, y: 0 },
          { componentId: 'comp-2', x: 200, y: 150 },
        ],
      };

      useAppStore.setState({
        currentView: mockView,
        components: mockComponents as any,
        relations: mockRelations as any,
      });

      vi.mocked(apiClient.updateMultiplePositions).mockResolvedValueOnce(undefined as any);
      vi.mocked(apiClient.getViewById).mockResolvedValueOnce(mockUpdatedView);

      await useAppStore.getState().applyAutoLayout();

      expect(apiClient.updateMultiplePositions).toHaveBeenCalledWith('view-1', {
        positions: [
          { componentId: 'comp-1', x: 0, y: 0 },
          { componentId: 'comp-2', x: 200, y: 150 },
        ],
      });
      expect(apiClient.getViewById).toHaveBeenCalledWith('view-1');
      expect(useAppStore.getState().currentView).toEqual(mockUpdatedView);
      expect(mockToast.success).toHaveBeenCalledWith('Layout applied');
    });

    it('should handle API error', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [{ componentId: 'comp-1', x: 0, y: 0 }],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({
        currentView: mockView,
        components: [{ id: 'comp-1', name: 'Component 1' }] as any,
        relations: [],
      });

      const apiError = new ApiError('Failed to update positions', 500);
      vi.mocked(apiClient.updateMultiplePositions).mockRejectedValueOnce(apiError);

      await expect(useAppStore.getState().applyAutoLayout()).rejects.toThrow(apiError);

      expect(mockToast.error).toHaveBeenCalledWith('Failed to update positions');
    });

    it('should handle generic error', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [{ componentId: 'comp-1', x: 0, y: 0 }],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({
        currentView: mockView,
        components: [{ id: 'comp-1', name: 'Component 1' }] as any,
        relations: [],
      });

      const genericError = new Error('Network error');
      vi.mocked(apiClient.updateMultiplePositions).mockRejectedValueOnce(genericError);

      await expect(useAppStore.getState().applyAutoLayout()).rejects.toThrow(genericError);

      expect(mockToast.error).toHaveBeenCalledWith('Failed to apply layout');
    });

    it('should do nothing if no current view', async () => {
      useAppStore.setState({ currentView: null });

      await useAppStore.getState().applyAutoLayout();

      expect(apiClient.updateMultiplePositions).not.toHaveBeenCalled();
      expect(mockToast.success).not.toHaveBeenCalled();
    });

    it('should only layout components in current view', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [
          { componentId: 'comp-1', x: 0, y: 0 },
        ],
        edgeType: 'default',
        layoutDirection: 'TB',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      const mockComponents = [
        { id: 'comp-1', name: 'Component 1' },
        { id: 'comp-2', name: 'Component 2' },
      ];

      useAppStore.setState({
        currentView: mockView,
        components: mockComponents as any,
        relations: [],
      });

      vi.mocked(apiClient.updateMultiplePositions).mockResolvedValueOnce(undefined as any);
      vi.mocked(apiClient.getViewById).mockResolvedValueOnce(mockView);

      await useAppStore.getState().applyAutoLayout();

      const call = vi.mocked(apiClient.updateMultiplePositions).mock.calls[0];
      expect(call[1].positions).toHaveLength(1);
      expect(call[1].positions[0].componentId).toBe('comp-1');
    });

    it('should respect layout direction from view', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [
          { componentId: 'comp-1', x: 0, y: 0 },
          { componentId: 'comp-2', x: 0, y: 0 },
        ],
        edgeType: 'default',
        layoutDirection: 'LR',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      const mockComponents = [
        { id: 'comp-1', name: 'Component 1' },
        { id: 'comp-2', name: 'Component 2' },
      ];

      useAppStore.setState({
        currentView: mockView,
        components: mockComponents as any,
        relations: [],
      });

      vi.mocked(apiClient.updateMultiplePositions).mockResolvedValueOnce(undefined as any);
      vi.mocked(apiClient.getViewById).mockResolvedValueOnce(mockView);

      const { calculateDagreLayout } = await import('../utils/layout');

      await useAppStore.getState().applyAutoLayout();

      expect(calculateDagreLayout).toHaveBeenCalledWith(
        expect.any(Array),
        expect.any(Array),
        expect.objectContaining({ direction: 'LR' })
      );
    });
  });
});
