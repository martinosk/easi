import { fireEvent, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../../test/helpers/entityBuilders';
import type { CapabilityTreeNode } from '../../hooks/useCapabilityTree';
import { buildCapabilityTree } from '../../hooks/useCapabilityTree';
import { CapabilityTree } from './CapabilityTree';

const tree = buildCapabilityTree([
  cap('a', 'Customer Management', 'L1'),
  cap('a1', 'Onboarding', 'L2', 'a'),
  cap('a1x', 'Identity Checks', 'L3', 'a1'),
  cap('b', 'Billing', 'L1'),
]);

const renderTree = (ui: React.ReactElement) => renderWithProviders(ui, { withRouter: false });

describe('CapabilityTree', () => {
  it('renders only collapsed roots with level labels by default', () => {
    renderTree(<CapabilityTree tree={tree} />);

    expect(screen.getByText('Customer Management')).toBeInTheDocument();
    expect(screen.getByText('Billing')).toBeInTheDocument();
    expect(screen.queryByText('Onboarding')).not.toBeInTheDocument();
    expect(screen.getAllByText('L1')).toHaveLength(2);
    expect(screen.queryByText('L2')).not.toBeInTheDocument();
  });

  it('expands and collapses nodes via their toggles', async () => {
    renderTree(<CapabilityTree tree={tree} />);

    const rootRow = screen.getByText('Customer Management').closest('div')!;
    await userEvent.click(within(rootRow).getByLabelText('Expand'));
    expect(screen.getByText('Onboarding')).toBeInTheDocument();

    const onboardingRow = screen.getByText('Onboarding').closest('div')!;
    await userEvent.click(within(onboardingRow).getByLabelText('Expand'));
    expect(screen.getByText('Identity Checks')).toBeInTheDocument();

    await userEvent.click(within(onboardingRow).getByLabelText('Collapse'));
    expect(screen.queryByText('Identity Checks')).not.toBeInTheDocument();
  });

  it('filters by name, keeps ancestors, auto-expands, and bolds the match', async () => {
    renderTree(<CapabilityTree tree={tree} />);

    await userEvent.type(screen.getByPlaceholderText('Search capabilities...'), 'identity');

    expect(screen.queryByText('Billing')).not.toBeInTheDocument();
    expect(screen.getByText('Customer Management')).toBeInTheDocument();
    expect(screen.getByText('Onboarding')).toBeInTheDocument();

    const mark = screen.getByText('Identity');
    expect(mark.tagName).toBe('MARK');
    expect(mark).toHaveStyle({ fontWeight: 700 });
    expect(mark).toHaveStyle({ backgroundColor: 'transparent' });
  });

  it('shows a no-match message when the search finds nothing', async () => {
    renderTree(<CapabilityTree tree={tree} />);

    await userEvent.type(screen.getByPlaceholderText('Search capabilities...'), 'zzz');

    expect(screen.getByText('No matches')).toBeInTheDocument();
  });

  it('shows an empty message when there are no capabilities', () => {
    renderTree(<CapabilityTree tree={[]} />);

    expect(screen.getByText('No capabilities')).toBeInTheDocument();
  });

  it('shows a loading state', () => {
    renderTree(<CapabilityTree tree={[]} isLoading />);

    expect(screen.getByText('Loading capabilities...')).toBeInTheDocument();
  });

  it('applies row props from getRowProps and fires row handlers', async () => {
    const onClick = vi.fn();
    const onDragStart = vi.fn();
    const onContextMenu = vi.fn();
    renderTree(
      <CapabilityTree
        tree={tree}
        getRowProps={(node) => ({
          draggable: node.capability.level === 'L1',
          testId: `row-${node.capability.id}`,
          title: node.capability.name,
          selected: node.capability.id === 'a',
          dimmed: node.capability.id === 'b',
          onClick,
          onDragStart,
          onContextMenu,
        })}
      />,
    );

    const rowA = screen.getByTestId('row-a');
    await userEvent.click(within(rowA).getByLabelText('Expand'));
    const rowA1 = screen.getByTestId('row-a1');
    expect(rowA).toHaveAttribute('draggable', 'true');
    expect(rowA1).toHaveAttribute('draggable', 'false');
    expect(rowA).toHaveAttribute('title', 'Customer Management');
    expect(rowA.className).toContain('rowSelected');
    expect(screen.getByTestId('row-b').className).toContain('rowDimmed');

    await userEvent.click(rowA);
    expect(onClick).toHaveBeenCalled();
    fireEvent.contextMenu(rowA);
    expect(onContextMenu).toHaveBeenCalled();
    fireEvent.dragStart(rowA);
    expect(onDragStart).toHaveBeenCalled();
  });

  it('renders per-row content from renderRight', () => {
    renderTree(
      <CapabilityTree
        tree={tree}
        renderRight={(node) => (node.capability.id === 'b' ? <span>Mapped</span> : null)}
      />,
    );

    expect(screen.getByText('Mapped')).toBeInTheDocument();
  });

  it('supports controlled expansion', async () => {
    const onToggleExpanded = vi.fn();
    renderTree(<CapabilityTree tree={tree} expandedIds={new Set(['a', 'a1'])} onToggleExpanded={onToggleExpanded} />);

    expect(screen.getByText('Identity Checks')).toBeInTheDocument();

    const onboardingRow = screen.getByText('Onboarding').closest('div')!;
    await userEvent.click(within(onboardingRow).getByLabelText('Collapse'));
    expect(onToggleExpanded).toHaveBeenCalledWith('a1');
    expect(screen.getByText('Identity Checks')).toBeInTheDocument();
  });

  it('reports visible nodes in document order as expansion changes', async () => {
    const onVisibleNodesChange = vi.fn();
    renderTree(<CapabilityTree tree={tree} onVisibleNodesChange={onVisibleNodesChange} />);

    const visibleIds = (calls: CapabilityTreeNode[][][]) =>
      calls[calls.length - 1][0].map((n) => n.capability.id as string);

    expect(visibleIds(onVisibleNodesChange.mock.calls)).toEqual(['b', 'a']);

    const rootRow = screen.getByText('Customer Management').closest('div')!;
    await userEvent.click(within(rootRow).getByLabelText('Expand'));

    expect(visibleIds(onVisibleNodesChange.mock.calls)).toEqual(['b', 'a', 'a1']);
  });

  it('activates a clickable row with the keyboard', async () => {
    const onClick = vi.fn();
    renderTree(<CapabilityTree tree={tree} getRowProps={() => ({ onClick })} />);

    const row = screen.getAllByRole('treeitem')[0];
    row.focus();
    await userEvent.keyboard('{Enter}');

    expect(onClick).toHaveBeenCalled();
  });

  it('exposes container and search test ids and a root className', () => {
    renderTree(
      <CapabilityTree tree={tree} testId="capability-sidebar" searchTestId="capability-filter" className="custom" />,
    );

    expect(screen.getByTestId('capability-sidebar')).toBeInTheDocument();
    expect(screen.getByTestId('capability-filter')).toBeInTheDocument();
    expect(screen.getByTestId('capability-sidebar').className).toContain('custom');
  });
});
