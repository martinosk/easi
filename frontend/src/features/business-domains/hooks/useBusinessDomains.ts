import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { businessDomainsApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import { invalidateFor } from '../../../lib/invalidateFor';
import { mutationEffects } from '../../../lib/mutationEffects';
import type {
  BusinessDomain,
  BusinessDomainId,
  CreateBusinessDomainRequest,
  UpdateBusinessDomainRequest,
} from '../../../api/types';
import toast from 'react-hot-toast';

export interface UseBusinessDomainsResult {
  domains: BusinessDomain[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  createDomain: (name: string, description?: string) => Promise<BusinessDomain>;
  updateDomain: (domain: BusinessDomain, name: string, description?: string) => Promise<BusinessDomain>;
  deleteDomain: (domain: BusinessDomain) => Promise<void>;
}

export function useBusinessDomains(): UseBusinessDomainsResult {
  const query = useBusinessDomainsQuery();
  const createMutation = useCreateBusinessDomain();
  const updateMutation = useUpdateBusinessDomain();
  const deleteMutation = useDeleteBusinessDomain();

  const createDomain = useCallback(
    async (name: string, description?: string): Promise<BusinessDomain> => {
      return createMutation.mutateAsync({ name, description });
    },
    [createMutation]
  );

  const updateDomain = useCallback(
    async (domain: BusinessDomain, name: string, description?: string): Promise<BusinessDomain> => {
      return updateMutation.mutateAsync({ id: domain.id, request: { name, description } });
    },
    [updateMutation]
  );

  const deleteDomain = useCallback(
    async (domain: BusinessDomain): Promise<void> => {
      await deleteMutation.mutateAsync(domain.id);
    },
    [deleteMutation]
  );

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    domains: query.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    refetch,
    createDomain,
    updateDomain,
    deleteDomain,
  };
}

export function useBusinessDomainsQuery() {
  return useQuery({
    queryKey: queryKeys.businessDomains.lists(),
    queryFn: () => businessDomainsApi.getAll(),
  });
}

export function useBusinessDomain(id: BusinessDomainId | undefined) {
  return useQuery({
    queryKey: queryKeys.businessDomains.detail(id!),
    queryFn: () => businessDomainsApi.getById(id!),
    enabled: !!id,
  });
}

export function useDomainCapabilities(capabilitiesLink: string | undefined) {
  return useQuery({
    queryKey: queryKeys.businessDomains.capabilitiesByLink(capabilitiesLink!),
    queryFn: () => businessDomainsApi.getCapabilities(capabilitiesLink!),
    enabled: !!capabilitiesLink,
  });
}

export function useCapabilityRealizationsByDomain(
  domainId: BusinessDomainId | undefined,
  depth: number = 4
) {
  return useQuery({
    queryKey: queryKeys.businessDomains.realizations(domainId!, depth),
    queryFn: () => businessDomainsApi.getCapabilityRealizations(domainId!, depth),
    enabled: !!domainId,
  });
}

export function useCreateBusinessDomain() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateBusinessDomainRequest) =>
      businessDomainsApi.create(request),
    onSuccess: (newDomain) => {
      invalidateFor(queryClient, mutationEffects.businessDomains.create());
      toast.success(`Business domain "${newDomain.name}" created`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create business domain');
    },
  });
}

export function useUpdateBusinessDomain() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      request,
    }: {
      id: BusinessDomainId;
      request: UpdateBusinessDomainRequest;
    }) => businessDomainsApi.update(id, request),
    onSuccess: (updatedDomain) => {
      invalidateFor(queryClient, mutationEffects.businessDomains.update(updatedDomain.id));
      toast.success(`Business domain "${updatedDomain.name}" updated`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update business domain');
    },
  });
}

export function useDeleteBusinessDomain() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: BusinessDomainId) => businessDomainsApi.delete(id),
    onSuccess: (_, deletedId) => {
      invalidateFor(queryClient, mutationEffects.businessDomains.delete(deletedId));
      toast.success('Business domain deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete business domain');
    },
  });
}

