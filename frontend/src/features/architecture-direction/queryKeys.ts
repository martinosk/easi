export const directionQueryKeys = {
  all: ['directions'] as const,
  byEnterpriseCapability: (id: string) => [...directionQueryKeys.all, 'byEC', id] as const,
  sourceCandidates: (id: string, query: { q: string; domainId?: string }) =>
    [...directionQueryKeys.all, 'sourceCandidates', id, query] as const,
  compositionPreview: (id: string, sourceCapabilityIds: string[]) =>
    [...directionQueryKeys.all, 'compositionPreview', id, [...sourceCapabilityIds].sort()] as const,
};
