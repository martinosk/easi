import { useQuery } from '@tanstack/react-query';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { enterpriseCapabilitiesQueryKeys } from '../queryKeys';

export function useComposition(enterpriseCapabilityId: EnterpriseCapabilityId | undefined) {
  return useQuery({
    queryKey: enterpriseCapabilitiesQueryKeys.composition(enterpriseCapabilityId ?? ''),
    queryFn: () => enterpriseArchApi.getComposition(enterpriseCapabilityId!),
    enabled: !!enterpriseCapabilityId,
  });
}
