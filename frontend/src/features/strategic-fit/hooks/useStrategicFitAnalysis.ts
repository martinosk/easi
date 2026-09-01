import { useQuery } from '@tanstack/react-query';
import { strategicFitApi } from '../api/strategicFitApi';
import { strategicFitAnalysisQueryKeys } from '../queryKeys';

export function useStrategicFitAnalysis(pillarId: string | null) {
  return useQuery({
    queryKey: strategicFitAnalysisQueryKeys.byPillar(pillarId!),
    queryFn: () => strategicFitApi.getStrategicFitAnalysis(pillarId!),
    enabled: !!pillarId,
  });
}
