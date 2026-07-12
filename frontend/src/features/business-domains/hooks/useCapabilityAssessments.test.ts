import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { describe, expect, it } from 'vitest';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { buildCapability, buildCapabilityRealization } from '../../../test/helpers';
import { seedSpec180Db } from '../../../test/mocks/spec180/store';
import { useCapabilityAssessments } from './useCapabilityAssessments';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

const capability = buildCapability({ id: toCapabilityId('cap-1'), name: 'Booking management' });

describe('useCapabilityAssessments', () => {
  it('exposes the current assessment for a Direct realization by component id', async () => {
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
        },
      ],
    });
    const realizations = [buildCapabilityRealization({ componentId: toComponentId('comp-1'), origin: 'Direct' })];
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityAssessments(capability, realizations), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.getAssessment('comp-1')?.grade).toBe('Migrate'));
  });

  it('returns undefined for an unassessed component', () => {
    const realizations = [buildCapabilityRealization({ componentId: toComponentId('comp-2'), origin: 'Direct' })];
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityAssessments(capability, realizations), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.getAssessment('comp-2')).toBeUndefined();
  });

  it('exposes the landscape rollup counts per component id', async () => {
    seedSpec180Db({
      assessments: [
        {
          id: 'ta-1',
          capabilityId: 'cap-1',
          capabilityName: 'Booking management',
          componentId: 'comp-1',
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
          componentId: 'comp-1',
          componentName: 'Seabook',
          grade: 'Tolerate',
          rationale: '',
          assessedBy: 'user-1',
          assessedAt: '2026-06-01T00:00:00Z',
        },
      ],
    });
    const realizations = [buildCapabilityRealization({ componentId: toComponentId('comp-1'), origin: 'Direct' })];
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityAssessments(capability, realizations), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() =>
      expect(result.current.getRollup('comp-1')).toEqual({ Invest: 1, Tolerate: 1, Migrate: 0, Eliminate: 0 }),
    );
  });

  it('reports canAssess true when the assessments collection carries the x-assess link', async () => {
    seedSpec180Db({ canWrite: true });
    const realizations = [buildCapabilityRealization({ componentId: toComponentId('comp-1'), origin: 'Direct' })];
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityAssessments(capability, realizations), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.canAssess).toBe(true));
  });

  it('reports canAssess false for a read-only caller with no x-assess link', async () => {
    seedSpec180Db({ canWrite: false });
    const realizations = [buildCapabilityRealization({ componentId: toComponentId('comp-1'), origin: 'Direct' })];
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityAssessments(capability, realizations), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.canAssess).toBe(false));
  });

  it('does not fetch when no capability is open', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCapabilityAssessments(null, []), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.getAssessment('comp-1')).toBeUndefined();
    expect(result.current.canAssess).toBe(false);
  });
});
