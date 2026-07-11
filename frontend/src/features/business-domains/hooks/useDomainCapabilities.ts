import { useMutation, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import type { BusinessDomainId, Capability, CapabilityId } from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { businessDomainsApi } from '../api';
import { businessDomainsMutationEffects } from '../mutationEffects';

export function useAssociateCapability() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ domainId, capabilityId }: { domainId: BusinessDomainId; capabilityId: CapabilityId }) =>
      businessDomainsApi.associateCapabilityByDomainId(domainId, { capabilityId }),
    onSuccess: (_, { domainId, capabilityId }) => {
      invalidateFor(queryClient, businessDomainsMutationEffects.associateCapability(domainId, capabilityId));
      toast.success('Capability associated with domain');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to associate capability');
    },
  });
}

export function useDissociateCapability() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ domainId, capability }: { domainId: BusinessDomainId; capability: Capability }) =>
      businessDomainsApi.dissociateCapabilityByDomainId(domainId, capability.id),
    onSuccess: (_, { domainId, capability }) => {
      invalidateFor(queryClient, businessDomainsMutationEffects.dissociateCapability(domainId, capability.id));
      toast.success('Capability removed from domain');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to remove capability');
    },
  });
}
