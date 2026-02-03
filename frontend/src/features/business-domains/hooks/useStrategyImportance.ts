import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { strategyImportanceApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import { invalidateFor } from '../../../lib/invalidateFor';
import { mutationEffects } from '../../../lib/mutationEffects';
import type {
  BusinessDomainId,
  CapabilityId,
  StrategyImportanceId,
  SetStrategyImportanceRequest,
  UpdateStrategyImportanceRequest,
  CollectionResponse,
  StrategyImportance,
} from '../../../api/types';
import toast from 'react-hot-toast';

export function useStrategyImportanceByDomainAndCapability(
  domainId: BusinessDomainId | undefined,
  capabilityId: CapabilityId | undefined
) {
  return useQuery<CollectionResponse<StrategyImportance>>({
    queryKey: queryKeys.strategyImportance.byDomainAndCapability(domainId!, capabilityId!),
    queryFn: () => strategyImportanceApi.getByDomainAndCapability(domainId!, capabilityId!),
    enabled: !!domainId && !!capabilityId,
  });
}

export function useStrategyImportanceByDomain(domainId: BusinessDomainId | undefined) {
  return useQuery({
    queryKey: queryKeys.strategyImportance.byDomain(domainId!),
    queryFn: () => strategyImportanceApi.getByDomain(domainId!),
    enabled: !!domainId,
  });
}

export function useStrategyImportanceByCapability(capabilityId: CapabilityId | undefined) {
  return useQuery({
    queryKey: queryKeys.strategyImportance.byCapability(capabilityId!),
    queryFn: () => strategyImportanceApi.getByCapability(capabilityId!),
    enabled: !!capabilityId,
  });
}

export function useSetStrategyImportance() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      domainId,
      capabilityId,
      request,
    }: {
      domainId: BusinessDomainId;
      capabilityId: CapabilityId;
      request: SetStrategyImportanceRequest;
    }) => strategyImportanceApi.setImportance(domainId, capabilityId, request),
    onSuccess: (newImportance) => {
      invalidateFor(queryClient, mutationEffects.strategyImportance.set(
        newImportance.businessDomainId,
        newImportance.capabilityId
      ));
      toast.success('Strategic importance set');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to set importance');
    },
  });
}

export function useUpdateStrategyImportance() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      domainId,
      capabilityId,
      importanceId,
      request,
    }: {
      domainId: BusinessDomainId;
      capabilityId: CapabilityId;
      importanceId: StrategyImportanceId;
      request: UpdateStrategyImportanceRequest;
    }) => strategyImportanceApi.updateImportance(domainId, capabilityId, importanceId, request),
    onSuccess: (updated) => {
      invalidateFor(queryClient, mutationEffects.strategyImportance.update(
        updated.businessDomainId,
        updated.capabilityId
      ));
      toast.success('Strategic importance updated');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update importance');
    },
  });
}

export function useRemoveStrategyImportance() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      domainId,
      capabilityId,
      importanceId,
    }: {
      domainId: BusinessDomainId;
      capabilityId: CapabilityId;
      importanceId: StrategyImportanceId;
    }) => strategyImportanceApi.removeImportance(domainId, capabilityId, importanceId),
    onSuccess: (_, { domainId, capabilityId }) => {
      invalidateFor(queryClient, mutationEffects.strategyImportance.remove(domainId, capabilityId));
      toast.success('Strategic importance removed');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to remove importance');
    },
  });
}
