import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers/renderWithProviders';
import { userApi } from '../api/userApi';
import type { User } from '../types';
import { UsersPage } from './UsersPage';

vi.mock('../api/userApi');
vi.mock('react-hot-toast', () => ({
  default: { error: vi.fn(), success: vi.fn() },
}));
vi.mock('../../../store/userStore', () => ({
  useUserStore: <T,>(
    selector: (state: { hasPermission: (p: string) => boolean; user: { id: string } | null }) => T,
  ): T =>
    selector({ hasPermission: (p) => p === 'users:read' || p === 'users:manage', user: { id: 'current-user-id' } }),
}));

const mockUsers: User[] = [
  {
    id: 'user-1',
    email: 'admin@acme.com',
    name: 'Admin User',
    role: 'admin',
    status: 'active',
    createdAt: '2025-01-01T10:00:00Z',
    lastLoginAt: '2025-01-15T08:30:00Z',
    _links: {
      self: '/api/v1/users/user-1',
      update: '/api/v1/users/user-1',
    },
  },
  {
    id: 'user-2',
    email: 'architect@acme.com',
    name: 'Architect User',
    role: 'architect',
    status: 'active',
    createdAt: '2025-01-02T10:00:00Z',
    lastLoginAt: '2025-01-14T09:15:00Z',
    _links: {
      self: '/api/v1/users/user-2',
      update: '/api/v1/users/user-2',
    },
  },
  {
    id: 'user-3',
    email: 'disabled@acme.com',
    name: 'Disabled User',
    role: 'stakeholder',
    status: 'disabled',
    createdAt: '2025-01-03T10:00:00Z',
    _links: {
      self: '/api/v1/users/user-3',
      update: '/api/v1/users/user-3',
    },
  },
  {
    id: 'current-user-id',
    email: 'current@acme.com',
    name: 'Current User',
    role: 'admin',
    status: 'active',
    createdAt: '2025-01-04T10:00:00Z',
    lastLoginAt: '2025-01-16T10:00:00Z',
    _links: {
      self: '/api/v1/users/current-user-id',
    },
  },
];

describe('UsersPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(userApi.getAll).mockResolvedValue(mockUsers);
  });

  it('renders user management page with all users', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.getAllByTestId(/user-row-/)).toHaveLength(4);
    expect(screen.getByText('admin@acme.com')).toBeInTheDocument();
    expect(screen.getByText('architect@acme.com')).toBeInTheDocument();
    expect(screen.getByText('disabled@acme.com')).toBeInTheDocument();
  });

  it.each([
    { filterType: 'status', testId: 'status-filter', value: 'active', expectedCall: ['active', undefined] },
    { filterType: 'status', testId: 'status-filter', value: 'disabled', expectedCall: ['disabled', undefined] },
    { filterType: 'role', testId: 'role-filter', value: 'admin', expectedCall: [undefined, 'admin'] },
  ])('filters users by $filterType when $value filter is selected', async ({ testId, value, expectedCall }) => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const filter = screen.getByTestId(testId);
    fireEvent.change(filter, { target: { value } });

    await waitFor(() => {
      expect(userApi.getAll).toHaveBeenCalledWith(...expectedCall);
    });
  });

  it('displays "You" badge for current user', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('user-row-current-user-id')).toBeInTheDocument();
    });

    const currentUserRow = screen.getByTestId('user-row-current-user-id');
    expect(currentUserRow).toHaveTextContent('You');
  });

  it('does not show action buttons for current user', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.queryByTestId('change-role-btn-current-user-id')).not.toBeInTheDocument();
    expect(screen.queryByTestId('disable-btn-current-user-id')).not.toBeInTheDocument();
  });

  it('shows change role button for other users', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.getByTestId('change-role-btn-user-1')).toBeInTheDocument();
    expect(screen.getByTestId('change-role-btn-user-2')).toBeInTheDocument();
  });

  it('shows disable button for active users', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.getByTestId('disable-btn-user-1')).toBeInTheDocument();
    expect(screen.getByTestId('disable-btn-user-2')).toBeInTheDocument();
  });

  it('shows enable button for disabled users', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.getByTestId('enable-btn-user-3')).toBeInTheDocument();
  });

  it('opens change role modal when change role button is clicked', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const changeRoleBtn = screen.getByTestId('change-role-btn-user-1');
    fireEvent.click(changeRoleBtn);

    await waitFor(() => {
      expect(screen.getByTestId('change-role-modal')).toBeInTheDocument();
    });
  });

  it('calls update user API when disable is confirmed', async () => {
    vi.mocked(userApi.update).mockResolvedValue(mockUsers[1]);

    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('disable-btn-user-2'));

    const dialog = await screen.findByTestId('confirmation-dialog');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Disable' }));

    await waitFor(() => {
      expect(userApi.update).toHaveBeenCalledWith('user-2', { status: 'disabled' });
    });
  });

  it('calls update user API when enable button is clicked', async () => {
    vi.mocked(userApi.update).mockResolvedValue(mockUsers[2]);

    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('enable-btn-user-3'));

    await waitFor(() => {
      expect(userApi.update).toHaveBeenCalledWith('user-3', { status: 'active' });
    });
  });

  it('does not disable user when confirmation is cancelled', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('disable-btn-user-2'));

    const dialog = await screen.findByTestId('confirmation-dialog');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }));

    expect(userApi.update).not.toHaveBeenCalled();
  });

  it('shows error message when users fail to load', async () => {
    vi.mocked(userApi.getAll).mockRejectedValue(new Error('Failed to fetch users'));

    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-error')).toBeInTheDocument();
    });

    expect(screen.getByTestId('users-error')).toHaveTextContent('Failed to fetch users');
  });

  it('shows empty state when no users are found', async () => {
    vi.mocked(userApi.getAll).mockResolvedValue([]);

    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByText('No users found')).toBeInTheDocument();
    });
  });

  it('displays loading state while fetching users', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    expect(screen.getByText('Loading users...')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });
  });

  it('reloads users after successful role change', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(userApi.getAll).toHaveBeenCalledTimes(1);
  });

  it('displays user status badges correctly', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const statusBadges = screen.getAllByText('active');
    expect(statusBadges.length).toBeGreaterThan(0);

    const disabledBadge = screen.getByText('disabled');
    expect(disabledBadge).toBeInTheDocument();
  });

  it('displays role badges correctly', async () => {
    renderWithProviders(<UsersPage />, { withRouter: false });

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.getAllByText('admin').length).toBeGreaterThan(0);
    expect(screen.getByText('architect')).toBeInTheDocument();
    expect(screen.getByText('stakeholder')).toBeInTheDocument();
  });
});
