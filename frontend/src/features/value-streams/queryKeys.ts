export const valueStreamsQueryKeys = {
  all: ['valueStreams'] as const,
  lists: () => [...valueStreamsQueryKeys.all, 'list'] as const,
  list: (filters?: Record<string, unknown>) =>
    [...valueStreamsQueryKeys.lists(), filters] as const,
  details: () => [...valueStreamsQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...valueStreamsQueryKeys.details(), id] as const,
};
