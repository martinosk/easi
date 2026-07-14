import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { onePagerQualityApi } from '../api/onePagerQualityApi';
import { onePagerQualityQueryKeys } from '../queryKeys';
import type { OnePagerQualityOrder, OnePagerQualitySort } from '../types';

export interface UseOnePagerQualityListParams {
  sort: OnePagerQualitySort;
  order: OnePagerQualityOrder;
  cursor?: string;
  limit?: number;
}

export function useOnePagerQualityList({ sort, order, cursor, limit }: UseOnePagerQualityListParams) {
  return useQuery({
    queryKey: onePagerQualityQueryKeys.list(sort, order, cursor),
    queryFn: () => onePagerQualityApi.getList({ sort, order, cursor, limit }),
    placeholderData: keepPreviousData,
  });
}
