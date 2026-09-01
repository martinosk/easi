import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ComponentId, OwnershipStatistics } from '../../../api/types';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { componentsQueryKeys } from '../queryKeys';
import {
  useAssignComponentOwner,
  useClearComponentOwnership,
  useConfirmComponentOwnership,
  useNominateComponentOwner,
  useOwnershipStatistics,
} from './useComponentOwnership';

vi.mock('../api', () => ({
  componentsApi: {
    nominateOwner: vi.fn(),
    confirmOwnership: vi.fn(),
    assignOwner: vi.fn(),
    clearOwnership: vi.fn(),
    getOwnershipStatistics: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import toast from 'react-hot-toast';
import { componentsApi } from '../api';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useComponentOwnership hooks', () => {
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

  const renderMutation = <T,>(hook: () => T) => {
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(hook, { wrapper: createWrapper(queryClient) });
    return { result, invalidateQueriesSpy };
  };

  const expectOwnershipInvalidation = (spy: ReturnType<typeof vi.spyOn>, componentId: string) => {
    expect(spy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.lists() });
    expect(spy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.detail(componentId) });
    expect(spy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.ownershipStatistics() });
  };

  const transitionCases = [
    {
      name: 'nominate owner',
      arrange: (component: ReturnType<typeof buildComponent>) =>
        vi.mocked(componentsApi.nominateOwner).mockResolvedValue({ ...component, ownershipState: 'nominated' }),
      hook: useNominateComponentOwner,
      mutateArg: (component: ReturnType<typeof buildComponent>) => ({
        component,
        request: { ownerKind: 'user' as const, ownerId: 'u-1' },
      }),
      verifyCall: (component: ReturnType<typeof buildComponent>) =>
        expect(componentsApi.nominateOwner).toHaveBeenCalledWith(component, { ownerKind: 'user', ownerId: 'u-1' }),
      successToast: 'Owner nominated',
    },
    {
      name: 'confirm ownership',
      arrange: (component: ReturnType<typeof buildComponent>) =>
        vi.mocked(componentsApi.confirmOwnership).mockResolvedValue({ ...component, ownershipState: 'owned' }),
      hook: useConfirmComponentOwnership,
      mutateArg: (component: ReturnType<typeof buildComponent>) => component,
      verifyCall: (component: ReturnType<typeof buildComponent>) =>
        expect(componentsApi.confirmOwnership).toHaveBeenCalledWith(component),
      successToast: 'Ownership confirmed',
    },
    {
      name: 'assign owner',
      arrange: (component: ReturnType<typeof buildComponent>) =>
        vi.mocked(componentsApi.assignOwner).mockResolvedValue({ ...component, ownershipState: 'managed' }),
      hook: useAssignComponentOwner,
      mutateArg: (component: ReturnType<typeof buildComponent>) => ({
        component,
        request: { ownerKind: 'team' as const, ownerId: 't-1' },
      }),
      verifyCall: (component: ReturnType<typeof buildComponent>) =>
        expect(componentsApi.assignOwner).toHaveBeenCalledWith(component, { ownerKind: 'team', ownerId: 't-1' }),
      successToast: 'Owner assigned',
    },
    {
      name: 'clear ownership',
      arrange: () => vi.mocked(componentsApi.clearOwnership).mockResolvedValue(undefined),
      hook: useClearComponentOwnership,
      mutateArg: (component: ReturnType<typeof buildComponent>) => component,
      verifyCall: (component: ReturnType<typeof buildComponent>) =>
        expect(componentsApi.clearOwnership).toHaveBeenCalledWith(component),
      successToast: 'Ownership cleared',
    },
  ];

  for (const tc of transitionCases) {
    it(`${tc.name} invalidates ownership caches and toasts`, async () => {
      const component = buildComponent({ id: 'comp-1' as ComponentId });
      tc.arrange(component);

      const { result, invalidateQueriesSpy } = renderMutation(() => tc.hook());

      await act(async () => {
        await (result.current.mutateAsync as (arg: unknown) => Promise<unknown>)(tc.mutateArg(component));
      });

      tc.verifyCall(component);
      expectOwnershipInvalidation(invalidateQueriesSpy, 'comp-1');
      expect(toast.success).toHaveBeenCalledWith(tc.successToast);
    });
  }

  it('shows an error toast when a transition fails', async () => {
    const component = buildComponent({ id: 'comp-1' as ComponentId });
    vi.mocked(componentsApi.clearOwnership).mockRejectedValue(new Error('Ownership is already unknown'));

    const { result } = renderHook(() => useClearComponentOwnership(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      try {
        await result.current.mutateAsync(component);
      } catch {
        void 0;
      }
    });

    expect(toast.error).toHaveBeenCalledWith('Ownership is already unknown');
  });

  it('fetches ownership statistics', async () => {
    const stats: OwnershipStatistics = { unknown: 3, nominated: 1, owned: 2, managed: 1, total: 7 };
    vi.mocked(componentsApi.getOwnershipStatistics).mockResolvedValue(stats);

    const { result } = renderHook(() => useOwnershipStatistics(), { wrapper: createWrapper(queryClient) });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toEqual(stats);
  });
});
