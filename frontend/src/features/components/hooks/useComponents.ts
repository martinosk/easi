import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { componentsApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import { invalidateFor } from '../../../lib/invalidateFor';
import { mutationEffects } from '../../../lib/mutationEffects';
import type { ComponentId, CreateComponentRequest } from '../../../api/types';
import toast from 'react-hot-toast';

export function useComponents() {
  return useQuery({
    queryKey: queryKeys.components.lists(),
    queryFn: () => componentsApi.getAll(),
  });
}

export function useComponent(id: ComponentId | undefined) {
  return useQuery({
    queryKey: queryKeys.components.detail(id!),
    queryFn: () => componentsApi.getById(id!),
    enabled: !!id,
  });
}

export function useCreateComponent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateComponentRequest) => componentsApi.create(request),
    onSuccess: (newComponent) => {
      invalidateFor(queryClient, mutationEffects.components.create());
      toast.success(`Component "${newComponent.name}" created`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create component');
    },
  });
}

export function useUpdateComponent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, request }: { id: ComponentId; request: CreateComponentRequest }) =>
      componentsApi.update(id, request),
    onSuccess: (updatedComponent) => {
      invalidateFor(queryClient, mutationEffects.components.update(updatedComponent.id));
      toast.success(`Component "${updatedComponent.name}" updated`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update component');
    },
  });
}

export function useDeleteComponent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: ComponentId) => componentsApi.delete(id),
    onSuccess: (_, deletedId) => {
      invalidateFor(queryClient, mutationEffects.components.delete(deletedId));
      toast.success('Component deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete component');
    },
  });
}
