import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { chatApi } from '../api/chatApi';
import { chatQueryKeys } from '../queryKeys';

export function useConversations() {
  const queryClient = useQueryClient();

  const listQuery = useQuery({
    queryKey: chatQueryKeys.conversations(),
    queryFn: () => chatApi.listConversations(20),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => chatApi.deleteConversation(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: chatQueryKeys.conversations() });
    },
  });

  return {
    conversations: listQuery.data?.data ?? [],
    isLoading: listQuery.isLoading,
    deleteConversation: deleteMutation.mutate,
    invalidateList: () => queryClient.invalidateQueries({ queryKey: chatQueryKeys.conversations() }),
  };
}
