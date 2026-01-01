import { useQuery } from '@tanstack/react-query';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { queryKeys } from '../../../lib/queryClient';

export function useStrategicFitAnalysis(pillarId: string | null) {
  return useQuery({
    queryKey: queryKeys.strategicFitAnalysis.byPillar(pillarId!),
    queryFn: () => enterpriseArchApi.getStrategicFitAnalysis(pillarId!),
    enabled: !!pillarId,
  });
}
