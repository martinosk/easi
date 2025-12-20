import { useQuery } from '@tanstack/react-query';
import { metadataApi } from '../api/metadata';
import { queryKeys } from '../lib/queryClient';

export function useMaturityLevels() {
  return useQuery({
    queryKey: queryKeys.metadata.maturityLevels(),
    queryFn: () => metadataApi.getMaturityLevels(),
    staleTime: Infinity,
  });
}

export function useStatuses() {
  return useQuery({
    queryKey: queryKeys.metadata.statuses(),
    queryFn: () => metadataApi.getStatuses(),
    staleTime: Infinity,
  });
}

export function useOwnershipModels() {
  return useQuery({
    queryKey: queryKeys.metadata.ownershipModels(),
    queryFn: () => metadataApi.getOwnershipModels(),
    staleTime: Infinity,
  });
}

export function useStrategyPillars() {
  return useQuery({
    queryKey: queryKeys.metadata.strategyPillars(),
    queryFn: () => metadataApi.getStrategyPillars(),
    staleTime: Infinity,
  });
}

export function useVersion() {
  return useQuery({
    queryKey: queryKeys.metadata.version(),
    queryFn: () => metadataApi.getVersion(),
    staleTime: 1000 * 60 * 60,
  });
}

export function useReleases() {
  return useQuery({
    queryKey: queryKeys.releases.lists(),
    queryFn: () => metadataApi.getReleases(),
    staleTime: 1000 * 60 * 5,
  });
}

export function useLatestRelease() {
  return useQuery({
    queryKey: queryKeys.releases.latest(),
    queryFn: () => metadataApi.getLatestRelease(),
    staleTime: 1000 * 60 * 5,
  });
}
