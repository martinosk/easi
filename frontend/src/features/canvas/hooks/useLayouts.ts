import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { layoutsApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import type {
  LayoutContextType,
  LayoutContainer,
  UpsertLayoutRequest,
  ElementPositionInput,
  BatchUpdateItem,
} from '../../../api/types';

export function useLayout(
  contextType: LayoutContextType | undefined,
  contextRef: string | undefined
) {
  return useQuery({
    queryKey: queryKeys.layouts.detail(contextType!, contextRef!),
    queryFn: () => layoutsApi.get(contextType!, contextRef!),
    enabled: !!contextType && !!contextRef,
  });
}

export function useUpsertLayout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      contextType,
      contextRef,
      request,
    }: {
      contextType: LayoutContextType;
      contextRef: string;
      request?: UpsertLayoutRequest;
    }) => layoutsApi.upsert(contextType, contextRef, request),
    onSuccess: (data, { contextType, contextRef }) => {
      queryClient.setQueryData(
        queryKeys.layouts.detail(contextType, contextRef),
        data
      );
    },
  });
}

export function useDeleteLayout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      contextType,
      contextRef,
    }: {
      contextType: LayoutContextType;
      contextRef: string;
    }) => layoutsApi.delete(contextType, contextRef),
    onSuccess: (_, { contextType, contextRef }) => {
      queryClient.removeQueries({
        queryKey: queryKeys.layouts.detail(contextType, contextRef),
      });
    },
  });
}

export function useUpdateLayoutPreferences() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      contextType,
      contextRef,
      preferences,
      version,
    }: {
      contextType: LayoutContextType;
      contextRef: string;
      preferences: Record<string, unknown>;
      version: number;
    }) => layoutsApi.updatePreferences(contextType, contextRef, preferences, version),
    onSuccess: (data, { contextType, contextRef }) => {
      queryClient.setQueryData<LayoutContainer | null>(
        queryKeys.layouts.detail(contextType, contextRef),
        (old) =>
          old
            ? {
                ...old,
                preferences: data.preferences,
                version: data.version,
              }
            : null
      );
    },
  });
}

export function useUpsertElementPosition() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      contextType,
      contextRef,
      elementId,
      position,
    }: {
      contextType: LayoutContextType;
      contextRef: string;
      elementId: string;
      position: ElementPositionInput;
    }) => layoutsApi.upsertElement(contextType, contextRef, elementId, position),
    onSuccess: (_, { contextType, contextRef }) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.layouts.detail(contextType, contextRef),
      });
    },
  });
}

export function useDeleteElementPosition() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      contextType,
      contextRef,
      elementId,
    }: {
      contextType: LayoutContextType;
      contextRef: string;
      elementId: string;
    }) => layoutsApi.deleteElement(contextType, contextRef, elementId),
    onSuccess: (_, { contextType, contextRef }) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.layouts.detail(contextType, contextRef),
      });
    },
  });
}

export function useBatchUpdateElements() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      contextType,
      contextRef,
      updates,
    }: {
      contextType: LayoutContextType;
      contextRef: string;
      updates: BatchUpdateItem[];
    }) => layoutsApi.batchUpdateElements(contextType, contextRef, updates),
    onSuccess: (_, { contextType, contextRef }) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.layouts.detail(contextType, contextRef),
      });
    },
  });
}
