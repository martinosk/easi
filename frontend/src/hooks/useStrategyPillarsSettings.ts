import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { strategyPillarsApi, type BatchUpdateRequest } from '../api/metadata';
import { metadataQueryKeys } from '../lib/appQueryKeys';
import { invalidateFor } from '../lib/invalidateFor';
import { ApiError } from '../api/types';

export function useStrategyPillarsConfig() {
  return useQuery({
    queryKey: metadataQueryKeys.strategyPillarsConfig(),
    queryFn: () => strategyPillarsApi.getConfiguration(true),
    staleTime: 1000 * 60 * 5,
  });
}

export function useBatchUpdateStrategyPillars() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ request, version }: { request: BatchUpdateRequest; version: number }) =>
      strategyPillarsApi.batchUpdate(request, version),
    onSuccess: () => {
      invalidateFor(queryClient, [
        metadataQueryKeys.strategyPillarsConfig(),
      ]);
      toast.success('Strategy pillars updated successfully');
    },
    onError: (error: unknown) => {
      if (error instanceof ApiError && (error.statusCode === 409 || error.statusCode === 412)) {
        return;
      }
      const message = error instanceof Error ? error.message : 'Failed to update strategy pillars';
      toast.error(message);
    },
  });
}
