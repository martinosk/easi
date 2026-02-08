import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { originEntitiesApi } from '../api/originEntitiesApi';
import { vendorsQueryKeys } from '../queryKeys';
import { invalidateFor } from '../../../lib/invalidateFor';
import { vendorsMutationEffects } from '../mutationEffects';
import type {
  Vendor,
  VendorId,
  CreateVendorRequest,
  UpdateVendorRequest,
  ComponentId,
} from '../../../api/types';
import toast from 'react-hot-toast';

export interface UseVendorsResult {
  vendors: Vendor[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  createVendor: (request: CreateVendorRequest) => Promise<Vendor>;
  updateVendor: (id: VendorId, request: UpdateVendorRequest) => Promise<Vendor>;
  deleteVendor: (id: VendorId, name: string) => Promise<void>;
}

export function useVendors(): UseVendorsResult {
  const query = useVendorsQuery();
  const createMutation = useCreateVendor();
  const updateMutation = useUpdateVendor();
  const deleteMutation = useDeleteVendor();

  const createVendor = useCallback(
    async (request: CreateVendorRequest): Promise<Vendor> => {
      return createMutation.mutateAsync(request);
    },
    [createMutation]
  );

  const updateVendor = useCallback(
    async (id: VendorId, request: UpdateVendorRequest): Promise<Vendor> => {
      return updateMutation.mutateAsync({ id, request });
    },
    [updateMutation]
  );

  const deleteVendor = useCallback(
    async (id: VendorId, name: string): Promise<void> => {
      await deleteMutation.mutateAsync({ id, name });
    },
    [deleteMutation]
  );

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    vendors: query.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    refetch,
    createVendor,
    updateVendor,
    deleteVendor,
  };
}

export function useVendorsQuery() {
  return useQuery({
    queryKey: vendorsQueryKeys.lists(),
    queryFn: () => originEntitiesApi.vendors.getAll(),
  });
}

export function useVendor(id: VendorId | undefined) {
  return useQuery({
    queryKey: vendorsQueryKeys.detail(id!),
    queryFn: () => originEntitiesApi.vendors.getById(id!),
    enabled: !!id,
  });
}

export function useCreateVendor() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateVendorRequest) =>
      originEntitiesApi.vendors.create(request),
    onSuccess: (newVendor) => {
      invalidateFor(queryClient, vendorsMutationEffects.create());
      toast.success(`Vendor "${newVendor.name}" created successfully`);
    },
    onError: () => {
      toast.error('Failed to create vendor');
    },
  });
}

export function useUpdateVendor() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, request }: { id: VendorId; request: UpdateVendorRequest }) =>
      originEntitiesApi.vendors.update(id, request),
    onSuccess: (updatedVendor, { id }) => {
      invalidateFor(queryClient, vendorsMutationEffects.update(id));
      toast.success(`Vendor "${updatedVendor.name}" updated`);
    },
    onError: () => {
      toast.error('Failed to update vendor');
    },
  });
}

export function useDeleteVendor() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id }: { id: VendorId; name: string }) =>
      originEntitiesApi.vendors.delete(id),
    onSuccess: (_, { id, name }) => {
      invalidateFor(queryClient, vendorsMutationEffects.delete(id));
      toast.success(`Vendor "${name}" deleted`);
    },
    onError: () => {
      toast.error('Failed to delete vendor');
    },
  });
}

export function useLinkComponentToVendor() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      componentId,
      vendorId,
      notes,
    }: {
      componentId: ComponentId;
      vendorId: VendorId;
      notes?: string;
    }) => originEntitiesApi.vendors.linkComponent(componentId, vendorId, notes),
    onSuccess: (_, { vendorId, componentId }) => {
      invalidateFor(queryClient, vendorsMutationEffects.linkComponent(vendorId, componentId));
      toast.success('Component linked to vendor');
    },
    onError: () => {
      toast.error('Failed to link component to vendor');
    },
  });
}

export function useUnlinkComponentFromVendor() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ componentId }: { vendorId: VendorId; componentId: ComponentId }) =>
      originEntitiesApi.vendors.unlinkComponent(componentId),
    onSuccess: (_, { vendorId, componentId }) => {
      invalidateFor(queryClient, vendorsMutationEffects.unlinkComponent(vendorId, componentId));
      toast.success('Component unlinked');
    },
    onError: () => {
      toast.error('Failed to unlink component');
    },
  });
}
