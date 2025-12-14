import { useState, useEffect, useMemo, useCallback } from 'react';
import toast from 'react-hot-toast';
import { invitationApi } from '../api/invitationApi';
import { InviteUserModal } from '../components/InviteUserModal';
import type { Invitation, InvitationStatus, CreateInvitationRequest } from '../types';
import { useUserStore } from '../../../store/userStore';
import './InvitationsPage.css';

export function InvitationsPage() {
  const [invitations, setInvitations] = useState<Invitation[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<InvitationStatus | 'all'>('all');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const hasPermission = useUserStore((state) => state.hasPermission);

  const canManageInvitations = hasPermission('invitations:manage');

  const loadInvitations = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await invitationApi.listInvitations();
      setInvitations(response.data ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load invitations');
      toast.error('Failed to load invitations');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (canManageInvitations) {
      loadInvitations();
    }
  }, [canManageInvitations, loadInvitations]);

  const filteredInvitations = useMemo(() => {
    if (statusFilter === 'all') {
      return invitations;
    }
    return invitations.filter((inv) => inv.status === statusFilter);
  }, [invitations, statusFilter]);

  const handleCreateInvitation = async (request: CreateInvitationRequest) => {
    await invitationApi.createInvitation(request);
    toast.success(`Invitation created for ${request.email}. Please notify them to log in.`);
    await loadInvitations();
  };

  const handleRevokeInvitation = async (invitation: Invitation) => {
    if (!window.confirm(`Are you sure you want to revoke the invitation for ${invitation.email}?`)) {
      return;
    }

    try {
      await invitationApi.revokeInvitation(invitation.id);
      toast.success('Invitation revoked');
      await loadInvitations();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to revoke invitation');
    }
  };

  const getStatusBadgeClass = (status: InvitationStatus): string => {
    switch (status) {
      case 'pending':
        return 'status-badge-pending';
      case 'accepted':
        return 'status-badge-accepted';
      case 'expired':
        return 'status-badge-expired';
      case 'revoked':
        return 'status-badge-revoked';
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
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (!canManageInvitations) {
    return (
      <div className="invitations-page">
        <div className="invitations-container">
          <div className="error-message">
            You do not have permission to manage invitations.
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="invitations-page">
      <div className="invitations-container">
        <div className="invitations-header">
          <div>
            <h1 className="invitations-title">User Invitations</h1>
            <p className="invitations-subtitle">Create invitations for users to join. Once invited, users can log in using their company email.</p>
          </div>
          <button
            type="button"
            className="btn btn-primary"
            onClick={() => setIsModalOpen(true)}
            data-testid="invite-user-btn"
          >
            <svg className="btn-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            Invite User
          </button>
        </div>

        <div className="invitations-filters">
          <div className="filter-group">
            <label htmlFor="status-filter" className="filter-label">Filter by status:</label>
            <select
              id="status-filter"
              className="filter-select"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as InvitationStatus | 'all')}
              data-testid="status-filter"
            >
              <option value="all">All</option>
              <option value="pending">Pending</option>
              <option value="accepted">Accepted</option>
              <option value="expired">Expired</option>
              <option value="revoked">Revoked</option>
            </select>
          </div>
        </div>

        {isLoading && (
          <div className="loading-state">
            <div className="loading-spinner" />
            <p>Loading invitations...</p>
          </div>
        )}

        {error && !isLoading && (
          <div className="error-message" data-testid="invitations-error">
            {error}
          </div>
        )}

        {!isLoading && !error && filteredInvitations.length === 0 && (
          <div className="empty-state">
            <svg className="empty-state-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M20 21V19C20 17.9391 19.5786 16.9217 18.8284 16.1716C18.0783 15.4214 17.0609 15 16 15H8C6.93913 15 5.92172 15.4214 5.17157 16.1716C4.42143 16.9217 4 17.9391 4 19V21M16 7C16 9.20914 14.2091 11 12 11C9.79086 11 8 9.20914 8 7C8 4.79086 9.79086 3 12 3C14.2091 3 16 4.79086 16 7Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            <p className="empty-state-text">
              {statusFilter === 'all' ? 'No invitations found' : `No ${statusFilter} invitations`}
            </p>
            {statusFilter === 'all' && (
              <button
                type="button"
                className="btn btn-primary"
                onClick={() => setIsModalOpen(true)}
              >
                Create your first invitation
              </button>
            )}
          </div>
        )}

        {!isLoading && !error && filteredInvitations.length > 0 && (
          <div className="invitations-table-container">
            <table className="invitations-table" data-testid="invitations-table">
              <thead>
                <tr>
                  <th>Email</th>
                  <th>Role</th>
                  <th>Status</th>
                  <th>Invited By</th>
                  <th>Created</th>
                  <th>Expires</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredInvitations.map((invitation) => (
                  <tr key={invitation.id} data-testid={`invitation-row-${invitation.id}`}>
                    <td className="invitation-email">{invitation.email}</td>
                    <td>
                      <span className="role-badge">{invitation.role}</span>
                    </td>
                    <td>
                      <span className={`status-badge ${getStatusBadgeClass(invitation.status)}`}>
                        {invitation.status}
                      </span>
                    </td>
                    <td className="invited-by">{invitation.invitedBy ?? '-'}</td>
                    <td className="date-cell">{formatDate(invitation.createdAt)}</td>
                    <td className="date-cell">{formatDate(invitation.expiresAt)}</td>
                    <td className="actions-cell">
                      {invitation.status === 'pending' && invitation._links.revoke && (
                        <button
                          type="button"
                          className="btn btn-small btn-secondary"
                          onClick={() => handleRevokeInvitation(invitation)}
                          data-testid={`revoke-btn-${invitation.id}`}
                        >
                          Revoke
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <InviteUserModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateInvitation}
      />
    </div>
  );
}
