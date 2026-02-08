import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { enterpriseCapabilitiesQueryKeys } from '../queryKeys';
import { enterpriseCapabilitiesMutationEffects } from '../mutationEffects';
import { invalidateFor } from '../../../lib/invalidateFor';
import { getErrorMessage } from '../utils/errorMessages';
import type {
  EnterpriseCapability,
  EnterpriseCapabilityId,
  EnterpriseCapabilityLinkId,
  CreateEnterpriseCapabilityRequest,
  LinkCapabilityRequest,
} from '../types';
import toast from 'react-hot-toast';

export interface UseEnterpriseCapabilitiesResult {
  capabilities: EnterpriseCapability[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  createCapability: (request: CreateEnterpriseCapabilityRequest) => Promise<EnterpriseCapability>;
  deleteCapability: (id: EnterpriseCapabilityId, name: string) => Promise<void>;
}

export function useEnterpriseCapabilities(): UseEnterpriseCapabilitiesResult {
  const query = useEnterpriseCapabilitiesQuery();
  const createMutation = useCreateEnterpriseCapability();
  const deleteMutation = useDeleteEnterpriseCapability();

  const createCapability = useCallback(
    async (request: CreateEnterpriseCapabilityRequest): Promise<EnterpriseCapability> => {
      return createMutation.mutateAsync(request);
    },
    [createMutation]
  );

  const deleteCapability = useCallback(
    async (id: EnterpriseCapabilityId, name: string): Promise<void> => {
      await deleteMutation.mutateAsync({ id, name });
    },
    [deleteMutation]
  );

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    capabilities: query.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    refetch,
    createCapability,
    deleteCapability,
  };
}

export function useEnterpriseCapabilitiesQuery() {
  return useQuery({
    queryKey: enterpriseCapabilitiesQueryKeys.lists(),
    queryFn: () => enterpriseArchApi.getAll(),
  });
}

export function useEnterpriseCapability(id: EnterpriseCapabilityId | undefined) {
  return useQuery({
    queryKey: enterpriseCapabilitiesQueryKeys.detail(id!),
    queryFn: () => enterpriseArchApi.getById(id!),
    enabled: !!id,
  });
}

export function useCreateEnterpriseCapability() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateEnterpriseCapabilityRequest) =>
      enterpriseArchApi.create(request),
    onSuccess: (newCapability) => {
      invalidateFor(queryClient, enterpriseCapabilitiesMutationEffects.create());
      toast.success(`Enterprise capability "${newCapability.name}" created successfully`);
    },
    onError: (error: unknown) => {
      toast.error(getErrorMessage(error, 'Failed to create enterprise capability'));
    },
  });
}

export function useDeleteEnterpriseCapability() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id }: { id: EnterpriseCapabilityId; name: string }) =>
      enterpriseArchApi.delete(id),
    onSuccess: (_, { id, name }) => {
      invalidateFor(queryClient, enterpriseCapabilitiesMutationEffects.delete(id));
      toast.success(`Enterprise capability "${name}" deleted`);
    },
    onError: (error: unknown) => {
      toast.error(getErrorMessage(error, 'Failed to delete capability'));
    },
  });
}

export function useEnterpriseCapabilityLinks(enterpriseCapabilityId: EnterpriseCapabilityId | undefined) {
  return useQuery({
    queryKey: enterpriseCapabilitiesQueryKeys.links(enterpriseCapabilityId!),
    queryFn: () => enterpriseArchApi.getLinks(enterpriseCapabilityId!),
    enabled: !!enterpriseCapabilityId,
  });
}

export function useLinkDomainCapability() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      enterpriseCapabilityId,
      request,
    }: {
      enterpriseCapabilityId: EnterpriseCapabilityId;
      request: LinkCapabilityRequest;
    }) => enterpriseArchApi.linkDomainCapability(enterpriseCapabilityId, request),
    onSuccess: (_, { enterpriseCapabilityId }) => {
      invalidateFor(queryClient, enterpriseCapabilitiesMutationEffects.link(enterpriseCapabilityId));
      toast.success('Capability linked successfully');
    },
    onError: (error: unknown) => {
      toast.error(getErrorMessage(error, 'Failed to link capability'));
    },
  });
}

export function useUnlinkDomainCapability() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      enterpriseCapabilityId,
      linkId,
    }: {
      enterpriseCapabilityId: EnterpriseCapabilityId;
      linkId: EnterpriseCapabilityLinkId;
    }) => enterpriseArchApi.unlinkDomainCapability(enterpriseCapabilityId, linkId),
    onSuccess: (_, { enterpriseCapabilityId }) => {
      invalidateFor(queryClient, enterpriseCapabilitiesMutationEffects.unlink(enterpriseCapabilityId));
      toast.success('Capability unlinked successfully');
    },
    onError: (error: unknown) => {
      toast.error(getErrorMessage(error, 'Failed to unlink capability'));
    },
  });
}
