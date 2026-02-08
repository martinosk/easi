import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { businessDomainsApi } from '../api';
import { businessDomainsQueryKeys } from '../queryKeys';
import { businessDomainsMutationEffects } from '../mutationEffects';
import { invalidateFor } from '../../../lib/invalidateFor';
import type {
  BusinessDomain,
  BusinessDomainId,
  CreateBusinessDomainRequest,
  UpdateBusinessDomainRequest,
  BusinessDomainsResponse,
  HATEOASLinks,
} from '../../../api/types';
import toast from 'react-hot-toast';

export interface UseBusinessDomainsResult {
  domains: BusinessDomain[];
  collectionLinks: HATEOASLinks | undefined;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  createDomain: (name: string, description?: string, domainArchitectId?: string) => Promise<BusinessDomain>;
  updateDomain: (domain: BusinessDomain, name: string, description?: string, domainArchitectId?: string) => Promise<BusinessDomain>;
  deleteDomain: (domain: BusinessDomain) => Promise<void>;
}

export function useBusinessDomains(): UseBusinessDomainsResult {
  const query = useBusinessDomainsQuery();
  const createMutation = useCreateBusinessDomain();
  const updateMutation = useUpdateBusinessDomain();
  const deleteMutation = useDeleteBusinessDomain();

  const createDomain = useCallback(
    async (name: string, description?: string, domainArchitectId?: string): Promise<BusinessDomain> => {
      return createMutation.mutateAsync({ name, description, domainArchitectId });
    },
    [createMutation]
  );

  const updateDomain = useCallback(
    async (domain: BusinessDomain, name: string, description?: string, domainArchitectId?: string): Promise<BusinessDomain> => {
      return updateMutation.mutateAsync({ domain, request: { name, description, domainArchitectId } });
    },
    [updateMutation]
  );

  const deleteDomain = useCallback(
    async (domain: BusinessDomain): Promise<void> => {
      await deleteMutation.mutateAsync(domain);
    },
    [deleteMutation]
  );

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    domains: query.data?.data ?? [],
    collectionLinks: query.data?._links,
    isLoading: query.isLoading,
    error: query.error,
    refetch,
    createDomain,
    updateDomain,
    deleteDomain,
  };
}

export function useBusinessDomainsQuery() {
  return useQuery<BusinessDomainsResponse>({
    queryKey: businessDomainsQueryKeys.lists(),
    queryFn: () => businessDomainsApi.getAll(),
  });
}

export function useBusinessDomain(id: BusinessDomainId | undefined) {
  return useQuery({
    queryKey: businessDomainsQueryKeys.detail(id!),
    queryFn: () => businessDomainsApi.getById(id!),
    enabled: !!id,
  });
}

export function useDomainCapabilities(capabilitiesLink: string | undefined) {
  return useQuery({
    queryKey: businessDomainsQueryKeys.capabilitiesByLink(capabilitiesLink!),
    queryFn: () => businessDomainsApi.getCapabilities(capabilitiesLink!),
    enabled: !!capabilitiesLink,
  });
}

export function useCapabilityRealizationsByDomain(
  domainId: BusinessDomainId | undefined,
  depth: number = 4
) {
  return useQuery({
    queryKey: businessDomainsQueryKeys.realizations(domainId!, depth),
    queryFn: () => businessDomainsApi.getCapabilityRealizations(domainId!, depth),
    enabled: !!domainId,
  });
}

function useDomainMutation<TArgs, TResult>(
  mutationFn: (args: TArgs) => Promise<TResult>,
  onMutationSuccess: (queryClient: ReturnType<typeof useQueryClient>, result: TResult, args: TArgs) => void,
  errorMessage: string
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (result, args) => onMutationSuccess(queryClient, result, args),
    onError: (error: Error) => toast.error(error.message || errorMessage),
  });
}

export function useCreateBusinessDomain() {
  return useDomainMutation(
    (request: CreateBusinessDomainRequest) => businessDomainsApi.create(request),
    (qc, newDomain) => {
      invalidateFor(qc, businessDomainsMutationEffects.create());
      toast.success(`Business domain "${newDomain.name}" created`);
    },
    'Failed to create business domain'
  );
}

export function useUpdateBusinessDomain() {
  return useDomainMutation(
    ({ domain, request }: { domain: BusinessDomain; request: UpdateBusinessDomainRequest }) =>
      businessDomainsApi.update(domain, request),
    (qc, updatedDomain) => {
      invalidateFor(qc, businessDomainsMutationEffects.update(updatedDomain.id));
      toast.success(`Business domain "${updatedDomain.name}" updated`);
    },
    'Failed to update business domain'
  );
}

export function useDeleteBusinessDomain() {
  return useDomainMutation(
    (domain: BusinessDomain) => businessDomainsApi.delete(domain),
    (qc, _, deletedDomain) => {
      invalidateFor(qc, businessDomainsMutationEffects.delete(deletedDomain.id));
      toast.success('Business domain deleted');
    },
    'Failed to delete business domain'
  );
}

