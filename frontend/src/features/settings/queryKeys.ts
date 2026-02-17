export const assistantConfigQueryKeys = {
  all: ['assistantConfig'] as const,
  config: () => [...assistantConfigQueryKeys.all, 'config'] as const,
};
