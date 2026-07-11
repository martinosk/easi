import { useQuery } from '@tanstack/react-query';
import { onePagersApi } from '../api/onePagersApi';
import { onePagersQueryKeys } from '../queryKeys';
import type { OnePagerSubjectType } from '../types';

export function useOnePagerFacts(subjectType: OnePagerSubjectType, subjectId: string) {
  return useQuery({
    queryKey: onePagersQueryKeys.factsForSubject(subjectType, subjectId),
    queryFn: () => onePagersApi.getFacts(subjectType, subjectId),
    enabled: !!subjectId,
  });
}
