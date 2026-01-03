import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useViews,
  useView,
  useCreateView,
  useDeleteView,
  useRenameView,
  useSetDefaultView,
  useAddComponentToView,
  useRemoveComponentFromView,
} from './useViews';
import { queryKeys } from '../../../lib/queryClient';
import { buildView } from '../../../test/helpers/entityBuilders';
import type { ViewId, ComponentId } from '../../../api/types';

vi.mock('../api', () => ({
  viewsApi: {
    getAll: vi.fn(),
    getById: vi.fn(),
    getComponents: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
    rename: vi.fn(),
    setDefault: vi.fn(),
    addComponent: vi.fn(),
    removeComponent: vi.fn(),
    updateComponentPosition: vi.fn(),
    updateMultiplePositions: vi.fn(),
    addCapability: vi.fn(),
    removeCapability: vi.fn(),
    updateCapabilityPosition: vi.fn(),
    updateEdgeType: vi.fn(),
    updateColorScheme: vi.fn(),
    updateComponentColor: vi.fn(),
    clearComponentColor: vi.fn(),
    updateCapabilityColor: vi.fn(),
    clearCapabilityColor: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { viewsApi } from '../api';
import toast from 'react-hot-toast';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useViews hooks', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('useViews', () => {
    it('should fetch all views', async () => {
      const views = [
        buildView({ id: 'view-1' as ViewId, name: 'View 1' }),
        buildView({ id: 'view-2' as ViewId, name: 'View 2' }),
      ];
      vi.mocked(viewsApi.getAll).mockResolvedValue(views);

      const { result } = renderHook(() => useViews(), {
        wrapper: createWrapper(queryClient),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(views);
      expect(viewsApi.getAll).toHaveBeenCalledTimes(1);
    });

    it('should handle fetch error', async () => {
      const error = new Error('Failed to fetch views');
      vi.mocked(viewsApi.getAll).mockRejectedValue(error);

      const { result } = renderHook(() => useViews(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('useView', () => {
    it('should fetch a single view by id', async () => {
      const view = buildView({ id: 'view-1' as ViewId, name: 'Test View' });
      vi.mocked(viewsApi.getById).mockResolvedValue(view);

      const { result } = renderHook(() => useView('view-1' as ViewId), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(view);
      expect(viewsApi.getById).toHaveBeenCalledWith('view-1');
    });

    it('should not fetch when id is undefined', async () => {
      const { result } = renderHook(() => useView(undefined), {
        wrapper: createWrapper(queryClient),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.isFetching).toBe(false);
      expect(viewsApi.getById).not.toHaveBeenCalled();
    });
  });

  describe('useCreateView', () => {
    it('should create a view and update cache', async () => {
      const existingViews = [buildView({ id: 'view-1' as ViewId, name: 'Existing' })];
      const newView = buildView({ id: 'view-2' as ViewId, name: 'New View' });

      queryClient.setQueryData(queryKeys.views.lists(), existingViews);
      vi.mocked(viewsApi.create).mockResolvedValue(newView);

      const { result } = renderHook(() => useCreateView(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          name: 'New View',
          description: 'Test description',
        });
      });

      expect(viewsApi.create).toHaveBeenCalledWith({
        name: 'New View',
        description: 'Test description',
      });

      const cachedViews = queryClient.getQueryData<View[]>(queryKeys.views.lists());
      expect(cachedViews).toHaveLength(2);
      expect(cachedViews?.[1]).toEqual(newView);
      expect(toast.success).toHaveBeenCalledWith('View "New View" created');
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Name is required');
      vi.mocked(viewsApi.create).mockRejectedValue(error);

      const { result } = renderHook(() => useCreateView(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({ name: '' });
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Name is required');
    });
  });

  describe('useDeleteView', () => {
    it('should delete view and remove from cache', async () => {
      const view = buildView({ id: 'view-1' as ViewId, name: 'To Delete' });

      queryClient.setQueryData(queryKeys.views.lists(), [view]);
      queryClient.setQueryData(queryKeys.views.detail('view-1'), view);
      vi.mocked(viewsApi.delete).mockResolvedValue(undefined);

      const removeQueriesSpy = vi.spyOn(queryClient, 'removeQueries');

      const { result } = renderHook(() => useDeleteView(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync('view-1' as ViewId);
      });

      const cachedViews = queryClient.getQueryData<View[]>(queryKeys.views.lists());
      expect(cachedViews).toHaveLength(0);

      expect(removeQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.views.detail('view-1'),
      });

      expect(toast.success).toHaveBeenCalledWith('View deleted');
    });
  });

  describe('useRenameView', () => {
    it('should rename view and update cache', async () => {
      const view = buildView({ id: 'view-1' as ViewId, name: 'Original Name' });

      queryClient.setQueryData(queryKeys.views.lists(), [view]);
      vi.mocked(viewsApi.rename).mockResolvedValue(undefined);

      const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const { result } = renderHook(() => useRenameView(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          viewId: 'view-1' as ViewId,
          request: { name: 'New Name' },
        });
      });

      const cachedViews = queryClient.getQueryData<View[]>(queryKeys.views.lists());
      expect(cachedViews?.[0].name).toBe('New Name');

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.views.detail('view-1'),
      });

      expect(toast.success).toHaveBeenCalledWith('View renamed');
    });
  });

  describe('useSetDefaultView', () => {
    it('should set default view and update all views in cache', async () => {
      const views = [
        buildView({ id: 'view-1' as ViewId, name: 'View 1', isDefault: true }),
        buildView({ id: 'view-2' as ViewId, name: 'View 2', isDefault: false }),
      ];

      queryClient.setQueryData(queryKeys.views.lists(), views);
      vi.mocked(viewsApi.setDefault).mockResolvedValue(undefined);

      const { result } = renderHook(() => useSetDefaultView(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync('view-2' as ViewId);
      });

      const cachedViews = queryClient.getQueryData<View[]>(queryKeys.views.lists());
      expect(cachedViews?.[0].isDefault).toBe(false);
      expect(cachedViews?.[1].isDefault).toBe(true);

      expect(toast.success).toHaveBeenCalledWith('Default view updated');
    });
  });

  describe('useAddComponentToView', () => {
    it('should add component to view and invalidate view detail', async () => {
      vi.mocked(viewsApi.addComponent).mockResolvedValue(undefined);
      const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const { result } = renderHook(() => useAddComponentToView(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          viewId: 'view-1' as ViewId,
          request: {
            componentId: 'comp-1' as ComponentId,
            x: 100,
            y: 200,
          },
        });
      });

      expect(viewsApi.addComponent).toHaveBeenCalledWith('view-1', {
        componentId: 'comp-1',
        x: 100,
        y: 200,
      });

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.views.detail('view-1'),
      });
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Component already in view');
      vi.mocked(viewsApi.addComponent).mockRejectedValue(error);

      const { result } = renderHook(() => useAddComponentToView(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({
            viewId: 'view-1' as ViewId,
            request: { componentId: 'comp-1' as ComponentId, x: 0, y: 0 },
          });
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Component already in view');
    });
  });

  describe('useRemoveComponentFromView', () => {
    it('should remove component from view and invalidate view detail', async () => {
      vi.mocked(viewsApi.removeComponent).mockResolvedValue(undefined);
      const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const { result } = renderHook(() => useRemoveComponentFromView(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          viewId: 'view-1' as ViewId,
          componentId: 'comp-1' as ComponentId,
        });
      });

      expect(viewsApi.removeComponent).toHaveBeenCalledWith('view-1', 'comp-1');

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.views.detail('view-1'),
      });
    });
  });
});
