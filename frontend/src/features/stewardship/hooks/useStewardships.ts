import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import type { HATEOASLink } from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { stewardshipApi } from '../api/stewardshipApi';
import { stewardshipMutationEffects } from '../mutationEffects';
import { stewardshipQueryKeys } from '../queryKeys';
import type { DomainStewardships } from '../types';

export function useDomainStewardships(domainId: string, href: string | undefined) {
  return useQuery<DomainStewardships>({
    queryKey: stewardshipQueryKeys.domain(domainId),
    queryFn: () => stewardshipApi.getDomainStewardships(href!),
    enabled: !!href,
  });
}

interface AssignStewardInput {
  link: HATEOASLink;
  stewardId: string;
}

export function useAssignSteward(domainId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ link, stewardId }: AssignStewardInput) => stewardshipApi.assign(link, stewardId),
    onSuccess: () => {
      invalidateFor(queryClient, stewardshipMutationEffects.assign(domainId));
      toast.success('Steward assigned');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to assign steward');
    },
  });
}

export function useReleaseSteward(domainId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (link: HATEOASLink) => stewardshipApi.release(link),
    onSuccess: () => {
      invalidateFor(queryClient, stewardshipMutationEffects.release(domainId));
      toast.success('Steward released');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to release steward');
    },
  });
}
