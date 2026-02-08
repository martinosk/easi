import { useQuery } from '@tanstack/react-query';
import { useCallback } from 'react';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { timeSuggestionsQueryKeys } from '../queryKeys';
import type { TimeSuggestion } from '../types';

export interface TimeSuggestionsFilters {
  capabilityId?: string;
  componentId?: string;
}

export interface UseTimeSuggestionsResult {
  suggestions: TimeSuggestion[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useTimeSuggestions(filters?: TimeSuggestionsFilters): UseTimeSuggestionsResult {
  const query = useQuery({
    queryKey: timeSuggestionsQueryKeys.list(filters),
    queryFn: () => enterpriseArchApi.getTimeSuggestions(filters),
  });

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    suggestions: query.data?.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    refetch,
  };
}
