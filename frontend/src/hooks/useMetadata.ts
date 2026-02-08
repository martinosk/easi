import { useQuery } from '@tanstack/react-query';
import { metadataApi } from '../api/metadata';
import { metadataQueryKeys, releasesQueryKeys } from '../lib/appQueryKeys';

export function useMaturityLevels() {
  return useQuery({
    queryKey: metadataQueryKeys.maturityLevels(),
    queryFn: () => metadataApi.getMaturityLevels(),
    staleTime: Infinity,
  });
}

export function useStatuses() {
  return useQuery({
    queryKey: metadataQueryKeys.statuses(),
    queryFn: () => metadataApi.getStatuses(),
    staleTime: Infinity,
  });
}

export function useOwnershipModels() {
  return useQuery({
    queryKey: metadataQueryKeys.ownershipModels(),
    queryFn: () => metadataApi.getOwnershipModels(),
    staleTime: Infinity,
  });
}

export function useVersion() {
  return useQuery({
    queryKey: metadataQueryKeys.version(),
    queryFn: () => metadataApi.getVersion(),
    staleTime: 1000 * 60 * 60,
  });
}

export function useReleases() {
  return useQuery({
    queryKey: releasesQueryKeys.lists(),
    queryFn: () => metadataApi.getReleases(),
    staleTime: 1000 * 60 * 5,
  });
}

export function useLatestRelease() {
  return useQuery({
    queryKey: releasesQueryKeys.latest(),
    queryFn: () => metadataApi.getLatestRelease(),
    staleTime: 1000 * 60 * 5,
  });
}
