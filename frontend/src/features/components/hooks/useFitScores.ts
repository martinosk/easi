import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fitScoresApi } from '../api';
import { fitScoresQueryKeys, fitComparisonsQueryKeys } from '../queryKeys';
import { invalidateFor } from '../../../lib/invalidateFor';
import { fitScoresMutationEffects } from '../mutationEffects';
import type {
  ComponentId,
  CapabilityId,
  BusinessDomainId,
  SetApplicationFitScoreRequest,
  ApplicationFitScoresResponse,
  ApiError,
} from '../../../api/types';
import toast from 'react-hot-toast';

const fitScoreErrorMessages: Record<number, string> = {
  400: 'Invalid input. Please check the score value.',
  403: 'You do not have permission to modify fit scores.',
  404: 'Strategy pillar not found.',
  409: 'Fit scoring is not enabled for this pillar.',
  429: 'Too many requests. Please wait a moment and try again.',
};

function getFitScoreErrorMessage(error: unknown, defaultMessage: string): string {
  if (error instanceof Error && 'statusCode' in error) {
    const apiError = error as ApiError;
    return fitScoreErrorMessages[apiError.statusCode] ?? apiError.message ?? defaultMessage;
  }
  return error instanceof Error ? error.message : defaultMessage;
}

export function useComponentFitScores(componentId: ComponentId | undefined) {
  return useQuery<ApplicationFitScoresResponse>({
    queryKey: fitScoresQueryKeys.byComponent(componentId!),
    queryFn: () => fitScoresApi.getByComponent(componentId!),
    enabled: !!componentId,
  });
}

export function useSetFitScore() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      componentId,
      pillarId,
      request,
    }: {
      componentId: ComponentId;
      pillarId: string;
      request: SetApplicationFitScoreRequest;
    }) => fitScoresApi.setScore(componentId, pillarId, request),
    onSuccess: (_, { componentId }) => {
      invalidateFor(queryClient, fitScoresMutationEffects.set(componentId));
      toast.success('Fit score saved');
    },
    onError: (error: unknown) => {
      toast.error(getFitScoreErrorMessage(error, 'Failed to save fit score'));
    },
  });
}

export function useDeleteFitScore() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      componentId,
      pillarId,
    }: {
      componentId: ComponentId;
      pillarId: string;
    }) => fitScoresApi.deleteScore(componentId, pillarId),
    onSuccess: (_, { componentId }) => {
      invalidateFor(queryClient, fitScoresMutationEffects.delete(componentId));
      toast.success('Fit score removed');
    },
    onError: (error: unknown) => {
      toast.error(getFitScoreErrorMessage(error, 'Failed to remove fit score'));
    },
  });
}

export function useFitComparisons(
  componentId: ComponentId | undefined,
  capabilityId: CapabilityId | undefined,
  businessDomainId: BusinessDomainId | undefined
) {
  return useQuery({
    queryKey: fitComparisonsQueryKeys.byContext(
      componentId!,
      capabilityId!,
      businessDomainId!
    ),
    queryFn: () => fitScoresApi.getFitComparisons(componentId!, capabilityId!, businessDomainId!),
    enabled: !!componentId && !!capabilityId && !!businessDomainId,
  });
}
