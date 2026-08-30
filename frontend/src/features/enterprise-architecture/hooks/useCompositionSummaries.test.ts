import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import React from 'react';
import { describe, expect, it } from 'vitest';
import { createTestQueryClient } from '../../../test/helpers';
import { seedSpec172Db } from '../../../test/mocks/spec172/store';
import { useCompositionSummaries } from './useCompositionSummaries';

function wrapper({ children }: { children: ReactNode }) {
  const client = createTestQueryClient();
  return React.createElement(QueryClientProvider, { client }, children);
}

function seedTwoEcsOneDirected() {
  seedSpec172Db({
    enterpriseCapabilities: [
      { id: 'ec-crm', name: 'CRM', active: true, createdAt: '2026-01-01T00:00:00Z' },
      { id: 'ec-idle', name: 'Idle', active: true, createdAt: '2026-01-01T00:00:00Z' },
    ],
    directions: [
      {
        id: 'dir-crm',
        enterpriseCapabilityId: 'ec-crm',
        type: 'consolidate',
        status: 'proposed',
        horizon: 'next',
        sourceCapabilityIds: ['cap-cim'],
        createdAt: '2026-01-01T00:00:00Z',
      },
    ],
    capabilities: [
      {
        id: 'cap-cim',
        name: 'Customer Identity Mgmt',
        level: 'L1',
        parentId: null,
        businessDomainId: 'bd-c',
        businessDomainName: 'Customer',
      },
      {
        id: 'cap-consent',
        name: 'Customer Consent',
        level: 'L2',
        parentId: 'cap-cim',
        businessDomainId: 'bd-c',
        businessDomainName: 'Customer',
      },
    ],
  });
}

describe('useCompositionSummaries', () => {
  it('indexes one summary per enterprise capability by id', async () => {
    seedTwoEcsOneDirected();
    const { result } = renderHook(() => useCompositionSummaries(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const summaries = result.current.data!;
    expect(summaries.size).toBe(2);
    expect(summaries.get('ec-crm')).toMatchObject({
      enterpriseCapabilityId: 'ec-crm',
      sourceCount: 1,
      includedCount: 2,
      carvedOutCount: 0,
      domainCount: 1,
      hasActiveDirection: true,
      directionStatus: 'proposed',
    });
  });

  it('reports zero counts and no active direction for an undirected enterprise capability', async () => {
    seedTwoEcsOneDirected();
    const { result } = renderHook(() => useCompositionSummaries(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data!.get('ec-idle')).toMatchObject({
      sourceCount: 0,
      includedCount: 0,
      carvedOutCount: 0,
      domainCount: 0,
      hasActiveDirection: false,
    });
  });
});
