import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { onePagersQueryKeys } from '../queryKeys';
import type { FieldValue, OnePagerFacts } from '../types';
import { useSaveOnePagerFacts } from './useSaveOnePagerFacts';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    recordFieldValue: vi.fn(),
    clearFieldValue: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import toast from 'react-hot-toast';
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
    _links: {
      self: { href: '/api/v1/one-pagers/vendor/vendor-1/facts', method: 'GET' },
      'x-record': { href: '/api/v1/one-pagers/vendor/vendor-1/facts', method: 'PUT' },
    },
    ...overrides,
  };
}

function buildFieldValue(overrides: Partial<FieldValue> = {}): FieldValue {
  return {
    fieldId: 'field-1',
    value: { type: 'text', version: 1, value: 'hello' },
    displayText: 'hello',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'ea@example.com',
    _links: {
      'x-record': { href: '/api/v1/one-pagers/vendor/vendor-1/facts/field-1', method: 'PUT' },
      'x-clear': { href: '/api/v1/one-pagers/vendor/vendor-1/facts/field-1', method: 'DELETE' },
    },
    ...overrides,
  };
}

describe('useSaveOnePagerFacts', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('records changed fields and clears removed fields, then invalidates the facts query', async () => {
    const facts = buildFacts();
    const cleared = buildFieldValue({ fieldId: 'field-2' });
    vi.mocked(onePagersApi.recordFieldValue).mockResolvedValue(facts);
    vi.mocked(onePagersApi.clearFieldValue).mockResolvedValue(facts);
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useSaveOnePagerFacts('vendor', 'vendor-1'), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({
        facts,
        records: [{ fieldId: 'field-1', value: { type: 'text', version: 1, value: 'updated' } }],
        clears: [cleared],
      });
    });

    expect(onePagersApi.recordFieldValue).toHaveBeenCalledWith(facts, 'field-1', {
      value: { type: 'text', version: 1, value: 'updated' },
    });
    expect(onePagersApi.clearFieldValue).toHaveBeenCalledWith(cleared);
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({
      queryKey: onePagersQueryKeys.factsForSubject('vendor', 'vendor-1'),
    });
    expect(toast.success).toHaveBeenCalledWith('One-Pager updated');
  });

  it('sends no clear requests when only records changed', async () => {
    const facts = buildFacts();
    vi.mocked(onePagersApi.recordFieldValue).mockResolvedValue(facts);

    const { result } = renderHook(() => useSaveOnePagerFacts('vendor', 'vendor-1'), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({
        facts,
        records: [{ fieldId: 'field-1', value: { type: 'number', version: 1, value: 42.5 } }],
        clears: [],
      });
    });

    expect(onePagersApi.recordFieldValue).toHaveBeenCalledTimes(1);
    expect(onePagersApi.clearFieldValue).not.toHaveBeenCalled();
  });

  it('invalidates and shows an error toast when a write fails', async () => {
    const facts = buildFacts();
    vi.mocked(onePagersApi.recordFieldValue).mockRejectedValue(new Error('Value rejected'));
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useSaveOnePagerFacts('vendor', 'vendor-1'), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await expect(
        result.current.mutateAsync({
          facts,
          records: [{ fieldId: 'field-1', value: { type: 'text', version: 1, value: 'x' } }],
          clears: [],
        }),
      ).rejects.toThrow();
    });

    expect(invalidateQueriesSpy).toHaveBeenCalledWith({
      queryKey: onePagersQueryKeys.factsForSubject('vendor', 'vendor-1'),
    });
    expect(toast.error).toHaveBeenCalledWith('Value rejected');
  });
});
