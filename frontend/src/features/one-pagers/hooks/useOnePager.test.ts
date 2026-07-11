import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { onePagersQueryKeys } from '../queryKeys';
import type { OnePagerView } from '../types';
import { useOnePager } from './useOnePager';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getOnePager: vi.fn(),
  },
}));

import { onePagersApi } from '../api/onePagersApi';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function buildView(overrides: Partial<OnePagerView> = {}): OnePagerView {
  return {
    subjectType: 'vendor',
    subjectId: 'vendor-1',
    subjectName: 'Acme Corp',
    fields: [],
    completeness: { requiredCount: 0, filledCount: 0, missingFields: [] },
    _links: { self: { href: '/api/v1/one-pagers/vendor/vendor-1', method: 'GET' } },
    ...overrides,
  };
}

describe('useOnePager', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('fetches the one-pager view for a subject', async () => {
    const view = buildView();
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(view);

    const { result } = renderHook(() => useOnePager('vendor', 'vendor-1'), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data).toEqual(view);
    expect(onePagersApi.getOnePager).toHaveBeenCalledWith('vendor', 'vendor-1');
  });

  it('caches under a key scoped to subject type and id', async () => {
    const view = buildView();
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(view);

    renderHook(() => useOnePager('vendor', 'vendor-1'), { wrapper: createWrapper(queryClient) });

    await waitFor(() => {
      expect(queryClient.getQueryData(onePagersQueryKeys.onePager('vendor', 'vendor-1'))).toEqual(view);
    });
  });

  it('does not fetch while the subject type is missing', () => {
    renderHook(() => useOnePager(undefined, 'vendor-1'), { wrapper: createWrapper(queryClient) });

    expect(onePagersApi.getOnePager).not.toHaveBeenCalled();
  });

  it('does not fetch while the subject id is missing', () => {
    renderHook(() => useOnePager('vendor', undefined), { wrapper: createWrapper(queryClient) });

    expect(onePagersApi.getOnePager).not.toHaveBeenCalled();
  });

  it('surfaces fetch errors', async () => {
    const error = new Error('Failed to fetch one-pager');
    vi.mocked(onePagersApi.getOnePager).mockRejectedValue(error);

    const { result } = renderHook(() => useOnePager('application', 'app-1'), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toEqual(error);
  });
});
