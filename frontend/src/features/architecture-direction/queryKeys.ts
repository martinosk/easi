export const directionQueryKeys = {
  all: ['directions'] as const,
  byEnterpriseCapability: (id: string) => [...directionQueryKeys.all, 'byEC', id] as const,
};
