import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useDomainCapabilities } from './useDomainCapabilities';
import { queryKeys } from '../../../lib/queryClient';
import type { Capability, CapabilityId } from '../../../api/types';

vi.mock('../../../api/client', () => ({
  apiClient: {
    getDomainCapabilities: vi.fn(),
    associateCapabilityWithDomain: vi.fn(),
    dissociateCapabilityFromDomain: vi.fn(),
  },
}));

import { apiClient } from '../../../api/client';

const createCapability = (id: string, name: string): Capability => ({
  id: id as CapabilityId,
  name,
  description: '',
  level: 'L1',
  status: 'Active',
  maturityLevel: 'Genesis',
  _links: {
    removeFromDomain: `/api/v1/capabilities/${id}/remove`,
  },
});

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useDomainCapabilities - Business Domain Query Invalidation', () => {
  let queryClient: QueryClient;
  let invalidateQueriesSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');
    vi.mocked(apiClient.getDomainCapabilities).mockResolvedValue([]);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('associateCapability', () => {
    it('should invalidate businessDomains query after successful association', async () => {
      vi.mocked(apiClient.associateCapabilityWithDomain).mockResolvedValue(undefined);

      const { result } = renderHook(
        () => useDomainCapabilities('/api/v1/domains/domain-1/capabilities'),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      const capability = createCapability('cap-1', 'Test Capability');
      await act(async () => {
        await result.current.associateCapability('cap-1' as CapabilityId, capability);
      });

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.businessDomains.all,
      });
    });

    it('should update local capabilities state after association', async () => {
      vi.mocked(apiClient.associateCapabilityWithDomain).mockResolvedValue(undefined);

      const { result } = renderHook(
        () => useDomainCapabilities('/api/v1/domains/domain-1/capabilities'),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      const capability = createCapability('cap-1', 'Test Capability');
      await act(async () => {
        await result.current.associateCapability('cap-1' as CapabilityId, capability);
      });

      expect(result.current.capabilities).toContainEqual(capability);
    });
  });

  describe('dissociateCapability', () => {
    it('should invalidate businessDomains query after successful dissociation', async () => {
      const capability = createCapability('cap-1', 'Test Capability');
      vi.mocked(apiClient.getDomainCapabilities).mockResolvedValue([capability]);
      vi.mocked(apiClient.dissociateCapabilityFromDomain).mockResolvedValue(undefined);

      const { result } = renderHook(
        () => useDomainCapabilities('/api/v1/domains/domain-1/capabilities'),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.capabilities).toHaveLength(1);
      });

      await act(async () => {
        await result.current.dissociateCapability(capability);
      });

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.businessDomains.all,
      });
    });

    it('should remove capability from local state after dissociation', async () => {
      const capability = createCapability('cap-1', 'Test Capability');
      vi.mocked(apiClient.getDomainCapabilities).mockResolvedValue([capability]);
      vi.mocked(apiClient.dissociateCapabilityFromDomain).mockResolvedValue(undefined);

      const { result } = renderHook(
        () => useDomainCapabilities('/api/v1/domains/domain-1/capabilities'),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.capabilities).toHaveLength(1);
      });

      await act(async () => {
        await result.current.dissociateCapability(capability);
      });

      expect(result.current.capabilities).toHaveLength(0);
    });
  });
});
