import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { layoutsApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import { invalidateFor } from '../../../lib/invalidateFor';
import { mutationEffects } from '../../../lib/mutationEffects';
import type {
  LayoutContextType,
  UpsertLayoutRequest,
  ElementPositionInput,
  BatchUpdateItem,
} from '../../../api/types';

interface LayoutContext {
  contextType: LayoutContextType;
  contextRef: string;
}

function useLayoutMutationWithInvalidation<TVariables extends LayoutContext>(
  mutationFn: (variables: TVariables) => Promise<unknown>
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (_, { contextType, contextRef }) => {
      invalidateFor(queryClient, mutationEffects.layouts.updateElement(contextType, contextRef));
    },
  });
}

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
    }: LayoutContext & { request?: UpsertLayoutRequest }) =>
      layoutsApi.upsert(contextType, contextRef, request),
    onSuccess: (_, { contextType, contextRef }) => {
      invalidateFor(queryClient, mutationEffects.layouts.upsert(contextType, contextRef));
    },
  });
}

export function useDeleteLayout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ contextType, contextRef }: LayoutContext) =>
      layoutsApi.delete(contextType, contextRef),
    onSuccess: (_, { contextType, contextRef }) => {
      invalidateFor(queryClient, mutationEffects.layouts.delete(contextType, contextRef));
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
    }: LayoutContext & { preferences: Record<string, unknown>; version: number }) =>
      layoutsApi.updatePreferences(contextType, contextRef, preferences, version),
    onSuccess: (_, { contextType, contextRef }) => {
      invalidateFor(queryClient, mutationEffects.layouts.updatePreferences(contextType, contextRef));
    },
  });
}

export function useUpsertElementPosition() {
  return useLayoutMutationWithInvalidation(
    ({ contextType, contextRef, elementId, position }: LayoutContext & {
      elementId: string;
      position: ElementPositionInput;
    }) => layoutsApi.upsertElement(contextType, contextRef, elementId, position)
  );
}

export function useDeleteElementPosition() {
  return useLayoutMutationWithInvalidation(
    ({ contextType, contextRef, elementId }: LayoutContext & { elementId: string }) =>
      layoutsApi.deleteElement(contextType, contextRef, elementId)
  );
}

export function useBatchUpdateElements() {
  return useLayoutMutationWithInvalidation(
    ({ contextType, contextRef, updates }: LayoutContext & { updates: BatchUpdateItem[] }) =>
      layoutsApi.batchUpdateElements(contextType, contextRef, updates)
  );
}
