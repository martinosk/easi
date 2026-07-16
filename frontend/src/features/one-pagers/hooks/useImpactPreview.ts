import { useQuery } from '@tanstack/react-query';
import { onePagersApi } from '../api/onePagersApi';
import { onePagersQueryKeys } from '../queryKeys';
import type { ImpactPreviewFieldKind, OnePagerConfiguration } from '../types';

export function useImpactPreview(
  configuration: OnePagerConfiguration,
  fieldId: string | undefined,
  enabled: boolean,
  fieldKind: ImpactPreviewFieldKind = 'custom',
) {
  return useQuery({
    queryKey: onePagersQueryKeys.impactPreview(configuration.subjectType, fieldId, fieldKind),
    queryFn: () => onePagersApi.getImpactPreview(configuration, fieldId, fieldKind),
    enabled,
    staleTime: 0,
    gcTime: 0,
  });
}
