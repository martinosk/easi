import type { UseQueryResult } from '@tanstack/react-query';
import { useMemo } from 'react';
import type { CapabilityRealizationsGroup } from '../../../api/types';
import { journeyApi } from '../../journeys/api/journeyApi';
import { journeyQueryKeys } from '../../journeys/queryKeys';
import type { CapabilityJourneysBulkResponse } from '../../journeys/types';
import { buildJourneyIndex, type JourneyIndex } from '../lens/journeyIndex';
import { type DomainBoardViewModel, flattenViewModelCapabilities } from './domainBoardViewModel';
import { useCapabilityIdQueries } from './useCapabilityIdQueries';

export function useJourneyQueries(realizationQueries: UseQueryResult<CapabilityRealizationsGroup[]>[]) {
  return useCapabilityIdQueries<CapabilityJourneysBulkResponse>(
    realizationQueries,
    journeyQueryKeys.byCapabilityIds,
    (capabilityIds) => journeyApi.getByCapabilityIds(capabilityIds),
  );
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
