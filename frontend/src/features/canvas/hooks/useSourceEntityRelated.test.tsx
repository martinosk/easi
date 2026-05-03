import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { HATEOASLinks } from '../../../api/types';
import { useSourceEntityRelated } from './useSourceEntityRelated';

const componentsData: { id: string; name: string; _links: HATEOASLinks }[] = [];
const capabilitiesData: { id: string; name: string; _links: HATEOASLinks }[] = [];
const acquiredData: { id: string; _links: HATEOASLinks }[] = [];
const vendorData: { id: string; _links: HATEOASLinks }[] = [];
const teamData: { id: string; _links: HATEOASLinks }[] = [];

vi.mock('../../components/hooks/useComponents', () => ({
  useComponents: () => ({ data: componentsData }),
}));

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: () => ({ data: capabilitiesData }),
}));

vi.mock('../../origin-entities/hooks/useAcquiredEntities', () => ({
  useAcquiredEntitiesQuery: () => ({ data: acquiredData }),
}));

vi.mock('../../origin-entities/hooks/useVendors', () => ({
  useVendorsQuery: () => ({ data: vendorData }),
}));

vi.mock('../../origin-entities/hooks/useInternalTeams', () => ({
  useInternalTeamsQuery: () => ({ data: teamData }),
}));

function wrapper() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children);
}

const linksWithRelated = (entries: unknown): HATEOASLinks =>
  ({ self: { href: '/x', method: 'GET' }, 'x-related': entries }) as unknown as HATEOASLinks;

beforeEach(() => {
  componentsData.length = 0;
  capabilitiesData.length = 0;
  acquiredData.length = 0;
  vendorData.length = 0;
  teamData.length = 0;
});

afterEach(() => vi.clearAllMocks());

interface SeedCase {
  label: string;
  seed: () => void;
  nodeId: string;
  expectedRelationType: string;
}

const seedCases: SeedCase[] = [
  {
    label: 'component lookup by raw id',
    seed: () => {
      componentsData.push({
        id: 'comp-1',
        name: 'A',
        _links: linksWithRelated([
          { href: '/x', methods: ['POST'], title: 'Component (related)', targetType: 'component', relationType: 'component-relation' },
          { href: '/y', methods: ['GET'], title: 'Hidden', targetType: 'component', relationType: 'capability-requires' },
        ]),
      });
    },
    nodeId: 'comp-1',
    expectedRelationType: 'component-relation',
  },
  {
    label: 'capability lookup via cap- prefix',
    seed: () => {
      capabilitiesData.push({
        id: 'cap-uuid',
        name: 'B',
        _links: linksWithRelated([
          { href: '/c', methods: ['POST'], title: 'Capability (child of)', targetType: 'capability', relationType: 'capability-parent' },
        ]),
      });
    },
    nodeId: 'cap-cap-uuid',
    expectedRelationType: 'capability-parent',
  },
  {
    label: 'acquired entity lookup via acq- prefix',
    seed: () => {
      acquiredData.push({
        id: 'a1',
        _links: linksWithRelated([
          { href: '/c', methods: ['POST'], title: 'Component (acquired-via)', targetType: 'component', relationType: 'origin-acquired-via' },
        ]),
      });
    },
    nodeId: 'acq-a1',
    expectedRelationType: 'origin-acquired-via',
  },
  {
    label: 'vendor lookup via vendor- prefix',
    seed: () => {
      vendorData.push({
        id: 'v1',
        _links: linksWithRelated([
          { href: '/c', methods: ['POST'], title: 'Component (purchased-from)', targetType: 'component', relationType: 'origin-purchased-from' },
        ]),
      });
    },
    nodeId: 'vendor-v1',
    expectedRelationType: 'origin-purchased-from',
  },
  {
    label: 'internal team lookup via team- prefix',
    seed: () => {
      teamData.push({
        id: 't1',
        _links: linksWithRelated([
          { href: '/c', methods: ['POST'], title: 'Component (built-by)', targetType: 'component', relationType: 'origin-built-by' },
        ]),
      });
    },
    nodeId: 'team-t1',
    expectedRelationType: 'origin-built-by',
  },
];

describe('useSourceEntityRelated', () => {
  it('returns empty when nodeId is null', () => {
    const { result } = renderHook(() => useSourceEntityRelated(null), { wrapper: wrapper() });
    expect(result.current).toEqual([]);
  });

  it.each(seedCases)('$label', ({ seed, nodeId, expectedRelationType }) => {
    seed();
    const { result } = renderHook(() => useSourceEntityRelated(nodeId), { wrapper: wrapper() });
    expect(result.current).toHaveLength(1);
    expect(result.current[0].relationType).toBe(expectedRelationType);
  });

  it('filters out GET-only entries', () => {
    componentsData.push({
      id: 'comp-1',
      name: 'A',
      _links: linksWithRelated([
        { href: '/x', methods: ['POST'], title: 'Component (related)', targetType: 'component', relationType: 'component-relation' },
        { href: '/y', methods: ['GET'], title: 'Read only', targetType: 'component', relationType: 'capability-requires' },
      ]),
    });
    const { result } = renderHook(() => useSourceEntityRelated('comp-1'), { wrapper: wrapper() });
    expect(result.current).toHaveLength(1);
    expect(result.current[0].relationType).toBe('component-relation');
  });

  it('returns empty array when entity is not found', () => {
    const { result } = renderHook(() => useSourceEntityRelated('comp-not-here'), { wrapper: wrapper() });
    expect(result.current).toEqual([]);
  });
});
