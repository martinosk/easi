import { useQuery } from '@tanstack/react-query';
import { userApi } from '../api/userApi';
import type { User } from '../types';

const USERS_QUERY_KEY = ['users'] as const;

export function useActiveUsers() {
  return useQuery<User[]>({
    queryKey: [...USERS_QUERY_KEY, 'active'],
    queryFn: () => userApi.getAll('active'),
    staleTime: 1000 * 60 * 5,
  });
}
