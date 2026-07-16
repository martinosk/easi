import { useQuery } from '@tanstack/react-query';
import { onePagersApi } from '../api/onePagersApi';
import { onePagersQueryKeys } from '../queryKeys';
import type { OnePagerSubjectType } from '../types';

export function useOnePagerConfiguration(subjectType: OnePagerSubjectType, enabled = true) {
  return useQuery({
    queryKey: onePagersQueryKeys.configuration(subjectType),
    queryFn: () => onePagersApi.getConfiguration(subjectType),
    staleTime: Infinity,
    enabled,
  });
}
