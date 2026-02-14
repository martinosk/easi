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

function useStageMutation<TVariables>(
  mutationFn: (vars: TVariables) => Promise<ValueStreamDetail>,
  effectKey: keyof typeof valueStreamsMutationEffects,
  successMsg?: string,
  errorMsg?: string,
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (result) => {
      invalidateFor(queryClient, valueStreamsMutationEffects[effectKey](result.id));
      if (successMsg) toast.success(successMsg);
    },
    onError: (error: Error) => toast.error(error.message || errorMsg || 'Operation failed'),
  });
}

export function useAddStage() {
  return useStageMutation(
    ({ valueStream, request }: { valueStream: ValueStream; request: CreateStageRequest }) =>
      valueStreamsApi.addStage(valueStream, request),
    'addStage', 'Stage added', 'Failed to add stage',
  );
}

export function useUpdateStage() {
  return useStageMutation(
    ({ stage, request }: { stage: ValueStreamStage; request: UpdateStageRequest }) =>
      valueStreamsApi.updateStage(stage, request),
    'updateStage', 'Stage updated', 'Failed to update stage',
  );
}

export function useDeleteStage() {
  return useStageMutation(
    (stage: ValueStreamStage) => valueStreamsApi.deleteStage(stage),
    'deleteStage', 'Stage removed', 'Failed to remove stage',
  );
}

export function useReorderStages() {
  return useStageMutation(
    ({ valueStream, request }: { valueStream: ValueStream; request: ReorderStagesRequest }) =>
      valueStreamsApi.reorderStages(valueStream, request),
    'reorderStages', undefined, 'Failed to reorder stages',
  );
}

export function useAddStageCapability() {
  return useStageMutation(
    ({ stage, capabilityId }: { stage: ValueStreamStage; capabilityId: string }) =>
      valueStreamsApi.addStageCapability(stage, capabilityId),
    'addStageCapability', 'Capability mapped to stage', 'Failed to map capability',
  );
}

export function useRemoveStageCapability() {
  return useStageMutation(
    (mapping: StageCapabilityMapping) => valueStreamsApi.removeStageCapability(mapping),
    'removeStageCapability', 'Capability removed from stage', 'Failed to remove capability',
  );
}
