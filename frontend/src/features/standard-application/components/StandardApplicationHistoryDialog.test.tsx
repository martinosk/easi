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

vi.mock('../../components/hooks/useComponents', () => ({
  useComponents: vi.fn(),
}));

import { standardApplicationApi } from '../api/standardApplicationApi';
import { useComponents } from '../../components/hooks/useComponents';
import { StandardApplicationHistoryDialog } from './StandardApplicationHistoryDialog';

const mockedGetHistory = vi.mocked(standardApplicationApi.getHistory);
const mockedComponents = vi.mocked(useComponents);

function renderDialog(history: StandardApplicationHistory | undefined, opened: boolean) {
  mockedComponents.mockReturnValue({
    data: [
      { id: toComponentId('app-a'), name: 'Acme ERP', createdAt: '2025-01-01', _links: {} },
      { id: toComponentId('app-b'), name: 'Beta Suite', createdAt: '2025-01-01', _links: {} },
    ],
    isLoading: false,
    error: null,
  } as never);
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

  it('resolves application names and renders previous in the table', async () => {
    renderDialog(
      {
        standardApplicationId: toStandardApplicationId('ec-1'),
        enterpriseCapabilityId: toEnterpriseCapabilityId('ec-1'),
        entries: [
          { applicationId: toComponentId('app-b'), previousApplicationId: toComponentId('app-a'), narrative: 'switched', setAt: '2026-05-12T10:00:00Z' },
          { applicationId: toComponentId('app-a'), narrative: 'first', setAt: '2026-04-01T10:00:00Z' },
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
});
