export const auditQueryKeys = {
  all: ['audit'] as const,
  history: (aggregateId: string) =>
    [...auditQueryKeys.all, 'history', aggregateId] as const,
};
