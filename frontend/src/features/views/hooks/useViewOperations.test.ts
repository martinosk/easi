import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useViewOperations } from './useViewOperations';
import { useAppStore } from '../../../store/appStore';
import type { ViewId, ComponentId } from '../../../api/types';

vi.mock('../../../api/client');
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

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    React.createElement(QueryClientProvider, { client: queryClient }, children)
  );
}

describe('useViewOperations', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAppStore.setState({
      currentViewId: 'view-1' as ViewId,
      isInitialized: true,
      selectedNodeId: null,
      selectedEdgeId: null,
      selectedCapabilityId: null,
    });
  });

  afterEach(() => {
    useAppStore.setState({
      currentViewId: null,
      isInitialized: false,
    });
  });

  describe('switchView', () => {
    it('should update currentViewId in store', () => {
      const { result } = renderHook(() => useViewOperations(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.switchView('view-2' as ViewId);
      });

      expect(useAppStore.getState().currentViewId).toBe('view-2');
    });
  });

  describe('operations requiring currentViewId', () => {
    it.each([
      {
        operation: 'addComponentToView',
        execute: (ops: ReturnType<typeof useViewOperations>) =>
          ops.addComponentToView('comp-1' as ComponentId, 100, 200),
      },
      {
        operation: 'removeComponentFromView',
        execute: (ops: ReturnType<typeof useViewOperations>) =>
          ops.removeComponentFromView('comp-1' as ComponentId),
      },
    ])('$operation should warn when currentViewId is null', async ({ execute }) => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      useAppStore.setState({ currentViewId: null });

      const { result } = renderHook(() => useViewOperations(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        await execute(result.current);
      });

      expect(consoleSpy).toHaveBeenCalledWith('No current view selected');
      consoleSpy.mockRestore();
    });
  });
});
