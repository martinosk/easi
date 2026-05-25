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

vi.mock('../../enterprise-architecture/hooks/useEnterpriseCapabilities', () => ({
  useEnterpriseCapabilityLinks: vi.fn(),
}));

vi.mock('../../business-domains/hooks/useBusinessDomains', () => ({
  useBusinessDomainsQuery: vi.fn(),
}));

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: vi.fn(),
}));

import { directionApi } from '../api/directionApi';
import { useEnterpriseCapabilityLinks } from '../../enterprise-architecture/hooks/useEnterpriseCapabilities';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { DirectionPanel } from './DirectionPanel';

const mocked = vi.mocked(directionApi.getForEnterpriseCapability);
const mockedLinks = vi.mocked(useEnterpriseCapabilityLinks);
const mockedDomains = vi.mocked(useBusinessDomainsQuery);
const mockedCapabilities = vi.mocked(useCapabilities);

interface LinkFixture {
  capabilityId: string;
  capabilityName?: string;
  businessDomainName?: string;
}

interface CapabilityFixture {
  id: string;
  name: string;
}

function renderPanel(response: ECDirectionResponse, links: LinkFixture[] = [], capabilities: CapabilityFixture[] = []) {
  mocked.mockResolvedValueOnce(response);
  mockedLinks.mockReturnValue({
    data: links.map((l, i) => ({
      id: `link-${i}`,
      enterpriseCapabilityId: 'ec-1',
      domainCapabilityId: l.capabilityId,
      domainCapabilityName: l.capabilityName,
      businessDomainName: l.businessDomainName,
    })),
  } as never);
  mockedDomains.mockReturnValue({ data: { data: [] } } as never);
  mockedCapabilities.mockReturnValue({ data: capabilities } as never);
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

  it('renders type, status, and narrative for a draft direction', async () => {
    renderPanel({
      direction: {
        id: 'd-1',
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        type: 'consolidate',
        status: 'draft',
        horizon: 'next',
        narrative: 'We are consolidating payroll into one.',
        sourceCapabilities: [
          { id: 'cap-1', stale: false },
          { id: 'cap-2', stale: false },
        ],
        placements: [{ targetBusinessDomainId: 'dom-1' }],
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
  });

  it('resolves source capability names via the global capabilities query when they are not linked to this EC', async () => {
    renderPanel(
      {
        direction: {
          id: 'd-1',
          enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
          type: 'consolidate',
          status: 'draft',
          horizon: 'next',
          narrative: 'n',
          sourceCapabilities: [{ id: 'cap-unlinked', stale: false }],
          placements: [],
          createdAt: '2025-01-01T00:00:00Z',
          _links: {},
        },
        _links: {},
      },
      [], // no ECLinks — the source is not linked to this EC
      [{ id: 'cap-unlinked', name: 'Payroll (Norway)' }],
    );

    await waitFor(() => {
      expect(screen.getByTestId('direction-sources')).toHaveTextContent('Payroll (Norway)');
    });
    expect(screen.queryByText('cap-unlinked')).not.toBeInTheDocument();
  });

  it('marks stale references when source capability has been deleted', async () => {
    renderPanel({
      direction: {
        id: 'd-1',
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        type: 'consolidate',
        status: 'proposed',
        horizon: 'next',
        narrative: 'narrative',
        sourceCapabilities: [
          { id: 'cap-1', stale: false },
          { id: 'cap-2', stale: true },
        ],
        placements: [],
        createdAt: '2025-01-01T00:00:00Z',
        _links: {
          'x-agree': { href: '/api/v1/enterprise-capabilities/ec-1/direction/agree', method: 'POST' },
        },
      },
      _links: {},
    });

    await waitFor(() => {
      expect(screen.getByTestId('stale-reference')).toBeInTheDocument();
    });
  });

  it('shows the business domain of each source capability', async () => {
    renderPanel(
      {
        direction: {
          id: 'd-1',
          enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
          type: 'consolidate',
          status: 'draft',
          horizon: 'next',
          narrative: 'n',
          sourceCapabilities: [
            { id: 'cap-1', stale: false },
            { id: 'cap-2', stale: false },
          ],
          placements: [{ targetBusinessDomainId: 'dom-1' }],
            createdAt: '2025-01-01T00:00:00Z',
          _links: {},
        },
        _links: {},
      },
      [
        { capabilityId: 'cap-1', capabilityName: 'Customer Care', businessDomainName: 'Passenger' },
        { capabilityId: 'cap-2', capabilityName: 'Customer Service', businessDomainName: 'Terminal' },
      ],
    );

    await waitFor(() => {
      const sources = screen.getByTestId('direction-sources');
      expect(sources).toHaveTextContent(/Customer Care.*Passenger/);
      expect(sources).toHaveTextContent(/Customer Service.*Terminal/);
    });
  });

  it('offers reject (but not advance/edit) for an agreed direction', async () => {
    renderPanel({
      direction: {
        id: 'd-1',
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        type: 'stay',
        status: 'agreed',
        horizon: 'now',
        narrative: 'narrative',
        sourceCapabilities: [{ id: 'cap-1', stale: false }],
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
    expect(screen.queryByTestId('advance-to-proposed')).not.toBeInTheDocument();
    expect(screen.queryByTestId('advance-to-agreed')).not.toBeInTheDocument();
    // Per spec BDD: reject-and-replace remains the documented escape hatch from agreed.
    expect(screen.getByTestId('reject-direction')).toBeInTheDocument();
  });
});
