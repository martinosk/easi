import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { relationsApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
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
    onSuccess: (newRelation) => {
      queryClient.setQueryData<Relation[]>(
        queryKeys.relations.lists(),
        (old) => (old ? [...old, newRelation] : [newRelation])
      );
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
      id,
      request,
    }: {
      id: RelationId;
      request: Partial<CreateRelationRequest>;
    }) => relationsApi.update(id, request),
    onSuccess: (updatedRelation) => {
      queryClient.setQueryData<Relation[]>(
        queryKeys.relations.lists(),
        (old) =>
          old?.map((r) => (r.id === updatedRelation.id ? updatedRelation : r)) ?? []
      );
      queryClient.setQueryData(
        queryKeys.relations.detail(updatedRelation.id),
        updatedRelation
      );
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
    mutationFn: (id: RelationId) => relationsApi.delete(id),
    onSuccess: (_, deletedId) => {
      queryClient.setQueryData<Relation[]>(
        queryKeys.relations.lists(),
        (old) => old?.filter((r) => r.id !== deletedId) ?? []
      );
      queryClient.removeQueries({
        queryKey: queryKeys.relations.detail(deletedId),
      });
      toast.success('Relation deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete relation');
    },
  });
}
