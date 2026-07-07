import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from '../../../api/client';
import type { BusinessDomain, BusinessDomainId } from '../../../api/types';
import { businessDomainsApi } from '../../business-domains/api';
import { useDomainFilterData } from './useDomainFilterData';

vi.mock('../../business-domains/api', () => ({
  businessDomainsApi: {
    getAll: vi.fn(),
    getCapabilitiesByDomainId: vi.fn(),
  },
}));

vi.mock('../../../api/client', () => ({
  apiClient: {
    getCapabilityRealizationsByDomain: vi.fn(),
  },
}));

vi.mock('../../origin-entities/api/originEntitiesApi', () => ({
  originEntitiesApi: {
    getAllOriginRelationships: vi.fn().mockResolvedValue({ acquiredVia: [], purchasedFrom: [], builtBy: [] }),
  },
}));

const domains = [
  { id: 'domain-1' as BusinessDomainId, name: 'Domain 1', _links: {} },
  { id: 'domain-2' as BusinessDomainId, name: 'Domain 2', _links: {} },
] as unknown as BusinessDomain[];

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useDomainFilterData', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(businessDomainsApi.getAll).mockResolvedValue({ data: domains, _links: {} });
    vi.mocked(businessDomainsApi.getCapabilitiesByDomainId).mockResolvedValue([]);
    vi.mocked(apiClient.getCapabilityRealizationsByDomain).mockResolvedValue([]);
  });

  it('does not fetch per-domain data when disabled', async () => {
    const { result } = renderHook(() => useDomainFilterData([], { enabled: false }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.domains).toHaveLength(2);
    });

    expect(businessDomainsApi.getCapabilitiesByDomainId).not.toHaveBeenCalled();
    expect(apiClient.getCapabilityRealizationsByDomain).not.toHaveBeenCalled();
  });

  it('fetches per-domain data for every domain when enabled', async () => {
    renderHook(() => useDomainFilterData([], { enabled: true }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(businessDomainsApi.getCapabilitiesByDomainId).toHaveBeenCalledTimes(2);
      expect(apiClient.getCapabilityRealizationsByDomain).toHaveBeenCalledTimes(2);
    });
  });

  it('starts fetching per-domain data when enabled flips to true', async () => {
    const { result, rerender } = renderHook(({ enabled }) => useDomainFilterData([], { enabled }), {
      initialProps: { enabled: false },
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.domains).toHaveLength(2);
    });
    expect(businessDomainsApi.getCapabilitiesByDomainId).not.toHaveBeenCalled();

    rerender({ enabled: true });

    await waitFor(() => {
      expect(businessDomainsApi.getCapabilitiesByDomainId).toHaveBeenCalledTimes(2);
      expect(apiClient.getCapabilityRealizationsByDomain).toHaveBeenCalledTimes(2);
    });
  });

  it('still exposes domain list and filter data shape when disabled', async () => {
    const { result } = renderHook(() => useDomainFilterData([], { enabled: false }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.domains).toHaveLength(2);
    });

    expect(result.current.domainFilterData.allDomainIds).toEqual(['domain-1', 'domain-2']);
    expect(result.current.domainFilterData.domainCapabilityIds.size).toBe(0);
    expect(result.current.domainFilterData.domainComponentIds.size).toBe(0);
  });
});
