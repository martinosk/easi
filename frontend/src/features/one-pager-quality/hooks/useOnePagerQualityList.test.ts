import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { onePagerQualityQueryKeys } from '../queryKeys';
import type { OnePagerQualityResponse } from '../types';
import { useOnePagerQualityList } from './useOnePagerQualityList';

vi.mock('../api/onePagerQualityApi', () => ({
  onePagerQualityApi: {
    getList: vi.fn(),
  },
}));

import { onePagerQualityApi } from '../api/onePagerQualityApi';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function buildResponse(overrides: Partial<OnePagerQualityResponse> = {}): OnePagerQualityResponse {
  return {
    data: [],
    pagination: { hasMore: false, limit: 50 },
    _links: { self: { href: '/api/v1/one-pager-quality', method: 'GET' } },
    ...overrides,
  };
}

describe('useOnePagerQualityList', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('fetches the list with the default sort and order', async () => {
    const response = buildResponse();
    vi.mocked(onePagerQualityApi.getList).mockResolvedValue(response);

    const { result } = renderHook(() => useOnePagerQualityList({ sort: 'completeness', order: 'asc' }), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data).toEqual(response);
    expect(onePagerQualityApi.getList).toHaveBeenCalledWith({
      sort: 'completeness',
      order: 'asc',
      cursor: undefined,
      limit: undefined,
    });
  });

  it('calls the api with the provided cursor', async () => {
    vi.mocked(onePagerQualityApi.getList).mockResolvedValue(buildResponse());

    renderHook(() => useOnePagerQualityList({ sort: 'name', order: 'desc', cursor: 'abc123' }), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => {
      expect(onePagerQualityApi.getList).toHaveBeenCalledWith({
        sort: 'name',
        order: 'desc',
        cursor: 'abc123',
        limit: undefined,
      });
    });
  });

  it('caches under a query key scoped to sort, order, and cursor', async () => {
    const response = buildResponse();
    vi.mocked(onePagerQualityApi.getList).mockResolvedValue(response);

    renderHook(() => useOnePagerQualityList({ sort: 'creator', order: 'asc', cursor: 'xyz' }), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => {
      expect(queryClient.getQueryData(onePagerQualityQueryKeys.list('creator', 'asc', 'xyz'))).toEqual(response);
    });
  });

  it('uses a different query key when sort changes', async () => {
    vi.mocked(onePagerQualityApi.getList).mockResolvedValue(buildResponse());

    const { rerender } = renderHook(
      ({ sort }: { sort: 'completeness' | 'name' }) => useOnePagerQualityList({ sort, order: 'asc' }),
      {
        wrapper: createWrapper(queryClient),
        initialProps: { sort: 'completeness' },
      },
    );

    await waitFor(() => expect(onePagerQualityApi.getList).toHaveBeenCalledTimes(1));

    rerender({ sort: 'name' });

    await waitFor(() => expect(onePagerQualityApi.getList).toHaveBeenCalledTimes(2));
    expect(onePagerQualityApi.getList).toHaveBeenLastCalledWith({
      sort: 'name',
      order: 'asc',
      cursor: undefined,
      limit: undefined,
    });
  });

  it('surfaces fetch errors', async () => {
    const error = new Error('Failed to fetch one-pager quality');
    vi.mocked(onePagerQualityApi.getList).mockRejectedValue(error);

    const { result } = renderHook(() => useOnePagerQualityList({ sort: 'completeness', order: 'asc' }), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toEqual(error);
  });
});
