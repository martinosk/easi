import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { componentsApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import type { QueryKey } from '@tanstack/react-query';
import { invalidateFor } from '../../../lib/invalidateFor';
import { mutationEffects } from '../../../lib/mutationEffects';
import type {
  Component,
  ComponentId,
  CreateComponentRequest,
  AddComponentExpertRequest,
  Expert,
} from '../../../api/types';
import toast from 'react-hot-toast';

function useComponentMutation<TArgs, TResult>(
  mutationFn: (args: TArgs) => Promise<TResult>,
  getEffects: (result: TResult, args: TArgs) => QueryKey[],
  successMessage: string | ((result: TResult, args: TArgs) => string),
  errorMessage: string,
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (result, args) => {
      invalidateFor(queryClient, getEffects(result, args));
      toast.success(typeof successMessage === 'function' ? successMessage(result, args) : successMessage);
    },
    onError: (error: Error) => toast.error(error.message || errorMessage),
  });
}

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
  return useComponentMutation(
    (request: CreateComponentRequest) => componentsApi.create(request),
    () => mutationEffects.components.create(),
    (component) => `Component "${component.name}" created`,
    'Failed to create component',
  );
}

export function useUpdateComponent() {
  return useComponentMutation(
    ({ component, request }: { component: Component; request: CreateComponentRequest }) =>
      componentsApi.update(component, request),
    (updated) => mutationEffects.components.update(updated.id),
    (updated) => `Component "${updated.name}" updated`,
    'Failed to update component',
  );
}

export function useDeleteComponent() {
  return useComponentMutation(
    (component: Component) => componentsApi.delete(component),
    (_, component) => mutationEffects.components.delete(component.id),
    'Component deleted',
    'Failed to delete component',
  );
}

export function useComponentExpertRoles() {
  return useQuery({
    queryKey: queryKeys.components.expertRoles(),
    queryFn: () => componentsApi.getExpertRoles(),
  });
}

export function useAddComponentExpert() {
  return useComponentMutation(
    ({ id, request }: { id: ComponentId; request: AddComponentExpertRequest }) =>
      componentsApi.addExpert(id, request),
    (_, { id }) => mutationEffects.components.addExpert(id),
    'Expert added',
    'Failed to add expert',
  );
}

export function useRemoveComponentExpert() {
  return useComponentMutation(
    (params: { componentId: ComponentId; expert: Expert }) => componentsApi.removeExpert(params.expert),
    (_, params) => mutationEffects.components.removeExpert(params.componentId),
    'Expert removed',
    'Failed to remove expert',
  );
}
