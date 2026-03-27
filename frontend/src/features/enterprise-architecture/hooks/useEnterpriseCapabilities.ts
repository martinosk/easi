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
    (request: CreateEnterpriseCapabilityRequest) => createMutation.mutateAsync(request),
    [createMutation]
  );

  const deleteCapability = useCallback(
    async (id: EnterpriseCapabilityId, name: string) => { await deleteMutation.mutateAsync({ id, name }); },
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

interface MutationConfig<TArgs, TResult> {
  mutationFn: (args: TArgs) => Promise<TResult>;
  effects: (result: TResult, args: TArgs) => ReadonlyArray<readonly unknown[]>;
  successMessage: (result: TResult, args: TArgs) => string;
  errorMessage: string;
}

function useEnterpriseMutation<TArgs, TResult>(config: MutationConfig<TArgs, TResult>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: config.mutationFn,
    onSuccess: (result, args) => {
      invalidateFor(queryClient, config.effects(result, args));
      toast.success(config.successMessage(result, args));
    },
    onError: (error: unknown) => toast.error(getErrorMessage(error, config.errorMessage)),
  });
}

export function useCreateEnterpriseCapability() {
  return useEnterpriseMutation({
    mutationFn: (request: CreateEnterpriseCapabilityRequest) => enterpriseArchApi.create(request),
    effects: () => enterpriseCapabilitiesMutationEffects.create(),
    successMessage: (cap) => `Enterprise capability "${cap.name}" created successfully`,
    errorMessage: 'Failed to create enterprise capability',
  });
}

export function useDeleteEnterpriseCapability() {
  return useEnterpriseMutation({
    mutationFn: ({ id }: { id: EnterpriseCapabilityId; name: string }) => enterpriseArchApi.delete(id),
    effects: (_, { id }) => enterpriseCapabilitiesMutationEffects.delete(id),
    successMessage: (_, { name }) => `Enterprise capability "${name}" deleted`,
    errorMessage: 'Failed to delete capability',
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
  return useEnterpriseMutation({
    mutationFn: ({ enterpriseCapabilityId, request }: { enterpriseCapabilityId: EnterpriseCapabilityId; request: LinkCapabilityRequest }) =>
      enterpriseArchApi.linkDomainCapability(enterpriseCapabilityId, request),
    effects: (_, { enterpriseCapabilityId }) => enterpriseCapabilitiesMutationEffects.link(enterpriseCapabilityId),
    successMessage: () => 'Capability linked successfully',
    errorMessage: 'Failed to link capability',
  });
}

export function useUnlinkDomainCapability() {
  return useEnterpriseMutation({
    mutationFn: ({ enterpriseCapabilityId, linkId }: { enterpriseCapabilityId: EnterpriseCapabilityId; linkId: EnterpriseCapabilityLinkId }) =>
      enterpriseArchApi.unlinkDomainCapability(enterpriseCapabilityId, linkId),
    effects: (_, { enterpriseCapabilityId }) => enterpriseCapabilitiesMutationEffects.unlink(enterpriseCapabilityId),
    successMessage: () => 'Capability unlinked successfully',
    errorMessage: 'Failed to unlink capability',
  });
}
