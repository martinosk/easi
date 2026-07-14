import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { toCapabilityId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap, buildCapabilityRealization } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import type { BoardLens } from '../lens/boardLens';
import { buildJourneyIndex } from '../lens/journeyIndex';
import { BoardLensProvider } from './BoardLensContext';
import { L1Group, type L1GroupProps } from './L1Group';
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

function renderInLens(
  node: ReturnType<typeof buildCapabilityTree>[0],
  lens: BoardLens,
  props: Partial<L1GroupProps> = {},
) {
  const index = buildJourneyIndex({ journeys: [], capabilityDomainNames: new Map() });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <BoardLensProvider lens={lens} changesOnly={false} index={index} openCapabilityById={vi.fn()}>
      {children}
    </BoardLensProvider>
  );
  return renderWithProviders(wrapper({ children: <L1Group node={node} {...defaultProps} {...props} /> }), {
    withRouter: false,
  });
}

describe('L1Group', () => {
  it('is collapsed by default, showing a compact header with the name, sub count, and app count', () => {
    const [node] = buildL1Tree();
    renderInLens(node, 'now');

    const group = screen.getByTestId('l1-group-l1-a');
    expect(group).toHaveTextContent('Ferry Booking');
    expect(group).toHaveTextContent('L1 · 2 sub');
    expect(group).toHaveTextContent('2 apps');
    expect(screen.queryByTestId('capability-card-l1-a')).not.toBeInTheDocument();
  });

  it('expands (rather than opening the drawer) when the collapsed header is clicked', async () => {
    const [node] = buildL1Tree();
    const own = buildCapabilityRealization({ capabilityId: toCapabilityId('l1-a'), componentName: 'Seabook' });
    const getRealizationsForCapability = (id: ReturnType<typeof toCapabilityId>) =>
      id === toCapabilityId('l1-a') ? [own] : [];
    const onCapabilityClick = vi.fn();
    renderInLens(node, 'now', { getRealizationsForCapability, onCapabilityClick });

    await userEvent.click(screen.getByTestId('l1-toggle-l1-a'));

    const card = screen.getByTestId('capability-card-l1-a');
    expect(card).toHaveTextContent('Seabook');
    expect(screen.getByTestId('capability-card-l2-a1')).toBeInTheDocument();
    expect(screen.getByTestId('capability-card-l2-a2')).toBeInTheDocument();
    expect(onCapabilityClick).not.toHaveBeenCalled();
  });

  it('opens the drawer via onCapabilityClick when the expanded card is clicked', async () => {
    const [node] = buildL1Tree();
    const onCapabilityClick = vi.fn();
    renderInLens(node, 'now', { forceOpen: true, onCapabilityClick });

    await userEvent.click(screen.getByTestId('capability-card-l1-a'));

    expect(onCapabilityClick).toHaveBeenCalledWith(node.capability, expect.anything());
  });

  it('collapses the full card back to the compact header via the toggle', async () => {
    const [node] = buildL1Tree();
    renderInLens(node, 'now');

    await userEvent.click(screen.getByTestId('l1-toggle-l1-a'));
    expect(screen.getByTestId('capability-card-l1-a')).toBeInTheDocument();

    await userEvent.click(screen.getByTestId('l1-toggle-l1-a'));
    expect(screen.queryByTestId('capability-card-l1-a')).not.toBeInTheDocument();
  });

  it('shows a childless L1 collapsed too, expanding into its own card with no children', async () => {
    const [node] = buildCapabilityTree([cap('l1-leaf', 'Invoicing', 'L1')]);
    renderInLens(node, 'now', { distinctAppCount: 0 });

    expect(screen.getByTestId('l1-toggle-l1-leaf')).toBeInTheDocument();
    expect(screen.queryByTestId('capability-card-l1-leaf')).not.toBeInTheDocument();

    await userEvent.click(screen.getByTestId('l1-toggle-l1-leaf'));

    expect(screen.getByTestId('capability-card-l1-leaf')).toBeInTheDocument();
  });

  it('flags the app count with the consolidation signal class when more than 3 apps realise the group', () => {
    const [node] = buildL1Tree();
    renderInLens(node, 'now', { distinctAppCount: 4 });

    expect(screen.getByText('4 apps').className).toContain(classes.appCountMulti);
  });

  it('does not apply the consolidation signal class at 3 or fewer apps', () => {
    const [node] = buildL1Tree();
    renderInLens(node, 'now', { distinctAppCount: 3 });

    expect(screen.getByText('3 apps').className).not.toContain(classes.appCountMulti);
  });

  it('auto-expands and filters to matching child cards while a search query is active', () => {
    const [node] = buildL1Tree();
    renderInLens(node, 'now', { searchQuery: 'capacity' });

    expect(screen.getByTestId('capability-card-l2-a2')).toBeInTheDocument();
    expect(screen.queryByTestId('capability-card-l2-a1')).not.toBeInTheDocument();
  });

  it('force-opens into the full card via the forceOpen prop (e.g. for a deep-linked capability)', () => {
    const [node] = buildL1Tree();
    renderInLens(node, 'now', { forceOpen: true });

    expect(screen.getByTestId('capability-card-l1-a')).toBeInTheDocument();
    expect(screen.getByTestId('capability-card-l2-a1')).toBeInTheDocument();
  });

  it('carries the collapse / expand behaviour into the Journey lens', async () => {
    const [node] = buildL1Tree();
    renderInLens(node, 'journey');

    expect(screen.queryByTestId('capability-card-l1-a')).not.toBeInTheDocument();

    await userEvent.click(screen.getByTestId('l1-toggle-l1-a'));

    expect(screen.getByTestId('capability-card-l1-a')).toBeInTheDocument();
  });
});
