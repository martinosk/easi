import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { type StubRealizationRole, seedSpec181Db } from '../../../test/mocks/spec181/store';
import { realizationRoleApi } from '../api/realizationRoleApi';
import { useAssignRealizationRole, useClearRealizationRole, useRealizationRolesByCapabilityIds } from './useRealizationRoles';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function seedRole(overrides: Partial<StubRealizationRole> = {}) {
  seedSpec181Db({
    roles: [
      {
        capabilityId: 'cap-1',
        capabilityName: 'Booking management',
        componentId: 'comp-1',
        componentName: 'Phoenix',
        role: 'standard',
        assignedBy: 'user-1',
        assignedAt: '2026-06-01T00:00:00Z',
        ...overrides,
      },
    ],
  });
}

describe('useRealizationRolesByCapabilityIds', () => {
  it('fetches current roles for the given capability ids', async () => {
    seedRole();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useRealizationRolesByCapabilityIds(['cap-1']), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0].role).toBe('standard');
  });

  it('does not fetch when there are no capability ids', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useRealizationRolesByCapabilityIds([]), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.fetchStatus).toBe('idle');
  });

  it('exposes the x-assign collection link when the caller has write permission', async () => {
    seedSpec181Db({ canWrite: true });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useRealizationRolesByCapabilityIds(['cap-1']), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?._links?.['x-assign']).toBeDefined();
  });

  it('omits the x-assign collection link for a read-only caller', async () => {
    seedSpec181Db({ canWrite: false });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useRealizationRolesByCapabilityIds(['cap-1']), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?._links?.['x-assign']).toBeUndefined();
  });
});

describe('realizationRoleApi.getOne', () => {
  it('returns null when the pair holds no current role (404)', async () => {
    const result = await realizationRoleApi.getOne('cap-unassigned', 'comp-unassigned');
    expect(result).toBeNull();
  });

  it('returns the role when one exists', async () => {
    seedRole();
    const result = await realizationRoleApi.getOne('cap-1', 'comp-1');
    expect(result?.role).toBe('standard');
  });
});

describe('useAssignRealizationRole', () => {
  let queryClient: QueryClient;
  let invalidateQueriesSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
    invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('assigns a role and invalidates realization role queries', async () => {
    const { result } = renderHook(() => useAssignRealizationRole(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({
        capabilityId: 'cap-1',
        componentId: 'comp-1',
        request: { role: 'legacy' },
      });
    });

    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: ['realizationRoles'] });

    const found = await realizationRoleApi.getOne('cap-1', 'comp-1');
    expect(found?.role).toBe('legacy');
  });
});

describe('useClearRealizationRole', () => {
  it('clears a role and returns the pair to unclassified', async () => {
    seedRole();
    const role = await realizationRoleApi.getOne('cap-1', 'comp-1');
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
    const { result } = renderHook(() => useClearRealizationRole(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({ role: role! });
    });

    const afterClear = await realizationRoleApi.getOne('cap-1', 'comp-1');
    expect(afterClear).toBeNull();
  });
});
