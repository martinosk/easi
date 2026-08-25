import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import { UserMenu } from './UserMenu';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => ({
  ...(await vi.importActual<typeof import('react-router-dom')>('react-router-dom')),
  useNavigate: () => mockNavigate,
}));

vi.mock('../../features/edit-grants/hooks/useEditGrants', () => ({
  useMyEditGrants: () => ({ data: [] }),
}));

const mockState = {
  user: { id: 'u1', name: 'Ada Lovelace', email: 'ada@acme.com', role: 'admin' },
  tenant: { id: 't1', name: 'Acme' },
  permissions: new Set<string>(),
  hasPermission: (permission: string) => mockState.permissions.has(permission),
  logout: vi.fn(),
};

vi.mock('../../store/userStore', () => ({
  useUserStore: <T,>(selector: (state: typeof mockState) => T): T => selector(mockState),
}));

describe('UserMenu', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
    mockState.permissions = new Set();
  });

  it('shows a Settings item for users with metamodel:write that navigates to settings', async () => {
    mockState.permissions = new Set(['metamodel:write']);
    const user = userEvent.setup();
    renderWithProviders(<UserMenu />);

    await user.click(screen.getByTestId('user-menu-trigger'));
    await user.click(await screen.findByTestId('user-menu-settings'));

    expect(mockNavigate).toHaveBeenCalledWith('/settings');
  });

  it('hides the Settings item without metamodel:write', async () => {
    const user = userEvent.setup();
    renderWithProviders(<UserMenu />);

    await user.click(screen.getByTestId('user-menu-trigger'));

    expect(await screen.findByTestId('user-menu-logout')).toBeInTheDocument();
    expect(screen.queryByTestId('user-menu-settings')).not.toBeInTheDocument();
  });

  it('shows a tooltip on the trigger', async () => {
    const user = userEvent.setup();
    renderWithProviders(<UserMenu />);

    await user.hover(screen.getByTestId('user-menu-trigger'));

    expect(await screen.findByRole('tooltip')).toHaveTextContent('Ada Lovelace');
  });
});
