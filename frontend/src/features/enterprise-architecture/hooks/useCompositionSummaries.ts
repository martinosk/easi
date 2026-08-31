import { useQuery } from '@tanstack/react-query';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { compositionSummariesQueryKeys } from '../queryKeys';
import type { CompositionSummary } from '../types';

export type CompositionSummaryLookup = ReadonlyMap<string, CompositionSummary>;

function indexByEnterpriseCapability(summaries: CompositionSummary[]): CompositionSummaryLookup {
  return new Map(summaries.map((summary) => [summary.enterpriseCapabilityId, summary]));
}

export function useCompositionSummaries() {
  return useQuery({
    queryKey: compositionSummariesQueryKeys.lists(),
    queryFn: () => enterpriseArchApi.getCompositionSummaries(),
    select: indexByEnterpriseCapability,
  });
}
