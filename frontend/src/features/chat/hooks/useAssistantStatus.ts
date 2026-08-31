import { useQuery } from '@tanstack/react-query';
import { chatApi } from '../api/chatApi';
import { chatQueryKeys } from '../queryKeys';

export function useAssistantStatus(enabled: boolean) {
  return useQuery({
    queryKey: chatQueryKeys.status(),
    queryFn: () => chatApi.getStatus(),
    enabled,
  });
}
