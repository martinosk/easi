import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import React from 'react';
import { describe, expect, it, vi } from 'vitest';
import { ApiError, toEnterpriseCapabilityId } from '../../../api/types';
import { createTestQueryClient } from '../../../test/helpers';
import { seedSpec172Db } from '../../../test/mocks/spec172/store';
import { directionApi } from '../api/directionApi';
import { useAddSource, useCompositionPreview, useExcludeSource, useSourceCandidates } from './useDirection';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function wrapper({ children }: { children: ReactNode }) {
  const client = createTestQueryClient();
  return React.createElement(QueryClientProvider, { client }, children);
}

function seedScenario() {
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
        sourceCapabilityIds: ['cap-consent'],
        createdAt: '2026-01-01T00:00:00Z',
      },
      {
        id: 'dir-tp',
        enterpriseCapabilityId: 'ec-tp',
        type: 'consolidate',
        status: 'proposed',
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

describe('useSourceCandidates', () => {
  it('returns candidates annotated with R1 eligibility', async () => {
    seedScenario();
    const { result } = renderHook(() => useSourceCandidates(toEnterpriseCapabilityId('ec-crm'), { q: 'customer' }), {
      wrapper,
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const byId = Object.fromEntries(result.current.data!.data.map((c) => [c.capabilityId, c]));
    expect(byId['cap-cim'].eligible).toBe(true);
    expect(byId['cap-fraud'].eligible).toBe(false);
    expect(byId['cap-fraud'].conflictingEnterpriseCapability?.name).toBe('Take Payment');
  });

  it('is disabled (does not fetch) when the search term is empty', () => {
    seedScenario();
    const { result } = renderHook(() => useSourceCandidates(toEnterpriseCapabilityId('ec-crm'), { q: '  ' }), {
      wrapper,
    });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCompositionPreview', () => {
  it('previews the included set and per-source eligibility for a candidate source set', async () => {
    seedScenario();
    const { result } = renderHook(() => useCompositionPreview(toEnterpriseCapabilityId('ec-crm'), ['cap-cim']), {
      wrapper,
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const preview = result.current.data!;
    const roles = Object.fromEntries(preview.includedCapabilities.map((i) => [i.capabilityId, i.role]));
    expect(roles['cap-cim']).toBe('source');
    expect(roles['cap-consent']).toBe('implicit');
    expect(roles['cap-fraud']).toBe('carved-out');
    expect(preview.sourceEligibility).toEqual([
      { capabilityId: 'cap-cim', eligible: true, ineligibilityReason: null, conflictingEnterpriseCapability: null },
    ]);
  });
});

describe('useExcludeSource', () => {
  it('removes an explicit source from the active direction', async () => {
    seedScenario();
    const { result } = renderHook(() => useExcludeSource(), { wrapper });

    await result.current.mutateAsync({
      enterpriseCapabilityId: toEnterpriseCapabilityId('ec-crm'),
      capabilityId: 'cap-consent',
    });

    const composition = await directionApi.getForEnterpriseCapability(toEnterpriseCapabilityId('ec-crm'));
    expect(composition.direction?.sourceCapabilities).toHaveLength(0);
  });
});

describe('useAddSource', () => {
  it('adds an eligible source to the active direction', async () => {
    seedScenario();
    const { result } = renderHook(() => useAddSource(), { wrapper });

    const updated = await result.current.mutateAsync({
      enterpriseCapabilityId: toEnterpriseCapabilityId('ec-crm'),
      capabilityId: 'cap-cim',
    });

    expect(updated.sourceCapabilities.map((s) => s.id).sort()).toEqual(['cap-cim', 'cap-consent']);
  });

  it('rejects adding a source already owned by another active direction (R1) with a 409', async () => {
    seedScenario();
    const { result } = renderHook(() => useAddSource(), { wrapper });

    await expect(
      result.current.mutateAsync({
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-crm'),
        capabilityId: 'cap-fraud',
      }),
    ).rejects.toMatchObject({ statusCode: 409 });
  });
});

describe('capture R1 enforcement', () => {
  it('rejects capturing a direction sourcing a capability already sourced elsewhere', async () => {
    seedSpec172Db({
      enterpriseCapabilities: [
        { id: 'ec-ip', name: 'Identity Platform', active: true, createdAt: '2026-01-01T00:00:00Z' },
        { id: 'ec-tp', name: 'Take Payment', active: true, createdAt: '2026-01-01T00:00:00Z' },
      ],
      directions: [
        {
          id: 'dir-tp',
          enterpriseCapabilityId: 'ec-tp',
          type: 'consolidate',
          status: 'proposed',
          horizon: 'next',
          sourceCapabilityIds: ['cap-fraud'],
          createdAt: '2026-01-01T00:00:00Z',
        },
      ],
      capabilities: [
        {
          id: 'cap-fraud',
          name: 'Customer Fraud Prevention',
          level: 'L2',
          parentId: null,
          businessDomainId: 'bd-c',
          businessDomainName: 'Customer',
        },
      ],
    });
    await expect(
      directionApi.capture(toEnterpriseCapabilityId('ec-ip'), {
        type: 'consolidate',
        sourceCapabilityIds: ['cap-fraud'],
        horizon: 'next',
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });
});
