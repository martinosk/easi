import type { UseQueryResult } from '@tanstack/react-query';
import { useQueries } from '@tanstack/react-query';
import { useCallback, useMemo } from 'react';
import type { BusinessDomain, BusinessDomainId, Capability, CapabilityRealizationsGroup } from '../../../api/types';
import { canCreate } from '../../../utils/hateoas';
import { realizationRoleApi } from '../../architecture-direction/api/realizationRoleApi';
import { timeAssessmentApi } from '../../architecture-direction/api/timeAssessmentApi';
import { realizationRoleQueryKeys, timeAssessmentQueryKeys } from '../../architecture-direction/queryKeys';
import type { RealizationRolesResponse, TimeAssessmentsResponse } from '../../architecture-direction/types';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { useCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import { businessDomainsApi } from '../api';
import { businessDomainsQueryKeys } from '../queryKeys';
import { buildDomainBoardViewModel, type DomainBoardViewModel } from './domainBoardViewModel';
import { useBoardJourneyIndex, useJourneyQueries } from './useBoardJourneys';
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

function useAssessmentQueries(realizationQueries: UseQueryResult<CapabilityRealizationsGroup[]>[]) {
  return useQueries({
    queries: realizationQueries.map((realizationQuery) => {
      const capabilityIds = (realizationQuery.data ?? []).map((g) => g.capabilityId);
      return {
        queryKey: timeAssessmentQueryKeys.byCapabilityIds(capabilityIds),
        queryFn: () => timeAssessmentApi.getByCapabilityIds(capabilityIds),
        enabled: capabilityIds.length > 0,
      };
    }),
  });
}

function useRoleQueries(realizationQueries: UseQueryResult<CapabilityRealizationsGroup[]>[]) {
  return useQueries({
    queries: realizationQueries.map((realizationQuery) => {
      const capabilityIds = (realizationQuery.data ?? []).map((g) => g.capabilityId);
      return {
        queryKey: realizationRoleQueryKeys.byCapabilityIds(capabilityIds),
        queryFn: () => realizationRoleApi.getByCapabilityIds(capabilityIds),
        enabled: capabilityIds.length > 0,
      };
    }),
  });
}

function assembleBoardDomains(
  domains: BusinessDomain[],
  tree: CapabilityTreeNode[],
  capabilityQueries: UseQueryResult<Capability[]>[],
  realizationQueries: UseQueryResult<CapabilityRealizationsGroup[]>[],
  assessmentQueries: UseQueryResult<TimeAssessmentsResponse>[],
  roleQueries: UseQueryResult<RealizationRolesResponse>[],
): DomainBoardViewModel[] {
  return domains.map((domain, index) =>
    buildDomainBoardViewModel({
      domain,
      assignedCapabilities: capabilityQueries[index]?.data ?? [],
      tree,
      realizationGroups: realizationQueries[index]?.data ?? [],
      isLoading: Boolean(capabilityQueries[index]?.isLoading || realizationQueries[index]?.isLoading),
      assessments: assessmentQueries[index]?.data?.data ?? [],
      roles: roleQueries[index]?.data?.data ?? [],
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
  const assessmentQueries = useAssessmentQueries(realizationQueries);
  const roleQueries = useRoleQueries(realizationQueries);
  const journeyQueries = useJourneyQueries(realizationQueries);

  const boardDomains = useMemo(
    () => assembleBoardDomains(domains, tree, capabilityQueries, realizationQueries, assessmentQueries, roleQueries),
    [domains, tree, capabilityQueries, realizationQueries, assessmentQueries, roleQueries],
  );

  const journeyIndex = useBoardJourneyIndex(boardDomains, journeyQueries);

  const refetchDomain = useCallback(
    async (domainId: BusinessDomainId) => {
      const index = domains.findIndex((domain) => domain.id === domainId);
      if (index === -1) return;
      await Promise.all([
        capabilityQueries[index]?.refetch(),
        realizationQueries[index]?.refetch(),
        assessmentQueries[index]?.refetch(),
        roleQueries[index]?.refetch(),
        journeyQueries[index]?.refetch(),
      ]);
    },
    [domains, capabilityQueries, realizationQueries, assessmentQueries, roleQueries, journeyQueries],
  );

  return {
    domains,
    boardDomains,
    journeyIndex,
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
