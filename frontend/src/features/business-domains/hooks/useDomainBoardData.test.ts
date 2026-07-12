import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { describe, expect, it, vi } from 'vitest';
import { toBusinessDomainId, toCapabilityId, toComponentId } from '../../../api/types';
import {
  buildBusinessDomain,
  buildCapabilityRealization,
  buildCapabilityAt as cap,
} from '../../../test/helpers/entityBuilders';
import { timeAssessmentApi } from '../../architecture-direction/api/timeAssessmentApi';
import { capabilitiesApi } from '../../capabilities/api';
import { businessDomainsApi } from '../api';
import { businessDomainsQueryKeys } from '../queryKeys';
import { useDomainBoardData } from './useDomainBoardData';

vi.mock('../api', () => ({
  businessDomainsApi: {
    getAll: vi.fn(),
    getCapabilitiesByDomainId: vi.fn(),
    getCapabilityRealizations: vi.fn(),
  },
}));

vi.mock('../../capabilities/api', () => ({
  capabilitiesApi: {
    getAll: vi.fn(),
  },
}));

vi.mock('../../architecture-direction/api/timeAssessmentApi', () => ({
  timeAssessmentApi: {
    getByCapabilityIds: vi.fn(),
  },
}));

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useDomainBoardData', () => {
  it('fans out one capabilities query and one realizations query per domain, keyed for mutation invalidation', async () => {
    const domainA = buildBusinessDomain({ id: toBusinessDomainId('domain-a'), name: 'Ferry Freight' });
    const domainB = buildBusinessDomain({ id: toBusinessDomainId('domain-b'), name: 'Logistics' });

    vi.mocked(businessDomainsApi.getAll).mockResolvedValue({
      data: [domainA, domainB],
      _links: { create: { href: '/api/v1/business-domains', method: 'POST' } },
    });
    vi.mocked(capabilitiesApi.getAll).mockResolvedValue([cap('l1-a', 'Alpha', 'L1'), cap('l1-b', 'Bravo', 'L1')]);
    vi.mocked(businessDomainsApi.getCapabilitiesByDomainId).mockImplementation(async (domainId) =>
      domainId === domainA.id ? [cap('l1-a', 'Alpha', 'L1')] : [cap('l1-b', 'Bravo', 'L1')],
    );
    vi.mocked(businessDomainsApi.getCapabilityRealizations).mockResolvedValue([]);

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useDomainBoardData(), { wrapper: createWrapper(queryClient) });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    await waitFor(() => expect(result.current.boardDomains).toHaveLength(2));

    expect(businessDomainsApi.getCapabilitiesByDomainId).toHaveBeenCalledWith(domainA.id);
    expect(businessDomainsApi.getCapabilitiesByDomainId).toHaveBeenCalledWith(domainB.id);
    expect(businessDomainsApi.getCapabilityRealizations).toHaveBeenCalledWith(domainA.id, 4);
    expect(businessDomainsApi.getCapabilityRealizations).toHaveBeenCalledWith(domainB.id, 4);

    expect(queryClient.getQueryData(businessDomainsQueryKeys.capabilities(domainA.id))).toEqual([
      cap('l1-a', 'Alpha', 'L1'),
    ]);
    expect(queryClient.getQueryData(businessDomainsQueryKeys.realizations(domainA.id, 4))).toEqual([]);

    const boardA = result.current.boardDomains.find((d) => d.domain.id === domainA.id);
    expect(boardA?.l1Groups.map((g) => g.node.capability.name)).toEqual(['Alpha']);
    const boardB = result.current.boardDomains.find((d) => d.domain.id === domainB.id);
    expect(boardB?.l1Groups.map((g) => g.node.capability.name)).toEqual(['Bravo']);
  });

  it('refetchDomain re-fetches only the targeted domain queries', async () => {
    const domainA = buildBusinessDomain({ id: toBusinessDomainId('domain-a') });

    vi.mocked(businessDomainsApi.getAll).mockResolvedValue({ data: [domainA], _links: {} });
    vi.mocked(capabilitiesApi.getAll).mockResolvedValue([]);
    vi.mocked(businessDomainsApi.getCapabilitiesByDomainId).mockResolvedValue([]);
    vi.mocked(businessDomainsApi.getCapabilityRealizations).mockResolvedValue([]);

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useDomainBoardData(), { wrapper: createWrapper(queryClient) });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    vi.mocked(businessDomainsApi.getCapabilitiesByDomainId).mockClear();
    await result.current.refetchDomain(domainA.id);

    expect(businessDomainsApi.getCapabilitiesByDomainId).toHaveBeenCalledTimes(1);
    expect(businessDomainsApi.getCapabilitiesByDomainId).toHaveBeenCalledWith(domainA.id);

    await result.current.refetchDomain(toBusinessDomainId('unknown-domain'));
    expect(businessDomainsApi.getCapabilitiesByDomainId).toHaveBeenCalledTimes(1);
  });

  it("fans out one assessments query per domain scoped to that domain's realized capability ids, enriching the returned realizations with their current grade", async () => {
    const domainA = buildBusinessDomain({ id: toBusinessDomainId('domain-a') });

    vi.mocked(businessDomainsApi.getAll).mockResolvedValue({ data: [domainA], _links: {} });
    vi.mocked(capabilitiesApi.getAll).mockResolvedValue([cap('l1-a', 'Alpha', 'L1')]);
    vi.mocked(businessDomainsApi.getCapabilitiesByDomainId).mockResolvedValue([cap('l1-a', 'Alpha', 'L1')]);
    vi.mocked(businessDomainsApi.getCapabilityRealizations).mockResolvedValue([
      {
        capabilityId: toCapabilityId('l1-a'),
        capabilityName: 'Alpha',
        level: 'L1',
        realizations: [
          buildCapabilityRealization({
            capabilityId: toCapabilityId('l1-a'),
            componentId: toComponentId('comp-1'),
            origin: 'Direct',
          }),
        ],
      },
    ]);
    vi.mocked(timeAssessmentApi.getByCapabilityIds).mockResolvedValue({
      data: [
        {
          id: 'ta-1',
          capabilityId: 'l1-a',
          capabilityName: 'Alpha',
          componentId: 'comp-1',
          componentName: 'Component 1',
          grade: 'Migrate',
          rationale: '',
          assessedBy: 'user-1',
          assessedAt: '2026-01-01T00:00:00Z',
          stale: false,
          _links: {},
        },
      ],
      _links: {},
    });

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useDomainBoardData(), { wrapper: createWrapper(queryClient) });

    await waitFor(() => expect(timeAssessmentApi.getByCapabilityIds).toHaveBeenCalledWith(['l1-a']));

    await waitFor(() => {
      const boardA = result.current.boardDomains.find((d) => d.domain.id === domainA.id);
      expect(boardA?.getRealizationsForCapability(toCapabilityId('l1-a'))[0]?.timeGrade).toBe('Migrate');
    });
  });

  it('surfaces canCreateDomain from the collection HATEOAS links', async () => {
    vi.mocked(businessDomainsApi.getAll).mockResolvedValue({
      data: [],
      _links: { create: { href: '/api/v1/business-domains', method: 'POST' } },
    });
    vi.mocked(capabilitiesApi.getAll).mockResolvedValue([]);

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useDomainBoardData(), { wrapper: createWrapper(queryClient) });

    await waitFor(() => expect(result.current.canCreateDomain).toBe(true));
  });
});
