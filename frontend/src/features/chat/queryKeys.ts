export const chatQueryKeys = {
  all: ['chat'] as const,
  conversations: () => [...chatQueryKeys.all, 'conversations'] as const,
  status: () => [...chatQueryKeys.all, 'status'] as const,
};
