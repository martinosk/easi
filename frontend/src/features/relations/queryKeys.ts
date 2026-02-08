export const relationsQueryKeys = {
  all: ['relations'] as const,
  lists: () => [...relationsQueryKeys.all, 'list'] as const,
  list: (filters?: Record<string, unknown>) =>
    [...relationsQueryKeys.lists(), filters] as const,
  details: () => [...relationsQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...relationsQueryKeys.details(), id] as const,
};
