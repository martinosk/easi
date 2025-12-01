import { useState, useEffect, useCallback } from 'react';
import { apiClient } from '../../../api/client';
import type { CapabilityId, CapabilityRealization } from '../../../api/types';

export interface UseCapabilityRealizationsResult {
  realizations: CapabilityRealization[];
  isLoading: boolean;
  error: Error | null;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
}

export function useCapabilityRealizations(
  capabilityIds: CapabilityId[],
  enabled: boolean
): UseCapabilityRealizationsResult {
  const [realizations, setRealizations] = useState<CapabilityRealization[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!enabled || capabilityIds.length === 0) {
      setIsLoading(false);
      setRealizations([]);
      setError(null);
      return;
    }

    const fetchRealizations = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const response = await apiClient.getCapabilityRealizations(capabilityIds);
        setRealizations(response.data || []);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch capability realizations'));
        setRealizations([]);
      } finally {
        setIsLoading(false);
      }
    };

    fetchRealizations();
  }, [enabled, capabilityIds]);

  const getRealizationsForCapability = useCallback(
    (capabilityId: CapabilityId): CapabilityRealization[] => {
      return realizations.filter((r) => r.capabilityId === capabilityId);
    },
    [realizations]
  );

  return {
    realizations,
    isLoading,
    error,
    getRealizationsForCapability,
  };
}
