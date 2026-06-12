import { type QueryKey, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { directionApi } from '../api/directionApi';
import { directionMutationEffects } from '../mutationEffects';
import { directionQueryKeys } from '../queryKeys';
import type { CaptureDirectionRequest, Direction, UpdateDirectionRequest } from '../types';

interface SourceMutationArgs {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  capabilityId: string;
}

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

function useDirectionMutation<TVars>({
  call,
  invalidate,
  successMessage,
  failureMessage,
}: DirectionMutationConfig<TVars>) {
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

export function useRevertDirection() {
  return useDirectionMutation<ByECArgs>({
    call: ({ enterpriseCapabilityId }) => directionApi.revert(enterpriseCapabilityId),
    invalidate: ({ enterpriseCapabilityId }) => directionMutationEffects.revert(enterpriseCapabilityId),
    successMessage: 'Direction returned to draft',
    failureMessage: 'Failed to return direction to draft',
  });
}

export function useAddSource() {
  return useDirectionMutation<SourceMutationArgs>({
    call: ({ enterpriseCapabilityId, capabilityId }) => directionApi.addSource(enterpriseCapabilityId, capabilityId),
    invalidate: ({ enterpriseCapabilityId }) => directionMutationEffects.addSource(enterpriseCapabilityId),
    successMessage: 'Source added',
    failureMessage: 'Failed to add source',
  });
}

export function useExcludeSource() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ enterpriseCapabilityId, capabilityId }: SourceMutationArgs) =>
      directionApi.removeSource(enterpriseCapabilityId, capabilityId),
    onSuccess: (_result, { enterpriseCapabilityId }) => {
      invalidateFor(queryClient, directionMutationEffects.removeSource(enterpriseCapabilityId));
      toast.success('Capability excluded');
    },
    onError: (err) => toast.error(getErrorMessage(err, 'Failed to exclude capability')),
  });
}

export function useSourceCandidates(enterpriseCapabilityId: EnterpriseCapabilityId | undefined) {
  return useQuery({
    queryKey: directionQueryKeys.sourceCandidates(enterpriseCapabilityId ?? '', { q: '' }),
    queryFn: () => directionApi.searchSourceCandidates(enterpriseCapabilityId!, { q: '' }),
    enabled: !!enterpriseCapabilityId,
    staleTime: 5 * 60 * 1000,
  });
}

export function useCompositionPreview(
  enterpriseCapabilityId: EnterpriseCapabilityId | undefined,
  sourceCapabilityIds: string[],
) {
  return useQuery({
    queryKey: directionQueryKeys.compositionPreview(enterpriseCapabilityId ?? '', sourceCapabilityIds),
    queryFn: () => directionApi.previewComposition(enterpriseCapabilityId!, sourceCapabilityIds),
    enabled: !!enterpriseCapabilityId && sourceCapabilityIds.length > 0,
  });
}
