import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { layoutsApi } from '../api';
import { layoutsQueryKeys } from '../queryKeys';
import { layoutsMutationEffects } from '../mutationEffects';
import { invalidateFor } from '../../../lib/invalidateFor';
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
  mutationFn: (variables: TVariables) => Promise<unknown>,
  getEffects: (contextType: LayoutContextType, contextRef: string) => ReadonlyArray<readonly unknown[]> = layoutsMutationEffects.updateElement
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (_, { contextType, contextRef }) => {
      invalidateFor(queryClient, getEffects(contextType, contextRef));
    },
  });
}

export function useLayout(
  contextType: LayoutContextType | undefined,
  contextRef: string | undefined
) {
  return useQuery({
    queryKey: layoutsQueryKeys.detail(contextType!, contextRef!),
    queryFn: () => layoutsApi.get(contextType!, contextRef!),
    enabled: !!contextType && !!contextRef,
  });
}

export function useUpsertLayout() {
  return useLayoutMutationWithInvalidation(
    ({ contextType, contextRef, request }: LayoutContext & { request?: UpsertLayoutRequest }) =>
      layoutsApi.upsert(contextType, contextRef, request),
    layoutsMutationEffects.upsert
  );
}

export function useDeleteLayout() {
  return useLayoutMutationWithInvalidation(
    ({ contextType, contextRef }: LayoutContext) =>
      layoutsApi.delete(contextType, contextRef),
    layoutsMutationEffects.delete
  );
}

export function useUpdateLayoutPreferences() {
  return useLayoutMutationWithInvalidation(
    ({ contextType, contextRef, preferences, version }: LayoutContext & { preferences: Record<string, unknown>; version: number }) =>
      layoutsApi.updatePreferences(contextType, contextRef, preferences, version),
    layoutsMutationEffects.updatePreferences
  );
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
