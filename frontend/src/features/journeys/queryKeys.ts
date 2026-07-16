export const journeyQueryKeys = {
  all: ['journeys'] as const,
  active: (capabilityId: string) => [...journeyQueryKeys.all, 'active', capabilityId] as const,
  history: (capabilityId: string) => [...journeyQueryKeys.all, 'history', capabilityId] as const,
  collection: () => [...journeyQueryKeys.all, 'collection'] as const,
};
