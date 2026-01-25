import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useAppInitialization } from './useAppInitialization';
import { useAppStore } from '../store/appStore';
import type { View, ViewId } from '../api/types';

const mockCreateViewMutateAsync = vi.fn();
const mockGetParamValue = vi.fn();
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
  getParamValue: (...args: unknown[]) => mockGetParamValue(...args),
  clearParams: (...args: unknown[]) => mockClearParams(...args),
  deepLinkParams: { VIEW: { param: 'view', routes: ['*'] } },
}));

const { useViews } = await import('../features/views/hooks/useViews');
const mockUseViews = vi.mocked(useViews);
const mockToast = await import('react-hot-toast').then(m => m.default);

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
    mockGetParamValue.mockReturnValue(null);
    useAppStore.setState({
      currentViewId: null,
      isInitialized: false,
    });
  });

  afterEach(() => {
    useAppStore.setState({
      currentViewId: null,
      isInitialized: false,
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
      mockGetParamValue.mockReturnValue('view-linked');
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
      const views = [
        createMockView({ id: 'view-default' as ViewId, isDefault: true }),
      ];
      mockGetParamValue.mockReturnValue('non-existent-view');
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
  });
});
