import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { onePagersQueryKeys } from '../queryKeys';
import type { OnePagerConfiguration } from '../types';
import { useOnePagerConfiguration } from './useOnePagerConfiguration';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getConfiguration: vi.fn(),
  },
}));

import { onePagersApi } from '../api/onePagersApi';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function buildConfiguration(overrides: Partial<OnePagerConfiguration> = {}): OnePagerConfiguration {
  return {
    id: 'config-1',
    subjectType: 'vendor',
    builtInFields: [],
    customFields: [],
    displayOrder: [],
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'admin@example.com',
    _links: { self: { href: '/api/v1/one-pagers/configurations/vendor', method: 'GET' } },
    ...overrides,
  };
}

describe('useOnePagerConfiguration', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('fetches the configuration for a subject type', async () => {
    const configuration = buildConfiguration();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);

    const { result } = renderHook(() => useOnePagerConfiguration('vendor'), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data).toEqual(configuration);
    expect(onePagersApi.getConfiguration).toHaveBeenCalledWith('vendor');
  });

  it('uses a query key scoped to the subject type', async () => {
    const configuration = buildConfiguration();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);

    renderHook(() => useOnePagerConfiguration('vendor'), { wrapper: createWrapper(queryClient) });

    await waitFor(() => {
      expect(queryClient.getQueryData(onePagersQueryKeys.configuration('vendor'))).toEqual(configuration);
    });
  });

  it('surfaces fetch errors', async () => {
    const error = new Error('Failed to fetch configuration');
    vi.mocked(onePagersApi.getConfiguration).mockRejectedValue(error);

    const { result } = renderHook(() => useOnePagerConfiguration('application'), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toEqual(error);
  });
});
