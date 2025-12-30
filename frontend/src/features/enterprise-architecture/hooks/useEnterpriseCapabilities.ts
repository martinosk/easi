import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { queryKeys } from '../../../lib/queryClient';
import { getErrorMessage } from '../utils/errorMessages';
import type {
  EnterpriseCapability,
  EnterpriseCapabilityId,
  CreateEnterpriseCapabilityRequest,
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
    queryKey: queryKeys.enterpriseCapabilities.lists(),
    queryFn: () => enterpriseArchApi.getAll(),
  });
}

export function useEnterpriseCapability(id: EnterpriseCapabilityId | undefined) {
  return useQuery({
    queryKey: queryKeys.enterpriseCapabilities.detail(id!),
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
      queryClient.setQueryData<EnterpriseCapability[]>(
        queryKeys.enterpriseCapabilities.lists(),
        (old) => (old ? [...old, newCapability] : [newCapability])
      );
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
      queryClient.setQueryData<EnterpriseCapability[]>(
        queryKeys.enterpriseCapabilities.lists(),
        (old) => old?.filter((c) => c.id !== id) ?? []
      );
      queryClient.removeQueries({
        queryKey: queryKeys.enterpriseCapabilities.detail(id),
      });
      toast.success(`Enterprise capability "${name}" deleted`);
    },
    onError: (error: unknown) => {
      toast.error(getErrorMessage(error, 'Failed to delete capability'));
    },
  });
}
