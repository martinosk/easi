import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useRelations,
  useRelation,
  useCreateRelation,
  useUpdateRelation,
  useDeleteRelation,
} from './useRelations';
import { queryKeys } from '../../../lib/queryClient';
import { buildRelation } from '../../../test/helpers/entityBuilders';
import type { Relation, RelationId, ComponentId } from '../../../api/types';

vi.mock('../api', () => ({
  relationsApi: {
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

import { relationsApi } from '../api';
import toast from 'react-hot-toast';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useRelations hooks', () => {
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

  describe('useRelations', () => {
    it('should fetch all relations', async () => {
      const relations = [
        buildRelation({
          id: 'rel-1' as RelationId,
          sourceComponentId: 'comp-1' as ComponentId,
          targetComponentId: 'comp-2' as ComponentId,
        }),
        buildRelation({
          id: 'rel-2' as RelationId,
          sourceComponentId: 'comp-2' as ComponentId,
          targetComponentId: 'comp-3' as ComponentId,
        }),
      ];
      vi.mocked(relationsApi.getAll).mockResolvedValue(relations);

      const { result } = renderHook(() => useRelations(), {
        wrapper: createWrapper(queryClient),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(relations);
      expect(relationsApi.getAll).toHaveBeenCalledTimes(1);
    });

    it('should handle fetch error', async () => {
      const error = new Error('Failed to fetch relations');
      vi.mocked(relationsApi.getAll).mockRejectedValue(error);

      const { result } = renderHook(() => useRelations(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('useRelation', () => {
    it('should fetch a single relation by id', async () => {
      const relation = buildRelation({
        id: 'rel-1' as RelationId,
        name: 'Test Relation',
      });
      vi.mocked(relationsApi.getById).mockResolvedValue(relation);

      const { result } = renderHook(() => useRelation('rel-1' as RelationId), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(relation);
      expect(relationsApi.getById).toHaveBeenCalledWith('rel-1');
    });

    it('should not fetch when id is undefined', async () => {
      const { result } = renderHook(() => useRelation(undefined), {
        wrapper: createWrapper(queryClient),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.isFetching).toBe(false);
      expect(relationsApi.getById).not.toHaveBeenCalled();
    });
  });

  describe('useCreateRelation', () => {
    it('should create a relation and update cache', async () => {
      const existingRelations = [
        buildRelation({ id: 'rel-1' as RelationId }),
      ];
      const newRelation = buildRelation({
        id: 'rel-2' as RelationId,
        sourceComponentId: 'comp-a' as ComponentId,
        targetComponentId: 'comp-b' as ComponentId,
        relationType: 'Triggers',
        name: 'New Relation',
      });

      queryClient.setQueryData(queryKeys.relations.lists(), existingRelations);
      vi.mocked(relationsApi.create).mockResolvedValue(newRelation);

      const { result } = renderHook(() => useCreateRelation(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          sourceComponentId: 'comp-a' as ComponentId,
          targetComponentId: 'comp-b' as ComponentId,
          relationType: 'Triggers',
          name: 'New Relation',
        });
      });

      expect(relationsApi.create).toHaveBeenCalledWith({
        sourceComponentId: 'comp-a',
        targetComponentId: 'comp-b',
        relationType: 'Triggers',
        name: 'New Relation',
      });

      const cachedRelations = queryClient.getQueryData<Relation[]>(
        queryKeys.relations.lists()
      );
      expect(cachedRelations).toHaveLength(2);
      expect(cachedRelations?.[1]).toEqual(newRelation);
      expect(toast.success).toHaveBeenCalledWith('Relation created');
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Source and target must be different');
      vi.mocked(relationsApi.create).mockRejectedValue(error);

      const { result } = renderHook(() => useCreateRelation(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({
            sourceComponentId: 'comp-1' as ComponentId,
            targetComponentId: 'comp-1' as ComponentId,
            relationType: 'Triggers',
          });
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Source and target must be different');
    });

    it('should handle cache when no existing data', async () => {
      const newRelation = buildRelation({
        id: 'rel-1' as RelationId,
        name: 'First Relation',
      });

      vi.mocked(relationsApi.create).mockResolvedValue(newRelation);

      const { result } = renderHook(() => useCreateRelation(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          sourceComponentId: 'comp-1' as ComponentId,
          targetComponentId: 'comp-2' as ComponentId,
          relationType: 'Serves',
        });
      });

      const cachedRelations = queryClient.getQueryData<Relation[]>(
        queryKeys.relations.lists()
      );
      expect(cachedRelations).toHaveLength(1);
      expect(cachedRelations?.[0]).toEqual(newRelation);
    });
  });

  describe('useUpdateRelation', () => {
    it('should update relation and update both list and detail cache', async () => {
      const existingRelation = buildRelation({
        id: 'rel-1' as RelationId,
        name: 'Original Name',
        description: 'Original description',
      });
      const updatedRelation = buildRelation({
        id: 'rel-1' as RelationId,
        name: 'Updated Name',
        description: 'Updated description',
      });

      queryClient.setQueryData(queryKeys.relations.lists(), [existingRelation]);
      queryClient.setQueryData(queryKeys.relations.detail('rel-1'), existingRelation);
      vi.mocked(relationsApi.update).mockResolvedValue(updatedRelation);

      const { result } = renderHook(() => useUpdateRelation(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          id: 'rel-1' as RelationId,
          request: { name: 'Updated Name', description: 'Updated description' },
        });
      });

      const cachedRelations = queryClient.getQueryData<Relation[]>(
        queryKeys.relations.lists()
      );
      expect(cachedRelations?.[0].name).toBe('Updated Name');

      const cachedDetail = queryClient.getQueryData<Relation>(
        queryKeys.relations.detail('rel-1')
      );
      expect(cachedDetail?.name).toBe('Updated Name');

      expect(toast.success).toHaveBeenCalledWith('Relation updated');
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Update failed');
      vi.mocked(relationsApi.update).mockRejectedValue(error);

      const { result } = renderHook(() => useUpdateRelation(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({
            id: 'rel-1' as RelationId,
            request: { name: 'New Name' },
          });
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Update failed');
    });
  });

  describe('useDeleteRelation', () => {
    it('should delete relation and remove from cache', async () => {
      const relation = buildRelation({
        id: 'rel-1' as RelationId,
        name: 'To Delete',
      });

      queryClient.setQueryData(queryKeys.relations.lists(), [relation]);
      queryClient.setQueryData(queryKeys.relations.detail('rel-1'), relation);
      vi.mocked(relationsApi.delete).mockResolvedValue(undefined);

      const removeQueriesSpy = vi.spyOn(queryClient, 'removeQueries');

      const { result } = renderHook(() => useDeleteRelation(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync('rel-1' as RelationId);
      });

      const cachedRelations = queryClient.getQueryData<Relation[]>(
        queryKeys.relations.lists()
      );
      expect(cachedRelations).toHaveLength(0);

      expect(removeQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.relations.detail('rel-1'),
      });

      expect(toast.success).toHaveBeenCalledWith('Relation deleted');
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Cannot delete relation');
      vi.mocked(relationsApi.delete).mockRejectedValue(error);

      const { result } = renderHook(() => useDeleteRelation(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync('rel-1' as RelationId);
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Cannot delete relation');
    });
  });
});
