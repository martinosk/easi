import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MantineProvider } from '@mantine/core';
import React from 'react';
import { StrategyPillarsSettings } from './StrategyPillarsSettings';
import { strategyPillarsApi } from '../../../api/metadata';
import type { StrategyPillarsConfiguration } from '../../../api/types';
import { ApiError } from '../../../api/types';

vi.mock('../../../api/metadata', () => ({
  strategyPillarsApi: {
    getConfiguration: vi.fn(),
    createPillar: vi.fn(),
    updatePillar: vi.fn(),
    deletePillar: vi.fn(),
    batchUpdate: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { error: vi.fn(), success: vi.fn() },
}));

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
}

function renderWithProviders(ui: React.ReactElement, queryClient?: QueryClient) {
  const client = queryClient || createQueryClient();
  return render(
    <QueryClientProvider client={client}>
      <MantineProvider>
        {ui}
      </MantineProvider>
    </QueryClientProvider>
  );
}

const mockPillarsConfig: StrategyPillarsConfiguration = {
  data: [
    { id: 'pillar-1', name: 'Always On', description: 'Core capabilities', active: true, _links: {} },
    { id: 'pillar-2', name: 'Grow', description: 'Growth initiatives', active: true, _links: {} },
    { id: 'pillar-3', name: 'Transform', description: 'Transformation projects', active: true, _links: {} },
  ],
  version: 1,
  _links: {},
};

describe('StrategyPillarsSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(strategyPillarsApi.getConfiguration).mockResolvedValue(mockPillarsConfig);
  });

  it('renders loading state initially', () => {
    vi.mocked(strategyPillarsApi.getConfiguration).mockImplementation(
      () => new Promise(() => {})
    );

    renderWithProviders(<StrategyPillarsSettings />);

    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('renders all pillars after loading', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByText('Always On')).toBeInTheDocument();
    });

    expect(screen.getByText('Grow')).toBeInTheDocument();
    expect(screen.getByText('Transform')).toBeInTheDocument();
    expect(screen.getByText('Core capabilities')).toBeInTheDocument();
  });

  it('renders edit button in view mode', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });
  });

  it('enters edit mode when edit button is clicked', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    expect(screen.getByTestId('save-pillars-btn')).toBeInTheDocument();
    expect(screen.getByTestId('cancel-pillars-btn')).toBeInTheDocument();
  });

  it('shows input fields in edit mode', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    expect(screen.getByTestId('pillar-name-input-0')).toBeInTheDocument();
    expect(screen.getByTestId('pillar-description-input-0')).toBeInTheDocument();
  });

  it('shows add pillar button in edit mode', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    expect(screen.getByTestId('add-pillar-btn')).toBeInTheDocument();
  });

  it('shows delete buttons for each pillar in edit mode when more than one pillar exists', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    expect(screen.getByTestId('delete-pillar-btn-0')).toBeInTheDocument();
    expect(screen.getByTestId('delete-pillar-btn-1')).toBeInTheDocument();
    expect(screen.getByTestId('delete-pillar-btn-2')).toBeInTheDocument();
  });

  it('cancels edit mode and reverts changes when cancel is clicked', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    const nameInput = screen.getByTestId('pillar-name-input-0') as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'Modified Name' } });
    expect(nameInput.value).toBe('Modified Name');

    fireEvent.click(screen.getByTestId('cancel-pillars-btn'));

    expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    expect(screen.getByText('Always On')).toBeInTheDocument();
  });

  it('validates that pillar name cannot be empty', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    const nameInput = screen.getByTestId('pillar-name-input-0');
    fireEvent.change(nameInput, { target: { value: '' } });

    expect(screen.getByText(/name cannot be empty/i)).toBeInTheDocument();
  });

  it('validates pillar name uniqueness (case insensitive)', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    const nameInput = screen.getByTestId('pillar-name-input-0');
    fireEvent.change(nameInput, { target: { value: 'grow' } });

    expect(screen.getByText(/name must be unique/i)).toBeInTheDocument();
  });

  it('disables save button when validation errors exist', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    const nameInput = screen.getByTestId('pillar-name-input-0');
    fireEvent.change(nameInput, { target: { value: '' } });

    expect(screen.getByTestId('save-pillars-btn')).toBeDisabled();
  });

  it('adds a new pillar when add button is clicked', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));
    fireEvent.click(screen.getByTestId('add-pillar-btn'));

    expect(screen.getByTestId('pillar-name-input-3')).toBeInTheDocument();
  });

  it('disables add pillar button when max pillars (20) reached', async () => {
    const maxPillarsConfig: StrategyPillarsConfiguration = {
      data: Array.from({ length: 20 }, (_, i) => ({
        id: `pillar-${i}`,
        name: `Pillar ${i + 1}`,
        description: '',
        active: true,
        _links: {},
      })),
      version: 1,
      _links: {},
    };
    vi.mocked(strategyPillarsApi.getConfiguration).mockResolvedValue(maxPillarsConfig);

    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    expect(screen.getByTestId('add-pillar-btn')).toBeDisabled();
    expect(screen.getByText(/maximum 20 pillars/i)).toBeInTheDocument();
  });

  it('marks pillar for deletion when delete button is clicked', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));
    fireEvent.click(screen.getByTestId('delete-pillar-btn-0'));

    const pillarRow = screen.getByTestId('pillar-row-0');
    expect(pillarRow).toHaveClass('pillar-marked-for-deletion');
  });

  it('disables delete button when only one active pillar remains', async () => {
    const singlePillarConfig: StrategyPillarsConfiguration = {
      data: [
        { id: 'pillar-1', name: 'Always On', description: '', active: true, _links: {} },
      ],
      version: 1,
      _links: {},
    };
    vi.mocked(strategyPillarsApi.getConfiguration).mockResolvedValue(singlePillarConfig);

    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    expect(screen.getByTestId('delete-pillar-btn-0')).toBeDisabled();
  });

  it('shows error state when loading fails', async () => {
    vi.mocked(strategyPillarsApi.getConfiguration).mockRejectedValue(
      new Error('Failed to load')
    );

    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByText(/failed to load/i)).toBeInTheDocument();
    });
  });

  it('shows conflict dialog when concurrent modification detected', async () => {
    vi.mocked(strategyPillarsApi.getConfiguration).mockResolvedValue(mockPillarsConfig);
    vi.mocked(strategyPillarsApi.batchUpdate).mockRejectedValue(
      new ApiError('Conflict', 409)
    );

    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByTestId('edit-pillars-btn')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('edit-pillars-btn'));

    const nameInput = screen.getByTestId('pillar-name-input-0');
    fireEvent.change(nameInput, { target: { value: 'Modified Name' } });

    fireEvent.click(screen.getByTestId('save-pillars-btn'));

    await waitFor(() => {
      expect(screen.getAllByText(/modified by another user/i)).toHaveLength(2);
    });
  });

  it('shows pillar order numbers', async () => {
    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByText('1.')).toBeInTheDocument();
    });

    expect(screen.getByText('2.')).toBeInTheDocument();
    expect(screen.getByText('3.')).toBeInTheDocument();
  });

  it('filters out inactive pillars in view mode by default', async () => {
    const configWithInactive: StrategyPillarsConfiguration = {
      data: [
        { id: 'pillar-1', name: 'Always On', description: '', active: true, _links: {} },
        { id: 'pillar-2', name: 'Grow', description: '', active: false, _links: {} },
        { id: 'pillar-3', name: 'Transform', description: '', active: true, _links: {} },
      ],
      version: 1,
      _links: {},
    };
    vi.mocked(strategyPillarsApi.getConfiguration).mockResolvedValue(configWithInactive);

    renderWithProviders(<StrategyPillarsSettings />);

    await waitFor(() => {
      expect(screen.getByText('Always On')).toBeInTheDocument();
    });

    expect(screen.getByText('Transform')).toBeInTheDocument();
    expect(screen.queryByText('Grow')).not.toBeInTheDocument();
  });
});
