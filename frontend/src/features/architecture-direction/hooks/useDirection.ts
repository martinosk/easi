import { type QueryKey, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { invalidateFor } from '../../../lib/invalidateFor';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { directionApi } from '../api/directionApi';
import { directionMutationEffects } from '../mutationEffects';
import { directionQueryKeys } from '../queryKeys';
import type {
  CaptureDirectionRequest,
  Direction,
  UpdateDirectionRequest,
} from '../types';

function getErrorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error) return err.message;
  return fallback;
}

export function useDirectionForEnterpriseCapability(enterpriseCapabilityId: EnterpriseCapabilityId | undefined) {
  return useQuery({
    queryKey: directionQueryKeys.byEnterpriseCapability(enterpriseCapabilityId ?? ''),
    queryFn: () => directionApi.getForEnterpriseCapability(enterpriseCapabilityId!),
    enabled: !!enterpriseCapabilityId,
  });
}

interface DirectionMutationConfig<TVars> {
  call: (vars: TVars) => Promise<Direction>;
  invalidate: (vars: TVars) => QueryKey[];
  successMessage: string;
  failureMessage: string;
}

function useDirectionMutation<TVars>({ call, invalidate, successMessage, failureMessage }: DirectionMutationConfig<TVars>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: call,
    onSuccess: (_result, vars) => {
      invalidateFor(queryClient, invalidate(vars));
      toast.success(successMessage);
    },
    onError: (err) => toast.error(getErrorMessage(err, failureMessage)),
  });
}

interface ByECArgs {
  enterpriseCapabilityId: EnterpriseCapabilityId;
}

interface CaptureArgs extends ByECArgs {
  request: CaptureDirectionRequest;
}

interface UpdateArgs extends ByECArgs {
  request: UpdateDirectionRequest;
}

export function useCaptureDirection() {
  return useDirectionMutation<CaptureArgs>({
    call: ({ enterpriseCapabilityId, request }) => directionApi.capture(enterpriseCapabilityId, request),
    invalidate: ({ enterpriseCapabilityId }) => directionMutationEffects.capture(enterpriseCapabilityId),
    successMessage: 'Direction captured as draft',
    failureMessage: 'Failed to capture direction',
  });
}

export function useUpdateDirection() {
  return useDirectionMutation<UpdateArgs>({
    call: ({ enterpriseCapabilityId, request }) => directionApi.update(enterpriseCapabilityId, request),
    invalidate: ({ enterpriseCapabilityId }) => directionMutationEffects.update(enterpriseCapabilityId),
    successMessage: 'Direction updated',
    failureMessage: 'Failed to update direction',
  });
}

export function useProposeDirection() {
  return useDirectionMutation<ByECArgs>({
    call: ({ enterpriseCapabilityId }) => directionApi.propose(enterpriseCapabilityId),
    invalidate: ({ enterpriseCapabilityId }) => directionMutationEffects.propose(enterpriseCapabilityId),
    successMessage: 'Direction advanced to proposed',
    failureMessage: 'Failed to propose direction',
  });
}

export function useAgreeDirection() {
  return useDirectionMutation<ByECArgs>({
    call: ({ enterpriseCapabilityId }) => directionApi.agree(enterpriseCapabilityId),
    invalidate: ({ enterpriseCapabilityId }) => directionMutationEffects.agree(enterpriseCapabilityId),
    successMessage: 'Direction advanced to agreed',
    failureMessage: 'Failed to agree direction',
  });
}

export function useRejectDirection() {
  return useDirectionMutation<ByECArgs>({
    call: ({ enterpriseCapabilityId }) => directionApi.reject(enterpriseCapabilityId),
    invalidate: ({ enterpriseCapabilityId }) => directionMutationEffects.reject(enterpriseCapabilityId),
    successMessage: 'Direction rejected',
    failureMessage: 'Failed to reject direction',
  });
}
