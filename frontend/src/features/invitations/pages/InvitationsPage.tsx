import { Alert, Button, Center, Container, Group, Loader, Stack, Text, Title } from '@mantine/core';
import { useCallback, useEffect, useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { useUserStore } from '../../../store/userStore';
import { invitationApi } from '../api/invitationApi';
import { InvitationsEmptyState } from '../components/InvitationsEmptyState';
import { InvitationsFilters } from '../components/InvitationsFilters';
import { InvitationsTable } from '../components/InvitationsTable';
import { InviteUserModal } from '../components/InviteUserModal';
import type { CreateInvitationRequest, Invitation, InvitationStatus } from '../types';
import classes from './InvitationsPage.module.css';

function useInvitations(enabled: boolean) {
  const [invitations, setInvitations] = useState<Invitation[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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
    if (enabled) {
      loadInvitations();
    }
  }, [enabled, loadInvitations]);

  return { invitations, isLoading, error, loadInvitations };
}

export function InvitationsPage() {
  const hasPermission = useUserStore((state) => state.hasPermission);
  const canManageInvitations = hasPermission('invitations:manage');

  const { invitations, isLoading, error, loadInvitations } = useInvitations(canManageInvitations);
  const [statusFilter, setStatusFilter] = useState<InvitationStatus | 'all'>('all');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [revokeTarget, setRevokeTarget] = useState<Invitation | null>(null);
  const [isRevoking, setIsRevoking] = useState(false);
  const [revokeError, setRevokeError] = useState<string | null>(null);

  const filteredInvitations = useMemo(() => {
    if (statusFilter === 'all') return invitations;
    return invitations.filter((inv) => inv.status === statusFilter);
  }, [invitations, statusFilter]);

  const handleCreateInvitation = async (request: CreateInvitationRequest) => {
    await invitationApi.create(request);
    toast.success(`Invitation created for ${request.email}. Please notify them to log in.`);
    await loadInvitations();
  };

  const confirmRevoke = async () => {
    if (!revokeTarget) return;
    setIsRevoking(true);
    setRevokeError(null);
    try {
      await invitationApi.revoke(revokeTarget.id);
      toast.success('Invitation revoked');
      await loadInvitations();
      setRevokeTarget(null);
    } catch (err) {
      setRevokeError(err instanceof Error ? err.message : 'Failed to revoke invitation');
      toast.error(err instanceof Error ? err.message : 'Failed to revoke invitation');
    } finally {
      setIsRevoking(false);
    }
  };

  if (!canManageInvitations) {
    return (
      <PageShell>
        <Alert color="red">You do not have permission to manage invitations.</Alert>
      </PageShell>
    );
  }

  return (
    <PageShell>
      <InvitationsHeader onInvite={() => setIsModalOpen(true)} />
      <InvitationsFilters statusFilter={statusFilter} onFilterChange={setStatusFilter} />
      <InvitationsContent
        isLoading={isLoading}
        error={error}
        invitations={filteredInvitations}
        statusFilter={statusFilter}
        onInvite={() => setIsModalOpen(true)}
        onRevoke={setRevokeTarget}
      />
      <InviteUserModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateInvitation}
      />
      {revokeTarget && (
        <ConfirmationDialog
          title="Revoke invitation?"
          message="Are you sure you want to revoke the invitation for"
          itemName={revokeTarget.email}
          confirmText="Revoke"
          onConfirm={confirmRevoke}
          onCancel={() => {
            setRevokeTarget(null);
            setRevokeError(null);
          }}
          isLoading={isRevoking}
          error={revokeError}
        />
      )}
    </PageShell>
  );
}

function PageShell({ children }: { children: React.ReactNode }) {
  return (
    <div className={classes.page}>
      <Container size="xl" py="xl">
        {children}
      </Container>
    </div>
  );
}

interface InvitationsHeaderProps {
  onInvite: () => void;
}

function InvitationsHeader({ onInvite }: InvitationsHeaderProps) {
  return (
    <Group justify="space-between" align="flex-start" mb="xl">
      <Stack gap="xs">
        <Title order={1}>User Invitations</Title>
        <Text c="dimmed">
          Create invitations for users to join. Once invited, users can log in using their company
          email.
        </Text>
      </Stack>
      <Button onClick={onInvite} data-testid="invite-user-btn">
        Invite User
      </Button>
    </Group>
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

function InvitationsContent({
  isLoading,
  error,
  invitations,
  statusFilter,
  onInvite,
  onRevoke,
}: InvitationsContentProps) {
  if (isLoading) {
    return (
      <Center py="xl">
        <Stack align="center" gap="md">
          <Loader />
          <Text>Loading invitations...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Alert color="red" data-testid="invitations-error">
        {error}
      </Alert>
    );
  }

  if (invitations.length === 0) {
    return <InvitationsEmptyState statusFilter={statusFilter} onInvite={onInvite} />;
  }

  return <InvitationsTable invitations={invitations} onRevoke={onRevoke} />;
}
