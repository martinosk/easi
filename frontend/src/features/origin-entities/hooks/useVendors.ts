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

function useVendorMutation<TArgs, TResult>(
  mutationFn: (args: TArgs) => Promise<TResult>,
  onMutationSuccess: (queryClient: ReturnType<typeof useQueryClient>, result: TResult, args: TArgs) => void,
  errorMessage: string
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (result, args) => onMutationSuccess(queryClient, result, args),
    onError: () => toast.error(errorMessage),
  });
}

export function useCreateVendor() {
  return useVendorMutation(
    (request: CreateVendorRequest) => originEntitiesApi.vendors.create(request),
    (qc, newVendor) => {
      invalidateFor(qc, vendorsMutationEffects.create());
      toast.success(`Vendor "${newVendor.name}" created successfully`);
    },
    'Failed to create vendor'
  );
}

export function useUpdateVendor() {
  return useVendorMutation(
    ({ id, request }: { id: VendorId; request: UpdateVendorRequest }) =>
      originEntitiesApi.vendors.update(id, request),
    (qc, updatedVendor, { id }) => {
      invalidateFor(qc, vendorsMutationEffects.update(id));
      toast.success(`Vendor "${updatedVendor.name}" updated`);
    },
    'Failed to update vendor'
  );
}

export function useDeleteVendor() {
  return useVendorMutation(
    ({ id }: { id: VendorId; name: string }) =>
      originEntitiesApi.vendors.delete(id),
    (qc, _, { id, name }) => {
      invalidateFor(qc, vendorsMutationEffects.delete(id));
      toast.success(`Vendor "${name}" deleted`);
    },
    'Failed to delete vendor'
  );
}

export function useLinkComponentToVendor() {
  return useVendorMutation(
    ({ componentId, vendorId, notes }: { componentId: ComponentId; vendorId: VendorId; notes?: string }) =>
      originEntitiesApi.vendors.linkComponent(componentId, vendorId, notes),
    (qc, _, { vendorId, componentId }) => {
      invalidateFor(qc, vendorsMutationEffects.linkComponent(vendorId, componentId));
      toast.success('Component linked to vendor');
    },
    'Failed to link component to vendor'
  );
}

export function useUnlinkComponentFromVendor() {
  return useVendorMutation(
    ({ componentId }: { vendorId: VendorId; componentId: ComponentId }) =>
      originEntitiesApi.vendors.unlinkComponent(componentId),
    (qc, _, { vendorId, componentId }) => {
      invalidateFor(qc, vendorsMutationEffects.unlinkComponent(vendorId, componentId));
      toast.success('Component unlinked');
    },
    'Failed to unlink component'
  );
}
