import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { toComponentId } from '../../../api/types';
import { buildBusinessDomain, buildCapabilityRealization, renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { CapabilityDrawer } from './CapabilityDrawer';

vi.mock('../../../hooks/useStrategyPillarsSettings', () => ({
  useStrategyPillarsConfig: () => ({ data: { data: [] } }),
}));

vi.mock('../hooks/useStrategyImportance', () => ({
  useStrategyImportanceByDomainAndCapability: () => ({ data: { data: [] }, isLoading: false }),
  useSetStrategyImportance: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateStrategyImportance: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useRemoveStrategyImportance: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

const domain = buildBusinessDomain({ name: 'Ferry Freight' });

describe('CapabilityDrawer', () => {
  it('renders no capability content when no capability is selected', () => {
    renderWithProviders(
      <CapabilityDrawer
        capability={null}
        domain={null}
        l1Name={null}
        getRealizationsForCapability={() => []}
        onClose={vi.fn()}
        onChipClick={vi.fn()}
      />,
    );

    expect(screen.queryByRole('heading')).not.toBeInTheDocument();
    expect(screen.queryByText('no realising application mapped')).not.toBeInTheDocument();
  });

  it('renders the breadcrumb, name, level, and empty realisations state', () => {
    const capability = cap('l2-a', 'Booking Management', 'L2');
    renderWithProviders(
      <CapabilityDrawer
        capability={capability}
        domain={domain}
        l1Name="Ferry Booking"
        getRealizationsForCapability={() => []}
        onClose={vi.fn()}
        onChipClick={vi.fn()}
      />,
    );

    expect(screen.getByTestId('capability-drawer')).toHaveTextContent('Ferry Freight');
    expect(screen.getByTestId('capability-drawer')).toHaveTextContent('Ferry Booking');
    expect(screen.getByRole('heading', { name: 'Booking Management' })).toBeInTheDocument();
    expect(screen.getByText('L2')).toBeInTheDocument();
    expect(screen.getByText('no realising application mapped')).toBeInTheDocument();
  });

  it('renders a row per realising application with level, origin note, and notes', () => {
    const capability = cap('l2-a', 'Booking Management', 'L2');
    const realizations = [
      buildCapabilityRealization({
        componentId: toComponentId('comp-1'),
        componentName: 'Phoenix',
        realizationLevel: 'Full',
        origin: 'Direct',
        notes: 'Primary booking engine',
      }),
      buildCapabilityRealization({
        componentId: toComponentId('comp-2'),
        componentName: 'Seabook',
        realizationLevel: 'Partial',
        origin: 'Inherited',
        sourceCapabilityName: 'Ferry Booking',
      }),
    ];

    renderWithProviders(
      <CapabilityDrawer
        capability={capability}
        domain={domain}
        l1Name="Ferry Booking"
        getRealizationsForCapability={() => realizations}
        onClose={vi.fn()}
        onChipClick={vi.fn()}
      />,
    );

    expect(screen.getByTestId('drawer-realization-real-1')).toHaveTextContent('Primary booking engine');
    expect(screen.getByTestId('drawer-realization-real-2')).toHaveTextContent('Inherited from Ferry Booking');
  });

  it('calls onChipClick when a realising application chip is clicked', async () => {
    const capability = cap('l2-a', 'Booking Management', 'L2');
    const onChipClick = vi.fn();
    const realizations = [
      buildCapabilityRealization({ componentId: toComponentId('comp-1'), componentName: 'Phoenix' }),
    ];

    renderWithProviders(
      <CapabilityDrawer
        capability={capability}
        domain={domain}
        l1Name="Ferry Booking"
        getRealizationsForCapability={() => realizations}
        onClose={vi.fn()}
        onChipClick={onChipClick}
      />,
    );

    await userEvent.click(screen.getByTestId('app-chip-comp-1'));

    expect(onChipClick).toHaveBeenCalledWith('comp-1');
  });

  it('renders description, tags and owners under Details', () => {
    const capability = {
      ...cap('l2-a', 'Booking Management', 'L2'),
      description: 'Handles route bookings',
      primaryOwner: 'Jane Doe',
      tags: ['core', 'freight'],
    };

    renderWithProviders(
      <CapabilityDrawer
        capability={capability}
        domain={domain}
        l1Name="Ferry Booking"
        getRealizationsForCapability={() => []}
        onClose={vi.fn()}
        onChipClick={vi.fn()}
      />,
    );

    expect(screen.getByText('Handles route bookings')).toBeInTheDocument();
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
    expect(screen.getByText('core')).toBeInTheDocument();
    expect(screen.getByText('freight')).toBeInTheDocument();
  });
});
