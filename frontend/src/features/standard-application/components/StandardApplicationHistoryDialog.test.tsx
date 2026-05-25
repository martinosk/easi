import { MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { toComponentId, toEnterpriseCapabilityId, toStandardApplicationId } from '../../../api/types';
import type { StandardApplicationHistory } from '../types';

vi.mock('../api/standardApplicationApi', () => ({
  standardApplicationApi: {
    getForEnterpriseCapability: vi.fn(),
    getHistory: vi.fn(),
    set: vi.fn(),
  },
}));

import { standardApplicationApi } from '../api/standardApplicationApi';
import { StandardApplicationHistoryDialog } from './StandardApplicationHistoryDialog';

const mockedGetHistory = vi.mocked(standardApplicationApi.getHistory);

function renderDialog(history: StandardApplicationHistory | undefined, opened: boolean) {
  if (history !== undefined) {
    mockedGetHistory.mockResolvedValueOnce(history);
  }
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <MantineProvider>
      <QueryClientProvider client={queryClient}>
        <StandardApplicationHistoryDialog
          enterpriseCapabilityId={toEnterpriseCapabilityId('ec-1')}
          opened={opened}
          onClose={() => undefined}
        />
      </QueryClientProvider>
    </MantineProvider>,
  );
}

describe('StandardApplicationHistoryDialog', () => {
  it('does not fetch history while closed', () => {
    renderDialog(undefined, false);

    expect(mockedGetHistory).not.toHaveBeenCalled();
  });

  it('shows the empty-state message when no entries exist', async () => {
    renderDialog(
      {
        standardApplicationId: toStandardApplicationId('ec-1'),
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        entries: [],
        _links: {},
      },
      true,
    );

    await waitFor(() => {
      expect(screen.getByText(/no history yet/i)).toBeInTheDocument();
    });
    expect(screen.queryByTestId('standard-application-history-table')).not.toBeInTheDocument();
  });

  it('renders application names directly from the DTO', async () => {
    renderDialog(
      {
        standardApplicationId: toStandardApplicationId('ec-1'),
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        entries: [
          {
            applicationId: toComponentId('app-b'),
            previousApplicationId: toComponentId('app-a'),
            applicationName: 'Beta Suite',
            previousApplicationName: 'Acme ERP',
            narrative: 'switched',
            setAt: '2026-05-12T10:00:00Z',
          },
          {
            applicationId: toComponentId('app-a'),
            applicationName: 'Acme ERP',
            previousApplicationName: null,
            narrative: 'first',
            setAt: '2026-04-01T10:00:00Z',
          },
        ],
        _links: {},
      },
      true,
    );

    await waitFor(() => {
      expect(screen.getByTestId('standard-application-history-table')).toBeInTheDocument();
    });
    expect(screen.getByText('Beta Suite')).toBeInTheDocument();
    expect(screen.getAllByText('Acme ERP').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('switched')).toBeInTheDocument();
    expect(screen.getByText('first')).toBeInTheDocument();
  });

  it('renders placeholders for null application names', async () => {
    renderDialog(
      {
        standardApplicationId: toStandardApplicationId('ec-1'),
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        entries: [
          {
            applicationId: toComponentId('app-a'),
            applicationName: null,
            previousApplicationName: null,
            narrative: 'set during lag window',
            setAt: '2026-04-01T10:00:00Z',
          },
        ],
        _links: {},
      },
      true,
    );

    await waitFor(() => {
      expect(screen.getByTestId('standard-application-history-table')).toBeInTheDocument();
    });
    expect(screen.queryByText('app-a')).not.toBeInTheDocument();
  });
});
