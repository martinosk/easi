import { useMemo } from 'react';
import { useQueries } from '@tanstack/react-query';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { useOriginRelationshipsQuery } from '../../origin-entities/hooks/useOriginRelationships';
import { businessDomainsQueryKeys } from '../../business-domains/queryKeys';
import { businessDomainsApi } from '../../business-domains/api';
import { apiClient } from '../../../api/client';
import type { BusinessDomainId, Capability, CapabilityRealizationsGroup } from '../../../api/types';
import type { DomainFilterData } from '../utils/filterByDomain';

export function useDomainFilterData(allCapabilities: Capability[]) {
  const { data: domainsResponse } = useBusinessDomainsQuery();
  const domains = useMemo(() => domainsResponse?.data ?? [], [domainsResponse]);

  const capabilityQueries = useQueries({
    queries: domains.map((domain) => ({
      queryKey: businessDomainsQueryKeys.capabilities(domain.id),
      queryFn: () => businessDomainsApi.getCapabilitiesByDomainId(domain.id as BusinessDomainId),
      staleTime: 1000 * 60 * 5,
    })),
  });

  const realizationQueries = useQueries({
    queries: domains.map((domain) => ({
      queryKey: businessDomainsQueryKeys.realizations(domain.id, 4),
      queryFn: () => apiClient.getCapabilityRealizationsByDomain(domain.id as BusinessDomainId, 4),
      staleTime: 1000 * 60 * 5,
    })),
  });

  const { data: originRelationships = [] } = useOriginRelationshipsQuery();

  const domainFilterData: DomainFilterData = useMemo(() => {
    const domainCapabilityIds = new Map<string, string[]>();
    const domainComponentIds = new Map<string, string[]>();
    const allDomainIds: string[] = [];

    domains.forEach((domain, index) => {
      allDomainIds.push(domain.id);

      const caps = capabilityQueries[index]?.data;
      if (caps) {
        domainCapabilityIds.set(domain.id, caps.map((c) => c.id));
      }

      const groups: CapabilityRealizationsGroup[] | undefined = realizationQueries[index]?.data;
      if (groups) {
        const componentIds = new Set<string>();
        for (const group of groups) {
          for (const r of group.realizations) {
            componentIds.add(r.componentId);
          }
        }
        domainComponentIds.set(domain.id, [...componentIds]);
      }
    });

    return {
      domainCapabilityIds,
      allCapabilities,
      domainComponentIds,
      originRelationships,
      allDomainIds,
    };
  }, [domains, capabilityQueries, realizationQueries, allCapabilities, originRelationships]);

  return { domains, domainFilterData };
}
