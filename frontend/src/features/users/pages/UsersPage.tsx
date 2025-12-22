import { useState, useEffect, useMemo, useCallback } from 'react';
import toast from 'react-hot-toast';
import { userApi } from '../api/userApi';
import { ChangeRoleModal } from '../components/ChangeRoleModal';
import type { User, UserStatus } from '../types';
import type { UserRole } from '../../auth/types';
import { useUserStore } from '../../../store/userStore';
import './UsersPage.css';

export function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<UserStatus | 'all'>('all');
  const [roleFilter, setRoleFilter] = useState<UserRole | 'all'>('all');
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [isChangeRoleModalOpen, setIsChangeRoleModalOpen] = useState(false);

  const hasPermission = useUserStore((state) => state.hasPermission);
  const currentUser = useUserStore((state) => state.user);

  const canReadUsers = hasPermission('users:read');
  const canManageUsers = hasPermission('users:manage');

  const loadUsers = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const statusParam = statusFilter !== 'all' ? statusFilter : undefined;
      const roleParam = roleFilter !== 'all' ? roleFilter : undefined;
      const allUsers = await userApi.getAll(statusParam, roleParam);
      setUsers(allUsers);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users');
      toast.error('Failed to load users');
    } finally {
      setIsLoading(false);
    }
  }, [statusFilter, roleFilter]);

  useEffect(() => {
    if (canReadUsers) {
      loadUsers();
    }
  }, [canReadUsers, loadUsers]);

  const filteredUsers = useMemo(() => {
    return users;
  }, [users]);

  const handleChangeRole = async (newRole: UserRole) => {
    if (!selectedUser) return;

    try {
      await userApi.update(selectedUser.id, { role: newRole });
      toast.success(`Role changed to ${newRole} for ${selectedUser.email}`);
      setIsChangeRoleModalOpen(false);
      setSelectedUser(null);
      await loadUsers();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to change role');
    }
  };

  const handleDisableUser = async (user: User) => {
    if (!window.confirm(`Are you sure you want to disable the account for ${user.email}?`)) {
      return;
    }

    try {
      await userApi.update(user.id, { status: 'disabled' });
      toast.success(`Account disabled for ${user.email}`);
      await loadUsers();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to disable user');
    }
  };

  const handleEnableUser = async (user: User) => {
    try {
      await userApi.update(user.id, { status: 'active' });
      toast.success(`Account enabled for ${user.email}`);
      await loadUsers();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to enable user');
    }
  };

  const openChangeRoleModal = (user: User) => {
    setSelectedUser(user);
    setIsChangeRoleModalOpen(true);
  };

  const getStatusBadgeClass = (status: UserStatus): string => {
    switch (status) {
      case 'active':
        return 'status-badge-active';
      case 'disabled':
        return 'status-badge-disabled';
      default:
        return '';
    }
  };

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const formatDateTime = (dateString: string | undefined): string => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const isCurrentUser = (user: User): boolean => {
    return currentUser?.id === user.id;
  };

  if (!canReadUsers) {
    return (
      <div className="users-page">
        <div className="users-container">
          <div className="error-message">
            You do not have permission to view users.
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="users-page">
      <div className="users-container">
        <div className="users-header">
          <div>
            <h1 className="users-title">User Management</h1>
            <p className="users-subtitle">View and manage users in your organization.</p>
          </div>
        </div>

        <div className="users-filters">
          <div className="filter-group">
            <label htmlFor="status-filter" className="filter-label">Status:</label>
            <select
              id="status-filter"
              className="filter-select"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as UserStatus | 'all')}
              data-testid="status-filter"
            >
              <option value="all">All</option>
              <option value="active">Active</option>
              <option value="disabled">Disabled</option>
            </select>
          </div>
          <div className="filter-group">
            <label htmlFor="role-filter" className="filter-label">Role:</label>
            <select
              id="role-filter"
              className="filter-select"
              value={roleFilter}
              onChange={(e) => setRoleFilter(e.target.value as UserRole | 'all')}
              data-testid="role-filter"
            >
              <option value="all">All</option>
              <option value="admin">Admin</option>
              <option value="architect">Architect</option>
              <option value="stakeholder">Stakeholder</option>
            </select>
          </div>
        </div>

        {isLoading && (
          <div className="loading-state">
            <div className="loading-spinner" />
            <p>Loading users...</p>
          </div>
        )}

        {error && !isLoading && (
          <div className="error-message" data-testid="users-error">
            {error}
          </div>
        )}

        {!isLoading && !error && filteredUsers.length === 0 && (
          <div className="empty-state">
            <svg className="empty-state-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M20 21V19C20 17.9391 19.5786 16.9217 18.8284 16.1716C18.0783 15.4214 17.0609 15 16 15H8C6.93913 15 5.92172 15.4214 5.17157 16.1716C4.42143 16.9217 4 17.9391 4 19V21M16 7C16 9.20914 14.2091 11 12 11C9.79086 11 8 9.20914 8 7C8 4.79086 9.79086 3 12 3C14.2091 3 16 4.79086 16 7Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            <p className="empty-state-text">No users found</p>
          </div>
        )}

        {!isLoading && !error && filteredUsers.length > 0 && (
          <div className="users-table-container">
            <table className="users-table" data-testid="users-table">
              <thead>
                <tr>
                  <th>User</th>
                  <th>Role</th>
                  <th>Status</th>
                  <th>Created</th>
                  <th>Last Login</th>
                  {canManageUsers && <th>Actions</th>}
                </tr>
              </thead>
              <tbody>
                {filteredUsers.map((user) => (
                  <tr key={user.id} data-testid={`user-row-${user.id}`}>
                    <td className="user-info">
                      <div className="user-email">{user.email}</div>
                      {user.name && <div className="user-name">{user.name}</div>}
                      {isCurrentUser(user) && <span className="current-user-badge">You</span>}
                    </td>
                    <td>
                      <span className="role-badge">{user.role}</span>
                    </td>
                    <td>
                      <span className={`status-badge ${getStatusBadgeClass(user.status)}`}>
                        {user.status}
                      </span>
                    </td>
                    <td className="date-cell">{formatDate(user.createdAt)}</td>
                    <td className="date-cell">{formatDateTime(user.lastLoginAt)}</td>
                    {canManageUsers && (
                      <td className="actions-cell">
                        {user._links.update && (
                          <>
                            <button
                              type="button"
                              className="btn btn-small btn-secondary"
                              onClick={() => openChangeRoleModal(user)}
                              data-testid={`change-role-btn-${user.id}`}
                            >
                              Change Role
                            </button>
                            {user.status === 'active' ? (
                              <button
                                type="button"
                                className="btn btn-small btn-danger"
                                onClick={() => handleDisableUser(user)}
                                data-testid={`disable-btn-${user.id}`}
                              >
                                Disable
                              </button>
                            ) : (
                              <button
                                type="button"
                                className="btn btn-small btn-success"
                                onClick={() => handleEnableUser(user)}
                                data-testid={`enable-btn-${user.id}`}
                              >
                                Enable
                              </button>
                            )}
                          </>
                        )}
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {selectedUser && (
        <ChangeRoleModal
          isOpen={isChangeRoleModalOpen}
          onClose={() => {
            setIsChangeRoleModalOpen(false);
            setSelectedUser(null);
          }}
          onSubmit={handleChangeRole}
          user={selectedUser}
        />
      )}
    </div>
  );
}
