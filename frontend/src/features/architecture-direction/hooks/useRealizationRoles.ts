import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { invalidateFor } from '../../../lib/invalidateFor';
import { realizationRoleApi } from '../api/realizationRoleApi';
import { realizationRoleMutationEffects } from '../mutationEffects';
import { realizationRoleQueryKeys } from '../queryKeys';
import type { AssignRealizationRoleRequest, RealizationRoleAssignment } from '../types';

function getErrorMessage(err: unknown, fallback: string): string {
  return err instanceof Error ? err.message : fallback;
}

export function useRealizationRolesByCapabilityIds(capabilityIds: string[]) {
  return useQuery({
    queryKey: realizationRoleQueryKeys.byCapabilityIds(capabilityIds),
    queryFn: () => realizationRoleApi.getByCapabilityIds(capabilityIds),
    enabled: capabilityIds.length > 0,
  });
}

interface AssignArgs {
  capabilityId: string;
  componentId: string;
  request: AssignRealizationRoleRequest;
}

export function useAssignRealizationRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ capabilityId, componentId, request }: AssignArgs) =>
      realizationRoleApi.assign(capabilityId, componentId, request),
    onSuccess: () => {
      invalidateFor(queryClient, realizationRoleMutationEffects.assign());
      toast.success('Role saved');
    },
    onError: (err) => toast.error(getErrorMessage(err, 'Failed to save role')),
  });
}

interface ClearArgs {
  role: RealizationRoleAssignment;
}

export function useClearRealizationRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ role }: ClearArgs) => realizationRoleApi.clear(role),
    onSuccess: () => {
      invalidateFor(queryClient, realizationRoleMutationEffects.clear());
      toast.success('Role cleared');
    },
    onError: (err) => toast.error(getErrorMessage(err, 'Failed to clear role')),
  });
}
