import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import type { ReactNode } from 'react';
import React from 'react';
import { describe, expect, it } from 'vitest';
import { toEnterpriseCapabilityId } from '../../../api/types';
import { createTestQueryClient, server } from '../../../test/helpers';
import { seedSpec172Db } from '../../../test/mocks/spec172/store';
import type { ECDirectionResponse } from '../types';
import { useDirectionForEnterpriseCapability } from './useDirection';

function wrapper({ children }: { children: ReactNode }) {
  const client = createTestQueryClient();
  return React.createElement(QueryClientProvider, { client }, children);
}

describe('useDirectionForEnterpriseCapability', () => {
  it('fetches from the derived URL when no href is supplied', async () => {
    seedSpec172Db({
      enterpriseCapabilities: [
        { id: 'ec-crm', name: 'CRM', active: true, createdAt: '2026-01-01T00:00:00Z' },
      ],
    });

    const { result } = renderHook(() => useDirectionForEnterpriseCapability(toEnterpriseCapabilityId('ec-crm')), {
      wrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data!.direction).toBeNull();
  });

  it('fetches from a supplied href instead of the derived direction URL', async () => {
    seedSpec172Db({
      enterpriseCapabilities: [
        { id: 'ec-crm', name: 'CRM', active: true, createdAt: '2026-01-01T00:00:00Z' },
      ],
    });
    const linkedResponse: ECDirectionResponse = {
      direction: null,
      _links: { self: { href: '/api/v1/_custom/direction', method: 'GET' } },
    };
    server.use(http.get('*/api/v1/_custom/direction', () => HttpResponse.json(linkedResponse)));

    const { result } = renderHook(
      () =>
        useDirectionForEnterpriseCapability(toEnterpriseCapabilityId('ec-crm'), '/api/v1/_custom/direction'),
      { wrapper },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data!._links.self?.href).toBe('/api/v1/_custom/direction');
  });
});
