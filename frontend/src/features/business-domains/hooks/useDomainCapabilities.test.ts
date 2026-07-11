import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { BusinessDomainId, Capability, CapabilityId } from '../../../api/types';
import { businessDomainsQueryKeys } from '../queryKeys';
import { useAssociateCapability, useDissociateCapability } from './useDomainCapabilities';

vi.mock('../api', () => ({
  businessDomainsApi: {
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
    'x-remove-from-domain': { href: `/api/v1/business-domains/domain-1/capabilities/${id}`, method: 'DELETE' },
  },
});

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useAssociateCapability / useDissociateCapability', () => {
  let queryClient: QueryClient;
  let invalidateQueriesSpy: ReturnType<typeof vi.spyOn>;
  const domainId = 'domain-1' as BusinessDomainId;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('associates a capability with a domain given at call time (not bound to the hook)', async () => {
    vi.mocked(businessDomainsApi.associateCapabilityByDomainId).mockResolvedValue(undefined);
    const { result } = renderHook(() => useAssociateCapability(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({ domainId, capabilityId: 'cap-1' as CapabilityId });
    });

    expect(businessDomainsApi.associateCapabilityByDomainId).toHaveBeenCalledWith(domainId, {
      capabilityId: 'cap-1',
    });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: businessDomainsQueryKeys.capabilities(domainId) });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({
      queryKey: businessDomainsQueryKeys.realizations(domainId, 4),
    });
  });

  it('can associate the same capability with a different domain on a subsequent call', async () => {
    vi.mocked(businessDomainsApi.associateCapabilityByDomainId).mockResolvedValue(undefined);
    const { result } = renderHook(() => useAssociateCapability(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({
        domainId: 'domain-1' as BusinessDomainId,
        capabilityId: 'cap-1' as CapabilityId,
      });
    });
    await act(async () => {
      await result.current.mutateAsync({
        domainId: 'domain-2' as BusinessDomainId,
        capabilityId: 'cap-1' as CapabilityId,
      });
    });

    expect(businessDomainsApi.associateCapabilityByDomainId).toHaveBeenNthCalledWith(1, 'domain-1', {
      capabilityId: 'cap-1',
    });
    expect(businessDomainsApi.associateCapabilityByDomainId).toHaveBeenNthCalledWith(2, 'domain-2', {
      capabilityId: 'cap-1',
    });
  });

  it('dissociates a capability from a domain and invalidates that domain query', async () => {
    vi.mocked(businessDomainsApi.dissociateCapabilityByDomainId).mockResolvedValue(undefined);
    const capability = createCapability('cap-1', 'Test Capability');
    const { result } = renderHook(() => useDissociateCapability(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({ domainId, capability });
    });

    expect(businessDomainsApi.dissociateCapabilityByDomainId).toHaveBeenCalledWith(domainId, 'cap-1');
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: businessDomainsQueryKeys.capabilities(domainId) });

    await waitFor(() => expect(result.current.isPending).toBe(false));
  });
});
