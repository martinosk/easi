import { Button, Stack, Text } from '@mantine/core';

interface InvitationsEmptyStateProps {
  statusFilter: string;
  onInvite: () => void;
}

export function InvitationsEmptyState({ statusFilter, onInvite }: InvitationsEmptyStateProps) {
  const message =
    statusFilter === 'all' ? 'No invitations found' : `No ${statusFilter} invitations`;

  return (
    <Stack align="center" gap="lg" py="xl">
      <Text size="lg" c="dimmed">
        {message}
      </Text>
      {statusFilter === 'all' && (
        <Button onClick={onInvite}>Create your first invitation</Button>
      )}
    </Stack>
  );
}
