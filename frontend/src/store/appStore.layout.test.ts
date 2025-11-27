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

  describe('setColorScheme', () => {
    it('should update color scheme successfully', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        colorScheme: 'maturity',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({ currentView: mockView });
      vi.mocked(apiClient.updateViewColorScheme).mockResolvedValueOnce(undefined as any);

      await useAppStore.getState().setColorScheme('archimate');

      expect(useAppStore.getState().currentView?.colorScheme).toBe('archimate');
      expect(apiClient.updateViewColorScheme).toHaveBeenCalledWith('view-1', { colorScheme: 'archimate' });
      expect(mockToast.success).toHaveBeenCalledWith('Color scheme updated');
    });

    it('should rollback on API error', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        colorScheme: 'maturity',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({ currentView: mockView });
      const apiError = new ApiError('Failed to update', 500);
      vi.mocked(apiClient.updateViewColorScheme).mockRejectedValueOnce(apiError);

      await expect(useAppStore.getState().setColorScheme('archimate')).rejects.toThrow(apiError);

      expect(useAppStore.getState().currentView?.colorScheme).toBe('maturity');
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
        colorScheme: 'maturity',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      useAppStore.setState({ currentView: mockView });
      const genericError = new Error('Network error');
      vi.mocked(apiClient.updateViewColorScheme).mockRejectedValueOnce(genericError);

      await expect(useAppStore.getState().setColorScheme('archimate')).rejects.toThrow(genericError);

      expect(useAppStore.getState().currentView?.colorScheme).toBe('maturity');
      expect(mockToast.error).toHaveBeenCalledWith('Failed to update color scheme');
    });

    it('should do nothing if no current view', async () => {
      useAppStore.setState({ currentView: null });

      await useAppStore.getState().setColorScheme('archimate');

      expect(apiClient.updateViewColorScheme).not.toHaveBeenCalled();
      expect(mockToast.success).not.toHaveBeenCalled();
    });

    it('should handle all valid color schemes', async () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: false,
        components: [],
        edgeType: 'default',
        layoutDirection: 'TB',
        colorScheme: 'maturity',
        createdAt: new Date().toISOString(),
        _links: { self: { href: '/api/views/view-1' } },
      };

      const colorSchemes = ['maturity', 'archimate', 'archimate-classic'];

      for (const colorScheme of colorSchemes) {
        useAppStore.setState({ currentView: { ...mockView, colorScheme: 'maturity' } });
        vi.mocked(apiClient.updateViewColorScheme).mockResolvedValueOnce(undefined as any);

        await useAppStore.getState().setColorScheme(colorScheme);

        expect(useAppStore.getState().currentView?.colorScheme).toBe(colorScheme);
      }
    });
  });
});
