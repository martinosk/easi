import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { invalidateFor } from '../../../lib/invalidateFor';
import { timeAssessmentApi } from '../api/timeAssessmentApi';
import { timeAssessmentMutationEffects } from '../mutationEffects';
import { timeAssessmentQueryKeys } from '../queryKeys';
import type { AssessRealizationRequest, TimeAssessment } from '../types';

function getErrorMessage(err: unknown, fallback: string): string {
  return err instanceof Error ? err.message : fallback;
}

export function useAllTimeAssessments() {
  return useQuery({
    queryKey: timeAssessmentQueryKeys.collection(),
    queryFn: () => timeAssessmentApi.getAll(),
  });
}

export function useTimeAssessmentsByCapabilityIds(capabilityIds: string[]) {
  return useQuery({
    queryKey: timeAssessmentQueryKeys.byCapabilityIds(capabilityIds),
    queryFn: () => timeAssessmentApi.getByCapabilityIds(capabilityIds),
    enabled: capabilityIds.length > 0,
  });
}

export function useTimeAssessmentRollups(componentIds: string[]) {
  return useQuery({
    queryKey: timeAssessmentQueryKeys.rollups(componentIds),
    queryFn: () => timeAssessmentApi.getRollups(componentIds),
    enabled: componentIds.length > 0,
  });
}

interface AssessArgs {
  capabilityId: string;
  componentId: string;
  request: AssessRealizationRequest;
}

export function useAssessRealization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ capabilityId, componentId, request }: AssessArgs) =>
      timeAssessmentApi.assess(capabilityId, componentId, request),
    onSuccess: () => {
      invalidateFor(queryClient, timeAssessmentMutationEffects.assess());
      toast.success('Assessment saved');
    },
    onError: (err) => toast.error(getErrorMessage(err, 'Failed to save assessment')),
  });
}

interface RemoveArgs {
  assessment: TimeAssessment;
}

export function useRemoveTimeAssessment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ assessment }: RemoveArgs) => timeAssessmentApi.remove(assessment),
    onSuccess: () => {
      invalidateFor(queryClient, timeAssessmentMutationEffects.remove());
      toast.success('Assessment removed');
    },
    onError: (err) => toast.error(getErrorMessage(err, 'Failed to remove assessment')),
  });
}
