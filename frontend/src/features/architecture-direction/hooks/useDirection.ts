import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { invalidateFor } from '../../../lib/invalidateFor';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { directionApi } from '../api/directionApi';
import { directionQueryKeys } from '../queryKeys';
import type {
  CaptureDirectionRequest,
  Direction,
  DirectionId,
  UpdateDirectionRequest,
} from '../types';

function hasStringMessage(err: unknown): err is { message: string } {
  if (!err || typeof err !== 'object') return false;
  const candidate = (err as { message?: unknown }).message;
  return typeof candidate === 'string';
}

function getMessage(err: unknown, fallback: string): string {
  return hasStringMessage(err) ? err.message : fallback;
}

export function useDirectionForEnterpriseCapability(enterpriseCapabilityId: EnterpriseCapabilityId | undefined) {
  return useQuery({
    queryKey: directionQueryKeys.forEnterpriseCapability(enterpriseCapabilityId ?? ''),
    queryFn: () => directionApi.getForEnterpriseCapability(enterpriseCapabilityId!),
    enabled: !!enterpriseCapabilityId,
  });
}

interface CaptureArgs {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  request: CaptureDirectionRequest;
}

export function useCaptureDirection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ enterpriseCapabilityId, request }: CaptureArgs) =>
      directionApi.capture(enterpriseCapabilityId, request),
    onSuccess: (_result: Direction, { enterpriseCapabilityId }) => {
      invalidateFor(queryClient, [directionQueryKeys.forEnterpriseCapability(enterpriseCapabilityId)]);
      toast.success('Direction captured as draft');
    },
    onError: (err) => toast.error(getMessage(err, 'Failed to capture direction')),
  });
}

interface IDArgs {
  directionId: DirectionId;
  enterpriseCapabilityId: EnterpriseCapabilityId;
}

interface AdvanceArgs extends IDArgs {
  target: 'proposed' | 'agreed';
}

export function useAdvanceDirection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ directionId, target }: AdvanceArgs) => directionApi.advance(directionId, target),
    onSuccess: (_result, { enterpriseCapabilityId, target }) => {
      invalidateFor(queryClient, [directionQueryKeys.forEnterpriseCapability(enterpriseCapabilityId)]);
      toast.success(`Direction advanced to ${target}`);
    },
    onError: (err) => toast.error(getMessage(err, 'Failed to advance direction')),
  });
}

export function useRejectDirection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ directionId }: IDArgs) => directionApi.reject(directionId),
    onSuccess: (_result, { enterpriseCapabilityId }) => {
      invalidateFor(queryClient, [directionQueryKeys.forEnterpriseCapability(enterpriseCapabilityId)]);
      toast.success('Direction rejected');
    },
    onError: (err) => toast.error(getMessage(err, 'Failed to reject direction')),
  });
}

interface UpdateArgs extends IDArgs {
  request: UpdateDirectionRequest;
}

export function useUpdateDirection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ directionId, request }: UpdateArgs) => directionApi.update(directionId, request),
    onSuccess: (_result, { enterpriseCapabilityId }) => {
      invalidateFor(queryClient, [directionQueryKeys.forEnterpriseCapability(enterpriseCapabilityId)]);
      toast.success('Direction updated');
    },
    onError: (err) => toast.error(getMessage(err, 'Failed to update direction')),
  });
}
