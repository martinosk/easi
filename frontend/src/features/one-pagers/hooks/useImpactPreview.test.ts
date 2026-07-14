import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { OnePagerConfiguration } from '../types';
import { useImpactPreview } from './useImpactPreview';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getImpactPreview: vi.fn(),
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
    subjectType: 'application',
    builtInFields: [],
    customFields: [],
    displayOrder: [],
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'admin@example.com',
    _links: {
      self: { href: '/api/v1/one-pagers/configurations/application', method: 'GET' },
      'x-impact-preview': { href: '/api/v1/one-pagers/configurations/application/impact-preview', method: 'GET' },
    },
    ...overrides,
  };
}

describe('useImpactPreview', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it.each([
    { label: 'an existing field', fieldId: 'contract-link' as string | undefined, fieldKind: 'custom' as const, affectedSubjectCount: 37 },
    { label: 'a new field with no fieldId', fieldId: undefined, fieldKind: 'custom' as const, affectedSubjectCount: 120 },
    { label: 'a built-in field through the builtIn field kind', fieldId: 'experts' as string | undefined, fieldKind: 'builtIn' as const, affectedSubjectCount: 40 },
  ])('fetches the preview for $label', async ({ fieldId, fieldKind, affectedSubjectCount }) => {
    const configuration = buildConfiguration();
    vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({
      subjectType: 'application',
      fieldId,
      affectedSubjectCount,
    });

    const { result } = renderHook(() => useImpactPreview(configuration, fieldId, true, fieldKind), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(onePagersApi.getImpactPreview).toHaveBeenCalledWith(configuration, fieldId, fieldKind);
    expect(result.current.data?.affectedSubjectCount).toBe(affectedSubjectCount);
  });

  it('does not fetch while disabled', () => {
    const configuration = buildConfiguration();

    renderHook(() => useImpactPreview(configuration, 'contract-link', false), {
      wrapper: createWrapper(queryClient),
    });

    expect(onePagersApi.getImpactPreview).not.toHaveBeenCalled();
  });

  it('refetches every time it is re-enabled instead of serving a stale cache', async () => {
    const configuration = buildConfiguration();
    vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({
      subjectType: 'application',
      fieldId: 'contract-link',
      affectedSubjectCount: 37,
    });

    const { unmount } = renderHook(() => useImpactPreview(configuration, 'contract-link', true), {
      wrapper: createWrapper(queryClient),
    });
    await waitFor(() => expect(onePagersApi.getImpactPreview).toHaveBeenCalledTimes(1));
    unmount();

    renderHook(() => useImpactPreview(configuration, 'contract-link', true), {
      wrapper: createWrapper(queryClient),
    });
    await waitFor(() => expect(onePagersApi.getImpactPreview).toHaveBeenCalledTimes(2));
  });
});
