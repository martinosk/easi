import { MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { toComponentId, toEnterpriseCapabilityId, toStandardApplicationId } from '../../../api/types';
import type { ECStandardApplicationResponse } from '../types';

vi.mock('../api/standardApplicationApi', () => ({
  standardApplicationApi: {
    getForEnterpriseCapability: vi.fn(),
    getHistory: vi.fn(),
    set: vi.fn(),
  },
}));

vi.mock('../../components/hooks/useComponents', () => ({
  useComponents: vi.fn(),
}));

import { standardApplicationApi } from '../api/standardApplicationApi';
import { useComponents } from '../../components/hooks/useComponents';
import { StandardApplicationPanel } from './StandardApplicationPanel';

const mockedGet = vi.mocked(standardApplicationApi.getForEnterpriseCapability);
const mockedComponents = vi.mocked(useComponents);

function renderPanel(response: ECStandardApplicationResponse) {
  mockedGet.mockResolvedValueOnce(response);
  mockedComponents.mockReturnValue({
    data: [{ id: toComponentId('app-1'), name: 'Acme ERP', createdAt: '2025-01-01', _links: {} }],
    isLoading: false,
    error: null,
  } as never);
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <MantineProvider>
      <QueryClientProvider client={queryClient}>
        <StandardApplicationPanel enterpriseCapabilityId={toEnterpriseCapabilityId('ec-1')} />
      </QueryClientProvider>
    </MantineProvider>,
  );
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

  it('renders application name and narrative for a current standard', async () => {
    renderPanel({
      standard: {
        id: toStandardApplicationId('sa-1'),
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        applicationId: toComponentId('app-1'),
        applicationStale: false,
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

  it('hides the Change standard button when the edit HATEOAS link is absent', async () => {
    renderPanel({
      standard: {
        id: toStandardApplicationId('sa-1'),
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        applicationId: toComponentId('app-1'),
        applicationStale: false,
        narrative: 'standard',
        setAt: '2025-04-12T10:00:00Z',
        _links: {},
      },
      _links: {
        'x-history': {
          href: '/api/v1/enterprise-capabilities/ec-1/standard-application/history',
          method: 'GET',
        },
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId('standard-application-name')).toBeInTheDocument();
    });
    expect(screen.queryByTestId('change-standard-application-button')).not.toBeInTheDocument();
    expect(screen.getByTestId('view-standard-application-history-button')).toBeInTheDocument();
  });

  it('shows the stale indicator when the application has been deleted', async () => {
    renderPanel({
      standard: {
        id: toStandardApplicationId('sa-1'),
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        applicationId: toComponentId('app-1'),
        applicationStale: true,
        narrative: 'standard',
        setAt: '2025-04-12T10:00:00Z',
        _links: {},
      },
      _links: {},
    });

    await waitFor(() => {
      expect(screen.getByTestId('standard-application-stale-indicator')).toBeInTheDocument();
    });
  });
});
