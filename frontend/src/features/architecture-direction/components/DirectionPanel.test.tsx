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

import { directionApi } from '../api/directionApi';
import { useEnterpriseCapabilityLinks } from '../../enterprise-architecture/hooks/useEnterpriseCapabilities';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { DirectionPanel } from './DirectionPanel';

const mocked = vi.mocked(directionApi.getForEnterpriseCapability);
const mockedLinks = vi.mocked(useEnterpriseCapabilityLinks);
const mockedDomains = vi.mocked(useBusinessDomainsQuery);

interface LinkFixture {
  capabilityId: string;
  capabilityName?: string;
  businessDomainName?: string;
}

function renderPanel(response: ECDirectionResponse, links: LinkFixture[] = []) {
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
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <DirectionPanel enterpriseCapabilityId={toEnterpriseCapabilityId('ec-1')} />
    </QueryClientProvider>,
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
        hasStaleReferences: false,
        createdAt: '2025-01-01T00:00:00Z',
        _links: {
          'x-advance-proposed': { href: '/api/v1/directions/d-1/advance/proposed', method: 'POST' },
          'x-reject': { href: '/api/v1/directions/d-1/reject', method: 'POST' },
        },
      },
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
        hasStaleReferences: true,
        createdAt: '2025-01-01T00:00:00Z',
        _links: {
          'x-advance-agreed': { href: '/api/v1/directions/d-1/advance/agreed', method: 'POST' },
        },
      },
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
          hasStaleReferences: false,
          createdAt: '2025-01-01T00:00:00Z',
          _links: {},
        },
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

  it('hides advance/reject actions for an agreed direction', async () => {
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
        hasStaleReferences: false,
        createdAt: '2025-01-01T00:00:00Z',
        _links: {},
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId('direction-status-badge')).toHaveTextContent(/agreed/i);
    });
    expect(screen.queryByTestId('advance-to-proposed')).not.toBeInTheDocument();
    expect(screen.queryByTestId('advance-to-agreed')).not.toBeInTheDocument();
    expect(screen.queryByTestId('reject-direction')).not.toBeInTheDocument();
  });
});
