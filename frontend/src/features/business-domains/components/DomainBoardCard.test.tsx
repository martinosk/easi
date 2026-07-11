import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { buildBusinessDomain, renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import { buildDomainBoardViewModel } from '../hooks/domainBoardViewModel';
import { DomainBoardCard } from './DomainBoardCard';
import classes from './DomainBoardCard.module.css';

const domain = buildBusinessDomain({ name: 'Ferry Freight' });

function buildViewModel() {
  const tree = buildCapabilityTree([
    cap('l1-a', 'Ferry Booking', 'L1'),
    cap('l2-a1', 'Booking Management', 'L2', 'l1-a'),
  ]);
  return buildDomainBoardViewModel({
    domain,
    assignedCapabilities: [cap('l1-a', 'Ferry Booking', 'L1')],
    tree,
    realizationGroups: [],
    isLoading: false,
  });
}

const noop = () => {};

function renderCard(overrides: Partial<React.ComponentProps<typeof DomainBoardCard>> = {}) {
  return renderWithProviders(
    <DomainBoardCard
      viewModel={buildViewModel()}
      searchQuery=""
      selectedCapabilities={new Set()}
      getColorForValue={() => '#000'}
      onCapabilityClick={vi.fn()}
      onCapabilityContextMenu={vi.fn()}
      onChipClick={vi.fn()}
      onDomainMenu={vi.fn()}
      onDragOver={noop}
      onDragLeave={noop}
      onDrop={noop}
      {...overrides}
    />,
    { withRouter: false },
  );
}

describe('DomainBoardCard', () => {
  it('renders the domain name, capability count, and app count', () => {
    renderCard();

    const card = screen.getByTestId(`domain-card-${domain.id}`);
    expect(card).toHaveTextContent('Ferry Freight');
    expect(card).toHaveTextContent('2');
    expect(card).toHaveTextContent('capabilities');
  });

  it('renders one L1Group per assigned L1 capability', () => {
    renderCard();
    expect(screen.getByTestId('l1-group-l1-a')).toBeInTheDocument();
  });

  it('shows an empty state when no capabilities are assigned', () => {
    const emptyViewModel = buildDomainBoardViewModel({
      domain,
      assignedCapabilities: [],
      tree: [],
      realizationGroups: [],
      isLoading: false,
    });
    renderCard({ viewModel: emptyViewModel });

    expect(screen.getByText('No capabilities assigned to this domain yet.')).toBeInTheDocument();
  });

  it('hides non-matching L1 groups entirely while searching', () => {
    const tree = buildCapabilityTree([cap('l1-a', 'Ferry Booking', 'L1'), cap('l1-b', 'Customs Compliance', 'L1')]);
    const vm = buildDomainBoardViewModel({
      domain,
      assignedCapabilities: [cap('l1-a', 'Ferry Booking', 'L1'), cap('l1-b', 'Customs Compliance', 'L1')],
      tree,
      realizationGroups: [],
      isLoading: false,
    });

    renderCard({ viewModel: vm, searchQuery: 'booking' });

    expect(screen.getByTestId('l1-group-l1-a')).toBeInTheDocument();
    expect(screen.queryByTestId('l1-group-l1-b')).not.toBeInTheDocument();
  });

  it('shows a no-matches message when search excludes every group', () => {
    renderCard({ searchQuery: 'zzz-no-match' });
    expect(screen.getByText('No matches in this domain.')).toBeInTheDocument();
  });

  it('calls onDomainMenu when the quiet actions button is clicked', async () => {
    const onDomainMenu = vi.fn();
    renderCard({ onDomainMenu });

    await userEvent.click(screen.getByRole('button', { name: `Actions for ${domain.name}` }));

    expect(onDomainMenu).toHaveBeenCalledWith(expect.anything(), domain);
  });

  it('calls onDomainMenu on right-click anywhere on the card', () => {
    const onDomainMenu = vi.fn();
    renderCard({ onDomainMenu });

    const card = screen.getByTestId(`domain-card-${domain.id}`);
    card.dispatchEvent(new MouseEvent('contextmenu', { bubbles: true, cancelable: true }));

    expect(onDomainMenu).toHaveBeenCalledWith(expect.anything(), domain);
  });

  it('applies the drop-target style when isDropTarget is true', () => {
    renderCard({ isDropTarget: true });
    expect(screen.getByTestId(`domain-card-${domain.id}`).className).toContain(classes.dropTarget);
  });

  it('applies the highlighted style when isHighlighted is true (deep-link scroll target)', () => {
    renderCard({ isHighlighted: true });
    expect(screen.getByTestId(`domain-card-${domain.id}`).className).toContain(classes.highlighted);
  });
});
