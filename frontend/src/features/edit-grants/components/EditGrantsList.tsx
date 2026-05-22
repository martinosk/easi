import { Badge, Button, Group, Loader, Table } from '@mantine/core';
import { useState } from 'react';
import { hasLink } from '../../../utils/hateoas';
import { useEditGrantsForArtifact, useRevokeEditGrant } from '../hooks/useEditGrants';
import type { ArtifactType, EditGrant, EditGrantStatus } from '../types';
import { EditGrantsEmptyState } from './EditGrantsEmptyState';

interface EditGrantsListProps {
  artifactType: ArtifactType;
  artifactId: string;
}

const STATUS_OPTIONS: { label: string; value: EditGrantStatus | 'all' }[] = [
  { label: 'All', value: 'all' },
  { label: 'Active', value: 'active' },
  { label: 'Revoked', value: 'revoked' },
  { label: 'Expired', value: 'expired' },
];

const STATUS_BADGE_COLORS: Record<EditGrantStatus, string> = {
  active: 'green',
  revoked: 'red',
  expired: 'yellow',
};

interface FilterBarProps {
  statusFilter: EditGrantStatus | 'all';
  onChange: (value: EditGrantStatus | 'all') => void;
}

function FilterBar({ statusFilter, onChange }: FilterBarProps) {
  return (
    <Group gap="xs" mb="md">
      {STATUS_OPTIONS.map((option) => (
        <Button
          key={option.value}
          size="xs"
          variant={statusFilter === option.value ? 'filled' : 'default'}
          onClick={() => onChange(option.value)}
          data-testid={`filter-${option.value}`}
        >
          {option.label}
        </Button>
      ))}
    </Group>
  );
}

interface GrantRowProps {
  grant: EditGrant;
  onRevoke: (id: string) => void;
  isRevoking: boolean;
}

function GrantRow({ grant, onRevoke, isRevoking }: GrantRowProps) {
  return (
    <Table.Tr data-testid={`edit-grant-row-${grant.id}`}>
      <Table.Td>{grant.granteeEmail}</Table.Td>
      <Table.Td>{grant.grantorEmail}</Table.Td>
      <Table.Td>
        <Badge variant="light" color={STATUS_BADGE_COLORS[grant.status]} tt="capitalize">
          {grant.status}
        </Badge>
      </Table.Td>
      <Table.Td>{new Date(grant.expiresAt).toLocaleDateString()}</Table.Td>
      <Table.Td>
        {hasLink(grant, 'delete') && (
          <Button
            size="xs"
            color="red"
            onClick={() => onRevoke(grant.id)}
            disabled={isRevoking}
            data-testid={`revoke-grant-${grant.id}`}
          >
            Revoke
          </Button>
        )}
      </Table.Td>
    </Table.Tr>
  );
}

export function EditGrantsList({ artifactType, artifactId }: EditGrantsListProps) {
  const [statusFilter, setStatusFilter] = useState<EditGrantStatus | 'all'>('all');
  const { data: grants, isLoading } = useEditGrantsForArtifact(artifactType, artifactId);
  const revokeGrant = useRevokeEditGrant();

  const filteredGrants = grants?.filter((g) => statusFilter === 'all' || g.status === statusFilter) ?? [];

  if (isLoading) {
    return <Loader data-testid="edit-grants-loading" />;
  }

  return (
    <div data-testid="edit-grants-list">
      <FilterBar statusFilter={statusFilter} onChange={setStatusFilter} />

      {filteredGrants.length === 0 ? (
        <EditGrantsEmptyState statusFilter={statusFilter} />
      ) : (
        <Table data-testid="edit-grants-table" striped highlightOnHover>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Grantee</Table.Th>
              <Table.Th>Granted By</Table.Th>
              <Table.Th>Status</Table.Th>
              <Table.Th>Expires</Table.Th>
              <Table.Th>Actions</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {filteredGrants.map((grant) => (
              <GrantRow
                key={grant.id}
                grant={grant}
                onRevoke={(id) => revokeGrant.mutate(id)}
                isRevoking={revokeGrant.isPending}
              />
            ))}
          </Table.Tbody>
        </Table>
      )}
    </div>
  );
}
