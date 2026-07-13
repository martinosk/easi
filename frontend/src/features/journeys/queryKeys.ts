export const journeyQueryKeys = {
  all: ['journeys'] as const,
  active: (capabilityId: string) => [...journeyQueryKeys.all, 'active', capabilityId] as const,
  history: (capabilityId: string) => [...journeyQueryKeys.all, 'history', capabilityId] as const,
  byCapabilityIds: (capabilityIds: string[]) =>
    [...journeyQueryKeys.all, 'byCapabilityIds', [...capabilityIds].sort()] as const,
};
