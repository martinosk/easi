import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { View, ViewId } from '../api/types';
import { useAppStore } from '../store/appStore';
import { useAppInitialization } from './useAppInitialization';

const mockCreateViewMutateAsync = vi.fn();
const mockReadDeepLink = vi.fn();
const mockClearParams = vi.fn();

vi.mock('../features/views/hooks/useViews', () => ({
  useViews: vi.fn(),
  useCreateView: () => ({
    mutateAsync: mockCreateViewMutateAsync,
    isPending: false,
  }),
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock('../lib/deepLinks', () => ({
  readDeepLink: (...args: unknown[]) => mockReadDeepLink(...args),
  clearParams: (...args: unknown[]) => mockClearParams(...args),
  deepLinkParams: { VIEW: { param: 'view', routes: ['/canvas'] } },
}));

const { useViews } = await import('../features/views/hooks/useViews');
const mockUseViews = vi.mocked(useViews);
const mockToast = await import('react-hot-toast').then((m) => m.default);

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

const createMockView = (overrides: Partial<View> = {}): View => ({
  id: 'view-1' as ViewId,
  name: 'Test View',
  description: 'Test description',
  isDefault: false,
  isPrivate: false,
  components: [],
  capabilities: [],
  originEntities: [],
  edgeType: 'default',
  colorScheme: 'maturity',
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1', method: 'GET' } },
  ...overrides,
});

interface MockUseViewsOptions {
  views?: View[];
  isLoading?: boolean;
  error?: Error | null;
}

function mockUseViewsReturn({ views, isLoading = false, error = null }: MockUseViewsOptions) {
  mockUseViews.mockReturnValue({
    data: views,
    isLoading,
    error,
  } as ReturnType<typeof useViews>);
}

function renderInitializationHook() {
  return renderHook(() => useAppInitialization(), {
    wrapper: createWrapper(),
  });
}

describe('useAppInitialization', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockReadDeepLink.mockReturnValue(null);
    mockClearParams.mockImplementation(() => mockReadDeepLink.mockReturnValue(null));
    useAppStore.setState({
      currentViewId: null,
      isInitialized: false,
      openViewIds: [],
    });
  });

  afterEach(() => {
    useAppStore.setState({
      currentViewId: null,
      isInitialized: false,
      openViewIds: [],
    });
  });

  describe('when views are loading', () => {
    it('should return isLoading true', async () => {
      mockUseViewsReturn({ isLoading: true });

      const { result, unmount } = renderInitializationHook();

      expect(result.current.isLoading).toBe(true);
      expect(result.current.isInitialized).toBe(false);

      await act(async () => {
        unmount();
      });
    });
  });

  describe('when views exist', () => {
    it.each([
      {
        scenario: 'default view available',
        views: [
          createMockView({ id: 'view-other' as ViewId, isDefault: false }),
          createMockView({ id: 'view-default' as ViewId, isDefault: true }),
        ],
        expectedViewId: 'view-default',
      },
      {
        scenario: 'no default view',
        views: [
          createMockView({ id: 'view-first' as ViewId, isDefault: false }),
          createMockView({ id: 'view-second' as ViewId, isDefault: false }),
        ],
        expectedViewId: 'view-first',
      },
    ])('should select correct view when $scenario', async ({ views, expectedViewId }) => {
      mockUseViewsReturn({ views });

      const { result, unmount } = renderInitializationHook();

      await waitFor(() => {
        expect(result.current.isInitialized).toBe(true);
      });

      expect(useAppStore.getState().currentViewId).toBe(expectedViewId);
      expect(mockToast.success).toHaveBeenCalledWith('Data loaded successfully');

      await act(async () => {
        unmount();
      });
    });
  });

  describe('when no views exist', () => {
    it('should create a default view', async () => {
      const createdView = createMockView({ id: 'new-view' as ViewId, name: 'Default View' });
      mockCreateViewMutateAsync.mockResolvedValue(createdView);
      mockUseViewsReturn({ views: [] });

      const { result, unmount } = renderInitializationHook();

      await waitFor(() => {
        expect(result.current.isInitialized).toBe(true);
      });

      expect(mockCreateViewMutateAsync).toHaveBeenCalledWith({
        name: 'Default View',
        description: 'Main application view',
      });
      expect(useAppStore.getState().currentViewId).toBe('new-view');
      expect(mockToast.success).toHaveBeenCalledWith('Created default view');

      await act(async () => {
        unmount();
      });
    });

    it('should handle view creation error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const error = new Error('Failed to create view');
      mockCreateViewMutateAsync.mockRejectedValue(error);
      mockUseViewsReturn({ views: [] });

      const { unmount } = renderInitializationHook();

      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalledWith('Failed to initialize:', expect.any(Error));
        expect(mockToast.error).toHaveBeenCalledWith('Failed to initialize application');
      });

      consoleSpy.mockRestore();

      await act(async () => {
        unmount();
      });
    });

    it('should not start a second creation when the canvas is reopened while the first is in flight', async () => {
      const createdView = createMockView({ id: 'new-view' as ViewId, name: 'Default View' });
      let finishCreation: (view: View) => void = () => {};
      mockCreateViewMutateAsync.mockReturnValue(
        new Promise<View>((resolve) => {
          finishCreation = resolve;
        }),
      );
      mockUseViewsReturn({ views: [] });

      const first = renderInitializationHook();
      await waitFor(() => {
        expect(mockCreateViewMutateAsync).toHaveBeenCalledTimes(1);
      });
      await act(async () => {
        first.unmount();
      });

      const second = renderInitializationHook();
      await act(async () => {
        await Promise.resolve();
      });
      expect(mockCreateViewMutateAsync).toHaveBeenCalledTimes(1);

      await act(async () => {
        finishCreation(createdView);
      });
      await waitFor(() => {
        expect(second.result.current.isInitialized).toBe(true);
      });

      expect(mockCreateViewMutateAsync).toHaveBeenCalledTimes(1);
      expect(useAppStore.getState().currentViewId).toBe('new-view');

      await act(async () => {
        second.unmount();
      });
    });

    it('should allow a new attempt after a failed creation', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      mockCreateViewMutateAsync.mockRejectedValueOnce(new Error('Failed to create view'));
      mockUseViewsReturn({ views: [] });

      const first = renderInitializationHook();
      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to initialize application');
      });
      await act(async () => {
        first.unmount();
      });

      mockCreateViewMutateAsync.mockResolvedValue(createMockView({ id: 'new-view' as ViewId }));
      const second = renderInitializationHook();
      await waitFor(() => {
        expect(second.result.current.isInitialized).toBe(true);
      });

      expect(mockCreateViewMutateAsync).toHaveBeenCalledTimes(2);
      consoleSpy.mockRestore();

      await act(async () => {
        second.unmount();
      });
    });
  });

  describe('when already initialized', () => {
    it('should not re-initialize', async () => {
      useAppStore.setState({
        currentViewId: 'existing-view' as ViewId,
        isInitialized: true,
      });
      mockUseViewsReturn({ views: [createMockView()] });

      const { result, unmount } = renderInitializationHook();

      expect(result.current.isInitialized).toBe(true);
      expect(result.current.currentViewId).toBe('existing-view');
      expect(useAppStore.getState().currentViewId).toBe('existing-view');

      await act(async () => {
        unmount();
      });
    });
  });

  describe('when there is an error loading views', () => {
    it('should return error and not be loading', async () => {
      const viewsError = new Error('Failed to load views');
      mockUseViewsReturn({ error: viewsError });

      const { result, unmount } = renderInitializationHook();

      expect(result.current.error).toBe(viewsError);
      expect(result.current.isLoading).toBe(false);

      await act(async () => {
        unmount();
      });
    });
  });

  describe('view deep linking', () => {
    it('should select view from URL parameter when valid', async () => {
      const views = [
        createMockView({ id: 'view-1' as ViewId, isDefault: true }),
        createMockView({ id: 'view-linked' as ViewId, isDefault: false }),
      ];
      mockReadDeepLink.mockReturnValue('view-linked');
      mockUseViewsReturn({ views });

      const { result, unmount } = renderInitializationHook();

      await waitFor(() => {
        expect(result.current.isInitialized).toBe(true);
      });

      expect(useAppStore.getState().currentViewId).toBe('view-linked');
      expect(mockClearParams).toHaveBeenCalledWith(['view']);

      await act(async () => {
        unmount();
      });
    });

    it('should show error and fall back to default when view ID is invalid', async () => {
      const views = [createMockView({ id: 'view-default' as ViewId, isDefault: true })];
      mockReadDeepLink.mockReturnValue('non-existent-view');
      mockUseViewsReturn({ views });

      const { result, unmount } = renderInitializationHook();

      await waitFor(() => {
        expect(result.current.isInitialized).toBe(true);
      });

      expect(mockToast.error).toHaveBeenCalledWith('The linked view does not exist');
      expect(useAppStore.getState().currentViewId).toBe('view-default');
      expect(mockClearParams).toHaveBeenCalledWith(['view']);

      await act(async () => {
        unmount();
      });
    });

    it('should clear an empty view parameter and select the default view without an error', async () => {
      const views = [createMockView({ id: 'view-default' as ViewId, isDefault: true })];
      mockReadDeepLink.mockReturnValue('');
      mockUseViewsReturn({ views });

      const { result, unmount } = renderInitializationHook();

      await waitFor(() => {
        expect(result.current.isInitialized).toBe(true);
      });

      expect(useAppStore.getState().currentViewId).toBe('view-default');
      expect(mockClearParams).toHaveBeenCalledWith(['view']);
      expect(mockToast.error).not.toHaveBeenCalled();

      await act(async () => {
        unmount();
      });
    });

    it('should leave the URL alone when there is no view parameter', async () => {
      mockUseViewsReturn({ views: [createMockView({ id: 'view-default' as ViewId, isDefault: true })] });

      const { result, unmount } = renderInitializationHook();

      await waitFor(() => {
        expect(result.current.isInitialized).toBe(true);
      });

      expect(mockClearParams).not.toHaveBeenCalled();

      await act(async () => {
        unmount();
      });
    });

    describe('when the canvas is reopened after initialisation', () => {
      const views = [
        createMockView({ id: 'view-1' as ViewId, isDefault: true }),
        createMockView({ id: 'view-linked' as ViewId, isDefault: false }),
      ];

      beforeEach(() => {
        useAppStore.setState({
          currentViewId: 'view-1' as ViewId,
          isInitialized: true,
          openViewIds: ['view-1' as ViewId],
        });
        mockUseViewsReturn({ views });
      });

      it('should open the linked view next to the open ones and clear the parameter', async () => {
        mockReadDeepLink.mockReturnValue('view-linked');

        const { unmount } = renderInitializationHook();

        await waitFor(() => {
          expect(useAppStore.getState().currentViewId).toBe('view-linked');
        });
        expect(useAppStore.getState().openViewIds).toEqual(['view-1', 'view-linked']);
        expect(mockClearParams).toHaveBeenCalledWith(['view']);
        expect(mockToast.success).not.toHaveBeenCalled();

        await act(async () => {
          unmount();
        });
      });

      it('should keep the current view and report a linked view that does not exist', async () => {
        mockReadDeepLink.mockReturnValue('non-existent-view');

        const { unmount } = renderInitializationHook();

        await waitFor(() => {
          expect(mockToast.error).toHaveBeenCalledWith('The linked view does not exist');
        });
        expect(useAppStore.getState().currentViewId).toBe('view-1');
        expect(mockClearParams).toHaveBeenCalledWith(['view']);

        await act(async () => {
          unmount();
        });
      });
    });
  });
});
