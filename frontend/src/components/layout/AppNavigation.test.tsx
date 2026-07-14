import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import { AppNavigation } from './AppNavigation';

const mockState = {
  hasPermission: (_permission: string) => true,
  sessionLinks: null as Record<string, string> | null,
  tenant: null,
};

vi.mock('../../store/userStore', () => ({
  useUserStore: <T,>(selector: (state: typeof mockState) => T): T => selector(mockState),
}));

describe('AppNavigation', () => {
  it('shows the One-Pager Quality nav entry when the session link is present', () => {
    mockState.sessionLinks = { 'x-one-pager-quality': '/api/v1/one-pager-quality' };

    renderWithProviders(<AppNavigation currentView="canvas" />);

    expect(screen.getByTestId('nav-one-pager-quality')).toBeInTheDocument();
  });

  it('hides the One-Pager Quality nav entry when the session link is absent', () => {
    mockState.sessionLinks = {};

    renderWithProviders(<AppNavigation currentView="canvas" />);

    expect(screen.queryByTestId('nav-one-pager-quality')).not.toBeInTheDocument();
  });

  it('hides the One-Pager Quality nav entry when sessionLinks is null', () => {
    mockState.sessionLinks = null;

    renderWithProviders(<AppNavigation currentView="canvas" />);

    expect(screen.queryByTestId('nav-one-pager-quality')).not.toBeInTheDocument();
  });
});
