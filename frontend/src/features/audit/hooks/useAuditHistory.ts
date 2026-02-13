import { useQuery } from '@tanstack/react-query';
import { auditApi } from '../api/auditApi';
import { auditQueryKeys } from '../queryKeys';

export function useAuditHistory(aggregateId: string | undefined) {
  return useQuery({
    queryKey: auditQueryKeys.history(aggregateId!),
    queryFn: () => auditApi.getHistory({ aggregateId: aggregateId! }),
    enabled: !!aggregateId,
    staleTime: 0,
    refetchOnMount: 'always',
    refetchOnWindowFocus: true,
  });
}
