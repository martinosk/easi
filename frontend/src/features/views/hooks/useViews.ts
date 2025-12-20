import { useQuery, useMutation, useQueryClient, QueryClient } from '@tanstack/react-query';
import { viewsApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import type {
  View,
  ViewId,
  CreateViewRequest,
  AddComponentToViewRequest,
  AddCapabilityToViewRequest,
  UpdatePositionRequest,
  UpdateMultiplePositionsRequest,
  RenameViewRequest,
  UpdateViewEdgeTypeRequest,
  UpdateViewColorSchemeRequest,
  ComponentId,
  CapabilityId,
  Position,
} from '../../../api/types';
import toast from 'react-hot-toast';

type ViewMutationOptions<TVariables extends { viewId: ViewId }, TData = void> = {
  mutationFn: (variables: TVariables) => Promise<TData>;
  errorMessage: string;
  showErrorToast?: boolean;
};

function createViewMutation<TVariables extends { viewId: ViewId }, TData = void>(
  queryClient: QueryClient,
  options: ViewMutationOptions<TVariables, TData>
) {
  const { mutationFn, errorMessage, showErrorToast = true } = options;
  return useMutation({
    mutationFn,
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(variables.viewId) });
    },
    onError: showErrorToast
      ? (error: Error) => {
          toast.error(error.message || errorMessage);
        }
      : undefined,
  });
}

export function useViews() {
  return useQuery({
    queryKey: queryKeys.views.lists(),
    queryFn: () => viewsApi.getAll(),
  });
}

export function useView(id: ViewId | undefined) {
  return useQuery({
    queryKey: queryKeys.views.detail(id!),
    queryFn: () => viewsApi.getById(id!),
    enabled: !!id,
  });
}

export function useViewComponents(viewId: ViewId | undefined) {
  return useQuery({
    queryKey: queryKeys.views.components(viewId!),
    queryFn: () => viewsApi.getComponents(viewId!),
    enabled: !!viewId,
  });
}

export function useCreateView() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateViewRequest) => viewsApi.create(request),
    onSuccess: (newView) => {
      queryClient.setQueryData<View[]>(
        queryKeys.views.lists(),
        (old) => (old ? [...old, newView] : [newView])
      );
      toast.success(`View "${newView.name}" created`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create view');
    },
  });
}

export function useDeleteView() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: ViewId) => viewsApi.delete(id),
    onSuccess: (_, deletedId) => {
      queryClient.setQueryData<View[]>(
        queryKeys.views.lists(),
        (old) => old?.filter((v) => v.id !== deletedId) ?? []
      );
      queryClient.removeQueries({
        queryKey: queryKeys.views.detail(deletedId),
      });
      toast.success('View deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete view');
    },
  });
}

export function useRenameView() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ viewId, request }: { viewId: ViewId; request: RenameViewRequest }) =>
      viewsApi.rename(viewId, request),
    onSuccess: (_, { viewId, request }) => {
      queryClient.setQueryData<View[]>(
        queryKeys.views.lists(),
        (old) => old?.map((v) => (v.id === viewId ? { ...v, name: request.name } : v)) ?? []
      );
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
      toast.success('View renamed');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to rename view');
    },
  });
}

export function useSetDefaultView() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (viewId: ViewId) => viewsApi.setDefault(viewId),
    onSuccess: (_, viewId) => {
      queryClient.setQueryData<View[]>(
        queryKeys.views.lists(),
        (old) =>
          old?.map((v) => ({
            ...v,
            isDefault: v.id === viewId,
          })) ?? []
      );
      toast.success('Default view updated');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to set default view');
    },
  });
}

export function useUpdateViewEdgeType() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; request: UpdateViewEdgeTypeRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.updateEdgeType(viewId, request),
    errorMessage: 'Failed to update edge type',
  });
}

export function useUpdateViewColorScheme() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; request: UpdateViewColorSchemeRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.updateColorScheme(viewId, request),
    errorMessage: 'Failed to update color scheme',
  });
}

export function useAddComponentToView() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; request: AddComponentToViewRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.addComponent(viewId, request),
    errorMessage: 'Failed to add component to view',
  });
}

export function useRemoveComponentFromView() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; componentId: ComponentId }>(queryClient, {
    mutationFn: ({ viewId, componentId }) => viewsApi.removeComponent(viewId, componentId),
    errorMessage: 'Failed to remove component from view',
  });
}

export function useUpdateComponentPosition() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; componentId: ComponentId; request: UpdatePositionRequest }>(queryClient, {
    mutationFn: ({ viewId, componentId, request }) => viewsApi.updateComponentPosition(viewId, componentId, request),
    errorMessage: 'Failed to update position',
    showErrorToast: false,
  });
}

export function useUpdateMultiplePositions() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; request: UpdateMultiplePositionsRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.updateMultiplePositions(viewId, request),
    errorMessage: 'Failed to update positions',
    showErrorToast: false,
  });
}

export function useAddCapabilityToView() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; request: AddCapabilityToViewRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.addCapability(viewId, request),
    errorMessage: 'Failed to add capability to view',
  });
}

export function useRemoveCapabilityFromView() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; capabilityId: CapabilityId }>(queryClient, {
    mutationFn: ({ viewId, capabilityId }) => viewsApi.removeCapability(viewId, capabilityId),
    errorMessage: 'Failed to remove capability from view',
  });
}

export function useUpdateCapabilityPosition() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; capabilityId: CapabilityId; position: Position }>(queryClient, {
    mutationFn: ({ viewId, capabilityId, position }) => viewsApi.updateCapabilityPosition(viewId, capabilityId, position),
    errorMessage: 'Failed to update position',
    showErrorToast: false,
  });
}

export function useUpdateComponentColor() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; componentId: ComponentId; color: string }>(queryClient, {
    mutationFn: ({ viewId, componentId, color }) => viewsApi.updateComponentColor(viewId, componentId, color),
    errorMessage: 'Failed to update component color',
  });
}

export function useClearComponentColor() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; componentId: ComponentId }>(queryClient, {
    mutationFn: ({ viewId, componentId }) => viewsApi.clearComponentColor(viewId, componentId),
    errorMessage: 'Failed to clear component color',
  });
}

export function useUpdateCapabilityColor() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; capabilityId: CapabilityId; color: string }>(queryClient, {
    mutationFn: ({ viewId, capabilityId, color }) => viewsApi.updateCapabilityColor(viewId, capabilityId, color),
    errorMessage: 'Failed to update capability color',
  });
}

export function useClearCapabilityColor() {
  const queryClient = useQueryClient();
  return createViewMutation<{ viewId: ViewId; capabilityId: CapabilityId }>(queryClient, {
    mutationFn: ({ viewId, capabilityId }) => viewsApi.clearCapabilityColor(viewId, capabilityId),
    errorMessage: 'Failed to clear capability color',
  });
}
