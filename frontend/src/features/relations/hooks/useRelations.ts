import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { relationsApi } from '../api';
import { relationsQueryKeys } from '../queryKeys';
import { relationsMutationEffects } from '../mutationEffects';
import { invalidateFor } from '../../../lib/invalidateFor';
import type { Relation, RelationId, CreateRelationRequest } from '../../../api/types';
import toast from 'react-hot-toast';

export function useRelations() {
  return useQuery({
    queryKey: relationsQueryKeys.lists(),
    queryFn: () => relationsApi.getAll(),
  });
}

export function useRelation(id: RelationId | undefined) {
  return useQuery({
    queryKey: relationsQueryKeys.detail(id!),
    queryFn: () => relationsApi.getById(id!),
    enabled: !!id,
  });
}

export function useCreateRelation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateRelationRequest) => relationsApi.create(request),
    onSuccess: () => {
      invalidateFor(queryClient, relationsMutationEffects.create());
      toast.success('Relation created');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create relation');
    },
  });
}

export function useUpdateRelation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      relation,
      request,
    }: {
      relation: Relation;
      request: Partial<CreateRelationRequest>;
    }) => relationsApi.update(relation, request),
    onSuccess: (updatedRelation) => {
      invalidateFor(queryClient, relationsMutationEffects.update(updatedRelation.id));
      toast.success('Relation updated');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update relation');
    },
  });
}

export function useDeleteRelation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (relation: Relation) => relationsApi.delete(relation),
    onSuccess: (_, deletedRelation) => {
      invalidateFor(queryClient, relationsMutationEffects.delete(deletedRelation.id));
      toast.success('Relation deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete relation');
    },
  });
}
