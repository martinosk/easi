import { useQuery, useMutation, useQueryClient, QueryClient } from '@tanstack/react-query';
import { viewsApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import { invalidateFor } from '../../../lib/invalidateFor';
import { mutationEffects } from '../../../lib/mutationEffects';
import type {
  View,
  ViewId,
  CreateViewRequest,
  AddComponentToViewRequest,
  AddCapabilityToViewRequest,
  AddOriginEntityToViewRequest,
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

function useViewMutation<TVariables extends { viewId: ViewId }, TData = void>(
  queryClient: QueryClient,
  options: ViewMutationOptions<TVariables, TData>
) {
  const { mutationFn, errorMessage, showErrorToast = true } = options;
  return useMutation({
    mutationFn,
    onSuccess: (_, variables) => {
      invalidateFor(queryClient, mutationEffects.views.updateDetail(variables.viewId));
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
      invalidateFor(queryClient, mutationEffects.views.create());
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
    mutationFn: (view: View) => viewsApi.delete(view),
    onSuccess: (_, deletedView) => {
      invalidateFor(queryClient, mutationEffects.views.delete(deletedView.id));
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
    onSuccess: (_, { viewId }) => {
      invalidateFor(queryClient, mutationEffects.views.rename(viewId));
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
    onSuccess: () => {
      invalidateFor(queryClient, mutationEffects.views.setDefault());
      toast.success('Default view updated');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to set default view');
    },
  });
}

export function useUpdateViewEdgeType() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; request: UpdateViewEdgeTypeRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.updateEdgeType(viewId, request),
    errorMessage: 'Failed to update edge type',
  });
}

export function useUpdateViewColorScheme() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; request: UpdateViewColorSchemeRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.updateColorScheme(viewId, request),
    errorMessage: 'Failed to update color scheme',
  });
}

export function useAddComponentToView() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; request: AddComponentToViewRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.addComponent(viewId, request),
    errorMessage: 'Failed to add component to view',
  });
}

export function useRemoveComponentFromView() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; componentId: ComponentId }>(queryClient, {
    mutationFn: ({ viewId, componentId }) => viewsApi.removeComponent(viewId, componentId),
    errorMessage: 'Failed to remove component from view',
  });
}

export function useUpdateComponentPosition() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; componentId: ComponentId; request: UpdatePositionRequest }>(queryClient, {
    mutationFn: ({ viewId, componentId, request }) => viewsApi.updateComponentPosition(viewId, componentId, request),
    errorMessage: 'Failed to update position',
    showErrorToast: false,
  });
}

export function useUpdateMultiplePositions() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; request: UpdateMultiplePositionsRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.updateMultiplePositions(viewId, request),
    errorMessage: 'Failed to update positions',
    showErrorToast: false,
  });
}

export function useAddCapabilityToView() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; request: AddCapabilityToViewRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.addCapability(viewId, request),
    errorMessage: 'Failed to add capability to view',
  });
}

export function useRemoveCapabilityFromView() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; capabilityId: CapabilityId }>(queryClient, {
    mutationFn: ({ viewId, capabilityId }) => viewsApi.removeCapability(viewId, capabilityId),
    errorMessage: 'Failed to remove capability from view',
  });
}

export function useUpdateCapabilityPosition() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; capabilityId: CapabilityId; position: Position }>(queryClient, {
    mutationFn: ({ viewId, capabilityId, position }) => viewsApi.updateCapabilityPosition(viewId, capabilityId, position),
    errorMessage: 'Failed to update position',
    showErrorToast: false,
  });
}

export function useUpdateComponentColor() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; componentId: ComponentId; color: string }>(queryClient, {
    mutationFn: ({ viewId, componentId, color }) => viewsApi.updateComponentColor(viewId, componentId, color),
    errorMessage: 'Failed to update component color',
  });
}

export function useClearComponentColor() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; componentId: ComponentId }>(queryClient, {
    mutationFn: ({ viewId, componentId }) => viewsApi.clearComponentColor(viewId, componentId),
    errorMessage: 'Failed to clear component color',
  });
}

export function useUpdateCapabilityColor() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; capabilityId: CapabilityId; color: string }>(queryClient, {
    mutationFn: ({ viewId, capabilityId, color }) => viewsApi.updateCapabilityColor(viewId, capabilityId, color),
    errorMessage: 'Failed to update capability color',
  });
}

export function useClearCapabilityColor() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; capabilityId: CapabilityId }>(queryClient, {
    mutationFn: ({ viewId, capabilityId }) => viewsApi.clearCapabilityColor(viewId, capabilityId),
    errorMessage: 'Failed to clear capability color',
  });
}

export function useAddOriginEntityToView() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; request: AddOriginEntityToViewRequest }>(queryClient, {
    mutationFn: ({ viewId, request }) => viewsApi.addOriginEntity(viewId, request),
    errorMessage: 'Failed to add origin entity to view',
  });
}

export function useRemoveOriginEntityFromView() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; originEntityId: string }>(queryClient, {
    mutationFn: ({ viewId, originEntityId }) => viewsApi.removeOriginEntity(viewId, originEntityId),
    errorMessage: 'Failed to remove origin entity from view',
  });
}

export function useUpdateOriginEntityPosition() {
  const queryClient = useQueryClient();
  return useViewMutation<{ viewId: ViewId; originEntityId: string; position: Position }>(queryClient, {
    mutationFn: ({ viewId, originEntityId, position }) => viewsApi.updateOriginEntityPosition(viewId, originEntityId, position),
    errorMessage: 'Failed to update position',
    showErrorToast: false,
  });
}

export function useChangeViewVisibility() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ viewId, isPrivate }: { viewId: ViewId; isPrivate: boolean }) =>
      viewsApi.changeVisibility(viewId, isPrivate),
    onSuccess: (_, { viewId, isPrivate }) => {
      invalidateFor(queryClient, mutationEffects.views.changeVisibility(viewId));
      toast.success(isPrivate ? 'View made private' : 'View made public');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to change view visibility');
    },
  });
}
