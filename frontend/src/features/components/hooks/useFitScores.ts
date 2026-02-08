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

function getFitScoreErrorMessage(error: unknown, defaultMessage: string): string {
  if (error instanceof Error && 'statusCode' in error) {
    const apiError = error as ApiError;
    switch (apiError.statusCode) {
      case 400:
        return apiError.message || 'Invalid input. Please check the score value.';
      case 403:
        return 'You do not have permission to modify fit scores.';
      case 404:
        return 'Strategy pillar not found.';
      case 409:
        return 'Fit scoring is not enabled for this pillar.';
      case 429:
        return 'Too many requests. Please wait a moment and try again.';
      default:
        return apiError.message || defaultMessage;
    }
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
