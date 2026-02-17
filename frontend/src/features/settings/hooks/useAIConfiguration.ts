import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { assistantConfigApi } from '../../../api/assistant/assistantConfigApi';
import type { UpdateAIConfigRequest } from '../../../api/assistant/types';
import { assistantConfigQueryKeys } from '../queryKeys';
import { invalidateFor } from '../../../lib/invalidateFor';

export function useAIConfiguration() {
  return useQuery({
    queryKey: assistantConfigQueryKeys.config(),
    queryFn: () => assistantConfigApi.getConfig(),
    staleTime: 1000 * 60 * 5,
  });
}

export function useUpdateAIConfiguration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: UpdateAIConfigRequest) =>
      assistantConfigApi.updateConfig(request),
    onSuccess: () => {
      invalidateFor(queryClient, [assistantConfigQueryKeys.config()]);
      toast.success('AI configuration saved successfully');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to save AI configuration');
    },
  });
}

export function useTestAIConnection() {
  return useMutation({
    mutationFn: () => assistantConfigApi.testConnection(),
  });
}
