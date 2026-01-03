import { useQuery } from '@tanstack/react-query';
import { auditApi } from '../api/auditApi';
import { queryKeys } from '../../../lib/queryClient';

export function useAuditHistory(aggregateId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.audit.history(aggregateId!),
    queryFn: () => auditApi.getHistory({ aggregateId: aggregateId! }),
    enabled: !!aggregateId,
  });
}
