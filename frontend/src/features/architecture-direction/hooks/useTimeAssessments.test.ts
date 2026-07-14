import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { type StubTimeAssessment, seedSpec180Db } from '../../../test/mocks/spec180/store';
import { timeAssessmentApi } from '../api/timeAssessmentApi';
import {
  useAssessRealization,
  useRemoveTimeAssessment,
  useTimeAssessmentRollups,
  useTimeAssessmentsByCapabilityIds,
} from './useTimeAssessments';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function seedAssessment(overrides: Partial<StubTimeAssessment> = {}) {
  seedSpec180Db({
    assessments: [
      {
        id: 'ta-1',
        capabilityId: 'cap-1',
        capabilityName: 'Booking management',
        componentId: 'comp-1',
        componentName: 'Seabook',
        grade: 'Migrate',
        rationale: '',
        assessedBy: 'user-1',
        assessedByName: 'Domain Architect',
        assessedAt: '2026-06-01T00:00:00Z',
        ...overrides,
      },
    ],
  });
}

describe('useTimeAssessmentsByCapabilityIds', () => {
  it('fetches current assessments for the given capability ids', async () => {
    seedAssessment();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useTimeAssessmentsByCapabilityIds(['cap-1']), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0].grade).toBe('Migrate');
  });

  it('does not fetch when there are no capability ids', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useTimeAssessmentsByCapabilityIds([]), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.fetchStatus).toBe('idle');
  });

  it('exposes the x-assess collection link when the caller has write permission', async () => {
    seedSpec180Db({ canWrite: true });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useTimeAssessmentsByCapabilityIds(['cap-1']), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?._links?.['x-assess']).toBeDefined();
  });

  it('omits the x-assess collection link for a read-only caller', async () => {
    seedSpec180Db({ canWrite: false });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useTimeAssessmentsByCapabilityIds(['cap-1']), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?._links?.['x-assess']).toBeUndefined();
  });
});

describe('useTimeAssessmentRollups', () => {
  it('counts current assessments per grade across the landscape for each component id', async () => {
    seedSpec180Db({
      assessments: [
        {
          id: 'ta-1',
          capabilityId: 'cap-1',
          capabilityName: 'Booking management',
          componentId: 'comp-seabook',
          componentName: 'Seabook',
          grade: 'Invest',
          rationale: '',
          assessedBy: 'user-1',
          assessedAt: '2026-06-01T00:00:00Z',
        },
        {
          id: 'ta-2',
          capabilityId: 'cap-2',
          capabilityName: 'Accounts receivable',
          componentId: 'comp-seabook',
          componentName: 'Seabook',
          grade: 'Migrate',
          rationale: '',
          assessedBy: 'user-1',
          assessedAt: '2026-06-01T00:00:00Z',
        },
      ],
    });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useTimeAssessmentRollups(['comp-seabook']), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.data[0].counts).toEqual({ Invest: 1, Tolerate: 0, Migrate: 1, Eliminate: 0 });
  });
});

describe('timeAssessmentApi.getOne', () => {
  it('returns null when the pair has no current assessment (404)', async () => {
    const result = await timeAssessmentApi.getOne('cap-unassessed', 'comp-unassessed');
    expect(result).toBeNull();
  });

  it('returns the assessment when one exists', async () => {
    seedAssessment();
    const result = await timeAssessmentApi.getOne('cap-1', 'comp-1');
    expect(result?.grade).toBe('Migrate');
  });
});

describe('useAssessRealization', () => {
  let queryClient: QueryClient;
  let invalidateQueriesSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
    invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('assesses a realisation and invalidates assessment and rollup queries', async () => {
    const { result } = renderHook(() => useAssessRealization(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({
        capabilityId: 'cap-1',
        componentId: 'comp-1',
        request: { grade: 'Migrate' },
      });
    });

    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: ['timeAssessments'] });

    const found = await timeAssessmentApi.getOne('cap-1', 'comp-1');
    expect(found?.grade).toBe('Migrate');
  });
});

describe('useRemoveTimeAssessment', () => {
  it('removes an assessment and returns the pair to unassessed', async () => {
    seedAssessment();
    const assessment = await timeAssessmentApi.getOne('cap-1', 'comp-1');
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
    const { result } = renderHook(() => useRemoveTimeAssessment(), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({ assessment: assessment! });
    });

    const afterRemoval = await timeAssessmentApi.getOne('cap-1', 'comp-1');
    expect(afterRemoval).toBeNull();
  });
});
