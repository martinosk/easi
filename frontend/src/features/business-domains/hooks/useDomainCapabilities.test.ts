import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useDomainCapabilities } from './useDomainCapabilities';
import { queryKeys } from '../../../lib/queryClient';
import type { BusinessDomainId, Capability, CapabilityId } from '../../../api/types';

vi.mock('../api', () => ({
  businessDomainsApi: {
    getCapabilitiesByDomainId: vi.fn(),
    associateCapabilityByDomainId: vi.fn(),
    dissociateCapabilityByDomainId: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { businessDomainsApi } from '../api';

const createCapability = (id: string, name: string): Capability => ({
  id: id as CapabilityId,
  name,
  description: '',
  level: 'L1',
  status: 'Active',
  maturityLevel: 'Genesis',
  createdAt: '2024-01-01',
  _links: {
    self: { href: `/api/v1/capabilities/${id}`, method: 'GET' },
    removeFromDomain: { href: `/api/v1/capabilities/${id}/remove`, method: 'DELETE' },
  },
});

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useDomainCapabilities - Business Domain Query Invalidation', () => {
  let queryClient: QueryClient;
  let invalidateQueriesSpy: ReturnType<typeof vi.spyOn>;
  const domainId = 'domain-1' as BusinessDomainId;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');
    vi.mocked(businessDomainsApi.getCapabilitiesByDomainId).mockResolvedValue([]);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('useQuery behavior', () => {
    it('should fetch capabilities using useQuery when domainId is provided', async () => {
      const capabilities = [createCapability('cap-1', 'Test Capability')];
      vi.mocked(businessDomainsApi.getCapabilitiesByDomainId).mockResolvedValue(capabilities);

      const { result } = renderHook(() => useDomainCapabilities(domainId), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(businessDomainsApi.getCapabilitiesByDomainId).toHaveBeenCalledWith(domainId);
      expect(result.current.capabilities).toEqual(capabilities);
    });

    it('should not fetch when domainId is undefined', async () => {
      const { result } = renderHook(() => useDomainCapabilities(undefined), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(businessDomainsApi.getCapabilitiesByDomainId).not.toHaveBeenCalled();
      expect(result.current.capabilities).toEqual([]);
    });
  });

  describe('associateCapability', () => {
    it('should invalidate domain capabilities query after successful association', async () => {
      vi.mocked(businessDomainsApi.associateCapabilityByDomainId).mockResolvedValue(undefined);

      const { result } = renderHook(() => useDomainCapabilities(domainId), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      await act(async () => {
        await result.current.associateCapability('cap-1' as CapabilityId);
      });

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.businessDomains.capabilities(domainId),
      });
    });

    it('should call associateCapabilityByDomainId with correct arguments', async () => {
      vi.mocked(businessDomainsApi.associateCapabilityByDomainId).mockResolvedValue(undefined);

      const { result } = renderHook(() => useDomainCapabilities(domainId), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      await act(async () => {
        await result.current.associateCapability('cap-1' as CapabilityId);
      });

      expect(businessDomainsApi.associateCapabilityByDomainId).toHaveBeenCalledWith(domainId, {
        capabilityId: 'cap-1',
      });
    });
  });

  describe('dissociateCapability', () => {
    it('should invalidate domain capabilities query after successful dissociation', async () => {
      const capability = createCapability('cap-1', 'Test Capability');
      vi.mocked(businessDomainsApi.getCapabilitiesByDomainId).mockResolvedValue([capability]);
      vi.mocked(businessDomainsApi.dissociateCapabilityByDomainId).mockResolvedValue(undefined);

      const { result } = renderHook(() => useDomainCapabilities(domainId), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.capabilities).toHaveLength(1);
      });

      await act(async () => {
        await result.current.dissociateCapability(capability);
      });

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.businessDomains.capabilities(domainId),
      });
    });

    it('should call dissociateCapabilityByDomainId with correct arguments', async () => {
      const capability = createCapability('cap-1', 'Test Capability');
      vi.mocked(businessDomainsApi.getCapabilitiesByDomainId).mockResolvedValue([capability]);
      vi.mocked(businessDomainsApi.dissociateCapabilityByDomainId).mockResolvedValue(undefined);

      const { result } = renderHook(() => useDomainCapabilities(domainId), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.capabilities).toHaveLength(1);
      });

      await act(async () => {
        await result.current.dissociateCapability(capability);
      });

      expect(businessDomainsApi.dissociateCapabilityByDomainId).toHaveBeenCalledWith(
        domainId,
        'cap-1'
      );
    });
  });
});
