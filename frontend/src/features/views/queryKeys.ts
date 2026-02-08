export const viewsQueryKeys = {
  all: ['views'] as const,
  lists: () => [...viewsQueryKeys.all, 'list'] as const,
  list: (filters?: Record<string, unknown>) =>
    [...viewsQueryKeys.lists(), filters] as const,
  details: () => [...viewsQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...viewsQueryKeys.details(), id] as const,
  components: (viewId: string) => [...viewsQueryKeys.detail(viewId), 'components'] as const,
};
