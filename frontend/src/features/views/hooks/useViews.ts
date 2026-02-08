import { useQuery, useMutation, useQueryClient, QueryClient } from '@tanstack/react-query';
import { viewsApi } from '../api';
import { viewsQueryKeys } from '../queryKeys';
import { viewsMutationEffects } from '../mutationEffects';
import { invalidateFor } from '../../../lib/invalidateFor';
import {
  ApiError,
  type View,
  type ViewId,
  type CreateViewRequest,
  type AddComponentToViewRequest,
  type AddCapabilityToViewRequest,
  type AddOriginEntityToViewRequest,
  type UpdatePositionRequest,
  type UpdateMultiplePositionsRequest,
  type RenameViewRequest,
  type UpdateViewEdgeTypeRequest,
  type UpdateViewColorSchemeRequest,
  type ComponentId,
  type CapabilityId,
  type Position,
} from '../../../api/types';
import toast from 'react-hot-toast';

const MAX_CONCURRENCY_RETRIES = 3;

function isConcurrencyConflict(error: unknown): boolean {
  return error instanceof ApiError && error.statusCode === 412;
}

type ViewActionOptions<TVariables, TData = void> = {
  mutationFn: (variables: TVariables) => Promise<TData>;
  onSuccess: (data: TData, variables: TVariables, qc: QueryClient) => void;
  errorMessage: string;
};

function useViewAction<TVariables, TData = void>(options: ViewActionOptions<TVariables, TData>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: options.mutationFn,
    onSuccess: (data, variables) => options.onSuccess(data, variables, queryClient),
    onError: (error: Error) => {
      toast.error(error.message || options.errorMessage);
    },
  });
}

type ViewMutationOptions<TVariables extends { viewId: ViewId }, TData = void> = {
  mutationFn: (variables: TVariables) => Promise<TData>;
  errorMessage: string;
  showErrorToast?: boolean;
};

function useViewMutation<TVariables extends { viewId: ViewId }, TData = void>(
  options: ViewMutationOptions<TVariables, TData>
) {
  const queryClient = useQueryClient();
  const { mutationFn, errorMessage, showErrorToast = true } = options;
  return useMutation({
    mutationFn,
    retry: (failureCount, error) =>
      isConcurrencyConflict(error) && failureCount < MAX_CONCURRENCY_RETRIES,
    retryDelay: (attempt) =>
      Math.min(100 * 2 ** attempt, 1000) + Math.random() * 100,
    onSuccess: (_, variables) => {
      invalidateFor(queryClient, viewsMutationEffects.updateDetail(variables.viewId));
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
    queryKey: viewsQueryKeys.lists(),
    queryFn: () => viewsApi.getAll(),
  });
}

export function useView(id: ViewId | undefined) {
  return useQuery({
    queryKey: viewsQueryKeys.detail(id!),
    queryFn: () => viewsApi.getById(id!),
    enabled: !!id,
  });
}

export function useViewComponents(viewId: ViewId | undefined) {
  return useQuery({
    queryKey: viewsQueryKeys.components(viewId!),
    queryFn: () => viewsApi.getComponents(viewId!),
    enabled: !!viewId,
  });
}

export function useCreateView() {
  return useViewAction<CreateViewRequest, View>({
    mutationFn: (request) => viewsApi.create(request),
    onSuccess: (newView, _, qc) => {
      invalidateFor(qc, viewsMutationEffects.create());
      toast.success(`View "${newView.name}" created`);
    },
    errorMessage: 'Failed to create view',
  });
}

export function useDeleteView() {
  return useViewAction<View>({
    mutationFn: (view) => viewsApi.delete(view),
    onSuccess: (_, deletedView, qc) => {
      invalidateFor(qc, viewsMutationEffects.delete(deletedView.id));
      toast.success('View deleted');
    },
    errorMessage: 'Failed to delete view',
  });
}

export function useRenameView() {
  return useViewAction<{ viewId: ViewId; request: RenameViewRequest }>({
    mutationFn: ({ viewId, request }) => viewsApi.rename(viewId, request),
    onSuccess: (_, { viewId }, qc) => {
      invalidateFor(qc, viewsMutationEffects.rename(viewId));
      toast.success('View renamed');
    },
    errorMessage: 'Failed to rename view',
  });
}

export function useSetDefaultView() {
  return useViewAction<ViewId>({
    mutationFn: (viewId) => viewsApi.setDefault(viewId),
    onSuccess: (_data, _vars, qc) => {
      invalidateFor(qc, viewsMutationEffects.setDefault());
      toast.success('Default view updated');
    },
    errorMessage: 'Failed to set default view',
  });
}

export function useUpdateViewEdgeType() {
  return useViewMutation<{ viewId: ViewId; request: UpdateViewEdgeTypeRequest }>({
    mutationFn: ({ viewId, request }) => viewsApi.updateEdgeType(viewId, request),
    errorMessage: 'Failed to update edge type',
  });
}

export function useUpdateViewColorScheme() {
  return useViewMutation<{ viewId: ViewId; request: UpdateViewColorSchemeRequest }>({
    mutationFn: ({ viewId, request }) => viewsApi.updateColorScheme(viewId, request),
    errorMessage: 'Failed to update color scheme',
  });
}

export function useAddComponentToView() {
  return useViewMutation<{ viewId: ViewId; request: AddComponentToViewRequest }>({
    mutationFn: ({ viewId, request }) => viewsApi.addComponent(viewId, request),
    errorMessage: 'Failed to add component to view',
  });
}

export function useRemoveComponentFromView() {
  return useViewMutation<{ viewId: ViewId; componentId: ComponentId }>({
    mutationFn: ({ viewId, componentId }) => viewsApi.removeComponent(viewId, componentId),
    errorMessage: 'Failed to remove component from view',
  });
}

export function useUpdateComponentPosition() {
  return useViewMutation<{ viewId: ViewId; componentId: ComponentId; request: UpdatePositionRequest }>({
    mutationFn: ({ viewId, componentId, request }) => viewsApi.updateComponentPosition(viewId, componentId, request),
    errorMessage: 'Failed to update position',
    showErrorToast: false,
  });
}

export function useUpdateMultiplePositions() {
  return useViewMutation<{ viewId: ViewId; request: UpdateMultiplePositionsRequest }>({
    mutationFn: ({ viewId, request }) => viewsApi.updateMultiplePositions(viewId, request),
    errorMessage: 'Failed to update positions',
    showErrorToast: false,
  });
}

export function useAddCapabilityToView() {
  return useViewMutation<{ viewId: ViewId; request: AddCapabilityToViewRequest }>({
    mutationFn: ({ viewId, request }) => viewsApi.addCapability(viewId, request),
    errorMessage: 'Failed to add capability to view',
  });
}

export function useRemoveCapabilityFromView() {
  return useViewMutation<{ viewId: ViewId; capabilityId: CapabilityId }>({
    mutationFn: ({ viewId, capabilityId }) => viewsApi.removeCapability(viewId, capabilityId),
    errorMessage: 'Failed to remove capability from view',
  });
}

export function useUpdateCapabilityPosition() {
  return useViewMutation<{ viewId: ViewId; capabilityId: CapabilityId; position: Position }>({
    mutationFn: ({ viewId, capabilityId, position }) => viewsApi.updateCapabilityPosition(viewId, capabilityId, position),
    errorMessage: 'Failed to update position',
    showErrorToast: false,
  });
}

export function useUpdateComponentColor() {
  return useViewMutation<{ viewId: ViewId; componentId: ComponentId; color: string }>({
    mutationFn: ({ viewId, componentId, color }) => viewsApi.updateComponentColor(viewId, componentId, color),
    errorMessage: 'Failed to update component color',
  });
}

export function useClearComponentColor() {
  return useViewMutation<{ viewId: ViewId; componentId: ComponentId }>({
    mutationFn: ({ viewId, componentId }) => viewsApi.clearComponentColor(viewId, componentId),
    errorMessage: 'Failed to clear component color',
  });
}

export function useUpdateCapabilityColor() {
  return useViewMutation<{ viewId: ViewId; capabilityId: CapabilityId; color: string }>({
    mutationFn: ({ viewId, capabilityId, color }) => viewsApi.updateCapabilityColor(viewId, capabilityId, color),
    errorMessage: 'Failed to update capability color',
  });
}

export function useClearCapabilityColor() {
  return useViewMutation<{ viewId: ViewId; capabilityId: CapabilityId }>({
    mutationFn: ({ viewId, capabilityId }) => viewsApi.clearCapabilityColor(viewId, capabilityId),
    errorMessage: 'Failed to clear capability color',
  });
}

export function useAddOriginEntityToView() {
  return useViewMutation<{ viewId: ViewId; request: AddOriginEntityToViewRequest }>({
    mutationFn: ({ viewId, request }) => viewsApi.addOriginEntity(viewId, request),
    errorMessage: 'Failed to add origin entity to view',
  });
}

export function useRemoveOriginEntityFromView() {
  return useViewMutation<{ viewId: ViewId; originEntityId: string }>({
    mutationFn: ({ viewId, originEntityId }) => viewsApi.removeOriginEntity(viewId, originEntityId),
    errorMessage: 'Failed to remove origin entity from view',
  });
}

export function useUpdateOriginEntityPosition() {
  return useViewMutation<{ viewId: ViewId; originEntityId: string; position: Position }>({
    mutationFn: ({ viewId, originEntityId, position }) => viewsApi.updateOriginEntityPosition(viewId, originEntityId, position),
    errorMessage: 'Failed to update position',
    showErrorToast: false,
  });
}

export function useChangeViewVisibility() {
  return useViewAction<{ viewId: ViewId; isPrivate: boolean }>({
    mutationFn: ({ viewId, isPrivate }) => viewsApi.changeVisibility(viewId, isPrivate),
    onSuccess: (_, { isPrivate, viewId }, qc) => {
      invalidateFor(qc, viewsMutationEffects.changeVisibility(viewId));
      toast.success(isPrivate ? 'View made private' : 'View made public');
    },
    errorMessage: 'Failed to change view visibility',
  });
}
