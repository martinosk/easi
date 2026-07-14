import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { BoardToolbar, type BoardToolbarProps } from './BoardToolbar';

function baseProps(overrides: Partial<BoardToolbarProps> = {}): BoardToolbarProps {
  return {
    searchQuery: '',
    onSearchChange: vi.fn(),
    canCreateDomain: true,
    onCreateDomain: vi.fn(),
    showAssignToggle: true,
    assignRailOpen: false,
    onToggleAssignRail: vi.fn(),
    lens: 'now',
    onLensChange: vi.fn(),
    changesOnly: false,
    onChangesOnlyChange: vi.fn(),
    summary: { settled: 0, inFlight: 0, notStarted: 0 },
    ...overrides,
  };
}

function BoardToolbarHarness(overrides: Partial<BoardToolbarProps>) {
  return <BoardToolbar {...baseProps(overrides)} />;
}

function renderToolbar(overrides: Partial<BoardToolbarProps> = {}) {
  const props = baseProps(overrides);
  return { props, ...renderWithProviders(<BoardToolbar {...props} />, { withRouter: false }) };
}

describe('BoardToolbar', () => {
  it('renders the search input and reports changes', async () => {
    const onSearchChange = vi.fn();
    renderToolbar({ onSearchChange });

    await userEvent.type(screen.getByTestId('board-search-input'), 'phoenix');

    expect(onSearchChange).toHaveBeenCalled();
  });

  it('renders the Full / Partial / Planned / Inherited legend', () => {
    renderToolbar();

    const legend = screen.getByTestId('board-legend');
    expect(legend).toHaveTextContent('Full');
    expect(legend).toHaveTextContent('Partial');
    expect(legend).toHaveTextContent('Planned');
    expect(legend).toHaveTextContent('Inherited');
  });

  it('shows New domain only when canCreateDomain is true', () => {
    const { rerender } = renderToolbar({ canCreateDomain: true });
    expect(screen.getByTestId('create-domain-button')).toBeInTheDocument();

    rerender(<BoardToolbarHarness canCreateDomain={false} />);
    expect(screen.queryByTestId('create-domain-button')).not.toBeInTheDocument();
  });

  it('offers Now, Journey, and Target lenses and reports a change', async () => {
    const onLensChange = vi.fn();
    renderToolbar({ onLensChange });

    expect(screen.getByTestId('lens-switcher')).toHaveTextContent('Now');
    expect(screen.getByTestId('lens-switcher')).toHaveTextContent('Journey');
    expect(screen.getByTestId('lens-switcher')).toHaveTextContent('Target');

    await userEvent.click(screen.getByText('Journey'));
    expect(onLensChange).toHaveBeenCalledWith('journey');
  });

  it('shows a one-line description of the active lens', () => {
    renderToolbar({ lens: 'journey' });
    expect(screen.getByTestId('lens-description')).toHaveTextContent('done');
  });

  it('hides the assign rail toggle outside the Now lens', () => {
    renderToolbar({ lens: 'journey' });
    expect(screen.queryByTestId('assign-rail-toggle')).not.toBeInTheDocument();
  });

  it('shows the changes-only toggle and summary only in the relevant lenses', () => {
    const { rerender } = renderToolbar({ lens: 'now' });
    expect(screen.queryByTestId('changes-only-toggle')).not.toBeInTheDocument();
    expect(screen.queryByTestId('board-summary')).not.toBeInTheDocument();

    rerender(<BoardToolbarHarness lens="journey" summary={{ settled: 3, inFlight: 2, notStarted: 1 }} />);
    expect(screen.getByTestId('changes-only-toggle')).toBeInTheDocument();
    expect(screen.getByTestId('board-summary')).toHaveTextContent('3');
    expect(screen.getByTestId('board-summary')).toHaveTextContent('settled');
  });

  it('shows the changes-only toggle but not the summary in the Target lens', () => {
    renderToolbar({ lens: 'target' });
    expect(screen.getByTestId('changes-only-toggle')).toBeInTheDocument();
    expect(screen.queryByTestId('board-summary')).not.toBeInTheDocument();
  });

  it('hides the assign rail toggle entirely when showAssignToggle is false (stakeholder role)', () => {
    renderToolbar({ showAssignToggle: false });
    expect(screen.queryByTestId('assign-rail-toggle')).not.toBeInTheDocument();
  });

  it('calls onCreateDomain and onToggleAssignRail when their buttons are clicked', async () => {
    const onCreateDomain = vi.fn();
    const onToggleAssignRail = vi.fn();
    renderToolbar({ onCreateDomain, onToggleAssignRail });

    await userEvent.click(screen.getByTestId('create-domain-button'));
    await userEvent.click(screen.getByTestId('assign-rail-toggle'));

    expect(onCreateDomain).toHaveBeenCalledTimes(1);
    expect(onToggleAssignRail).toHaveBeenCalledTimes(1);
  });
});
