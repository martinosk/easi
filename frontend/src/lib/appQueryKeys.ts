export const metadataQueryKeys = {
  all: ['metadata'] as const,
  maturityLevels: () => [...metadataQueryKeys.all, 'maturityLevels'] as const,
  maturityScale: () => [...metadataQueryKeys.all, 'maturityScale'] as const,
  statuses: () => [...metadataQueryKeys.all, 'statuses'] as const,
  ownershipModels: () => [...metadataQueryKeys.all, 'ownershipModels'] as const,
  strategyPillarsConfig: () => [...metadataQueryKeys.all, 'strategyPillarsConfig'] as const,
  version: () => [...metadataQueryKeys.all, 'version'] as const,
};

export const releasesQueryKeys = {
  all: ['releases'] as const,
  lists: () => [...releasesQueryKeys.all, 'list'] as const,
  latest: () => [...releasesQueryKeys.all, 'latest'] as const,
  detail: (version: string) => [...releasesQueryKeys.all, 'detail', version] as const,
};
