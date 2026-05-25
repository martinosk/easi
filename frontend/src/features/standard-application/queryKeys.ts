export const standardApplicationQueryKeys = {
  all: ['standard-applications'] as const,
  byEnterpriseCapability: (id: string) =>
    [...standardApplicationQueryKeys.all, 'byEC', id] as const,
  historyByEnterpriseCapability: (id: string) =>
    [...standardApplicationQueryKeys.all, 'history', 'byEC', id] as const,
};
