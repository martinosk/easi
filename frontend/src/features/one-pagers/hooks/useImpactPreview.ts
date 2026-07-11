import { useQuery } from '@tanstack/react-query';
import { onePagersApi } from '../api/onePagersApi';
import { onePagersQueryKeys } from '../queryKeys';
import type { OnePagerConfiguration } from '../types';

export function useImpactPreview(configuration: OnePagerConfiguration, fieldId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: onePagersQueryKeys.impactPreview(configuration.subjectType, fieldId),
    queryFn: () => onePagersApi.getImpactPreview(configuration, fieldId),
    enabled,
    staleTime: 0,
    gcTime: 0,
  });
}
