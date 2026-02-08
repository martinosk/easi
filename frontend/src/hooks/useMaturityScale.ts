import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { maturityScaleApi } from '../api/metadata';
import { metadataQueryKeys } from '../lib/appQueryKeys';
import { invalidateFor } from '../lib/invalidateFor';
import type { UpdateMaturityScaleRequest } from '../api/types';

export function useMaturityScale() {
  return useQuery({
    queryKey: metadataQueryKeys.maturityScale(),
    queryFn: () => maturityScaleApi.getConfiguration(),
    staleTime: 1000 * 60 * 5,
  });
}

export function useUpdateMaturityScale() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: UpdateMaturityScaleRequest) =>
      maturityScaleApi.updateConfiguration(request),
    onSuccess: () => {
      invalidateFor(queryClient, [
        metadataQueryKeys.maturityScale(),
        metadataQueryKeys.maturityLevels(),
      ]);
      toast.success('Maturity scale updated successfully');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update maturity scale');
    },
  });
}

export function useResetMaturityScale() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => maturityScaleApi.resetToDefaults(),
    onSuccess: () => {
      invalidateFor(queryClient, [
        metadataQueryKeys.maturityScale(),
        metadataQueryKeys.maturityLevels(),
      ]);
      toast.success('Maturity scale reset to defaults');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to reset maturity scale');
    },
  });
}
