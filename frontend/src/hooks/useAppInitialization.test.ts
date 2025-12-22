import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useAppInitialization } from './useAppInitialization';
import { useAppStore } from '../store/appStore';
import type { View, ViewId } from '../api/types';

const mockCreateViewMutateAsync = vi.fn();

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
  components: [],
  capabilities: [],
  edgeType: 'default',
  colorScheme: 'maturity',
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1' } },
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
    it('should return isLoading true', () => {
      mockUseViewsReturn({ isLoading: true });

      const { result } = renderInitializationHook();

      expect(result.current.isLoading).toBe(true);
      expect(result.current.isInitialized).toBe(false);
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

      renderInitializationHook();

      await waitFor(() => {
        expect(useAppStore.getState().currentViewId).toBe(expectedViewId);
        expect(useAppStore.getState().isInitialized).toBe(true);
      });

      expect(mockToast.success).toHaveBeenCalledWith('Data loaded successfully');
    });
  });

  describe('when no views exist', () => {
    it('should create a default view', async () => {
      const createdView = createMockView({ id: 'new-view' as ViewId, name: 'Default View' });
      mockCreateViewMutateAsync.mockResolvedValue(createdView);
      mockUseViewsReturn({ views: [] });

      renderInitializationHook();

      await waitFor(() => {
        expect(mockCreateViewMutateAsync).toHaveBeenCalledWith({
          name: 'Default View',
          description: 'Main application view',
        });
        expect(useAppStore.getState().currentViewId).toBe('new-view');
        expect(useAppStore.getState().isInitialized).toBe(true);
      });

      expect(mockToast.success).toHaveBeenCalledWith('Created default view');
    });

    it('should handle view creation error', async () => {
      const error = new Error('Failed to create view');
      mockCreateViewMutateAsync.mockRejectedValue(error);
      mockUseViewsReturn({ views: [] });

      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      renderInitializationHook();

      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalledWith('Failed to initialize:', error);
        expect(mockToast.error).toHaveBeenCalledWith('Failed to initialize application');
      });

      consoleSpy.mockRestore();
    });
  });

  describe('when already initialized', () => {
    it('should not re-initialize', async () => {
      useAppStore.setState({
        currentViewId: 'existing-view' as ViewId,
        isInitialized: true,
      });
      mockUseViewsReturn({ views: [createMockView()] });

      const { result } = renderInitializationHook();

      expect(result.current.isInitialized).toBe(true);
      expect(result.current.currentViewId).toBe('existing-view');
      expect(useAppStore.getState().currentViewId).toBe('existing-view');
    });
  });

  describe('when there is an error loading views', () => {
    it('should return error and not be loading', () => {
      const viewsError = new Error('Failed to load views');
      mockUseViewsReturn({ error: viewsError });

      const { result } = renderInitializationHook();

      expect(result.current.error).toBe(viewsError);
      expect(result.current.isLoading).toBe(false);
    });
  });
});
