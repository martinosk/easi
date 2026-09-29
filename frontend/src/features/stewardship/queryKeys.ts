export const stewardshipQueryKeys = {
  all: ['stewardships'] as const,
  domain: (domainId: string) => [...stewardshipQueryKeys.all, 'domain', domainId] as const,
};
