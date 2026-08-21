import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { buildHookData, type BusinessDomainsHookData } from '../testkit/hookData';
import { DomainBoard } from './DomainBoard';

function renderBoard(overrides: Partial<BusinessDomainsHookData> = {}) {
  return renderWithProviders(
    <DomainBoard hookData={buildHookData(overrides)} viewMode="board" onViewModeChange={vi.fn()} />,
  );
}

describe('DomainBoard', () => {
  it('renders the board container and one card per business domain', () => {
    renderBoard();

    expect(screen.getByTestId('domain-board')).toBeInTheDocument();
    expect(screen.getByTestId('domain-card-domain-1')).toBeInTheDocument();
    expect(screen.getByTestId('domain-card-domain-2')).toBeInTheDocument();
  });

  it('renders the toolbar with search and legend', () => {
    renderBoard();

    expect(screen.getByTestId('board-search-input')).toBeInTheDocument();
    expect(screen.getByTestId('board-legend')).toBeInTheDocument();
  });

  it('shows an empty-state message when there are no business domains', () => {
    renderBoard({ boardDomains: [] });

    expect(screen.queryByTestId('domain-board')).not.toBeInTheDocument();
    expect(screen.getByText(/No business domains yet/)).toBeInTheDocument();
  });

  it('hides the assign rail when the toggle is closed', () => {
    renderBoard({ assignRailOpen: false });
    expect(screen.queryByTestId('assign-rail')).not.toBeInTheDocument();
  });

  it('shows the assign rail when open and the role permits it', () => {
    renderBoard({ assignRailOpen: true, showAssignRail: true });
    expect(screen.getByTestId('assign-rail')).toBeInTheDocument();
  });

  it('never shows the assign rail for a stakeholder, even if assignRailOpen is somehow true', () => {
    renderBoard({ assignRailOpen: true, showAssignRail: false });
    expect(screen.queryByTestId('assign-rail')).not.toBeInTheDocument();
  });
});
