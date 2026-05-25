import { MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { toComponentId, toEnterpriseCapabilityId, toStandardApplicationId } from '../../../api/types';
import type { ECStandardApplicationResponse, StandardApplication } from '../types';

vi.mock('../api/standardApplicationApi', () => ({
  standardApplicationApi: {
    getForEnterpriseCapability: vi.fn(),
    getHistory: vi.fn(),
    set: vi.fn(),
  },
}));

import { standardApplicationApi } from '../api/standardApplicationApi';
import { StandardApplicationPanel } from './StandardApplicationPanel';

const mockedGet = vi.mocked(standardApplicationApi.getForEnterpriseCapability);

function renderPanel(response: ECStandardApplicationResponse) {
  mockedGet.mockResolvedValueOnce(response);
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <MantineProvider>
      <QueryClientProvider client={queryClient}>
        <StandardApplicationPanel enterpriseCapabilityId={toEnterpriseCapabilityId('ec-1')} />
      </QueryClientProvider>
    </MantineProvider>,
  );
}

function standardWith(overrides: Partial<StandardApplication> = {}): ECStandardApplicationResponse {
  return {
    standard: {
      id: toStandardApplicationId('sa-1'),
      enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
      applicationId: toComponentId('app-1'),
      applicationStale: false,
      applicationName: 'Acme ERP',
      narrative: 'standard',
      setAt: '2025-04-12T10:00:00Z',
      _links: {},
      ...overrides,
    },
    _links: {},
  };
}

describe('StandardApplicationPanel', () => {
  it('shows "No standard yet" empty state when no standard exists', async () => {
    renderPanel({ standard: null, _links: {} });

    await waitFor(() => {
      expect(screen.getByTestId('standard-application-empty-state')).toHaveTextContent('No standard yet');
    });
  });

  it('offers "Set standard" only when HATEOAS x-set-standard link is present', async () => {
    renderPanel({
      standard: null,
      _links: {
        'x-set-standard': { href: '/api/v1/enterprise-capabilities/ec-1/standard-application', method: 'PUT' },
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId('set-standard-application-button')).toBeInTheDocument();
    });
  });

  it('does not offer Set standard when the HATEOAS link is absent (read-only actor)', async () => {
    renderPanel({ standard: null, _links: {} });

    await waitFor(() => {
      expect(screen.getByTestId('standard-application-empty-state')).toBeInTheDocument();
    });
    expect(screen.queryByTestId('set-standard-application-button')).not.toBeInTheDocument();
  });

  it('renders application name from the DTO and narrative for a current standard', async () => {
    renderPanel({
      standard: {
        id: toStandardApplicationId('sa-1'),
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        applicationId: toComponentId('app-1'),
        applicationStale: false,
        applicationName: 'Acme ERP',
        narrative: 'Covers the operational and reporting layers.',
        setAt: '2025-04-12T10:00:00Z',
        _links: {
          edit: { href: '/api/v1/enterprise-capabilities/ec-1/standard-application', method: 'PUT' },
          'x-history': {
            href: '/api/v1/enterprise-capabilities/ec-1/standard-application/history',
            method: 'GET',
          },
        },
      },
      _links: {
        edit: { href: '/api/v1/enterprise-capabilities/ec-1/standard-application', method: 'PUT' },
        'x-history': {
          href: '/api/v1/enterprise-capabilities/ec-1/standard-application/history',
          method: 'GET',
        },
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId('standard-application-name')).toHaveTextContent('Acme ERP');
    });
    expect(screen.getByTestId('standard-application-narrative')).toHaveTextContent(/operational and reporting/i);
    expect(screen.getByTestId('change-standard-application-button')).toBeInTheDocument();
    expect(screen.getByTestId('view-standard-application-history-button')).toBeInTheDocument();
  });

  it('renders a placeholder when application name is null', async () => {
    renderPanel(standardWith({ applicationName: null }));

    await waitFor(() => {
      expect(screen.getByTestId('standard-application-name')).toHaveTextContent('—');
    });
  });

  it('hides the Change standard button when the edit HATEOAS link is absent', async () => {
    const response = standardWith();
    response._links = {
      'x-history': {
        href: '/api/v1/enterprise-capabilities/ec-1/standard-application/history',
        method: 'GET',
      },
    };
    renderPanel(response);

    await waitFor(() => {
      expect(screen.getByTestId('standard-application-name')).toBeInTheDocument();
    });
    expect(screen.queryByTestId('change-standard-application-button')).not.toBeInTheDocument();
    expect(screen.getByTestId('view-standard-application-history-button')).toBeInTheDocument();
  });

  it('shows the stale indicator when the application has been deleted', async () => {
    renderPanel(standardWith({ applicationStale: true }));

    await waitFor(() => {
      expect(screen.getByTestId('standard-application-stale-indicator')).toBeInTheDocument();
    });
  });
});
