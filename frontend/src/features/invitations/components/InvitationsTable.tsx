import type { Invitation, InvitationStatus } from '../types';

interface InvitationsTableProps {
  invitations: Invitation[];
  onRevoke: (invitation: Invitation) => void;
}

const STATUS_BADGE_CLASSES: Record<InvitationStatus, string> = {
  pending: 'status-badge-pending',
  accepted: 'status-badge-accepted',
  expired: 'status-badge-expired',
  revoked: 'status-badge-revoked',
};

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function InvitationsTable({ invitations, onRevoke }: InvitationsTableProps) {
  return (
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
          {invitations.map((invitation) => (
            <tr key={invitation.id} data-testid={`invitation-row-${invitation.id}`}>
              <td className="invitation-email">{invitation.email}</td>
              <td>
                <span className="role-badge">{invitation.role}</span>
              </td>
              <td>
                <span className={`status-badge ${STATUS_BADGE_CLASSES[invitation.status]}`}>
                  {invitation.status}
                </span>
              </td>
              <td className="invited-by">{invitation.invitedBy ?? '-'}</td>
              <td className="date-cell">{formatDate(invitation.createdAt)}</td>
              <td className="date-cell">{formatDate(invitation.expiresAt)}</td>
              <td className="actions-cell">
                {invitation.status === 'pending' && invitation._links.update && (
                  <button
                    type="button"
                    className="btn btn-small btn-secondary"
                    onClick={() => onRevoke(invitation)}
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
  );
}
