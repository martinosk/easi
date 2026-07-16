import type { UseQueryResult } from '@tanstack/react-query';
import { useMemo } from 'react';
import type { CapabilityJourneysBulkResponse } from '../../journeys/types';
import { buildJourneyIndex, type JourneyIndex } from '../lens/journeyIndex';
import { type DomainBoardViewModel, flattenViewModelCapabilities } from './domainBoardViewModel';

export function useBoardJourneyIndex(
  boardDomains: DomainBoardViewModel[],
  journeysQuery: UseQueryResult<CapabilityJourneysBulkResponse>,
): JourneyIndex {
  return useMemo(() => {
    const journeys = journeysQuery.data?.data ?? [];
    const capabilityDomainNames = new Map<string, string>();
    for (const viewModel of boardDomains) {
      for (const capability of flattenViewModelCapabilities(viewModel)) {
        capabilityDomainNames.set(capability.id, viewModel.domain.name);
      }
    }
    return buildJourneyIndex({ journeys, capabilityDomainNames });
  }, [boardDomains, journeysQuery.data]);
}
