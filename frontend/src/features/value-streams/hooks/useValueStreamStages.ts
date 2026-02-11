import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { valueStreamsApi } from '../api';
import { valueStreamsQueryKeys } from '../queryKeys';
import { valueStreamsMutationEffects } from '../mutationEffects';
import { invalidateFor } from '../../../lib/invalidateFor';
import type {
  ValueStreamId,
  ValueStreamDetail,
  ValueStream,
  ValueStreamStage,
  StageCapabilityMapping,
  CreateStageRequest,
  UpdateStageRequest,
  ReorderStagesRequest,
} from '../../../api/types';
import toast from 'react-hot-toast';

export function useValueStreamDetail(id: ValueStreamId | undefined) {
  return useQuery<ValueStreamDetail>({
    queryKey: valueStreamsQueryKeys.detail(id!),
    queryFn: () => valueStreamsApi.getById(id!),
    enabled: !!id,
  });
}

export function useAddStage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ valueStream, request }: { valueStream: ValueStream; request: CreateStageRequest }) =>
      valueStreamsApi.addStage(valueStream, request),
    onSuccess: (result) => {
      invalidateFor(queryClient, valueStreamsMutationEffects.addStage(result.id));
      toast.success('Stage added');
    },
    onError: (error: Error) => toast.error(error.message || 'Failed to add stage'),
  });
}

export function useUpdateStage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ stage, request }: { stage: ValueStreamStage; request: UpdateStageRequest }) =>
      valueStreamsApi.updateStage(stage, request),
    onSuccess: (result) => {
      invalidateFor(queryClient, valueStreamsMutationEffects.updateStage(result.id));
      toast.success('Stage updated');
    },
    onError: (error: Error) => toast.error(error.message || 'Failed to update stage'),
  });
}

export function useDeleteStage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (stage: ValueStreamStage) =>
      valueStreamsApi.deleteStage(stage),
    onSuccess: (result) => {
      invalidateFor(queryClient, valueStreamsMutationEffects.deleteStage(result.id));
      toast.success('Stage removed');
    },
    onError: (error: Error) => toast.error(error.message || 'Failed to remove stage'),
  });
}

export function useReorderStages() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ valueStream, request }: { valueStream: ValueStream; request: ReorderStagesRequest }) =>
      valueStreamsApi.reorderStages(valueStream, request),
    onSuccess: (result) => {
      invalidateFor(queryClient, valueStreamsMutationEffects.reorderStages(result.id));
    },
    onError: (error: Error) => toast.error(error.message || 'Failed to reorder stages'),
  });
}

export function useAddStageCapability() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ stage, capabilityId }: { stage: ValueStreamStage; capabilityId: string }) =>
      valueStreamsApi.addStageCapability(stage, capabilityId),
    onSuccess: (result) => {
      invalidateFor(queryClient, valueStreamsMutationEffects.addStageCapability(result.id));
      toast.success('Capability mapped to stage');
    },
    onError: (error: Error) => toast.error(error.message || 'Failed to map capability'),
  });
}

export function useRemoveStageCapability() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (mapping: StageCapabilityMapping) =>
      valueStreamsApi.removeStageCapability(mapping),
    onSuccess: (result) => {
      invalidateFor(queryClient, valueStreamsMutationEffects.removeStageCapability(result.id));
      toast.success('Capability removed from stage');
    },
    onError: (error: Error) => toast.error(error.message || 'Failed to remove capability'),
  });
}
