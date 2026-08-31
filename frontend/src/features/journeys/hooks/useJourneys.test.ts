import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { addComponent } from '../../../test/mocks/db';
import { buildStubJourney } from '../../../test/mocks/spec182/builders';
import { type StubJourney, seedSpec182Db } from '../../../test/mocks/spec182/store';
import { journeyApi } from '../api/journeyApi';
import type { CapabilityJourney } from '../types';
import {
  useAbandonJourney,
  useAddJourneyMilestone,
  useAllJourneys,
  useCaptureJourney,
  useChangeJourneySourceApplications,
  useCompleteJourney,
  useJourneyForCapability,
  useJourneyHistory,
  useRemoveJourneyMilestone,
  useStartJourney,
  useUpdateJourneyDetails,
  useUpdateJourneyMilestone,
  useUpdateJourneyProgress,
} from './useJourneys';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function newQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
}

describe('useJourneyForCapability', () => {
  it('returns the active journey wrapper for a capability', async () => {
    seedSpec182Db({ journeys: [buildStubJourney()] });
    const { result } = renderHook(() => useJourneyForCapability('cap-1'), { wrapper: createWrapper(newQueryClient()) });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.journeys[0]?.id).toBe('journey-1');
  });

  it('returns no journeys with an x-capture link when none is active and the caller can write', async () => {
    seedSpec182Db({ canWrite: true });
    const { result } = renderHook(() => useJourneyForCapability('cap-1'), { wrapper: createWrapper(newQueryClient()) });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.journeys).toEqual([]);
    expect(result.current.data?._links?.['x-capture']).toBeDefined();
  });

  it('omits the x-capture link for a read-only caller', async () => {
    seedSpec182Db({ canWrite: false });
    const { result } = renderHook(() => useJourneyForCapability('cap-1'), { wrapper: createWrapper(newQueryClient()) });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?._links?.['x-capture']).toBeUndefined();
  });

  it('does not fetch when no capability id is given', () => {
    const { result } = renderHook(() => useJourneyForCapability(undefined), {
      wrapper: createWrapper(newQueryClient()),
    });

    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useJourneyHistory', () => {
  it('returns all journeys for a capability, newest first', async () => {
    seedSpec182Db({
      journeys: [
        buildStubJourney({ id: 'journey-old', status: 'done', plannedAt: '2025-01-01T00:00:00Z' }),
        buildStubJourney({ id: 'journey-new', status: 'abandoned', plannedAt: '2026-01-01T00:00:00Z' }),
      ],
    });
    const { result } = renderHook(() => useJourneyHistory('cap-1'), { wrapper: createWrapper(newQueryClient()) });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.data.map((j) => j.id)).toEqual(['journey-new', 'journey-old']);
  });
});

describe('useAllJourneys', () => {
  it('returns the current journey for every capability in the tenant', async () => {
    seedSpec182Db({
      journeys: [
        buildStubJourney({ capabilityId: 'cap-1' }),
        buildStubJourney({ id: 'journey-2', capabilityId: 'cap-2' }),
      ],
    });
    const { result } = renderHook(() => useAllJourneys(), {
      wrapper: createWrapper(newQueryClient()),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.data).toHaveLength(2);
  });
});

describe('journey mutations', () => {
  let queryClient: QueryClient;
  let invalidateQueriesSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    queryClient = newQueryClient();
    invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  async function seedActiveJourney(overrides: Partial<StubJourney> = {}): Promise<CapabilityJourney> {
    seedSpec182Db({ journeys: [buildStubJourney(overrides)] });
    const wrapper = await journeyApi.getForCapability('cap-1');
    return wrapper.journeys[0];
  }

  async function runMutation<TArgs>(
    useHook: () => { mutateAsync: (args: TArgs) => Promise<unknown> },
    args: TArgs,
  ): Promise<void> {
    const { result } = renderHook(useHook, { wrapper: createWrapper(queryClient) });
    await act(async () => {
      await result.current.mutateAsync(args);
    });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: ['journeys'] });
  }

  async function currentJourney(): Promise<CapabilityJourney | null> {
    const wrapper = await journeyApi.getForCapability('cap-1');
    return wrapper.journeys[0] ?? null;
  }

  it('captures a journey and invalidates journey queries', async () => {
    const seabook = addComponent({ name: 'Seabook' });
    const phoenix = addComponent({ name: 'Phoenix' });

    await runMutation(useCaptureJourney, {
      capabilityId: 'cap-1',
      request: {
        kind: 'migration' as const,
        fromComponentIds: [String(seabook.id)],
        toComponentId: String(phoenix.id),
      },
    });

    expect((await currentJourney())?.kind).toBe('migration');
  });

  it('starts a journey and invalidates journey queries', async () => {
    const journey = await seedActiveJourney();

    await runMutation(useStartJourney, journey);

    expect((await currentJourney())?.status).toBe('in-flight');
  });

  it('completes an in-flight journey', async () => {
    const journey = await seedActiveJourney({ status: 'in-flight' });

    await runMutation(useCompleteJourney, journey);

    const history = await journeyApi.getHistory('cap-1');
    expect(history.data[0].status).toBe('done');
  });

  it('abandons a planned journey', async () => {
    const journey = await seedActiveJourney();

    await runMutation(useAbandonJourney, journey);

    const history = await journeyApi.getHistory('cap-1');
    expect(history.data[0].status).toBe('abandoned');
  });

  it('updates journey details', async () => {
    const journey = await seedActiveJourney();

    await runMutation(useUpdateJourneyDetails, {
      journey,
      request: { note: 'Updated note', targetPeriod: { year: 2028, quarter: 1 } },
    });

    const after = await currentJourney();
    expect(after?.note).toBe('Updated note');
    expect(after?.targetPeriod).toEqual({ year: 2028, quarter: 1 });
  });

  it('updates journey progress', async () => {
    const journey = await seedActiveJourney();

    await runMutation(useUpdateJourneyProgress, { journey, request: { progress: 60 } });

    expect((await currentJourney())?.progress).toBe(60);
  });

  it('changes journey source applications', async () => {
    const capacity = addComponent({ name: 'CapacityMgmt' });
    const journey = await seedActiveJourney();

    await runMutation(useChangeJourneySourceApplications, {
      journey,
      request: { componentIds: [String(capacity.id)] },
    });

    expect((await currentJourney())?.fromApplications.map((a) => a.componentId)).toEqual([String(capacity.id)]);
  });

  it('adds a milestone to a journey', async () => {
    const journey = await seedActiveJourney();

    await runMutation(useAddJourneyMilestone, { journey, request: { label: 'API live' } });

    expect((await currentJourney())?.milestones.map((m) => m.label)).toEqual(['API live']);
  });

  it('updates a milestone on a journey', async () => {
    const journey = await seedActiveJourney({
      milestones: [{ id: 'ms-1', label: 'API live', targetPeriod: null, status: 'planned' }],
    });

    await runMutation(useUpdateJourneyMilestone, {
      milestone: journey.milestones[0],
      request: { label: 'API live', targetPeriod: null, status: 'done' as const },
    });

    expect((await currentJourney())?.milestones[0].status).toBe('done');
  });

  it('removes a milestone from a journey', async () => {
    const journey = await seedActiveJourney({
      milestones: [{ id: 'ms-1', label: 'API live', targetPeriod: null, status: 'planned' }],
    });

    await runMutation(useRemoveJourneyMilestone, journey.milestones[0]);

    expect((await currentJourney())?.milestones).toHaveLength(0);
  });
});
