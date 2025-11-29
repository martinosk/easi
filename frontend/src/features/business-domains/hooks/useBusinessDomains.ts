import { useState, useEffect, useCallback } from 'react';
import { apiClient } from '../../../api/client';
import type { BusinessDomain } from '../../../api/types';

export interface UseBusinessDomainsResult {
  domains: BusinessDomain[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  createDomain: (name: string, description?: string) => Promise<BusinessDomain>;
  updateDomain: (domain: BusinessDomain, name: string, description?: string) => Promise<BusinessDomain>;
  deleteDomain: (domain: BusinessDomain) => Promise<void>;
}

export function useBusinessDomains(): UseBusinessDomainsResult {
  const [domains, setDomains] = useState<BusinessDomain[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchDomains = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await apiClient.getBusinessDomains();
      setDomains(data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch domains'));
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDomains();
  }, [fetchDomains]);

  const createDomain = useCallback(async (name: string, description?: string) => {
    const newDomain = await apiClient.createBusinessDomain({ name, description });
    setDomains((prev) => [...prev, newDomain]);
    return newDomain;
  }, []);

  const updateDomain = useCallback(async (domain: BusinessDomain, name: string, description?: string) => {
    const updated = await apiClient.updateBusinessDomain(domain.id, { name, description });
    setDomains((prev) => prev.map((d) => (d.id === domain.id ? updated : d)));
    return updated;
  }, []);

  const deleteDomain = useCallback(async (domain: BusinessDomain) => {
    await apiClient.deleteBusinessDomain(domain.id);
    setDomains((prev) => prev.filter((d) => d.id !== domain.id));
  }, []);

  return {
    domains,
    isLoading,
    error,
    refetch: fetchDomains,
    createDomain,
    updateDomain,
    deleteDomain,
  };
}
