import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
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

  return useMutation({
    mutationFn: ({
      viewId,
      request,
    }: {
      viewId: ViewId;
      request: UpdateViewEdgeTypeRequest;
    }) => viewsApi.updateEdgeType(viewId, request),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update edge type');
    },
  });
}

export function useUpdateViewColorScheme() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      viewId,
      request,
    }: {
      viewId: ViewId;
      request: UpdateViewColorSchemeRequest;
    }) => viewsApi.updateColorScheme(viewId, request),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update color scheme');
    },
  });
}

export function useAddComponentToView() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      viewId,
      request,
    }: {
      viewId: ViewId;
      request: AddComponentToViewRequest;
    }) => viewsApi.addComponent(viewId, request),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to add component to view');
    },
  });
}

export function useRemoveComponentFromView() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ viewId, componentId }: { viewId: ViewId; componentId: ComponentId }) =>
      viewsApi.removeComponent(viewId, componentId),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to remove component from view');
    },
  });
}

export function useUpdateComponentPosition() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      viewId,
      componentId,
      request,
    }: {
      viewId: ViewId;
      componentId: ComponentId;
      request: UpdatePositionRequest;
    }) => viewsApi.updateComponentPosition(viewId, componentId, request),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
  });
}

export function useUpdateMultiplePositions() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      viewId,
      request,
    }: {
      viewId: ViewId;
      request: UpdateMultiplePositionsRequest;
    }) => viewsApi.updateMultiplePositions(viewId, request),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
  });
}

export function useAddCapabilityToView() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      viewId,
      request,
    }: {
      viewId: ViewId;
      request: AddCapabilityToViewRequest;
    }) => viewsApi.addCapability(viewId, request),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to add capability to view');
    },
  });
}

export function useRemoveCapabilityFromView() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      viewId,
      capabilityId,
    }: {
      viewId: ViewId;
      capabilityId: CapabilityId;
    }) => viewsApi.removeCapability(viewId, capabilityId),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to remove capability from view');
    },
  });
}

export function useUpdateCapabilityPosition() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      viewId,
      capabilityId,
      position,
    }: {
      viewId: ViewId;
      capabilityId: CapabilityId;
      position: Position;
    }) => viewsApi.updateCapabilityPosition(viewId, capabilityId, position),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
  });
}

export function useUpdateComponentColor() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      viewId,
      componentId,
      color,
    }: {
      viewId: ViewId;
      componentId: ComponentId;
      color: string;
    }) => viewsApi.updateComponentColor(viewId, componentId, color),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update component color');
    },
  });
}

export function useClearComponentColor() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ viewId, componentId }: { viewId: ViewId; componentId: ComponentId }) =>
      viewsApi.clearComponentColor(viewId, componentId),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to clear component color');
    },
  });
}

export function useUpdateCapabilityColor() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      viewId,
      capabilityId,
      color,
    }: {
      viewId: ViewId;
      capabilityId: CapabilityId;
      color: string;
    }) => viewsApi.updateCapabilityColor(viewId, capabilityId, color),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update capability color');
    },
  });
}

export function useClearCapabilityColor() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      viewId,
      capabilityId,
    }: {
      viewId: ViewId;
      capabilityId: CapabilityId;
    }) => viewsApi.clearCapabilityColor(viewId, capabilityId),
    onSuccess: (_, { viewId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.views.detail(viewId) });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to clear capability color');
    },
  });
}
