import { type UseQueryResult, useQueries } from '@tanstack/react-query';
import { useMemo } from 'react';
import type { CapabilityRealizationsGroup } from '../../../api/types';
import { journeyApi } from '../../journeys/api/journeyApi';
import { journeyQueryKeys } from '../../journeys/queryKeys';
import type { CapabilityJourneysBulkResponse } from '../../journeys/types';
import { buildJourneyIndex, type JourneyIndex } from '../lens/journeyIndex';
import { type DomainBoardViewModel, flattenViewModelCapabilities } from './domainBoardViewModel';

export function useJourneyQueries(realizationQueries: UseQueryResult<CapabilityRealizationsGroup[]>[]) {
  return useQueries({
    queries: realizationQueries.map((realizationQuery) => {
      const capabilityIds = (realizationQuery.data ?? []).map((group) => group.capabilityId);
      return {
        queryKey: journeyQueryKeys.byCapabilityIds(capabilityIds),
        queryFn: () => journeyApi.getByCapabilityIds(capabilityIds),
        enabled: capabilityIds.length > 0,
      };
    }),
  });
}

export function useBoardJourneyIndex(
  boardDomains: DomainBoardViewModel[],
  journeyQueries: UseQueryResult<CapabilityJourneysBulkResponse>[],
): JourneyIndex {
  return useMemo(() => {
    const journeys = journeyQueries.flatMap((query) => query.data?.data ?? []);
    const capabilityDomainNames = new Map<string, string>();
    for (const viewModel of boardDomains) {
      for (const capability of flattenViewModelCapabilities(viewModel)) {
        capabilityDomainNames.set(capability.id, viewModel.domain.name);
      }
    }
    return buildJourneyIndex({ journeys, capabilityDomainNames });
  }, [boardDomains, journeyQueries]);
}
