import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ComponentStatistics } from '../../../api/types';
import { useComponentStatistics } from './useComponentStatistics';

vi.mock('../api', () => ({
  componentsApi: {
    getStatistics: vi.fn(),
  },
}));

import { componentsApi } from '../api';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useComponentStatistics', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
  });

  it('fetches component statistics', async () => {
    const stats: ComponentStatistics = {
      unknown: 3,
      nominated: 1,
      owned: 2,
      managed: 1,
      hosting: { 'on-premises': 1, cloud: 2, saas: 1, 'third-party-hosted': 0, unknown: 3 },
      total: 7,
    };
    vi.mocked(componentsApi.getStatistics).mockResolvedValue(stats);

    const { result } = renderHook(() => useComponentStatistics(), { wrapper: createWrapper(queryClient) });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toEqual(stats);
  });
});
