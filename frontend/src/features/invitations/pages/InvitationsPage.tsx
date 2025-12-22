import { useState, useEffect, useMemo, useCallback } from 'react';
import toast from 'react-hot-toast';
import { invitationApi } from '../api/invitationApi';
import { InviteUserModal } from '../components/InviteUserModal';
import { InvitationsTable } from '../components/InvitationsTable';
import { InvitationsEmptyState } from '../components/InvitationsEmptyState';
import { InvitationsFilters } from '../components/InvitationsFilters';
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
      const allInvitations = await invitationApi.getAll();
      setInvitations(allInvitations);
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
    if (statusFilter === 'all') return invitations;
    return invitations.filter((inv) => inv.status === statusFilter);
  }, [invitations, statusFilter]);

  const handleCreateInvitation = async (request: CreateInvitationRequest) => {
    await invitationApi.create(request);
    toast.success(`Invitation created for ${request.email}. Please notify them to log in.`);
    await loadInvitations();
  };

  const handleRevokeInvitation = async (invitation: Invitation) => {
    if (!window.confirm(`Are you sure you want to revoke the invitation for ${invitation.email}?`)) {
      return;
    }
    try {
      await invitationApi.revoke(invitation.id);
      toast.success('Invitation revoked');
      await loadInvitations();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to revoke invitation');
    }
  };

  if (!canManageInvitations) {
    return (
      <div className="invitations-page">
        <div className="invitations-container">
          <div className="error-message">You do not have permission to manage invitations.</div>
        </div>
      </div>
    );
  }

  return (
    <div className="invitations-page">
      <div className="invitations-container">
        <InvitationsHeader onInvite={() => setIsModalOpen(true)} />
        <InvitationsFilters statusFilter={statusFilter} onFilterChange={setStatusFilter} />
        <InvitationsContent
          isLoading={isLoading}
          error={error}
          invitations={filteredInvitations}
          statusFilter={statusFilter}
          onInvite={() => setIsModalOpen(true)}
          onRevoke={handleRevokeInvitation}
        />
      </div>
      <InviteUserModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateInvitation}
      />
    </div>
  );
}

interface InvitationsHeaderProps {
  onInvite: () => void;
}

function InvitationsHeader({ onInvite }: InvitationsHeaderProps) {
  return (
    <div className="invitations-header">
      <div>
        <h1 className="invitations-title">User Invitations</h1>
        <p className="invitations-subtitle">
          Create invitations for users to join. Once invited, users can log in using their company email.
        </p>
      </div>
      <button type="button" className="btn btn-primary" onClick={onInvite} data-testid="invite-user-btn">
        <svg className="btn-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
        Invite User
      </button>
    </div>
  );
}

interface InvitationsContentProps {
  isLoading: boolean;
  error: string | null;
  invitations: Invitation[];
  statusFilter: string;
  onInvite: () => void;
  onRevoke: (invitation: Invitation) => void;
}

function InvitationsContent({ isLoading, error, invitations, statusFilter, onInvite, onRevoke }: InvitationsContentProps) {
  if (isLoading) {
    return (
      <div className="loading-state">
        <div className="loading-spinner" />
        <p>Loading invitations...</p>
      </div>
    );
  }

  if (error) {
    return <div className="error-message" data-testid="invitations-error">{error}</div>;
  }

  if (invitations.length === 0) {
    return <InvitationsEmptyState statusFilter={statusFilter} onInvite={onInvite} />;
  }

  return <InvitationsTable invitations={invitations} onRevoke={onRevoke} />;
}
