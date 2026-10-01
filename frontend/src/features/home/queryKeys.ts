export const homeQueryKeys = {
  all: ['home'] as const,
  detail: () => [...homeQueryKeys.all, 'detail'] as const,
};
