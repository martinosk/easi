import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { strategyImportanceApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import type {
  BusinessDomainId,
  CapabilityId,
  StrategyImportance,
  StrategyImportanceId,
  SetStrategyImportanceRequest,
  UpdateStrategyImportanceRequest,
} from '../../../api/types';
import toast from 'react-hot-toast';

export function useStrategyImportanceByDomainAndCapability(
  domainId: BusinessDomainId | undefined,
  capabilityId: CapabilityId | undefined
) {
  return useQuery({
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
      queryClient.setQueryData<StrategyImportance[]>(
        queryKeys.strategyImportance.byDomainAndCapability(
          newImportance.businessDomainId,
          newImportance.capabilityId
        ),
        (old) => (old ? [...old, newImportance] : [newImportance])
      );
      queryClient.invalidateQueries({
        queryKey: queryKeys.strategyImportance.byDomain(newImportance.businessDomainId),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.strategyImportance.byCapability(newImportance.capabilityId),
      });
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
      queryClient.setQueryData<StrategyImportance[]>(
        queryKeys.strategyImportance.byDomainAndCapability(
          updated.businessDomainId,
          updated.capabilityId
        ),
        (old) => old?.map((i) => (i.id === updated.id ? updated : i)) ?? []
      );
      queryClient.invalidateQueries({
        queryKey: queryKeys.strategyImportance.byDomain(updated.businessDomainId),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.strategyImportance.byCapability(updated.capabilityId),
      });
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
    onSuccess: (_, { domainId, capabilityId, importanceId }) => {
      queryClient.setQueryData<StrategyImportance[]>(
        queryKeys.strategyImportance.byDomainAndCapability(domainId, capabilityId),
        (old) => old?.filter((i) => i.id !== importanceId) ?? []
      );
      queryClient.invalidateQueries({
        queryKey: queryKeys.strategyImportance.byDomain(domainId),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.strategyImportance.byCapability(capabilityId),
      });
      toast.success('Strategic importance removed');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to remove importance');
    },
  });
}
