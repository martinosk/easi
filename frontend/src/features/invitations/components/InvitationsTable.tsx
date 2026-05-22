import { Badge, Button, Paper, Table, Text } from '@mantine/core';
import type { Invitation, InvitationStatus } from '../types';

interface InvitationsTableProps {
  invitations: Invitation[];
  onRevoke: (invitation: Invitation) => void;
}

const STATUS_BADGE_COLORS: Record<InvitationStatus, string> = {
  pending: 'orange',
  accepted: 'green',
  expired: 'gray',
  revoked: 'red',
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
    <Paper shadow="sm" radius="lg" withBorder>
      <Table data-testid="invitations-table" striped highlightOnHover verticalSpacing="sm">
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Email</Table.Th>
            <Table.Th>Role</Table.Th>
            <Table.Th>Status</Table.Th>
            <Table.Th>Invited By</Table.Th>
            <Table.Th>Created</Table.Th>
            <Table.Th>Expires</Table.Th>
            <Table.Th>Actions</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {invitations.map((invitation) => (
            <Table.Tr key={invitation.id} data-testid={`invitation-row-${invitation.id}`}>
              <Table.Td>
                <Text fw={500}>{invitation.email}</Text>
              </Table.Td>
              <Table.Td>
                <Badge variant="light" color="gray" tt="capitalize">
                  {invitation.role}
                </Badge>
              </Table.Td>
              <Table.Td>
                <Badge variant="light" color={STATUS_BADGE_COLORS[invitation.status]} tt="capitalize">
                  {invitation.status}
                </Badge>
              </Table.Td>
              <Table.Td>
                <Text c="dimmed">{invitation.invitedBy ?? '-'}</Text>
              </Table.Td>
              <Table.Td>
                <Text c="dimmed" size="xs">
                  {formatDate(invitation.createdAt)}
                </Text>
              </Table.Td>
              <Table.Td>
                <Text c="dimmed" size="xs">
                  {formatDate(invitation.expiresAt)}
                </Text>
              </Table.Td>
              <Table.Td>
                {invitation.status === 'pending' && invitation._links.update && (
                  <Button
                    size="xs"
                    variant="default"
                    onClick={() => onRevoke(invitation)}
                    data-testid={`revoke-btn-${invitation.id}`}
                  >
                    Revoke
                  </Button>
                )}
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </Paper>
  );
}
