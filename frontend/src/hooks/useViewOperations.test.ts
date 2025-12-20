import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useViewOperations } from './useViewOperations';
import { useAppStore } from '../store/appStore';
import apiClient from '../api/client';
import type { View, ViewId } from '../api/types';

vi.mock('../api/client');
vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

const originalError = console.error;
beforeEach(() => {
  console.error = (...args: unknown[]) => {
    const message = args[0];
    if (typeof message === 'string' && message.includes('not wrapped in act')) {
      return;
    }
    originalError.apply(console, args);
  };
});

afterEach(() => {
  console.error = originalError;
});

describe('useViewOperations', () => {
  const mockView1: View = {
    id: 'view-1' as ViewId,
    name: 'View 1',
    description: 'First view',
    isDefault: true,
    components: [],
    capabilities: [],
    edgeType: 'default',
    colorScheme: 'maturity',
  };

  const mockView2: View = {
    id: 'view-2' as ViewId,
    name: 'View 2',
    description: 'Second view',
    isDefault: false,
    components: [],
    capabilities: [],
    edgeType: 'default',
    colorScheme: 'maturity',
  };

  beforeEach(() => {
    vi.clearAllMocks();
    useAppStore.setState({
      currentView: mockView1,
      views: [mockView1, mockView2],
      components: [],
      relations: [],
      capabilities: [],
      canvasCapabilities: [],
      selectedNodeId: null,
      selectedEdgeId: null,
      selectedCapabilityId: null,
      isLoading: false,
      error: null,
    });
  });

  afterEach(() => {
    useAppStore.setState({
      currentView: null,
      views: [],
    });
  });

  describe('switchView', () => {
    it('should update currentView in store after switching', async () => {
      vi.mocked(apiClient.getViewById).mockResolvedValue(mockView2);

      const { result } = renderHook(() => useViewOperations());

      await act(async () => {
        await result.current.switchView('view-2' as ViewId);
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      expect(useAppStore.getState().currentView).toEqual(mockView2);
    });
  });

  describe('addComponentToView after view switch', () => {
    it('should use the NEW view ID after switching views', async () => {
      vi.mocked(apiClient.getViewById).mockResolvedValue(mockView2);
      vi.mocked(apiClient.addComponentToView).mockResolvedValue(undefined);

      const { result, rerender } = renderHook(() => useViewOperations());

      await act(async () => {
        await result.current.switchView('view-2' as ViewId);
        await new Promise(resolve => setTimeout(resolve, 0));
        rerender();
      });

      await waitFor(() => {
        expect(useAppStore.getState().currentView?.id).toBe('view-2');
      });

      vi.mocked(apiClient.getViewById).mockResolvedValue({
        ...mockView2,
        components: [{ componentId: 'comp-1', x: 100, y: 200 }],
      });

      await act(async () => {
        await result.current.addComponentToView('comp-1' as import('../api/types').ComponentId, 100, 200);
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      expect(apiClient.addComponentToView).toHaveBeenCalledWith(
        'view-2',
        expect.objectContaining({ componentId: 'comp-1' })
      );
    });

    it('should NOT use the OLD view ID after switching views', async () => {
      vi.mocked(apiClient.getViewById).mockResolvedValue(mockView2);
      vi.mocked(apiClient.addComponentToView).mockResolvedValue(undefined);

      const { result, rerender } = renderHook(() => useViewOperations());

      expect(useAppStore.getState().currentView?.id).toBe('view-1');

      await act(async () => {
        await result.current.switchView('view-2' as ViewId);
        await new Promise(resolve => setTimeout(resolve, 0));
        rerender();
      });

      await waitFor(() => {
        expect(useAppStore.getState().currentView?.id).toBe('view-2');
      });

      vi.mocked(apiClient.getViewById).mockResolvedValue({
        ...mockView2,
        components: [{ componentId: 'comp-1', x: 100, y: 200 }],
      });

      await act(async () => {
        await result.current.addComponentToView('comp-1' as import('../api/types').ComponentId, 100, 200);
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      expect(apiClient.addComponentToView).not.toHaveBeenCalledWith(
        'view-1',
        expect.anything()
      );
    });

    it('should warn when currentView is null', async () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

      useAppStore.setState({ currentView: null });

      const { result } = renderHook(() => useViewOperations());

      await act(async () => {
        await result.current.addComponentToView('comp-1' as import('../api/types').ComponentId, 100, 200);
      });

      expect(consoleSpy).toHaveBeenCalledWith('No current view selected');
      expect(apiClient.addComponentToView).not.toHaveBeenCalled();

      consoleSpy.mockRestore();
    });
  });
});
