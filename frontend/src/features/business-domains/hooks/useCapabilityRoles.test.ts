import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { describe, expect, it } from 'vitest';
import { toCapabilityId } from '../../../api/types';
import { buildCapability } from '../../../test/helpers';
import { seedSpec181Db } from '../../../test/mocks/spec181/store';
import { useCapabilityRoles } from './useCapabilityRoles';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

const capability = buildCapability({ id: toCapabilityId('cap-1'), name: 'Booking management' });

describe('useCapabilityRoles', () => {
  it('exposes the current role for a Direct realization by component id', async () => {
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
        },
      ],
    });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityRoles(capability), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.getRole('comp-1')?.role).toBe('standard'));
  });

  it('returns undefined for an unclassified component', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityRoles(capability), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.getRole('comp-2')).toBeUndefined();
  });

  it('reports canAssign true when the roles collection carries the x-assign link', async () => {
    seedSpec181Db({ canWrite: true });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityRoles(capability), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.canAssign).toBe(true));
  });

  it('reports canAssign false for a read-only caller with no x-assign link', async () => {
    seedSpec181Db({ canWrite: false });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityRoles(capability), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.canAssign).toBe(false));
  });

  it('does not fetch when no capability is open', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityRoles(null), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.getRole('comp-1')).toBeUndefined();
    expect(result.current.canAssign).toBe(false);
  });
});
