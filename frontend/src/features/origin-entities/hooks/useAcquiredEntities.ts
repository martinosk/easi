import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import type {
  AcquiredEntityId,
  ComponentId,
  CreateAcquiredEntityRequest,
  UpdateAcquiredEntityRequest,
} from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { originEntitiesApi } from '../api/originEntitiesApi';
import { acquiredEntitiesMutationEffects } from '../mutationEffects';
import { acquiredEntitiesQueryKeys } from '../queryKeys';

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

interface MutationConfig<TArgs, TResult> {
  mutationFn: (args: TArgs) => Promise<TResult>;
  effects: (result: TResult, args: TArgs) => ReadonlyArray<readonly string[]>;
  successMessage: (result: TResult, args: TArgs) => string;
  errorMessage: string;
}

function useEntityMutation<TArgs, TResult>(config: MutationConfig<TArgs, TResult>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: config.mutationFn,
    onSuccess: (result, args) => {
      invalidateFor(queryClient, config.effects(result, args));
      toast.success(config.successMessage(result, args));
    },
    onError: () => toast.error(config.errorMessage),
  });
}

export function useCreateAcquiredEntity() {
  return useEntityMutation({
    mutationFn: (request: CreateAcquiredEntityRequest) => originEntitiesApi.acquiredEntities.create(request),
    effects: () => acquiredEntitiesMutationEffects.create(),
    successMessage: (entity) => `Acquired entity "${entity.name}" created successfully`,
    errorMessage: 'Failed to create acquired entity',
  });
}

export function useUpdateAcquiredEntity() {
  return useEntityMutation({
    mutationFn: ({ id, request }: { id: AcquiredEntityId; request: UpdateAcquiredEntityRequest }) =>
      originEntitiesApi.acquiredEntities.update(id, request),
    effects: (_, { id }) => acquiredEntitiesMutationEffects.update(id),
    successMessage: (entity) => `Acquired entity "${entity.name}" updated`,
    errorMessage: 'Failed to update acquired entity',
  });
}

export function useDeleteAcquiredEntity() {
  return useEntityMutation({
    mutationFn: ({ id }: { id: AcquiredEntityId; name: string }) => originEntitiesApi.acquiredEntities.delete(id),
    effects: (_, { id }) => acquiredEntitiesMutationEffects.delete(id),
    successMessage: (_, { name }) => `Acquired entity "${name}" deleted`,
    errorMessage: 'Failed to delete acquired entity',
  });
}

export function useLinkComponentToAcquiredEntity() {
  return useEntityMutation({
    mutationFn: ({
      componentId,
      entityId,
      notes,
    }: {
      componentId: ComponentId;
      entityId: AcquiredEntityId;
      notes?: string;
    }) => originEntitiesApi.acquiredEntities.linkComponent(componentId, entityId, notes),
    effects: (_, { entityId, componentId }) => acquiredEntitiesMutationEffects.linkComponent(entityId, componentId),
    successMessage: () => 'Component linked to acquired entity',
    errorMessage: 'Failed to link component to acquired entity',
  });
}

export function useUnlinkComponentFromAcquiredEntity() {
  return useEntityMutation({
    mutationFn: ({ componentId }: { entityId: AcquiredEntityId; componentId: ComponentId }) =>
      originEntitiesApi.acquiredEntities.unlinkComponent(componentId),
    effects: (_, { entityId, componentId }) => acquiredEntitiesMutationEffects.unlinkComponent(entityId, componentId),
    successMessage: () => 'Component unlinked',
    errorMessage: 'Failed to unlink component',
  });
}
