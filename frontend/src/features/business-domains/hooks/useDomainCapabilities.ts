import { useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { businessDomainsApi } from '../api';
import { businessDomainsQueryKeys } from '../queryKeys';
import { businessDomainsMutationEffects } from '../mutationEffects';
import { invalidateFor } from '../../../lib/invalidateFor';
import type { Capability, CapabilityId, BusinessDomainId } from '../../../api/types';
import toast from 'react-hot-toast';

export interface UseDomainCapabilitiesResult {
  capabilities: Capability[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  associateCapability: (capabilityId: CapabilityId) => Promise<void>;
  dissociateCapability: (capability: Capability) => Promise<void>;
}

export function useDomainCapabilities(
  domainId: BusinessDomainId | undefined
): UseDomainCapabilitiesResult {
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: businessDomainsQueryKeys.capabilities(domainId!),
    queryFn: () => businessDomainsApi.getCapabilitiesByDomainId(domainId!),
    enabled: !!domainId,
  });

  const associateMutation = useMutation({
    mutationFn: (capabilityId: CapabilityId) =>
      businessDomainsApi.associateCapabilityByDomainId(domainId!, { capabilityId }),
    onSuccess: (_, capabilityId) => {
      invalidateFor(
        queryClient,
        businessDomainsMutationEffects.associateCapability(domainId!, capabilityId)
      );
      toast.success('Capability associated with domain');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to associate capability');
    },
  });

  const dissociateMutation = useMutation({
    mutationFn: (capabilityId: CapabilityId) =>
      businessDomainsApi.dissociateCapabilityByDomainId(domainId!, capabilityId),
    onSuccess: (_, capabilityId) => {
      invalidateFor(
        queryClient,
        businessDomainsMutationEffects.dissociateCapability(domainId!, capabilityId)
      );
      toast.success('Capability removed from domain');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to remove capability');
    },
  });

  const associateCapability = useCallback(
    async (capabilityId: CapabilityId) => {
      if (!domainId) {
        throw new Error('Domain ID not available');
      }
      await associateMutation.mutateAsync(capabilityId);
    },
    [domainId, associateMutation]
  );

  const dissociateCapability = useCallback(
    async (capability: Capability) => {
      if (!domainId) {
        throw new Error('Domain ID not available');
      }
      await dissociateMutation.mutateAsync(capability.id);
    },
    [domainId, dissociateMutation]
  );

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    capabilities: query.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    refetch,
    associateCapability,
    dissociateCapability,
  };
}
