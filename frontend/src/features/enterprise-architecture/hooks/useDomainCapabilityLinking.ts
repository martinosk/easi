import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback, useMemo } from 'react';
import type { Capability } from '../../../api/types';
import { capabilitiesApi } from '../../capabilities/api/capabilitiesApi';
import { capabilitiesQueryKeys } from '../../capabilities/queryKeys';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { enterpriseCapabilitiesQueryKeys } from '../queryKeys';
import type { CapabilityLinkStatusResponse, EnterpriseCapabilityId } from '../types';
import { useLinkDomainCapability } from './useEnterpriseCapabilities';

export interface UseDomainCapabilityLinkingResult {
  domainCapabilities: Capability[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  isLoading: boolean;
  error: string | null;
  linkCapability: (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => Promise<void>;
}

function useDomainCapabilitiesQuery(enabled: boolean) {
  return useQuery({
    queryKey: capabilitiesQueryKeys.lists(),
    queryFn: () => capabilitiesApi.getAll(),
    enabled,
  });
}

function useLinkStatusQuery(enabled: boolean, capabilityIds: string[]) {
  return useQuery({
    queryKey: [...enterpriseCapabilitiesQueryKeys.linkStatuses(), capabilityIds] as const,
    queryFn: () => enterpriseArchApi.getBatchLinkStatus(capabilityIds),
    enabled: enabled && capabilityIds.length > 0,
  });
}

function toLinkStatusMap(data: CapabilityLinkStatusResponse[] | undefined): Map<string, CapabilityLinkStatusResponse> {
  if (!data) return new Map();
  return new Map(data.map((s) => [s.capabilityId, s]));
}

function getFirstError(...queries: { error: Error | null }[]): string | null {
  for (const q of queries) {
    if (q.error) return q.error.message;
  }
  return null;
}

export function useDomainCapabilityLinking(enabled: boolean): UseDomainCapabilityLinkingResult {
  const queryClient = useQueryClient();
  const domainQuery = useDomainCapabilitiesQuery(enabled);

  const capabilityIds = useMemo(() => domainQuery.data?.map((c) => c.id) ?? [], [domainQuery.data]);
  const linkStatusQuery = useLinkStatusQuery(enabled, capabilityIds);
  const linkStatuses = useMemo(() => toLinkStatusMap(linkStatusQuery.data), [linkStatusQuery.data]);

  const linkMutation = useLinkDomainCapability();

  const linkCapability = useCallback(
    async (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => {
      await linkMutation.mutateAsync({
        enterpriseCapabilityId,
        request: { domainCapabilityId: domainCapability.id },
      });
      queryClient.invalidateQueries({ queryKey: enterpriseCapabilitiesQueryKeys.linkStatuses() });
    },
    [linkMutation, queryClient],
  );

  return {
    domainCapabilities: domainQuery.data ?? [],
    linkStatuses,
    isLoading: domainQuery.isLoading || linkStatusQuery.isLoading,
    error: getFirstError(domainQuery, linkStatusQuery),
    linkCapability,
  };
}
