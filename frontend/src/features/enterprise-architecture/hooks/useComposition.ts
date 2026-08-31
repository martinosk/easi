import { useQuery } from '@tanstack/react-query';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { enterpriseCapabilitiesQueryKeys } from '../queryKeys';

export function useComposition(enterpriseCapabilityId: EnterpriseCapabilityId | undefined, href?: string) {
  return useQuery({
    queryKey: [...enterpriseCapabilitiesQueryKeys.composition(enterpriseCapabilityId ?? ''), href ?? 'derived-url'],
    queryFn: () => enterpriseArchApi.getComposition(enterpriseCapabilityId!, href),
    enabled: !!enterpriseCapabilityId,
  });
}
