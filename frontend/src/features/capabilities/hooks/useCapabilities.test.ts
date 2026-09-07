import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { CapabilityDependencyId, CapabilityId } from '../../../api/types';
import { buildCapability, buildCapabilityDependency } from '../../../test/helpers/entityBuilders';
import { businessDomainsQueryKeys } from '../../business-domains/queryKeys';
import { capabilitiesQueryKeys } from '../queryKeys';
import {
  useCapabilities,
  useCapability,
  useCapabilityDependencies,
  useCreateCapability,
  useCreateCapabilityDependency,
  useDeleteCapability,
  useDeleteCapabilityDependency,
  useUpdateCapability,
} from './useCapabilities';

vi.mock('../api', () => ({
  capabilitiesApi: {
    getAll: vi.fn(),
    getById: vi.fn(),
    getChildren: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    updateMetadata: vi.fn(),
    delete: vi.fn(),
    changeParent: vi.fn(),
    getAllDependencies: vi.fn(),
    getOutgoingDependencies: vi.fn(),
    getIncomingDependencies: vi.fn(),
    createDependency: vi.fn(),
    deleteDependency: vi.fn(),
    getSystemsByCapability: vi.fn(),
    getCapabilitiesByComponent: vi.fn(),
    linkSystem: vi.fn(),
    updateRealization: vi.fn(),
    deleteRealization: vi.fn(),
    addExpert: vi.fn(),
    addTag: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import toast from 'react-hot-toast';
import { capabilitiesApi } from '../api';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

type QueryKey = readonly unknown[];

describe('useCapabilities hooks', () => {
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

  const renderMutation = <T>(hook: () => T) => {
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(hook, { wrapper: createWrapper(queryClient) });
    return { result, invalidateQueriesSpy };
  };

  const expectInvalidated = (spy: ReturnType<typeof vi.spyOn>, ...keys: QueryKey[]) => {
    for (const queryKey of keys) {
      expect(spy).toHaveBeenCalledWith({ queryKey });
    }
  };

  describe('useCapabilities', () => {
    it('should fetch all capabilities', async () => {
      const capabilities = [
        buildCapability({ id: 'cap-1' as CapabilityId, name: 'Capability 1' }),
        buildCapability({ id: 'cap-2' as CapabilityId, name: 'Capability 2' }),
      ];
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue(capabilities);

      const { result } = renderHook(() => useCapabilities(), {
        wrapper: createWrapper(queryClient),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(capabilities);
      expect(capabilitiesApi.getAll).toHaveBeenCalledTimes(1);
    });

    it('should handle fetch error', async () => {
      const error = new Error('Failed to fetch capabilities');
      vi.mocked(capabilitiesApi.getAll).mockRejectedValue(error);

      const { result } = renderHook(() => useCapabilities(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('useCapability', () => {
    it('should fetch a single capability by id', async () => {
      const capability = buildCapability({ id: 'cap-1' as CapabilityId, name: 'Test Capability' });
      vi.mocked(capabilitiesApi.getById).mockResolvedValue(capability);

      const { result } = renderHook(() => useCapability('cap-1' as CapabilityId), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(capability);
      expect(capabilitiesApi.getById).toHaveBeenCalledWith('cap-1');
    });

    it('should not fetch when id is undefined', async () => {
      const { result } = renderHook(() => useCapability(undefined), {
        wrapper: createWrapper(queryClient),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.isFetching).toBe(false);
      expect(capabilitiesApi.getById).not.toHaveBeenCalled();
    });
  });

  describe('useCreateCapability', () => {
    it('should create a capability and invalidate cache', async () => {
      const newCapability = buildCapability({ id: 'cap-2' as CapabilityId, name: 'New Capability' });
      vi.mocked(capabilitiesApi.create).mockResolvedValue(newCapability);
      const { result, invalidateQueriesSpy } = renderMutation(() => useCreateCapability());

      await act(async () => {
        await result.current.mutateAsync({
          name: 'New Capability',
          description: 'Test description',
          level: 'L1',
        });
      });

      expect(capabilitiesApi.create).toHaveBeenCalledWith({
        name: 'New Capability',
        description: 'Test description',
        level: 'L1',
      });

      expectInvalidated(invalidateQueriesSpy, capabilitiesQueryKeys.lists());
      expect(toast.success).toHaveBeenCalledWith('Capability "New Capability" created');
    });

    it('should invalidate parent children query when parentId is set', async () => {
      const newCapability = buildCapability({
        id: 'cap-2' as CapabilityId,
        name: 'Child Capability',
        parentId: 'parent-1' as CapabilityId,
      });
      vi.mocked(capabilitiesApi.create).mockResolvedValue(newCapability);
      const { result, invalidateQueriesSpy } = renderMutation(() => useCreateCapability());

      await act(async () => {
        await result.current.mutateAsync({
          name: 'Child Capability',
          level: 'L2',
          parentId: 'parent-1' as CapabilityId,
        });
      });

      expectInvalidated(invalidateQueriesSpy, capabilitiesQueryKeys.children('parent-1'));
    });

    it('should show error toast on failure', async () => {
      const error = new Error('Validation failed');
      vi.mocked(capabilitiesApi.create).mockRejectedValue(error);

      const { result } = renderHook(() => useCreateCapability(), {
        wrapper: createWrapper(queryClient),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({
            name: 'Test',
            level: 'L1',
          });
        } catch {
          // Expected to throw
        }
      });

      expect(toast.error).toHaveBeenCalledWith('Validation failed');
    });
  });

  describe('useUpdateCapability', () => {
    it('should update capability and invalidate cache', async () => {
      const existingCapability = buildCapability({
        id: 'cap-1' as CapabilityId,
        name: 'Original Name',
      });
      const updatedCapability = buildCapability({
        id: 'cap-1' as CapabilityId,
        name: 'Updated Name',
      });

      queryClient.setQueryData(capabilitiesQueryKeys.lists(), [existingCapability]);
      queryClient.setQueryData(capabilitiesQueryKeys.detail('cap-1'), existingCapability);
      vi.mocked(capabilitiesApi.update).mockResolvedValue(updatedCapability);

      const { result, invalidateQueriesSpy } = renderMutation(() => useUpdateCapability());

      await act(async () => {
        await result.current.mutateAsync({
          capability: existingCapability,
          request: { name: 'Updated Name' },
        });
      });

      expectInvalidated(invalidateQueriesSpy, capabilitiesQueryKeys.lists(), capabilitiesQueryKeys.detail('cap-1'));

      expect(toast.success).toHaveBeenCalledWith('Capability "Updated Name" updated');
    });
  });

  describe('useDeleteCapability', () => {
    it('should delete capability and invalidate relevant queries', async () => {
      const capability = buildCapability({ id: 'cap-1' as CapabilityId, name: 'To Delete' });

      queryClient.setQueryData(capabilitiesQueryKeys.lists(), [capability]);
      queryClient.setQueryData(capabilitiesQueryKeys.detail('cap-1'), capability);
      vi.mocked(capabilitiesApi.delete).mockResolvedValue(undefined);

      const { result, invalidateQueriesSpy } = renderMutation(() => useDeleteCapability());

      await act(async () => {
        await result.current.mutateAsync({ capability });
      });

      expectInvalidated(invalidateQueriesSpy, capabilitiesQueryKeys.lists(), capabilitiesQueryKeys.detail('cap-1'));

      expect(toast.success).toHaveBeenCalledWith('Capability deleted');
    });

    it.each([
      {
        name: 'should invalidate parent children query when parentId provided',
        extra: { parentId: 'parent-1' },
        expectedKey: capabilitiesQueryKeys.children('parent-1'),
      },
      {
        name: 'should invalidate domain capabilities query when domainId provided',
        extra: { domainId: 'domain-1' },
        expectedKey: businessDomainsQueryKeys.capabilities('domain-1'),
      },
    ])('$name', async ({ extra, expectedKey }) => {
      const capability = buildCapability({ id: 'cap-1' as CapabilityId, name: 'To Delete' });
      vi.mocked(capabilitiesApi.delete).mockResolvedValue(undefined);
      const { result, invalidateQueriesSpy } = renderMutation(() => useDeleteCapability());

      await act(async () => {
        await result.current.mutateAsync({ capability, ...extra });
      });

      expectInvalidated(invalidateQueriesSpy, expectedKey);
    });
  });

  describe('useCapabilityDependencies', () => {
    it('should fetch all capability dependencies', async () => {
      const dependencies = [
        buildCapabilityDependency({
          id: 'dep-1' as CapabilityDependencyId,
          sourceCapabilityId: 'cap-1' as CapabilityId,
          targetCapabilityId: 'cap-2' as CapabilityId,
        }),
      ];
      vi.mocked(capabilitiesApi.getAllDependencies).mockResolvedValue(dependencies);

      const { result } = renderHook(() => useCapabilityDependencies(), {
        wrapper: createWrapper(queryClient),
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toEqual(dependencies);
    });
  });

  describe('useCreateCapabilityDependency', () => {
    it('should create dependency and invalidate related queries', async () => {
      const newDependency = buildCapabilityDependency({
        id: 'dep-1' as CapabilityDependencyId,
        sourceCapabilityId: 'cap-1' as CapabilityId,
        targetCapabilityId: 'cap-2' as CapabilityId,
      });

      vi.mocked(capabilitiesApi.createDependency).mockResolvedValue(newDependency);
      const { result, invalidateQueriesSpy } = renderMutation(() => useCreateCapabilityDependency());

      await act(async () => {
        await result.current.mutateAsync({
          sourceCapabilityId: 'cap-1' as CapabilityId,
          targetCapabilityId: 'cap-2' as CapabilityId,
          dependencyType: 'Requires',
        });
      });

      expectInvalidated(
        invalidateQueriesSpy,
        capabilitiesQueryKeys.dependencies(),
        capabilitiesQueryKeys.outgoing('cap-1'),
        capabilitiesQueryKeys.incoming('cap-2'),
      );
      expect(toast.success).toHaveBeenCalledWith('Dependency created');
    });
  });

  describe('useDeleteCapabilityDependency', () => {
    it('should delete dependency and invalidate queries', async () => {
      const dependency = buildCapabilityDependency({
        id: 'dep-1' as CapabilityDependencyId,
        sourceCapabilityId: 'cap-1' as CapabilityId,
        targetCapabilityId: 'cap-2' as CapabilityId,
      });
      vi.mocked(capabilitiesApi.deleteDependency).mockResolvedValue(undefined);
      const { result, invalidateQueriesSpy } = renderMutation(() => useDeleteCapabilityDependency());

      await act(async () => {
        await result.current.mutateAsync(dependency);
      });

      expectInvalidated(
        invalidateQueriesSpy,
        capabilitiesQueryKeys.dependencies(),
        capabilitiesQueryKeys.outgoing('cap-1'),
        capabilitiesQueryKeys.incoming('cap-2'),
      );
      expect(toast.success).toHaveBeenCalledWith('Dependency deleted');
    });
  });
});
