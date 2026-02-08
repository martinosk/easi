export const layoutsQueryKeys = {
  all: ['layouts'] as const,
  detail: (contextType: string, contextRef: string) =>
    [...layoutsQueryKeys.all, contextType, contextRef] as const,
};
