export const directionQueryKeys = {
  all: ['directions'] as const,
  forEnterpriseCapability: (id: string) => [...directionQueryKeys.all, 'forEC', id] as const,
  detail: (id: string) => [...directionQueryKeys.all, 'detail', id] as const,
};
