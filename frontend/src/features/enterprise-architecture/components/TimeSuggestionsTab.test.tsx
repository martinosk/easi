import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MantineProvider } from '@mantine/core';
import { TimeSuggestionsTab } from './TimeSuggestionsTab';
import * as useTimeSuggestionsModule from '../hooks/useTimeSuggestions';
import type { TimeSuggestion } from '../types';

vi.mock('../hooks/useTimeSuggestions');

const mockUseTimeSuggestions = vi.mocked(useTimeSuggestionsModule.useTimeSuggestions);

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <MantineProvider>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </MantineProvider>
  );
}

const mockSuggestions: TimeSuggestion[] = [
  {
    capabilityId: 'cap-1',
    capabilityName: 'Customer Management',
    componentId: 'comp-1',
    componentName: 'CRM System',
    suggestedTime: 'Tolerate',
    technicalGap: -5,
    functionalGap: 2,
  },
  {
    capabilityId: 'cap-1',
    capabilityName: 'Customer Management',
    componentId: 'comp-2',
    componentName: 'Legacy CRM',
    suggestedTime: 'Eliminate',
    technicalGap: 25,
    functionalGap: 30,
  },
  {
    capabilityId: 'cap-2',
    capabilityName: 'Order Processing',
    componentId: 'comp-3',
    componentName: 'Order Service',
    suggestedTime: 'Invest',
    technicalGap: 15,
    functionalGap: -3,
  },
  {
    capabilityId: 'cap-3',
    capabilityName: 'Inventory',
    componentId: 'comp-4',
    componentName: 'Warehouse System',
    suggestedTime: null,
    technicalGap: null,
    functionalGap: 10,
  },
];

describe('TimeSuggestionsTab', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state', () => {
    mockUseTimeSuggestions.mockReturnValue({
      suggestions: [],
      isLoading: true,
      error: null,
      refetch: vi.fn(),
    });

    render(<TimeSuggestionsTab />, { wrapper: createWrapper() });

    expect(screen.getByText('Loading TIME suggestions...')).toBeInTheDocument();
  });

  it('renders error state', () => {
    mockUseTimeSuggestions.mockReturnValue({
      suggestions: [],
      isLoading: false,
      error: new Error('Network error'),
      refetch: vi.fn(),
    });

    render(<TimeSuggestionsTab />, { wrapper: createWrapper() });

    expect(screen.getByText(/Failed to load TIME suggestions/)).toBeInTheDocument();
    expect(screen.getByText(/Network error/)).toBeInTheDocument();
  });

  it('renders empty state when no suggestions', () => {
    mockUseTimeSuggestions.mockReturnValue({
      suggestions: [],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(<TimeSuggestionsTab />, { wrapper: createWrapper() });

    expect(screen.getByText('No TIME Suggestions Available')).toBeInTheDocument();
  });

  it('renders suggestions table with data', () => {
    mockUseTimeSuggestions.mockReturnValue({
      suggestions: mockSuggestions,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(<TimeSuggestionsTab />, { wrapper: createWrapper() });

    expect(screen.getAllByText('Customer Management')).toHaveLength(2);
    expect(screen.getByText('CRM System')).toBeInTheDocument();
    expect(screen.getByText('Legacy CRM')).toBeInTheDocument();
    expect(screen.getByText('Order Processing')).toBeInTheDocument();
    expect(screen.getByText('Inventory')).toBeInTheDocument();
  });

  it('displays TIME badges correctly', () => {
    mockUseTimeSuggestions.mockReturnValue({
      suggestions: mockSuggestions,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(<TimeSuggestionsTab />, { wrapper: createWrapper() });

    expect(screen.getAllByText('Tolerate').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Eliminate').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Invest').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('N/A')).toBeInTheDocument();
  });

  it('displays summary statistics correctly', () => {
    mockUseTimeSuggestions.mockReturnValue({
      suggestions: mockSuggestions,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(<TimeSuggestionsTab />, { wrapper: createWrapper() });

    expect(screen.getByText('4')).toBeInTheDocument();
    expect(screen.getByText('Total Realizations')).toBeInTheDocument();
  });

  it('renders TIME legend with all classifications', () => {
    mockUseTimeSuggestions.mockReturnValue({
      suggestions: mockSuggestions,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(<TimeSuggestionsTab />, { wrapper: createWrapper() });

    expect(screen.getByText('TIME Classifications')).toBeInTheDocument();
    expect(screen.getByText('Tolerate')).toBeInTheDocument();
    expect(screen.getByText('Invest')).toBeInTheDocument();
    expect(screen.getByText('Migrate')).toBeInTheDocument();
    expect(screen.getByText('Eliminate')).toBeInTheDocument();
  });
});
