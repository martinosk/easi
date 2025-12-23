import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { componentsApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import type { Component, ComponentId, CreateComponentRequest } from '../../../api/types';
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
      queryClient.setQueryData<Component[]>(queryKeys.components.lists(), (old) => {
        const updated = old ? [...old, newComponent] : [newComponent];
        return updated.sort((a, b) => a.name.localeCompare(b.name));
      });
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
      queryClient.invalidateQueries({ queryKey: queryKeys.components.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.components.detail(updatedComponent.id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.businessDomains.all });
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
      queryClient.setQueryData<Component[]>(
        queryKeys.components.lists(),
        (old) => old?.filter((c) => c.id !== deletedId) ?? []
      );
      queryClient.removeQueries({
        queryKey: queryKeys.components.detail(deletedId),
      });
      toast.success('Component deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete component');
    },
  });
}
