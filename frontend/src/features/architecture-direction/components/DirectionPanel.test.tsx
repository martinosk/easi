import { MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { toEnterpriseCapabilityId } from '../../../api/types';
import type { ECDirectionResponse } from '../types';

vi.mock('../api/directionApi', () => ({
  directionApi: {
    getForEnterpriseCapability: vi.fn(),
  },
}));

import { directionApi } from '../api/directionApi';
import { DirectionPanel } from './DirectionPanel';

const mocked = vi.mocked(directionApi.getForEnterpriseCapability);

function renderPanel(response: ECDirectionResponse) {
  mocked.mockResolvedValueOnce(response);
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <MantineProvider>
      <QueryClientProvider client={queryClient}>
        <DirectionPanel enterpriseCapabilityId={toEnterpriseCapabilityId('ec-1')} />
      </QueryClientProvider>
    </MantineProvider>,
  );
}

describe('DirectionPanel', () => {
  it('shows "No direction set" empty state when no direction exists', async () => {
    renderPanel({ direction: null, _links: {} });

    await waitFor(() => {
      expect(screen.getByTestId('direction-empty-state')).toHaveTextContent('No direction set');
    });
  });

  it('offers capture button only when the HATEOAS link is present', async () => {
    renderPanel({
      direction: null,
      _links: {
        'x-capture-direction': { href: '/api/v1/enterprise-capabilities/ec-1/direction', method: 'POST' },
      },
    });

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /capture direction/i })).toBeInTheDocument();
    });
  });

  it('does not offer capture button when the HATEOAS link is absent', async () => {
    renderPanel({ direction: null, _links: {} });

    await waitFor(() => {
      expect(screen.getByTestId('direction-empty-state')).toBeInTheDocument();
    });
    expect(screen.queryByRole('button', { name: /capture direction/i })).not.toBeInTheDocument();
  });

  it('renders type, status, and narrative for a draft direction (sources live in the composition view)', async () => {
    renderPanel({
      direction: {
        id: 'd-1',
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        type: 'consolidate',
        status: 'draft',
        horizon: 'next',
        narrative: 'We are consolidating payroll into one.',
        sourceCapabilities: [{ id: 'cap-1', stale: false, name: 'Payroll DK' }],
        placements: [],
        createdAt: '2025-01-01T00:00:00Z',
        _links: {
          'x-propose': { href: '/api/v1/enterprise-capabilities/ec-1/direction/propose', method: 'POST' },
          'x-reject': { href: '/api/v1/enterprise-capabilities/ec-1/direction/reject', method: 'POST' },
        },
      },
      _links: {},
    });

    await waitFor(() => {
      expect(screen.getByTestId('direction-status-badge')).toHaveTextContent(/draft/i);
    });
    expect(screen.getByTestId('direction-type')).toHaveTextContent('Consolidate');
    expect(screen.getByTestId('direction-narrative')).toHaveTextContent(/consolidating payroll/i);
    expect(screen.getByTestId('advance-to-proposed')).toBeInTheDocument();
    expect(screen.getByTestId('reject-direction')).toBeInTheDocument();
    expect(screen.queryByTestId('advance-to-agreed')).not.toBeInTheDocument();
    // sources are no longer listed in the direction panel
    expect(screen.queryByTestId('direction-sources')).not.toBeInTheDocument();
  });

  it('offers reject (but not advance/edit) for an agreed direction and explains immutability', async () => {
    renderPanel({
      direction: {
        id: 'd-1',
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        type: 'stay',
        status: 'agreed',
        horizon: 'now',
        narrative: 'narrative',
        sourceCapabilities: [{ id: 'cap-1', stale: false, name: 'Some Cap' }],
        placements: [],
        createdAt: '2025-01-01T00:00:00Z',
        _links: {
          'x-reject': { href: '/api/v1/enterprise-capabilities/ec-1/direction/reject', method: 'POST' },
        },
      },
      _links: {},
    });

    await waitFor(() => {
      expect(screen.getByTestId('direction-status-badge')).toHaveTextContent(/agreed/i);
    });
    expect(screen.getByTestId('direction-frozen-callout')).toBeInTheDocument();
    expect(screen.queryByTestId('advance-to-proposed')).not.toBeInTheDocument();
    expect(screen.queryByTestId('advance-to-agreed')).not.toBeInTheDocument();
    expect(screen.getByTestId('reject-direction')).toBeInTheDocument();
  });
});
