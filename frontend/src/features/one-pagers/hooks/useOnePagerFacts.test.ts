import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { onePagersQueryKeys } from '../queryKeys';
import type { OnePagerFacts } from '../types';
import { useOnePagerFacts } from './useOnePagerFacts';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getFacts: vi.fn(),
  },
}));

import { onePagersApi } from '../api/onePagersApi';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function buildFacts(overrides: Partial<OnePagerFacts> = {}): OnePagerFacts {
  return {
    subjectType: 'vendor',
    subjectId: 'vendor-1',
    values: [],
    _links: { self: { href: '/api/v1/one-pagers/vendor/vendor-1/facts', method: 'GET' } },
    ...overrides,
  };
}

describe('useOnePagerFacts', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('fetches the facts for a subject', async () => {
    const facts = buildFacts();
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(facts);

    const { result } = renderHook(() => useOnePagerFacts('vendor', 'vendor-1'), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data).toEqual(facts);
    expect(onePagersApi.getFacts).toHaveBeenCalledWith('vendor', 'vendor-1');
  });

  it('caches under a key scoped to subject type and id', async () => {
    const facts = buildFacts();
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(facts);

    renderHook(() => useOnePagerFacts('vendor', 'vendor-1'), { wrapper: createWrapper(queryClient) });

    await waitFor(() => {
      expect(queryClient.getQueryData(onePagersQueryKeys.factsForSubject('vendor', 'vendor-1'))).toEqual(facts);
    });
  });

  it('does not fetch while the subject id is empty', () => {
    renderHook(() => useOnePagerFacts('vendor', ''), { wrapper: createWrapper(queryClient) });

    expect(onePagersApi.getFacts).not.toHaveBeenCalled();
  });

  it('surfaces fetch errors', async () => {
    const error = new Error('Failed to fetch facts');
    vi.mocked(onePagersApi.getFacts).mockRejectedValue(error);

    const { result } = renderHook(() => useOnePagerFacts('application', 'app-1'), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toEqual(error);
  });

  it('does not fetch while disabled, even with a subject id', () => {
    renderHook(() => useOnePagerFacts('vendor', 'vendor-1', false), { wrapper: createWrapper(queryClient) });

    expect(onePagersApi.getFacts).not.toHaveBeenCalled();
  });
});
