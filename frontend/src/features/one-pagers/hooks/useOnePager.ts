import { useQuery } from '@tanstack/react-query';
import { onePagersApi } from '../api/onePagersApi';
import { onePagersQueryKeys } from '../queryKeys';
import type { OnePagerSubjectType } from '../types';

export function useOnePager(subjectType: OnePagerSubjectType | undefined, subjectId: string | undefined) {
  return useQuery({
    queryKey: onePagersQueryKeys.onePager(subjectType!, subjectId!),
    queryFn: () => onePagersApi.getOnePager(subjectType!, subjectId!),
    enabled: !!subjectType && !!subjectId,
  });
}
