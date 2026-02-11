import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { valueStreamsApi } from '../api';
import { valueStreamsQueryKeys } from '../queryKeys';
import { valueStreamsMutationEffects } from '../mutationEffects';
import { invalidateFor } from '../../../lib/invalidateFor';
import type {
  ValueStream,
  ValueStreamId,
  CreateValueStreamRequest,
  UpdateValueStreamRequest,
  ValueStreamsResponse,
  HATEOASLinks,
} from '../../../api/types';
import toast from 'react-hot-toast';

export interface UseValueStreamsResult {
  valueStreams: ValueStream[];
  collectionLinks: HATEOASLinks | undefined;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  createValueStream: (name: string, description?: string) => Promise<ValueStream>;
  updateValueStream: (valueStream: ValueStream, name: string, description?: string) => Promise<ValueStream>;
  deleteValueStream: (valueStream: ValueStream) => Promise<void>;
}

export function useValueStreams(): UseValueStreamsResult {
  const query = useValueStreamsQuery();
  const createMutation = useCreateValueStream();
  const updateMutation = useUpdateValueStream();
  const deleteMutation = useDeleteValueStream();

  const createValueStream = useCallback(
    async (name: string, description?: string): Promise<ValueStream> => {
      return createMutation.mutateAsync({ name, description });
    },
    [createMutation]
  );

  const updateValueStream = useCallback(
    async (valueStream: ValueStream, name: string, description?: string): Promise<ValueStream> => {
      return updateMutation.mutateAsync({ valueStream, request: { name, description } });
    },
    [updateMutation]
  );

  const deleteValueStream = useCallback(
    async (valueStream: ValueStream): Promise<void> => {
      await deleteMutation.mutateAsync(valueStream);
    },
    [deleteMutation]
  );

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    valueStreams: query.data?.data ?? [],
    collectionLinks: query.data?._links,
    isLoading: query.isLoading,
    error: query.error,
    refetch,
    createValueStream,
    updateValueStream,
    deleteValueStream,
  };
}

export function useValueStreamsQuery() {
  return useQuery<ValueStreamsResponse>({
    queryKey: valueStreamsQueryKeys.lists(),
    queryFn: () => valueStreamsApi.getAll(),
  });
}

export function useValueStream(id: ValueStreamId | undefined) {
  return useQuery({
    queryKey: valueStreamsQueryKeys.detail(id!),
    queryFn: () => valueStreamsApi.getById(id!),
    enabled: !!id,
  });
}

function useValueStreamMutation<TArgs, TResult>(
  mutationFn: (args: TArgs) => Promise<TResult>,
  onMutationSuccess: (queryClient: ReturnType<typeof useQueryClient>, result: TResult, args: TArgs) => void,
  errorMessage: string
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (result, args) => onMutationSuccess(queryClient, result, args),
    onError: (error: Error) => toast.error(error.message || errorMessage),
  });
}

export function useCreateValueStream() {
  return useValueStreamMutation(
    (request: CreateValueStreamRequest) => valueStreamsApi.create(request),
    (qc, newValueStream) => {
      invalidateFor(qc, valueStreamsMutationEffects.create());
      toast.success(`Value stream "${newValueStream.name}" created`);
    },
    'Failed to create value stream'
  );
}

export function useUpdateValueStream() {
  return useValueStreamMutation(
    ({ valueStream, request }: { valueStream: ValueStream; request: UpdateValueStreamRequest }) =>
      valueStreamsApi.update(valueStream, request),
    (qc, updatedValueStream) => {
      invalidateFor(qc, valueStreamsMutationEffects.update(updatedValueStream.id));
      toast.success(`Value stream "${updatedValueStream.name}" updated`);
    },
    'Failed to update value stream'
  );
}

export function useDeleteValueStream() {
  return useValueStreamMutation(
    (valueStream: ValueStream) => valueStreamsApi.delete(valueStream),
    (qc, _, deletedValueStream) => {
      invalidateFor(qc, valueStreamsMutationEffects.delete(deletedValueStream.id));
      toast.success('Value stream deleted');
    },
    'Failed to delete value stream'
  );
}
