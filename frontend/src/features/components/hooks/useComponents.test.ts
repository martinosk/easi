import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useComponents,
  useComponent,
  useCreateComponent,
  useUpdateComponent,
  useDeleteComponent,
} from './useComponents';
import { queryKeys } from '../../../lib/queryClient';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import type { ComponentId } from '../../../api/types';

vi.mock('../api', () => ({
  componentsApi: {
    getAll: vi.fn(),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { componentsApi } from '../api';
import toast from 'react-hot-toast';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useComponents hooks', () => {
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

  describe('useComponents', () => {
    it('should fetch all components', async () => {
      const components = [
        buildComponent({ id: 'comp-1' as ComponentId, name: 'Component 1' }),
        buildComponent({ id: 'comp-2' as ComponentId, name: 'Component 2' }),
      ];
      vi.mocked(componentsApi.getAll).mockResolvedValue(components);

      const { result } = renderHook(() => useComponents(), {
        wrapper: createWrapper(queryClient),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(components);
      expect(componentsApi.getAll).toHaveBeenCalledTimes(1);
    });

    it('should handle fetch error', async () => {
      const error = new Error('Failed to fetch components');
      vi.mocked(componentsApi.getAll).mockRejectedValue(error);

      const { result } = renderHook(() => useComponents(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('useComponent', () => {
    it('should fetch a single component by id', async () => {
      const component = buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'Test Component',
      });
      vi.mocked(componentsApi.getById).mockResolvedValue(component);

      const { result } = renderHook(() => useComponent('comp-1' as ComponentId), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(component);
      expect(componentsApi.getById).toHaveBeenCalledWith('comp-1');
    });

    it('should not fetch when id is undefined', async () => {
      const { result } = renderHook(() => useComponent(undefined), {
        wrapper: createWrapper(queryClient),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.isFetching).toBe(false);
      expect(componentsApi.getById).not.toHaveBeenCalled();
    });
  });

  describe('useCreateComponent', () => {
    it('should create a component and invalidate cache', async () => {
      const newComponent = buildComponent({
        id: 'comp-2' as ComponentId,
        name: 'New Component',
      });

      vi.mocked(componentsApi.create).mockResolvedValue(newComponent);
      const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const { result } = renderHook(() => useCreateComponent(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          name: 'New Component',
          description: 'Test description',
        });
      });

      expect(componentsApi.create).toHaveBeenCalledWith({
        name: 'New Component',
        description: 'Test description',
      });

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.components.lists(),
      });
      expect(toast.success).toHaveBeenCalledWith('Component "New Component" created');
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Component name already exists');
      vi.mocked(componentsApi.create).mockRejectedValue(error);

      const { result } = renderHook(() => useCreateComponent(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({ name: 'Duplicate' });
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Component name already exists');
    });
  });

  describe('useUpdateComponent', () => {
    it('should update component and invalidate both list and detail cache', async () => {
      const existingComponent = buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'Original Name',
      });
      const updatedComponent = buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'Updated Name',
      });

      queryClient.setQueryData(queryKeys.components.lists(), [existingComponent]);
      queryClient.setQueryData(queryKeys.components.detail('comp-1'), existingComponent);
      vi.mocked(componentsApi.update).mockResolvedValue(updatedComponent);

      const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const { result } = renderHook(() => useUpdateComponent(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          id: 'comp-1' as ComponentId,
          request: { name: 'Updated Name' },
        });
      });

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.components.lists(),
      });
      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.components.detail('comp-1'),
      });

      expect(toast.success).toHaveBeenCalledWith('Component "Updated Name" updated');
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Update failed');
      vi.mocked(componentsApi.update).mockRejectedValue(error);

      const { result } = renderHook(() => useUpdateComponent(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({
            id: 'comp-1' as ComponentId,
            request: { name: 'New Name' },
          });
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Update failed');
    });
  });

  describe('useDeleteComponent', () => {
    it('should delete component and invalidate cache', async () => {
      const component = buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'To Delete',
      });

      queryClient.setQueryData(queryKeys.components.lists(), [component]);
      queryClient.setQueryData(queryKeys.components.detail('comp-1'), component);
      vi.mocked(componentsApi.delete).mockResolvedValue(undefined);

      const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const { result } = renderHook(() => useDeleteComponent(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync('comp-1' as ComponentId);
      });

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.components.lists(),
      });
      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.components.detail('comp-1'),
      });

      expect(toast.success).toHaveBeenCalledWith('Component deleted');
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Cannot delete component in use');
      vi.mocked(componentsApi.delete).mockRejectedValue(error);

      const { result } = renderHook(() => useDeleteComponent(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync('comp-1' as ComponentId);
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Cannot delete component in use');
    });
  });
});
