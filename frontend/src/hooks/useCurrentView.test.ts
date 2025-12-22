import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useCurrentView } from './useCurrentView';
import { useAppStore } from '../store/appStore';
import type { View, ViewId } from '../api/types';

vi.mock('../features/views/hooks/useViews', () => ({
  useView: vi.fn(),
}));

const { useView } = await import('../features/views/hooks/useViews');
const mockUseView = vi.mocked(useView);

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

interface MockUseViewOptions {
  view?: View;
  isLoading?: boolean;
  error?: Error | null;
}

function mockUseViewReturn({ view, isLoading = false, error = null }: MockUseViewOptions = {}) {
  mockUseView.mockReturnValue({
    data: view,
    isLoading,
    error,
  } as ReturnType<typeof useView>);
}

function renderCurrentViewHook() {
  return renderHook(() => useCurrentView(), { wrapper: createWrapper() });
}

describe('useCurrentView', () => {
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

  describe('when no view is selected', () => {
    it('should return null values and call useView with undefined', () => {
      mockUseViewReturn();

      const { result } = renderCurrentViewHook();

      expect(result.current.currentView).toBeNull();
      expect(result.current.currentViewId).toBeNull();
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(mockUseView).toHaveBeenCalledWith(undefined);
    });
  });

  describe('when a view is selected', () => {
    it('should return view data and call useView with viewId', () => {
      const mockView = createMockView({ id: 'view-123' as ViewId });
      useAppStore.setState({ currentViewId: 'view-123' as ViewId });
      mockUseViewReturn({ view: mockView });

      const { result } = renderCurrentViewHook();

      expect(result.current.currentView).toEqual(mockView);
      expect(result.current.currentViewId).toBe('view-123');
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(mockUseView).toHaveBeenCalledWith('view-123');
    });
  });

  describe('when view is loading', () => {
    it('should return isLoading true and null currentView', () => {
      useAppStore.setState({ currentViewId: 'view-1' as ViewId });
      mockUseViewReturn({ isLoading: true });

      const { result } = renderCurrentViewHook();

      expect(result.current.currentViewId).toBe('view-1');
      expect(result.current.isLoading).toBe(true);
      expect(result.current.currentView).toBeNull();
    });
  });

  describe('when there is an error', () => {
    it('should return the error with null currentView', () => {
      useAppStore.setState({ currentViewId: 'view-1' as ViewId });
      const viewError = new Error('Failed to load view');
      mockUseViewReturn({ error: viewError });

      const { result } = renderCurrentViewHook();

      expect(result.current.currentViewId).toBe('view-1');
      expect(result.current.error).toBe(viewError);
      expect(result.current.currentView).toBeNull();
      expect(result.current.isLoading).toBe(false);
    });
  });

  describe('when store currentViewId changes', () => {
    it('should update the view being fetched', () => {
      useAppStore.setState({ currentViewId: 'view-1' as ViewId });
      mockUseViewReturn({ view: createMockView({ id: 'view-1' as ViewId }) });

      const { result, rerender } = renderCurrentViewHook();

      expect(result.current.currentViewId).toBe('view-1');

      act(() => {
        useAppStore.setState({ currentViewId: 'view-2' as ViewId });
        mockUseViewReturn({ view: createMockView({ id: 'view-2' as ViewId }) });
        rerender();
      });

      expect(result.current.currentViewId).toBe('view-2');
    });
  });
});
