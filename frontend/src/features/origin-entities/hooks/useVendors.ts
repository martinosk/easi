import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { originEntitiesApi } from '../api/originEntitiesApi';
import { vendorsQueryKeys } from '../queryKeys';
import { invalidateFor } from '../../../lib/invalidateFor';
import { vendorsMutationEffects } from '../mutationEffects';
import type {
  VendorId,
  CreateVendorRequest,
  UpdateVendorRequest,
  ComponentId,
} from '../../../api/types';
import toast from 'react-hot-toast';

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

interface MutationConfig<TArgs, TResult> {
  mutationFn: (args: TArgs) => Promise<TResult>;
  effects: (result: TResult, args: TArgs) => ReadonlyArray<readonly string[]>;
  successMessage: (result: TResult, args: TArgs) => string;
  errorMessage: string;
}

function useVendorMutation<TArgs, TResult>(config: MutationConfig<TArgs, TResult>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: config.mutationFn,
    onSuccess: (result, args) => {
      invalidateFor(queryClient, config.effects(result, args));
      toast.success(config.successMessage(result, args));
    },
    onError: () => toast.error(config.errorMessage),
  });
}

export function useCreateVendor() {
  return useVendorMutation({
    mutationFn: (request: CreateVendorRequest) => originEntitiesApi.vendors.create(request),
    effects: () => vendorsMutationEffects.create(),
    successMessage: (vendor) => `Vendor "${vendor.name}" created successfully`,
    errorMessage: 'Failed to create vendor',
  });
}

export function useUpdateVendor() {
  return useVendorMutation({
    mutationFn: ({ id, request }: { id: VendorId; request: UpdateVendorRequest }) =>
      originEntitiesApi.vendors.update(id, request),
    effects: (_, { id }) => vendorsMutationEffects.update(id),
    successMessage: (vendor) => `Vendor "${vendor.name}" updated`,
    errorMessage: 'Failed to update vendor',
  });
}

export function useDeleteVendor() {
  return useVendorMutation({
    mutationFn: ({ id }: { id: VendorId; name: string }) => originEntitiesApi.vendors.delete(id),
    effects: (_, { id }) => vendorsMutationEffects.delete(id),
    successMessage: (_, { name }) => `Vendor "${name}" deleted`,
    errorMessage: 'Failed to delete vendor',
  });
}

export function useLinkComponentToVendor() {
  return useVendorMutation({
    mutationFn: ({ componentId, vendorId, notes }: { componentId: ComponentId; vendorId: VendorId; notes?: string }) =>
      originEntitiesApi.vendors.linkComponent(componentId, vendorId, notes),
    effects: (_, { vendorId, componentId }) => vendorsMutationEffects.linkComponent(vendorId, componentId),
    successMessage: () => 'Component linked to vendor',
    errorMessage: 'Failed to link component to vendor',
  });
}

export function useUnlinkComponentFromVendor() {
  return useVendorMutation({
    mutationFn: ({ componentId }: { vendorId: VendorId; componentId: ComponentId }) =>
      originEntitiesApi.vendors.unlinkComponent(componentId),
    effects: (_, { vendorId, componentId }) => vendorsMutationEffects.unlinkComponent(vendorId, componentId),
    successMessage: () => 'Component unlinked',
    errorMessage: 'Failed to unlink component',
  });
}
