import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import type { ReactNode } from 'react';
import React from 'react';
import { describe, expect, it } from 'vitest';
import { toEnterpriseCapabilityId } from '../../../api/types';
import { createTestQueryClient, server } from '../../../test/helpers';
import { seedSpec172Db } from '../../../test/mocks/spec172/store';
import type { CompositionResponse } from '../types';
import { useComposition } from './useComposition';

function wrapper({ children }: { children: ReactNode }) {
  const client = createTestQueryClient();
  return React.createElement(QueryClientProvider, { client }, children);
}

function seedCarveOutScenario() {
  seedSpec172Db({
    enterpriseCapabilities: [
      { id: 'ec-crm', name: 'CRM', active: true, createdAt: '2026-01-01T00:00:00Z' },
      { id: 'ec-tp', name: 'Take Payment', active: true, createdAt: '2026-01-01T00:00:00Z' },
    ],
    directions: [
      {
        id: 'dir-crm',
        enterpriseCapabilityId: 'ec-crm',
        type: 'consolidate',
        status: 'draft',
        horizon: 'next',
        sourceCapabilityIds: ['cap-cim'],
        createdAt: '2026-01-01T00:00:00Z',
      },
      {
        id: 'dir-tp',
        enterpriseCapabilityId: 'ec-tp',
        type: 'consolidate',
        status: 'draft',
        horizon: 'next',
        sourceCapabilityIds: ['cap-fraud'],
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
      {
        id: 'cap-fraud',
        name: 'Customer Fraud Prevention',
        level: 'L2',
        parentId: 'cap-cim',
        businessDomainId: 'bd-c',
        businessDomainName: 'Customer',
      },
    ],
  });
}

describe('useComposition', () => {
  it('fetches the EC composition grouped by domain with roles and carve-outs', async () => {
    seedCarveOutScenario();
    const { result } = renderHook(() => useComposition(toEnterpriseCapabilityId('ec-crm')), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const data = result.current.data!;
    expect(data.meta.sourceCount).toBe(1);
    expect(data.meta.includedCount).toBe(2); // cim (source) + consent (implicit); fraud carved out
    expect(data.meta.carvedOutCount).toBe(1);

    const items = data.data.flatMap((g) => g.items);
    const byId = Object.fromEntries(items.map((i) => [i.capabilityId, i]));
    expect(byId['cap-cim'].role).toBe('source');
    expect(byId['cap-cim']._links['x-exclude']).toBeDefined();
    expect(byId['cap-consent'].role).toBe('implicit');
    expect(byId['cap-fraud'].role).toBe('carved-out');
    expect(byId['cap-fraud'].carvedOutBy?.enterpriseCapabilityName).toBe('Take Payment');
  });

  it('omits x-exclude on sources when the direction is proposed (source set frozen)', async () => {
    seedSpec172Db({
      enterpriseCapabilities: [
        { id: 'ec-ci', name: 'Customer Identity', active: true, createdAt: '2026-01-01T00:00:00Z' },
      ],
      directions: [
        {
          id: 'dir-ci',
          enterpriseCapabilityId: 'ec-ci',
          type: 'consolidate',
          status: 'proposed',
          horizon: 'now',
          sourceCapabilityIds: ['cap-a'],
          createdAt: '2026-01-01T00:00:00Z',
        },
      ],
      capabilities: [
        { id: 'cap-a', name: 'A', level: 'L2', parentId: null, businessDomainId: 'bd-c', businessDomainName: 'Customer' },
      ],
    });

    const { result } = renderHook(() => useComposition(toEnterpriseCapabilityId('ec-ci')), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const source = result.current.data!.data.flatMap((g) => g.items).find((i) => i.capabilityId === 'cap-a')!;
    expect(source._links['x-exclude']).toBeUndefined();
  });

  it('omits x-exclude on sources when the direction is agreed (R5 immutability)', async () => {
    seedSpec172Db({
      enterpriseCapabilities: [
        { id: 'ec-ci', name: 'Customer Identity', active: true, createdAt: '2026-01-01T00:00:00Z' },
      ],
      directions: [
        {
          id: 'dir-ci',
          enterpriseCapabilityId: 'ec-ci',
          type: 'consolidate',
          status: 'agreed',
          horizon: 'now',
          sourceCapabilityIds: ['cap-a'],
          createdAt: '2026-01-01T00:00:00Z',
        },
      ],
      capabilities: [
        {
          id: 'cap-a',
          name: 'A',
          level: 'L2',
          parentId: null,
          businessDomainId: 'bd-c',
          businessDomainName: 'Customer',
        },
      ],
    });

    const { result } = renderHook(() => useComposition(toEnterpriseCapabilityId('ec-ci')), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const source = result.current.data!.data.flatMap((g) => g.items).find((i) => i.capabilityId === 'cap-a')!;
    expect(source.role).toBe('source');
    expect(source._links['x-exclude']).toBeUndefined();
  });

  it('fetches from a supplied href instead of the derived composition URL', async () => {
    seedCarveOutScenario();
    const linkedResponse: CompositionResponse = {
      data: [],
      meta: { sourceCount: 9, includedCount: 9, carvedOutCount: 0, domainCount: 9 },
      _links: { self: { href: '/api/v1/_custom/composition', method: 'GET' } },
    };
    server.use(http.get('*/api/v1/_custom/composition', () => HttpResponse.json(linkedResponse)));

    const { result } = renderHook(
      () => useComposition(toEnterpriseCapabilityId('ec-crm'), '/api/v1/_custom/composition'),
      { wrapper },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data!.meta.sourceCount).toBe(9);
  });
});
