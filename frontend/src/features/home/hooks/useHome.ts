import { useQuery } from '@tanstack/react-query';
import { homeApi } from '../api/homeApi';
import { homeQueryKeys } from '../queryKeys';
import type { HomeResponse } from '../types';

export function useHome() {
  return useQuery<HomeResponse>({
    queryKey: homeQueryKeys.detail(),
    queryFn: () => homeApi.getHome(),
    staleTime: 0,
    refetchOnMount: 'always',
    refetchOnWindowFocus: true,
  });
}
