import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useBusinessDomains,
  useBusinessDomainsQuery,
  useBusinessDomain,
  useCreateBusinessDomain,
  useUpdateBusinessDomain,
  useDeleteBusinessDomain,
  useAssociateCapabilityWithDomain,
  useDissociateCapabilityFromDomain,
} from './useBusinessDomains';
import { queryKeys } from '../../../lib/queryClient';
import { buildBusinessDomain } from '../../../test/helpers/entityBuilders';
import type { BusinessDomain, BusinessDomainId, CapabilityId } from '../../../api/types';

vi.mock('../api', () => ({
  businessDomainsApi: {
    getAll: vi.fn(),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    getCapabilities: vi.fn(),
    associateCapability: vi.fn(),
    dissociateCapability: vi.fn(),
    getCapabilityRealizations: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { businessDomainsApi } from '../api';
import toast from 'react-hot-toast';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useBusinessDomains hooks', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('useBusinessDomainsQuery', () => {
    it('should fetch all business domains', async () => {
      const domains = [
        buildBusinessDomain({ id: 'domain-1' as BusinessDomainId, name: 'Domain 1' }),
        buildBusinessDomain({ id: 'domain-2' as BusinessDomainId, name: 'Domain 2' }),
      ];
      vi.mocked(businessDomainsApi.getAll).mockResolvedValue(domains);

      const { result } = renderHook(() => useBusinessDomainsQuery(), {
        wrapper: createWrapper(queryClient),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(domains);
      expect(businessDomainsApi.getAll).toHaveBeenCalledTimes(1);
    });

    it('should handle fetch error', async () => {
      const error = new Error('Failed to fetch domains');
      vi.mocked(businessDomainsApi.getAll).mockRejectedValue(error);

      const { result } = renderHook(() => useBusinessDomainsQuery(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('useBusinessDomain', () => {
    it('should fetch a single domain by id', async () => {
      const domain = buildBusinessDomain({
        id: 'domain-1' as BusinessDomainId,
        name: 'Test Domain',
      });
      vi.mocked(businessDomainsApi.getById).mockResolvedValue(domain);

      const { result } = renderHook(
        () => useBusinessDomain('domain-1' as BusinessDomainId),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(domain);
      expect(businessDomainsApi.getById).toHaveBeenCalledWith('domain-1');
    });

    it('should not fetch when id is undefined', async () => {
      const { result } = renderHook(() => useBusinessDomain(undefined), {
        wrapper: createWrapper(queryClient),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.isFetching).toBe(false);
      expect(businessDomainsApi.getById).not.toHaveBeenCalled();
    });
  });

  describe('useBusinessDomains (composite hook)', () => {
    it('should provide domains data and mutation functions', async () => {
      const domains = [
        buildBusinessDomain({ id: 'domain-1' as BusinessDomainId, name: 'Domain 1' }),
      ];
      vi.mocked(businessDomainsApi.getAll).mockResolvedValue(domains);

      const { result } = renderHook(() => useBusinessDomains(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.domains).toEqual(domains);
      expect(typeof result.current.createDomain).toBe('function');
      expect(typeof result.current.updateDomain).toBe('function');
      expect(typeof result.current.deleteDomain).toBe('function');
      expect(typeof result.current.refetch).toBe('function');
    });

    it('should create domain via createDomain function', async () => {
      vi.mocked(businessDomainsApi.getAll).mockResolvedValue([]);
      const newDomain = buildBusinessDomain({
        id: 'domain-1' as BusinessDomainId,
        name: 'New Domain',
      });
      vi.mocked(businessDomainsApi.create).mockResolvedValue(newDomain);

      const { result } = renderHook(() => useBusinessDomains(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      let createdDomain: BusinessDomain;
      await act(async () => {
        createdDomain = await result.current.createDomain('New Domain', 'Description');
      });

      expect(createdDomain!).toEqual(newDomain);
      expect(businessDomainsApi.create).toHaveBeenCalledWith({
        name: 'New Domain',
        description: 'Description',
      });
    });
  });

  describe('useCreateBusinessDomain', () => {
    it('should create a domain and update cache', async () => {
      const existingDomains = [
        buildBusinessDomain({ id: 'domain-1' as BusinessDomainId, name: 'Existing' }),
      ];
      const newDomain = buildBusinessDomain({
        id: 'domain-2' as BusinessDomainId,
        name: 'New Domain',
      });

      queryClient.setQueryData(queryKeys.businessDomains.lists(), existingDomains);
      vi.mocked(businessDomainsApi.create).mockResolvedValue(newDomain);

      const { result } = renderHook(() => useCreateBusinessDomain(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          name: 'New Domain',
          description: 'Test description',
        });
      });

      expect(businessDomainsApi.create).toHaveBeenCalledWith({
        name: 'New Domain',
        description: 'Test description',
      });

      const cachedDomains = queryClient.getQueryData<BusinessDomain[]>(
        queryKeys.businessDomains.lists()
      );
      expect(cachedDomains).toHaveLength(2);
      expect(cachedDomains?.[1]).toEqual(newDomain);
      expect(toast.success).toHaveBeenCalledWith('Business domain "New Domain" created');
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Domain name already exists');
      vi.mocked(businessDomainsApi.create).mockRejectedValue(error);

      const { result } = renderHook(() => useCreateBusinessDomain(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({ name: 'Duplicate' });
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Domain name already exists');
    });
  });

  describe('useUpdateBusinessDomain', () => {
    it('should update domain and update both list and detail cache', async () => {
      const existingDomain = buildBusinessDomain({
        id: 'domain-1' as BusinessDomainId,
        name: 'Original Name',
      });
      const updatedDomain = buildBusinessDomain({
        id: 'domain-1' as BusinessDomainId,
        name: 'Updated Name',
      });

      queryClient.setQueryData(queryKeys.businessDomains.lists(), [existingDomain]);
      queryClient.setQueryData(
        queryKeys.businessDomains.detail('domain-1'),
        existingDomain
      );
      vi.mocked(businessDomainsApi.update).mockResolvedValue(updatedDomain);

      const { result } = renderHook(() => useUpdateBusinessDomain(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          id: 'domain-1' as BusinessDomainId,
          request: { name: 'Updated Name' },
        });
      });

      const cachedDomains = queryClient.getQueryData<BusinessDomain[]>(
        queryKeys.businessDomains.lists()
      );
      expect(cachedDomains?.[0].name).toBe('Updated Name');

      const cachedDetail = queryClient.getQueryData<BusinessDomain>(
        queryKeys.businessDomains.detail('domain-1')
      );
      expect(cachedDetail?.name).toBe('Updated Name');

      expect(toast.success).toHaveBeenCalledWith('Business domain "Updated Name" updated');
    });
  });

  describe('useDeleteBusinessDomain', () => {
    it('should delete domain and remove from cache', async () => {
      const domain = buildBusinessDomain({
        id: 'domain-1' as BusinessDomainId,
        name: 'To Delete',
      });

      queryClient.setQueryData(queryKeys.businessDomains.lists(), [domain]);
      queryClient.setQueryData(queryKeys.businessDomains.detail('domain-1'), domain);
      vi.mocked(businessDomainsApi.delete).mockResolvedValue(undefined);

      const removeQueriesSpy = vi.spyOn(queryClient, 'removeQueries');

      const { result } = renderHook(() => useDeleteBusinessDomain(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync('domain-1' as BusinessDomainId);
      });

      const cachedDomains = queryClient.getQueryData<BusinessDomain[]>(
        queryKeys.businessDomains.lists()
      );
      expect(cachedDomains).toHaveLength(0);

      expect(removeQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.businessDomains.detail('domain-1'),
      });

      expect(toast.success).toHaveBeenCalledWith('Business domain deleted');
    });
  });

  describe('useAssociateCapabilityWithDomain', () => {
    it('should associate capability and invalidate queries', async () => {
      vi.mocked(businessDomainsApi.associateCapability).mockResolvedValue(undefined);
      const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const { result } = renderHook(() => useAssociateCapabilityWithDomain(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync({
          associateLink: '/api/v1/business-domains/domain-1/capabilities',
          request: { capabilityId: 'cap-1' as CapabilityId },
        });
      });

      expect(businessDomainsApi.associateCapability).toHaveBeenCalledWith(
        '/api/v1/business-domains/domain-1/capabilities',
        { capabilityId: 'cap-1' }
      );

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.businessDomains.all,
      });

      expect(toast.success).toHaveBeenCalledWith('Capability associated with domain');
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Capability already associated');
      vi.mocked(businessDomainsApi.associateCapability).mockRejectedValue(error);

      const { result } = renderHook(() => useAssociateCapabilityWithDomain(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({
            associateLink: '/api/v1/business-domains/domain-1/capabilities',
            request: { capabilityId: 'cap-1' as CapabilityId },
          });
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Capability already associated');
    });
  });

  describe('useDissociateCapabilityFromDomain', () => {
    it('should dissociate capability and invalidate queries', async () => {
      vi.mocked(businessDomainsApi.dissociateCapability).mockResolvedValue(undefined);
      const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const { result } = renderHook(() => useDissociateCapabilityFromDomain(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        await result.current.mutateAsync(
          '/api/v1/business-domains/domain-1/capabilities/cap-1'
        );
      });

      expect(businessDomainsApi.dissociateCapability).toHaveBeenCalledWith(
        '/api/v1/business-domains/domain-1/capabilities/cap-1'
      );

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: queryKeys.businessDomains.all,
      });

      expect(toast.success).toHaveBeenCalledWith('Capability removed from domain');
    });
  });
});
