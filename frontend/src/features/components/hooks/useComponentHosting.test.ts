import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ComponentId } from '../../../api/types';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { componentsQueryKeys } from '../queryKeys';
import { useClassifyComponentHosting } from './useComponentHosting';

vi.mock('../api', () => ({
  componentsApi: {
    classifyHosting: vi.fn(),
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

describe('useClassifyComponentHosting', () => {
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

  it('classifies hosting, invalidates caches, and toasts', async () => {
    const component = buildComponent({ id: 'comp-1' as ComponentId });
    vi.mocked(componentsApi.classifyHosting).mockResolvedValue({ ...component, hosting: 'saas' });
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useClassifyComponentHosting(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({ component, hosting: 'saas' });
    });

    expect(componentsApi.classifyHosting).toHaveBeenCalledWith(component, 'saas');
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.lists() });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.detail('comp-1') });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.statistics() });
    expect(toast.success).toHaveBeenCalledWith('Hosting classified');
  });

  it('shows an error toast when classification fails', async () => {
    const component = buildComponent({ id: 'comp-1' as ComponentId });
    vi.mocked(componentsApi.classifyHosting).mockRejectedValue(new Error('Invalid hosting classification'));

    const { result } = renderHook(() => useClassifyComponentHosting(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      try {
        await result.current.mutateAsync({ component, hosting: 'saas' });
      } catch {
        void 0;
      }
    });

    expect(toast.error).toHaveBeenCalledWith('Invalid hosting classification');
  });
});
