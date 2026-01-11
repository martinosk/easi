import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { relationsApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import { invalidateFor } from '../../../lib/invalidateFor';
import { mutationEffects } from '../../../lib/mutationEffects';
import type { Relation, RelationId, CreateRelationRequest } from '../../../api/types';
import toast from 'react-hot-toast';

export function useRelations() {
  return useQuery({
    queryKey: queryKeys.relations.lists(),
    queryFn: () => relationsApi.getAll(),
  });
}

export function useRelation(id: RelationId | undefined) {
  return useQuery({
    queryKey: queryKeys.relations.detail(id!),
    queryFn: () => relationsApi.getById(id!),
    enabled: !!id,
  });
}

export function useCreateRelation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateRelationRequest) => relationsApi.create(request),
    onSuccess: () => {
      invalidateFor(queryClient, mutationEffects.relations.create());
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
      invalidateFor(queryClient, mutationEffects.relations.update(updatedRelation.id));
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
      invalidateFor(queryClient, mutationEffects.relations.delete(deletedRelation.id));
      toast.success('Relation deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete relation');
    },
  });
}
