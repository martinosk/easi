import { useQuery } from '@tanstack/react-query';
import { onePagersApi } from '../api/onePagersApi';
import { onePagersQueryKeys } from '../queryKeys';
import type { OnePagerCompletenessEntry, OnePagerSubjectType } from '../types';

export type OnePagerCompletenessLookup = ReadonlyMap<string, boolean>;

function indexBySubject(entries: OnePagerCompletenessEntry[]): OnePagerCompletenessLookup {
  return new Map(entries.map((entry) => [entry.subjectId, entry.complete]));
}

export function useOnePagerCompleteness(subjectType: OnePagerSubjectType) {
  return useQuery({
    queryKey: onePagersQueryKeys.completenessForSubjectType(subjectType),
    queryFn: () => onePagersApi.getCompleteness(subjectType),
    select: indexBySubject,
  });
}
