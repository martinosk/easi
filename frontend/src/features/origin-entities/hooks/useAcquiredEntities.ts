import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { originEntitiesApi } from '../api/originEntitiesApi';
import { acquiredEntitiesQueryKeys } from '../queryKeys';
import { invalidateFor } from '../../../lib/invalidateFor';
import { acquiredEntitiesMutationEffects } from '../mutationEffects';
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
    queryKey: acquiredEntitiesQueryKeys.lists(),
    queryFn: () => originEntitiesApi.acquiredEntities.getAll(),
  });
}

export function useAcquiredEntity(id: AcquiredEntityId | undefined) {
  return useQuery({
    queryKey: acquiredEntitiesQueryKeys.detail(id!),
    queryFn: () => originEntitiesApi.acquiredEntities.getById(id!),
    enabled: !!id,
  });
}

function useEntityMutation<TArgs, TResult>(
  mutationFn: (args: TArgs) => Promise<TResult>,
  onMutationSuccess: (queryClient: ReturnType<typeof useQueryClient>, result: TResult, args: TArgs) => void,
  errorMessage: string
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (result, args) => onMutationSuccess(queryClient, result, args),
    onError: () => toast.error(errorMessage),
  });
}

export function useCreateAcquiredEntity() {
  return useEntityMutation(
    (request: CreateAcquiredEntityRequest) => originEntitiesApi.acquiredEntities.create(request),
    (qc, newEntity) => {
      invalidateFor(qc, acquiredEntitiesMutationEffects.create());
      toast.success(`Acquired entity "${newEntity.name}" created successfully`);
    },
    'Failed to create acquired entity'
  );
}

export function useUpdateAcquiredEntity() {
  return useEntityMutation(
    ({ id, request }: { id: AcquiredEntityId; request: UpdateAcquiredEntityRequest }) =>
      originEntitiesApi.acquiredEntities.update(id, request),
    (qc, updatedEntity, { id }) => {
      invalidateFor(qc, acquiredEntitiesMutationEffects.update(id));
      toast.success(`Acquired entity "${updatedEntity.name}" updated`);
    },
    'Failed to update acquired entity'
  );
}

export function useDeleteAcquiredEntity() {
  return useEntityMutation(
    ({ id }: { id: AcquiredEntityId; name: string }) =>
      originEntitiesApi.acquiredEntities.delete(id),
    (qc, _, { id, name }) => {
      invalidateFor(qc, acquiredEntitiesMutationEffects.delete(id));
      toast.success(`Acquired entity "${name}" deleted`);
    },
    'Failed to delete acquired entity'
  );
}

export function useLinkComponentToAcquiredEntity() {
  return useEntityMutation(
    ({ componentId, entityId, notes }: { componentId: ComponentId; entityId: AcquiredEntityId; notes?: string }) =>
      originEntitiesApi.acquiredEntities.linkComponent(componentId, entityId, notes),
    (qc, _, { entityId, componentId }) => {
      invalidateFor(qc, acquiredEntitiesMutationEffects.linkComponent(entityId, componentId));
      toast.success('Component linked to acquired entity');
    },
    'Failed to link component to acquired entity'
  );
}

export function useUnlinkComponentFromAcquiredEntity() {
  return useEntityMutation(
    ({ componentId }: { entityId: AcquiredEntityId; componentId: ComponentId }) =>
      originEntitiesApi.acquiredEntities.unlinkComponent(componentId),
    (qc, _, { entityId, componentId }) => {
      invalidateFor(qc, acquiredEntitiesMutationEffects.unlinkComponent(entityId, componentId));
      toast.success('Component unlinked');
    },
    'Failed to unlink component'
  );
}
