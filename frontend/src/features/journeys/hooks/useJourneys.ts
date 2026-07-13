import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { invalidateFor } from '../../../lib/invalidateFor';
import { journeyApi } from '../api/journeyApi';
import { journeyMutationEffects } from '../mutationEffects';
import { journeyQueryKeys } from '../queryKeys';
import type {
  AddJourneyMilestoneRequest,
  CapabilityJourney,
  CaptureJourneyRequest,
  ChangeJourneySourceApplicationsRequest,
  JourneyMilestone,
  UpdateJourneyDetailsRequest,
  UpdateJourneyMilestoneRequest,
  UpdateJourneyProgressRequest,
} from '../types';

function getErrorMessage(err: unknown, fallback: string): string {
  return err instanceof Error ? err.message : fallback;
}

export function useJourneyForCapability(capabilityId: string | undefined) {
  return useQuery({
    queryKey: journeyQueryKeys.active(capabilityId ?? ''),
    queryFn: () => journeyApi.getForCapability(capabilityId!),
    enabled: !!capabilityId,
  });
}

export function useJourneyHistory(capabilityId: string | undefined) {
  return useQuery({
    queryKey: journeyQueryKeys.history(capabilityId ?? ''),
    queryFn: () => journeyApi.getHistory(capabilityId!),
    enabled: !!capabilityId,
  });
}

export function useJourneysByCapabilityIds(capabilityIds: string[]) {
  return useQuery({
    queryKey: journeyQueryKeys.byCapabilityIds(capabilityIds),
    queryFn: () => journeyApi.getByCapabilityIds(capabilityIds),
    enabled: capabilityIds.length > 0,
  });
}

function useJourneyMutation<TArgs, TResult>(
  mutationFn: (args: TArgs) => Promise<TResult>,
  effectKey: keyof typeof journeyMutationEffects,
  successMessage: string,
  errorMessage: string,
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: () => {
      invalidateFor(queryClient, journeyMutationEffects[effectKey]());
      toast.success(successMessage);
    },
    onError: (err: unknown) => toast.error(getErrorMessage(err, errorMessage)),
  });
}

export function useCaptureJourney() {
  return useJourneyMutation(
    ({ capabilityId, request }: { capabilityId: string; request: CaptureJourneyRequest }) =>
      journeyApi.capture(capabilityId, request),
    'capture',
    'Journey captured',
    'Failed to capture journey',
  );
}

export function useStartJourney() {
  return useJourneyMutation(
    (journey: CapabilityJourney) => journeyApi.start(journey),
    'start',
    'Journey started',
    'Failed to start journey',
  );
}

export function useCompleteJourney() {
  return useJourneyMutation(
    (journey: CapabilityJourney) => journeyApi.complete(journey),
    'complete',
    'Journey completed',
    'Failed to complete journey',
  );
}

export function useAbandonJourney() {
  return useJourneyMutation(
    (journey: CapabilityJourney) => journeyApi.abandon(journey),
    'abandon',
    'Journey abandoned',
    'Failed to abandon journey',
  );
}

export function useUpdateJourneyDetails() {
  return useJourneyMutation(
    ({ journey, request }: { journey: CapabilityJourney; request: UpdateJourneyDetailsRequest }) =>
      journeyApi.updateDetails(journey, request),
    'updateDetails',
    'Journey details updated',
    'Failed to update journey details',
  );
}

export function useUpdateJourneyProgress() {
  return useJourneyMutation(
    ({ journey, request }: { journey: CapabilityJourney; request: UpdateJourneyProgressRequest }) =>
      journeyApi.updateProgress(journey, request),
    'updateProgress',
    'Progress updated',
    'Failed to update progress',
  );
}

export function useChangeJourneySourceApplications() {
  return useJourneyMutation(
    ({ journey, request }: { journey: CapabilityJourney; request: ChangeJourneySourceApplicationsRequest }) =>
      journeyApi.changeSourceApplications(journey, request),
    'changeSourceApplications',
    'Source applications updated',
    'Failed to update source applications',
  );
}

export function useAddJourneyMilestone() {
  return useJourneyMutation(
    ({ journey, request }: { journey: CapabilityJourney; request: AddJourneyMilestoneRequest }) =>
      journeyApi.addMilestone(journey, request),
    'addMilestone',
    'Milestone added',
    'Failed to add milestone',
  );
}

export function useUpdateJourneyMilestone() {
  return useJourneyMutation(
    ({ milestone, request }: { milestone: JourneyMilestone; request: UpdateJourneyMilestoneRequest }) =>
      journeyApi.updateMilestone(milestone, request),
    'updateMilestone',
    'Milestone updated',
    'Failed to update milestone',
  );
}

export function useRemoveJourneyMilestone() {
  return useJourneyMutation(
    (milestone: JourneyMilestone) => journeyApi.removeMilestone(milestone),
    'removeMilestone',
    'Milestone removed',
    'Failed to remove milestone',
  );
}
