import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { toCapabilityId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap, buildCapabilityRealization } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import { L1Group } from './L1Group';
import classes from './L1Group.module.css';

function buildL1Tree() {
  return buildCapabilityTree([
    cap('l1-a', 'Ferry Booking', 'L1'),
    cap('l2-a1', 'Booking Management', 'L2', 'l1-a'),
    cap('l2-a2', 'Booked Capacity', 'L2', 'l1-a'),
  ]);
}

const defaultProps = {
  distinctAppCount: 2,
  searchQuery: '',
  selectedCapabilities: new Set<ReturnType<typeof toCapabilityId>>(),
  getColorForValue: () => '#000',
  getRealizationsForCapability: () => [],
  onCapabilityClick: vi.fn(),
  onCapabilityContextMenu: vi.fn(),
  onChipClick: vi.fn(),
};

describe('L1Group', () => {
  it('renders the L1 name, sub-capability count tag, and distinct app count', () => {
    const [node] = buildL1Tree();
    renderWithProviders(<L1Group node={node} {...defaultProps} />, { withRouter: false });

    const group = screen.getByTestId('l1-group-l1-a');
    expect(group).toHaveTextContent('Ferry Booking');
    expect(group).toHaveTextContent('L1 · 2 sub');
    expect(group).toHaveTextContent('2 apps');
  });

  it('is collapsed by default and expands the L2 cards via the chevron toggle', async () => {
    const [node] = buildL1Tree();
    renderWithProviders(<L1Group node={node} {...defaultProps} />, { withRouter: false });

    expect(screen.queryByTestId('capability-card-l2-a1')).not.toBeInTheDocument();

    await userEvent.click(screen.getByTestId('l1-toggle-l1-a'));

    expect(screen.getByTestId('capability-card-l2-a1')).toBeInTheDocument();
    expect(screen.getByTestId('capability-card-l2-a2')).toBeInTheDocument();

    await userEvent.click(screen.getByTestId('l1-toggle-l1-a'));
    expect(screen.queryByTestId('capability-card-l2-a1')).not.toBeInTheDocument();
  });

  it('opens the drawer via onCapabilityClick when the L1 label is clicked, without expanding', async () => {
    const [node] = buildL1Tree();
    const onCapabilityClick = vi.fn();
    renderWithProviders(<L1Group node={node} {...defaultProps} onCapabilityClick={onCapabilityClick} />, {
      withRouter: false,
    });

    await userEvent.click(screen.getByTestId('l1-open-l1-a'));

    expect(onCapabilityClick).toHaveBeenCalledWith(node.capability, expect.anything());
    expect(screen.queryByTestId('capability-card-l2-a1')).not.toBeInTheDocument();
  });

  it('highlights the app count with the consolidation signal class when more than 3 apps realise the group', () => {
    const [node] = buildL1Tree();
    renderWithProviders(<L1Group node={node} {...defaultProps} distinctAppCount={4} />, { withRouter: false });

    expect(screen.getByText('4 apps').className).toContain(classes.appCountMulti);
  });

  it('does not apply the consolidation signal class at 3 or fewer apps', () => {
    const [node] = buildL1Tree();
    renderWithProviders(<L1Group node={node} {...defaultProps} distinctAppCount={3} />, { withRouter: false });

    expect(screen.getByText('3 apps').className).not.toContain(classes.appCountMulti);
  });

  it('auto-expands and filters to matching L2 cards while a search query is active', () => {
    const [node] = buildL1Tree();
    renderWithProviders(<L1Group node={node} {...defaultProps} searchQuery="capacity" />, { withRouter: false });

    expect(screen.getByTestId('capability-card-l2-a2')).toBeInTheDocument();
    expect(screen.queryByTestId('capability-card-l2-a1')).not.toBeInTheDocument();
  });

  it('renders the L1 itself as a capability card (no expander) when it has no L2 children', () => {
    const [node] = buildCapabilityTree([cap('l1-leaf', 'Invoicing', 'L1')]);
    renderWithProviders(<L1Group node={node} {...defaultProps} distinctAppCount={0} />, { withRouter: false });

    expect(screen.getByTestId('capability-card-l1-leaf')).toBeInTheDocument();
    expect(screen.queryByTestId('l1-toggle-l1-leaf')).not.toBeInTheDocument();
  });

  it('opens the drawer when a leaf L1 card is clicked', async () => {
    const [node] = buildCapabilityTree([cap('l1-leaf', 'Invoicing', 'L1')]);
    const onCapabilityClick = vi.fn();
    renderWithProviders(
      <L1Group node={node} {...defaultProps} distinctAppCount={0} onCapabilityClick={onCapabilityClick} />,
      { withRouter: false },
    );

    await userEvent.click(screen.getByTestId('capability-card-l1-leaf'));

    expect(onCapabilityClick).toHaveBeenCalledWith(node.capability, expect.anything());
  });

  it('shows the L1 own realising applications as chips above the L2 cards when expanded', () => {
    const [node] = buildL1Tree();
    const ownRealization = buildCapabilityRealization({ capabilityId: toCapabilityId('l1-a'), componentName: 'Seabook' });
    const getRealizationsForCapability = (id: ReturnType<typeof toCapabilityId>) =>
      id === toCapabilityId('l1-a') ? [ownRealization] : [];
    renderWithProviders(
      <L1Group node={node} {...defaultProps} getRealizationsForCapability={getRealizationsForCapability} forceOpen />,
      { withRouter: false },
    );

    expect(screen.getByTestId('l1-own-apps-l1-a')).toHaveTextContent('Seabook');
  });

  it('does not render an own-apps row when the L1 has no realizations of its own', () => {
    const [node] = buildL1Tree();
    renderWithProviders(<L1Group node={node} {...defaultProps} forceOpen />, { withRouter: false });

    expect(screen.queryByTestId('l1-own-apps-l1-a')).not.toBeInTheDocument();
  });

  it('force-opens via the forceOpen prop (e.g. for a deep-linked capability)', () => {
    const [node] = buildL1Tree();
    renderWithProviders(<L1Group node={node} {...defaultProps} forceOpen />, { withRouter: false });

    expect(screen.getByTestId('capability-card-l2-a1')).toBeInTheDocument();
  });
});
