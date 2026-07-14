import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { buildCapabilityRealization, renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import type { BoardCapabilityCardProps } from './BoardCapabilityCard';
import { BoardCapabilityCard } from './BoardCapabilityCard';

function buildNode(overrides: Parameters<typeof cap> = ['l2-a', 'Booking Management', 'L2']) {
  const [tree] = buildCapabilityTree([cap(...overrides)], { orphanRoots: 'any-level' });
  return tree;
}

function renderCard(overrides: Partial<BoardCapabilityCardProps> = {}) {
  return renderWithProviders(
    <BoardCapabilityCard
      node={buildNode()}
      isSelected={false}
      getColorForValue={() => '#000'}
      getRealizationsForCapability={() => []}
      onClick={vi.fn()}
      onContextMenu={vi.fn()}
      onChipClick={vi.fn()}
      {...overrides}
    />,
    { withRouter: false },
  );
}

describe('BoardCapabilityCard', () => {
  it('renders the capability name and level tag', () => {
    renderCard();

    expect(screen.getByTestId('capability-card-l2-a')).toHaveTextContent('Booking Management');
    expect(screen.getByTestId('capability-card-l2-a')).toHaveTextContent('L2');
  });

  it('shows the italic empty state when no realising applications are mapped', () => {
    renderCard();

    expect(screen.getByTestId('capability-card-empty-realizations')).toHaveTextContent(
      'no realising application mapped',
    );
  });

  it('renders app chips and a "n apps" flag when there is more than one realising application', () => {
    const realizations = [
      buildCapabilityRealization({ componentId: toComponentId('comp-1'), componentName: 'Phoenix' }),
      buildCapabilityRealization({ componentId: toComponentId('comp-2'), componentName: 'Seabook' }),
    ];

    renderCard({ getRealizationsForCapability: () => realizations });

    expect(screen.getByTestId('app-chip-comp-1')).toBeInTheDocument();
    expect(screen.getByTestId('app-chip-comp-2')).toBeInTheDocument();
    expect(screen.getByText('2 apps')).toBeInTheDocument();
  });

  it('does not render a "n apps" flag for a single realisation', () => {
    const realizations = [buildCapabilityRealization({ componentId: toComponentId('comp-1') })];

    renderCard({ getRealizationsForCapability: () => realizations });

    expect(screen.queryByText(/apps$/)).not.toBeInTheDocument();
  });

  it('calls onClick with the capability when the card is clicked', async () => {
    const onClick = vi.fn();
    renderCard({ onClick });

    await userEvent.click(screen.getByTestId('capability-card-l2-a'));

    expect(onClick).toHaveBeenCalledWith(expect.objectContaining({ id: 'l2-a' }), expect.anything());
  });

  it('clicking an app chip does not also trigger the card onClick', async () => {
    const onClick = vi.fn();
    const onChipClick = vi.fn();
    const realizations = [buildCapabilityRealization({ componentId: toComponentId('comp-1') })];

    renderCard({ getRealizationsForCapability: () => realizations, onClick, onChipClick });

    await userEvent.click(screen.getByTestId('app-chip-comp-1'));

    expect(onChipClick).toHaveBeenCalledWith('comp-1');
    expect(onClick).not.toHaveBeenCalled();
  });

  it('reveals L3/L4 descendants behind an expander, each with its own app chips', async () => {
    const [node] = buildCapabilityTree(
      [
        cap('l2-a', 'Booking Management', 'L2'),
        cap('l3-a1', 'Quotation', 'L3', 'l2-a'),
        cap('l4-a1a', 'Rate Lookup', 'L4', 'l3-a1'),
      ],
      { orphanRoots: 'any-level' },
    );
    const l4Realization = buildCapabilityRealization({
      capabilityId: toCapabilityId('l4-a1a'),
      componentId: toComponentId('comp-9'),
    });

    renderCard({ node, getRealizationsForCapability: (id) => (id === 'l4-a1a' ? [l4Realization] : []) });

    expect(screen.getByText('2 sub-capabilities')).toBeInTheDocument();
    expect(screen.queryByTestId('capability-card-l3-a1')).not.toBeInTheDocument();

    await userEvent.click(screen.getByText('2 sub-capabilities'));

    expect(screen.getByTestId('capability-card-l3-a1')).toHaveTextContent('Quotation');
    expect(screen.getByTestId('capability-card-l4-a1a')).toHaveTextContent('Rate Lookup');
    expect(screen.getByTestId('app-chip-comp-9')).toBeInTheDocument();
  });

  it('opens the clicked descendant, not the parent card, when a sub-capability row is clicked', async () => {
    const [node] = buildCapabilityTree(
      [cap('l2-a', 'Booking Management', 'L2'), cap('l3-a1', 'Quotation', 'L3', 'l2-a')],
      { orphanRoots: 'any-level' },
    );
    const onClick = vi.fn();

    renderCard({ node, onClick });
    await userEvent.click(screen.getByText('1 sub-capability'));
    await userEvent.click(screen.getByTestId('capability-card-l3-a1'));

    expect(onClick).toHaveBeenCalledTimes(1);
    expect(onClick).toHaveBeenCalledWith(expect.objectContaining({ id: toCapabilityId('l3-a1') }), expect.anything());
  });

  it('renders a custom sub-capability slot inside the card in place of the built-in expander', () => {
    const [node] = buildCapabilityTree(
      [cap('l2-a', 'Booking Management', 'L2'), cap('l3-a1', 'Quotation', 'L3', 'l2-a')],
      { orphanRoots: 'any-level' },
    );

    renderCard({
      node,
      subCapabilities: (
        <button type="button" data-testid="custom-slot">
          custom
        </button>
      ),
    });

    expect(screen.getByTestId('capability-card-l2-a')).toContainElement(screen.getByTestId('custom-slot'));
    expect(screen.queryByText('1 sub-capability')).not.toBeInTheDocument();
  });

  it('applies a maturity-derived left border colour when maturityValue is present', () => {
    const [node] = buildCapabilityTree([{ ...cap('l2-a', 'Booking Management', 'L2'), maturityValue: 42 }], {
      orphanRoots: 'any-level',
    });
    const getColorForValue = vi.fn().mockReturnValue('rgb(1, 2, 3)');

    renderCard({ node, getColorForValue });

    expect(getColorForValue).toHaveBeenCalledWith(42);
    expect(screen.getByTestId('capability-card-l2-a')).toHaveStyle({ borderLeftColor: 'rgb(1, 2, 3)' });
  });

  it('marks the card selected via data-selected when isSelected is true', () => {
    renderCard({ isSelected: true });

    expect(screen.getByTestId('capability-card-l2-a')).toHaveAttribute('data-selected', 'true');
  });
});
