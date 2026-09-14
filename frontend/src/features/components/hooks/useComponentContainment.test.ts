import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ComponentId } from '../../../api/types';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { componentsQueryKeys } from '../queryKeys';
import { useAttachComponent, useDetachComponent } from './useComponentContainment';

vi.mock('../api', () => ({
  componentsApi: {
    attachToParent: vi.fn(),
    detach: vi.fn(),
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

describe('component containment hooks', () => {
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

  it('attaches, invalidates part and parent caches, and toasts', async () => {
    const part = buildComponent({ id: 'quoting' as ComponentId });
    vi.mocked(componentsApi.attachToParent).mockResolvedValue(part);
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useAttachComponent(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({ component: part, request: { parentId: 'crm' as ComponentId, kind: 'composition' } });
    });

    expect(componentsApi.attachToParent).toHaveBeenCalledWith(part, { parentId: 'crm', kind: 'composition' });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.lists() });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.detail('quoting') });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.detail('crm') });
    expect(toast.success).toHaveBeenCalledWith('Application attached');
  });

  it('detaches, invalidates the former parent, and toasts', async () => {
    const part = buildComponent({
      id: 'quoting' as ComponentId,
      partOf: { id: 'crm' as ComponentId, name: 'CRM Suite', kind: 'aggregation' },
    });
    vi.mocked(componentsApi.detach).mockResolvedValue({ ...part, partOf: undefined });
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useDetachComponent(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync(part);
    });

    expect(componentsApi.detach).toHaveBeenCalledWith(part);
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.detail('quoting') });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: componentsQueryKeys.detail('crm') });
    expect(toast.success).toHaveBeenCalledWith('Application detached');
  });

  it('shows the backend message when attaching fails', async () => {
    const part = buildComponent({ id: 'quoting' as ComponentId });
    vi.mocked(componentsApi.attachToParent).mockRejectedValue(new Error('Component is already part of another component'));

    const { result } = renderHook(() => useAttachComponent(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      try {
        await result.current.mutateAsync({ component: part, request: { parentId: 'crm' as ComponentId, kind: 'composition' } });
      } catch {
        void 0;
      }
    });

    expect(toast.error).toHaveBeenCalledWith('Component is already part of another component');
  });
});
