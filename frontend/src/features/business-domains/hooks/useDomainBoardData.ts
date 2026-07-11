import type { UseQueryResult } from '@tanstack/react-query';
import { useQueries } from '@tanstack/react-query';
import { useCallback, useMemo } from 'react';
import type { BusinessDomain, BusinessDomainId, Capability, CapabilityRealizationsGroup } from '../../../api/types';
import { canCreate } from '../../../utils/hateoas';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { useCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import { businessDomainsApi } from '../api';
import { businessDomainsQueryKeys } from '../queryKeys';
import { buildDomainBoardViewModel, type DomainBoardViewModel } from './domainBoardViewModel';
import { useBusinessDomains } from './useBusinessDomains';

export const REALIZATION_DEPTH = 4;

function useCapabilityQueries(domains: BusinessDomain[]) {
  return useQueries({
    queries: domains.map((domain) => ({
      queryKey: businessDomainsQueryKeys.capabilities(domain.id),
      queryFn: () => businessDomainsApi.getCapabilitiesByDomainId(domain.id),
    })),
  });
}

function useRealizationQueries(domains: BusinessDomain[]) {
  return useQueries({
    queries: domains.map((domain) => ({
      queryKey: businessDomainsQueryKeys.realizations(domain.id, REALIZATION_DEPTH),
      queryFn: () => businessDomainsApi.getCapabilityRealizations(domain.id, REALIZATION_DEPTH),
    })),
  });
}

function assembleBoardDomains(
  domains: BusinessDomain[],
  tree: CapabilityTreeNode[],
  capabilityQueries: UseQueryResult<Capability[]>[],
  realizationQueries: UseQueryResult<CapabilityRealizationsGroup[]>[],
): DomainBoardViewModel[] {
  return domains.map((domain, index) =>
    buildDomainBoardViewModel({
      domain,
      assignedCapabilities: capabilityQueries[index]?.data ?? [],
      tree,
      realizationGroups: realizationQueries[index]?.data ?? [],
      isLoading: Boolean(capabilityQueries[index]?.isLoading || realizationQueries[index]?.isLoading),
    }),
  );
}

export function useDomainBoardData() {
  const {
    domains,
    collectionLinks,
    isLoading: domainsLoading,
    error,
    createDomain,
    updateDomain,
    deleteDomain,
  } = useBusinessDomains();
  const { tree, isLoading: treeLoading } = useCapabilityTree();
  const { data: allCapabilities = [] } = useCapabilities();

  const capabilityQueries = useCapabilityQueries(domains);
  const realizationQueries = useRealizationQueries(domains);

  const boardDomains = useMemo(
    () => assembleBoardDomains(domains, tree, capabilityQueries, realizationQueries),
    [domains, tree, capabilityQueries, realizationQueries],
  );

  const refetchDomain = useCallback(
    async (domainId: BusinessDomainId) => {
      const index = domains.findIndex((domain) => domain.id === domainId);
      if (index === -1) return;
      await Promise.all([capabilityQueries[index]?.refetch(), realizationQueries[index]?.refetch()]);
    },
    [domains, capabilityQueries, realizationQueries],
  );

  return {
    domains,
    boardDomains,
    canCreateDomain: canCreate({ _links: collectionLinks }),
    isLoading: domainsLoading || treeLoading,
    error,
    tree,
    treeLoading,
    allCapabilities,
    refetchDomain,
    createDomain,
    updateDomain,
    deleteDomain,
  };
}
