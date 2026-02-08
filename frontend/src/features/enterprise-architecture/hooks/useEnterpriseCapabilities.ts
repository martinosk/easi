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

function useEnterpriseMutation<TArgs, TResult>(
  mutationFn: (args: TArgs) => Promise<TResult>,
  onMutationSuccess: (queryClient: ReturnType<typeof useQueryClient>, result: TResult, args: TArgs) => void,
  errorMessage: string
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (result, args) => onMutationSuccess(queryClient, result, args),
    onError: (error: unknown) => toast.error(getErrorMessage(error, errorMessage)),
  });
}

export function useCreateEnterpriseCapability() {
  return useEnterpriseMutation(
    (request: CreateEnterpriseCapabilityRequest) => enterpriseArchApi.create(request),
    (qc, newCapability) => {
      invalidateFor(qc, enterpriseCapabilitiesMutationEffects.create());
      toast.success(`Enterprise capability "${newCapability.name}" created successfully`);
    },
    'Failed to create enterprise capability'
  );
}

export function useDeleteEnterpriseCapability() {
  return useEnterpriseMutation(
    ({ id }: { id: EnterpriseCapabilityId; name: string }) => enterpriseArchApi.delete(id),
    (qc, _, { id, name }) => {
      invalidateFor(qc, enterpriseCapabilitiesMutationEffects.delete(id));
      toast.success(`Enterprise capability "${name}" deleted`);
    },
    'Failed to delete capability'
  );
}

export function useEnterpriseCapabilityLinks(enterpriseCapabilityId: EnterpriseCapabilityId | undefined) {
  return useQuery({
    queryKey: enterpriseCapabilitiesQueryKeys.links(enterpriseCapabilityId!),
    queryFn: () => enterpriseArchApi.getLinks(enterpriseCapabilityId!),
    enabled: !!enterpriseCapabilityId,
  });
}

export function useLinkDomainCapability() {
  return useEnterpriseMutation(
    ({ enterpriseCapabilityId, request }: { enterpriseCapabilityId: EnterpriseCapabilityId; request: LinkCapabilityRequest }) =>
      enterpriseArchApi.linkDomainCapability(enterpriseCapabilityId, request),
    (qc, _, { enterpriseCapabilityId }) => {
      invalidateFor(qc, enterpriseCapabilitiesMutationEffects.link(enterpriseCapabilityId));
      toast.success('Capability linked successfully');
    },
    'Failed to link capability'
  );
}

export function useUnlinkDomainCapability() {
  return useEnterpriseMutation(
    ({ enterpriseCapabilityId, linkId }: { enterpriseCapabilityId: EnterpriseCapabilityId; linkId: EnterpriseCapabilityLinkId }) =>
      enterpriseArchApi.unlinkDomainCapability(enterpriseCapabilityId, linkId),
    (qc, _, { enterpriseCapabilityId }) => {
      invalidateFor(qc, enterpriseCapabilitiesMutationEffects.unlink(enterpriseCapabilityId));
      toast.success('Capability unlinked successfully');
    },
    'Failed to unlink capability'
  );
}
