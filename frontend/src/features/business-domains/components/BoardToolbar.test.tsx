import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { BoardToolbar, type BoardToolbarProps } from './BoardToolbar';

function renderToolbar(overrides: Partial<BoardToolbarProps> = {}) {
  const props: BoardToolbarProps = {
    searchQuery: '',
    onSearchChange: vi.fn(),
    canCreateDomain: true,
    onCreateDomain: vi.fn(),
    showAssignToggle: true,
    assignRailOpen: false,
    onToggleAssignRail: vi.fn(),
    ...overrides,
  };
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

    rerender(
      <BoardToolbar
        searchQuery=""
        onSearchChange={vi.fn()}
        canCreateDomain={false}
        onCreateDomain={vi.fn()}
        showAssignToggle
        assignRailOpen={false}
        onToggleAssignRail={vi.fn()}
      />,
    );
    expect(screen.queryByTestId('create-domain-button')).not.toBeInTheDocument();
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
