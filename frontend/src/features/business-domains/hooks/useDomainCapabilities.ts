import { useState, useEffect, useCallback } from 'react';
import { apiClient } from '../../../api/client';
import type { Capability, CapabilityId } from '../../../api/types';

export interface UseDomainCapabilitiesResult {
  capabilities: Capability[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  associateCapability: (capabilityId: CapabilityId, capability: Capability) => Promise<void>;
  dissociateCapability: (capability: Capability) => Promise<void>;
}

export function useDomainCapabilities(
  capabilitiesLink: string | undefined
): UseDomainCapabilitiesResult {
  const [capabilities, setCapabilities] = useState<Capability[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchCapabilities = useCallback(async () => {
    if (!capabilitiesLink) {
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);
    try {
      const data = await apiClient.getDomainCapabilities(capabilitiesLink);
      setCapabilities(data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch capabilities'));
    } finally {
      setIsLoading(false);
    }
  }, [capabilitiesLink]);

  useEffect(() => {
    fetchCapabilities();
  }, [fetchCapabilities]);

  const associateCapability = useCallback(
    async (capabilityId: CapabilityId, capability: Capability) => {
      if (!capabilitiesLink) {
        throw new Error('Capabilities link not available');
      }
      await apiClient.associateCapabilityWithDomain(capabilitiesLink, { capabilityId });
      setCapabilities((prev) => [...prev, capability]);
    },
    [capabilitiesLink]
  );

  const dissociateCapability = useCallback(async (capability: Capability) => {
    const dissociateLink = capability._links.removeFromDomain;
    if (!dissociateLink) {
      throw new Error('Dissociate link not available');
    }
    await apiClient.dissociateCapabilityFromDomain(dissociateLink);
    setCapabilities((prev) => prev.filter((c) => c.id !== capability.id));
  }, []);

  return {
    capabilities,
    isLoading,
    error,
    refetch: fetchCapabilities,
    associateCapability,
    dissociateCapability,
  };
}
