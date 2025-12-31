import { useCallback, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { capabilitiesApi } from '../../capabilities/api/capabilitiesApi';
import { queryKeys } from '../../../lib/queryClient';
import { useLinkDomainCapability } from './useEnterpriseCapabilities';
import type { EnterpriseCapabilityId, CapabilityLinkStatusResponse } from '../types';
import type { Capability } from '../../../api/types';

export interface UseDomainCapabilityLinkingResult {
  domainCapabilities: Capability[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  isLoading: boolean;
  error: string | null;
  linkCapability: (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => Promise<void>;
}

export function useDomainCapabilityLinking(enabled: boolean): UseDomainCapabilityLinkingResult {
  const queryClient = useQueryClient();

  const domainQuery = useQuery({
    queryKey: queryKeys.capabilities.lists(),
    queryFn: () => capabilitiesApi.getAll(),
    enabled,
  });

  const domainCapabilityIds = useMemo(
    () => domainQuery.data?.map((c) => c.id) ?? [],
    [domainQuery.data]
  );

  const linkStatusQuery = useQuery({
    queryKey: ['linkStatuses', domainCapabilityIds],
    queryFn: () => enterpriseArchApi.getBatchLinkStatus(domainCapabilityIds),
    enabled: enabled && domainCapabilityIds.length > 0,
  });

  const linkStatuses = useMemo(() => {
    if (!linkStatusQuery.data) return new Map<string, CapabilityLinkStatusResponse>();
    return new Map(linkStatusQuery.data.map((s) => [s.capabilityId, s]));
  }, [linkStatusQuery.data]);

  const linkMutation = useLinkDomainCapability();

  const linkCapability = useCallback(
    async (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => {
      await linkMutation.mutateAsync({
        enterpriseCapabilityId,
        request: { domainCapabilityId: domainCapability.id },
      });
      queryClient.invalidateQueries({ queryKey: ['linkStatuses'] });
    },
    [linkMutation, queryClient]
  );

  return {
    domainCapabilities: domainQuery.data ?? [],
    linkStatuses,
    isLoading: domainQuery.isLoading || linkStatusQuery.isLoading,
    error: domainQuery.error?.message || linkStatusQuery.error?.message || null,
    linkCapability,
  };
}
