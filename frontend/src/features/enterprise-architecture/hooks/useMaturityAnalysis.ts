import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { queryKeys } from '../../../lib/queryClient';
import { invalidateFor } from '../../../lib/invalidateFor';
import { mutationEffects } from '../../../lib/mutationEffects';
import { getErrorMessage } from '../utils/errorMessages';
import type {
  EnterpriseCapabilityId,
  MaturityAnalysisCandidate,
  MaturityAnalysisSummary,
  MaturityGapDetail,
  UnlinkedCapability,
} from '../types';
import toast from 'react-hot-toast';

export function useMaturityAnalysisCandidates(sortBy: string = 'gap') {
  return useQuery({
    queryKey: queryKeys.maturityAnalysis.candidates(sortBy),
    queryFn: () => enterpriseArchApi.getMaturityAnalysisCandidates(sortBy),
  });
}

export interface UseMaturityAnalysisResult {
  candidates: MaturityAnalysisCandidate[];
  summary: MaturityAnalysisSummary | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useMaturityAnalysis(sortBy: string = 'gap'): UseMaturityAnalysisResult {
  const query = useMaturityAnalysisCandidates(sortBy);

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    candidates: query.data?.data ?? [],
    summary: query.data?.summary ?? null,
    isLoading: query.isLoading,
    error: query.error,
    refetch,
  };
}

export function useMaturityGapDetail(enterpriseCapabilityId: EnterpriseCapabilityId | undefined) {
  return useQuery({
    queryKey: queryKeys.enterpriseCapabilities.maturityGap(enterpriseCapabilityId!),
    queryFn: () => enterpriseArchApi.getMaturityGapDetail(enterpriseCapabilityId!),
    enabled: !!enterpriseCapabilityId,
  });
}

export interface UseMaturityGapDetailResult {
  detail: MaturityGapDetail | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useMaturityGapDetailHook(enterpriseCapabilityId: EnterpriseCapabilityId | undefined): UseMaturityGapDetailResult {
  const query = useMaturityGapDetail(enterpriseCapabilityId);

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    detail: query.data ?? null,
    isLoading: query.isLoading,
    error: query.error,
    refetch,
  };
}

export interface UseUnlinkedCapabilitiesResult {
  capabilities: UnlinkedCapability[];
  total: number;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useUnlinkedCapabilities(
  businessDomainId?: string,
  search?: string
): UseUnlinkedCapabilitiesResult {
  const filters = { businessDomainId, search };

  const query = useQuery({
    queryKey: queryKeys.maturityAnalysis.unlinked(filters),
    queryFn: () => enterpriseArchApi.getUnlinkedCapabilities(filters),
  });

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    capabilities: query.data?.data ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    error: query.error,
    refetch,
  };
}

export function useSetTargetMaturity() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      enterpriseCapabilityId,
      targetMaturity,
    }: {
      enterpriseCapabilityId: EnterpriseCapabilityId;
      targetMaturity: number;
    }) => enterpriseArchApi.setTargetMaturity(enterpriseCapabilityId, targetMaturity),
    onSuccess: (_, { enterpriseCapabilityId }) => {
      invalidateFor(
        queryClient,
        mutationEffects.enterpriseCapabilities.setTargetMaturity(enterpriseCapabilityId)
      );
      toast.success('Target maturity updated successfully');
    },
    onError: (error: unknown) => {
      toast.error(getErrorMessage(error, 'Failed to set target maturity'));
    },
  });
}
