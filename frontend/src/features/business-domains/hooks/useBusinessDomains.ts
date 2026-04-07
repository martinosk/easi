import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import toast from 'react-hot-toast';
import type {
  BusinessDomain,
  BusinessDomainId,
  BusinessDomainsResponse,
  CreateBusinessDomainRequest,
  HATEOASLinks,
  UpdateBusinessDomainRequest,
} from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { businessDomainsApi } from '../api';
import { businessDomainsMutationEffects } from '../mutationEffects';
import { businessDomainsQueryKeys } from '../queryKeys';

export interface UseBusinessDomainsResult {
  domains: BusinessDomain[];
  collectionLinks: HATEOASLinks | undefined;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  createDomain: (name: string, description?: string, domainArchitectId?: string) => Promise<BusinessDomain>;
  updateDomain: (
    domain: BusinessDomain,
    name: string,
    description?: string,
    domainArchitectId?: string,
  ) => Promise<BusinessDomain>;
  deleteDomain: (domain: BusinessDomain) => Promise<void>;
}

export function useBusinessDomains(): UseBusinessDomainsResult {
  const query = useBusinessDomainsQuery();
  const createMutation = useCreateBusinessDomain();
  const updateMutation = useUpdateBusinessDomain();
  const deleteMutation = useDeleteBusinessDomain();

  const createDomain = useCallback(
    (name: string, description?: string, domainArchitectId?: string) =>
      createMutation.mutateAsync({ name, description, domainArchitectId }),
    [createMutation],
  );

  const updateDomain = useCallback(
    (domain: BusinessDomain, name: string, description?: string, domainArchitectId?: string) =>
      updateMutation.mutateAsync({ domain, request: { name, description, domainArchitectId } }),
    [updateMutation],
  );

  const deleteDomain = useCallback(
    async (domain: BusinessDomain) => {
      await deleteMutation.mutateAsync(domain);
    },
    [deleteMutation],
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

export function useCapabilityRealizationsByDomain(domainId: BusinessDomainId | undefined, depth: number = 4) {
  return useQuery({
    queryKey: businessDomainsQueryKeys.realizations(domainId!, depth),
    queryFn: () => businessDomainsApi.getCapabilityRealizations(domainId!, depth),
    enabled: !!domainId,
  });
}

interface MutationConfig<TArgs, TResult> {
  mutationFn: (args: TArgs) => Promise<TResult>;
  effects: (result: TResult, args: TArgs) => ReadonlyArray<readonly unknown[]>;
  successMessage: (result: TResult, args: TArgs) => string;
  errorMessage: string;
}

function useDomainMutation<TArgs, TResult>(config: MutationConfig<TArgs, TResult>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: config.mutationFn,
    onSuccess: (result, args) => {
      invalidateFor(queryClient, config.effects(result, args));
      toast.success(config.successMessage(result, args));
    },
    onError: (error: Error) => toast.error(error.message || config.errorMessage),
  });
}

export function useCreateBusinessDomain() {
  return useDomainMutation({
    mutationFn: (request: CreateBusinessDomainRequest) => businessDomainsApi.create(request),
    effects: () => businessDomainsMutationEffects.create(),
    successMessage: (domain) => `Business domain "${domain.name}" created`,
    errorMessage: 'Failed to create business domain',
  });
}

export function useUpdateBusinessDomain() {
  return useDomainMutation({
    mutationFn: ({ domain, request }: { domain: BusinessDomain; request: UpdateBusinessDomainRequest }) =>
      businessDomainsApi.update(domain, request),
    effects: (updatedDomain) => businessDomainsMutationEffects.update(updatedDomain.id),
    successMessage: (domain) => `Business domain "${domain.name}" updated`,
    errorMessage: 'Failed to update business domain',
  });
}

export function useDeleteBusinessDomain() {
  return useDomainMutation({
    mutationFn: (domain: BusinessDomain) => businessDomainsApi.delete(domain),
    effects: (_, domain) => businessDomainsMutationEffects.delete(domain.id),
    successMessage: () => 'Business domain deleted',
    errorMessage: 'Failed to delete business domain',
  });
}
