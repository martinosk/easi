import { useQuery } from '@tanstack/react-query';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { strategicFitAnalysisQueryKeys } from '../queryKeys';

export function useStrategicFitAnalysis(pillarId: string | null) {
  return useQuery({
    queryKey: strategicFitAnalysisQueryKeys.byPillar(pillarId!),
    queryFn: () => enterpriseArchApi.getStrategicFitAnalysis(pillarId!),
    enabled: !!pillarId,
  });
}
