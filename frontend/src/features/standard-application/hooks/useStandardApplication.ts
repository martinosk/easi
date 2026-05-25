import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { standardApplicationApi } from '../api/standardApplicationApi';
import { standardApplicationMutationEffects } from '../mutationEffects';
import { standardApplicationQueryKeys } from '../queryKeys';
import type { SetStandardApplicationRequest } from '../types';

function getErrorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error) return err.message;
  return fallback;
}

export function useStandardApplicationForEnterpriseCapability(
  enterpriseCapabilityId: EnterpriseCapabilityId | undefined,
) {
  return useQuery({
    queryKey: standardApplicationQueryKeys.byEnterpriseCapability(enterpriseCapabilityId ?? ''),
    queryFn: () => standardApplicationApi.getForEnterpriseCapability(enterpriseCapabilityId!),
    enabled: !!enterpriseCapabilityId,
  });
}

export function useStandardApplicationHistory(
  enterpriseCapabilityId: EnterpriseCapabilityId | undefined,
  enabled: boolean,
) {
  return useQuery({
    queryKey: standardApplicationQueryKeys.historyByEnterpriseCapability(enterpriseCapabilityId ?? ''),
    queryFn: () => standardApplicationApi.getHistory(enterpriseCapabilityId!),
    enabled: !!enterpriseCapabilityId && enabled,
  });
}

interface SetArgs {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  request: SetStandardApplicationRequest;
}

export function useSetStandardApplication() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ enterpriseCapabilityId, request }: SetArgs) =>
      standardApplicationApi.set(enterpriseCapabilityId, request),
    onSuccess: (_result, { enterpriseCapabilityId }) => {
      invalidateFor(queryClient, standardApplicationMutationEffects.set(enterpriseCapabilityId));
      toast.success('Standard application updated');
    },
    onError: (err) => toast.error(getErrorMessage(err, 'Failed to update standard application')),
  });
}
