import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { originEntitiesApi } from '../api/originEntitiesApi';
import { queryKeys } from '../../../lib/queryClient';
import { invalidateFor } from '../../../lib/invalidateFor';
import { mutationEffects } from '../../../lib/mutationEffects';
import type {
  AcquiredEntity,
  AcquiredEntityId,
  CreateAcquiredEntityRequest,
  UpdateAcquiredEntityRequest,
  ComponentId,
} from '../../../api/types';
import toast from 'react-hot-toast';

export interface UseAcquiredEntitiesResult {
  acquiredEntities: AcquiredEntity[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  createEntity: (request: CreateAcquiredEntityRequest) => Promise<AcquiredEntity>;
  updateEntity: (id: AcquiredEntityId, request: UpdateAcquiredEntityRequest) => Promise<AcquiredEntity>;
  deleteEntity: (id: AcquiredEntityId, name: string) => Promise<void>;
}

export function useAcquiredEntities(): UseAcquiredEntitiesResult {
  const query = useAcquiredEntitiesQuery();
  const createMutation = useCreateAcquiredEntity();
  const updateMutation = useUpdateAcquiredEntity();
  const deleteMutation = useDeleteAcquiredEntity();

  const createEntity = useCallback(
    async (request: CreateAcquiredEntityRequest): Promise<AcquiredEntity> => {
      return createMutation.mutateAsync(request);
    },
    [createMutation]
  );

  const updateEntity = useCallback(
    async (id: AcquiredEntityId, request: UpdateAcquiredEntityRequest): Promise<AcquiredEntity> => {
      return updateMutation.mutateAsync({ id, request });
    },
    [updateMutation]
  );

  const deleteEntity = useCallback(
    async (id: AcquiredEntityId, name: string): Promise<void> => {
      await deleteMutation.mutateAsync({ id, name });
    },
    [deleteMutation]
  );

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    acquiredEntities: query.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    refetch,
    createEntity,
    updateEntity,
    deleteEntity,
  };
}

export function useAcquiredEntitiesQuery() {
  return useQuery({
    queryKey: queryKeys.acquiredEntities.lists(),
    queryFn: () => originEntitiesApi.acquiredEntities.getAll(),
  });
}

export function useAcquiredEntity(id: AcquiredEntityId | undefined) {
  return useQuery({
    queryKey: queryKeys.acquiredEntities.detail(id!),
    queryFn: () => originEntitiesApi.acquiredEntities.getById(id!),
    enabled: !!id,
  });
}

export function useCreateAcquiredEntity() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateAcquiredEntityRequest) =>
      originEntitiesApi.acquiredEntities.create(request),
    onSuccess: (newEntity) => {
      invalidateFor(queryClient, mutationEffects.acquiredEntities.create());
      toast.success(`Acquired entity "${newEntity.name}" created successfully`);
    },
    onError: () => {
      toast.error('Failed to create acquired entity');
    },
  });
}

export function useUpdateAcquiredEntity() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, request }: { id: AcquiredEntityId; request: UpdateAcquiredEntityRequest }) =>
      originEntitiesApi.acquiredEntities.update(id, request),
    onSuccess: (updatedEntity, { id }) => {
      invalidateFor(queryClient, mutationEffects.acquiredEntities.update(id));
      toast.success(`Acquired entity "${updatedEntity.name}" updated`);
    },
    onError: () => {
      toast.error('Failed to update acquired entity');
    },
  });
}

export function useDeleteAcquiredEntity() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id }: { id: AcquiredEntityId; name: string }) =>
      originEntitiesApi.acquiredEntities.delete(id),
    onSuccess: (_, { id, name }) => {
      invalidateFor(queryClient, mutationEffects.acquiredEntities.delete(id));
      toast.success(`Acquired entity "${name}" deleted`);
    },
    onError: () => {
      toast.error('Failed to delete acquired entity');
    },
  });
}

export function useLinkComponentToAcquiredEntity() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      componentId,
      entityId,
      notes,
    }: {
      componentId: ComponentId;
      entityId: AcquiredEntityId;
      notes?: string;
    }) => originEntitiesApi.acquiredEntities.linkComponent(componentId, entityId, notes),
    onSuccess: (_, { entityId, componentId }) => {
      invalidateFor(queryClient, mutationEffects.acquiredEntities.linkComponent(entityId, componentId));
      toast.success('Component linked to acquired entity');
    },
    onError: () => {
      toast.error('Failed to link component to acquired entity');
    },
  });
}

export function useUnlinkComponentFromAcquiredEntity() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ componentId }: { entityId: AcquiredEntityId; componentId: ComponentId }) =>
      originEntitiesApi.acquiredEntities.unlinkComponent(componentId),
    onSuccess: (_, { entityId, componentId }) => {
      invalidateFor(queryClient, mutationEffects.acquiredEntities.unlinkComponent(entityId, componentId));
      toast.success('Component unlinked');
    },
    onError: () => {
      toast.error('Failed to unlink component');
    },
  });
}
