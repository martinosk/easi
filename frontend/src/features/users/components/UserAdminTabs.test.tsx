import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { UserAdminTabs } from './UserAdminTabs';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => ({
  ...(await vi.importActual<typeof import('react-router-dom')>('react-router-dom')),
  useNavigate: () => mockNavigate,
}));

let permissions = new Set<string>();
vi.mock('../../../store/userStore', () => ({
  useUserStore: <T,>(selector: (state: { hasPermission: (p: string) => boolean }) => T): T =>
    selector({ hasPermission: (p) => permissions.has(p) }),
}));

describe('UserAdminTabs', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
    permissions = new Set(['users:read', 'invitations:manage']);
  });

  it('shows Users and Invitations tabs with the active one selected', () => {
    renderWithProviders(<UserAdminTabs active="invitations" />);

    expect(screen.getByTestId('user-admin-tab-users')).toHaveAttribute('aria-selected', 'false');
    expect(screen.getByTestId('user-admin-tab-invitations')).toHaveAttribute('aria-selected', 'true');
  });

  it('hides the Invitations tab without invitations:manage', () => {
    permissions = new Set(['users:read']);

    renderWithProviders(<UserAdminTabs active="users" />);

    expect(screen.getByTestId('user-admin-tab-users')).toBeInTheDocument();
    expect(screen.queryByTestId('user-admin-tab-invitations')).not.toBeInTheDocument();
  });

  it('navigates when a tab is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<UserAdminTabs active="users" />);

    await user.click(screen.getByTestId('user-admin-tab-invitations'));

    expect(mockNavigate).toHaveBeenCalledWith('/invitations');
  });
});
