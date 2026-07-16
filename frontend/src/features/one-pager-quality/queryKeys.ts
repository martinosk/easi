import type { OnePagerQualityOrder, OnePagerQualitySort } from './types';

export const onePagerQualityQueryKeys = {
  all: ['onePagerQuality'] as const,
  lists: () => [...onePagerQualityQueryKeys.all, 'list'] as const,
  list: (sort: OnePagerQualitySort, order: OnePagerQualityOrder, cursor?: string) =>
    [...onePagerQualityQueryKeys.lists(), sort, order, cursor ?? null] as const,
};
