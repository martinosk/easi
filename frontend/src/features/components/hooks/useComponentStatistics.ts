import { useQuery } from '@tanstack/react-query';
import { componentsApi } from '../api';
import { componentsQueryKeys } from '../queryKeys';

export function useComponentStatistics() {
  return useQuery({
    queryKey: componentsQueryKeys.statistics(),
    queryFn: () => componentsApi.getStatistics(),
  });
}
