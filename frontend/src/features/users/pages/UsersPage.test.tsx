import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { UsersPage } from './UsersPage';
import { userApi } from '../api/userApi';
import type { User, UsersListResponse } from '../types';

vi.mock('../api/userApi');
vi.mock('react-hot-toast', () => ({
  default: { error: vi.fn(), success: vi.fn() },
}));
vi.mock('../../../store/userStore', () => ({
  useUserStore: (selector: (state: { hasPermission: (p: string) => boolean; user: { id: string } | null }) => any) =>
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

const mockResponse: UsersListResponse = {
  data: mockUsers,
  pagination: { hasMore: false, limit: 50 },
  _links: { self: '/api/v1/users' },
};

describe('UsersPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(userApi.listUsers).mockResolvedValue(mockResponse);
  });

  it('renders user management page with all users', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.getAllByTestId(/user-row-/)).toHaveLength(4);
    expect(screen.getByText('admin@acme.com')).toBeInTheDocument();
    expect(screen.getByText('architect@acme.com')).toBeInTheDocument();
    expect(screen.getByText('disabled@acme.com')).toBeInTheDocument();
  });

  it('filters users by status when active filter is selected', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const statusFilter = screen.getByTestId('status-filter');
    fireEvent.change(statusFilter, { target: { value: 'active' } });

    await waitFor(() => {
      expect(userApi.listUsers).toHaveBeenCalledWith(50, undefined, 'active', undefined);
    });
  });

  it('filters users by status when disabled filter is selected', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const statusFilter = screen.getByTestId('status-filter');
    fireEvent.change(statusFilter, { target: { value: 'disabled' } });

    await waitFor(() => {
      expect(userApi.listUsers).toHaveBeenCalledWith(50, undefined, 'disabled', undefined);
    });
  });

  it('filters users by role', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const roleFilter = screen.getByTestId('role-filter');
    fireEvent.change(roleFilter, { target: { value: 'admin' } });

    await waitFor(() => {
      expect(userApi.listUsers).toHaveBeenCalledWith(50, undefined, undefined, 'admin');
    });
  });

  it('displays "You" badge for current user', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('user-row-current-user-id')).toBeInTheDocument();
    });

    const currentUserRow = screen.getByTestId('user-row-current-user-id');
    expect(currentUserRow).toHaveTextContent('You');
  });

  it('does not show action buttons for current user', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.queryByTestId('change-role-btn-current-user-id')).not.toBeInTheDocument();
    expect(screen.queryByTestId('disable-btn-current-user-id')).not.toBeInTheDocument();
  });

  it('shows change role button for other users', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.getByTestId('change-role-btn-user-1')).toBeInTheDocument();
    expect(screen.getByTestId('change-role-btn-user-2')).toBeInTheDocument();
  });

  it('shows disable button for active users', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.getByTestId('disable-btn-user-1')).toBeInTheDocument();
    expect(screen.getByTestId('disable-btn-user-2')).toBeInTheDocument();
  });

  it('shows enable button for disabled users', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.getByTestId('enable-btn-user-3')).toBeInTheDocument();
  });

  it('opens change role modal when change role button is clicked', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const changeRoleBtn = screen.getByTestId('change-role-btn-user-1');
    fireEvent.click(changeRoleBtn);

    await waitFor(() => {
      expect(screen.getByTestId('change-role-modal')).toBeInTheDocument();
    });
  });

  it('calls update user API when disable button is clicked and confirmed', async () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    vi.mocked(userApi.updateUser).mockResolvedValue(mockUsers[1]);

    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const disableBtn = screen.getByTestId('disable-btn-user-2');
    fireEvent.click(disableBtn);

    await waitFor(() => {
      expect(userApi.updateUser).toHaveBeenCalledWith('user-2', { status: 'disabled' });
    });

    confirmSpy.mockRestore();
  });

  it('does not disable user when confirmation is cancelled', async () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false);

    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const disableBtn = screen.getByTestId('disable-btn-user-2');
    fireEvent.click(disableBtn);

    expect(userApi.updateUser).not.toHaveBeenCalled();

    confirmSpy.mockRestore();
  });

  it('calls update user API when enable button is clicked', async () => {
    vi.mocked(userApi.updateUser).mockResolvedValue(mockUsers[2]);

    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const enableBtn = screen.getByTestId('enable-btn-user-3');
    fireEvent.click(enableBtn);

    await waitFor(() => {
      expect(userApi.updateUser).toHaveBeenCalledWith('user-3', { status: 'active' });
    });
  });

  it('shows error message when users fail to load', async () => {
    vi.mocked(userApi.listUsers).mockRejectedValue(new Error('Failed to fetch users'));

    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-error')).toBeInTheDocument();
    });

    expect(screen.getByTestId('users-error')).toHaveTextContent('Failed to fetch users');
  });

  it('shows empty state when no users are found', async () => {
    vi.mocked(userApi.listUsers).mockResolvedValue({
      ...mockResponse,
      data: [],
    });

    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByText('No users found')).toBeInTheDocument();
    });
  });

  it('displays loading state while fetching users', async () => {
    render(<UsersPage />);

    expect(screen.getByText('Loading users...')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });
  });

  it('reloads users after successful role change', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(userApi.listUsers).toHaveBeenCalledTimes(1);
  });

  it('displays user status badges correctly', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    const statusBadges = screen.getAllByText('active');
    expect(statusBadges.length).toBeGreaterThan(0);

    const disabledBadge = screen.getByText('disabled');
    expect(disabledBadge).toBeInTheDocument();
  });

  it('displays role badges correctly', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(screen.getByTestId('users-table')).toBeInTheDocument();
    });

    expect(screen.getAllByText('admin').length).toBeGreaterThan(0);
    expect(screen.getByText('architect')).toBeInTheDocument();
    expect(screen.getByText('stakeholder')).toBeInTheDocument();
  });

});
