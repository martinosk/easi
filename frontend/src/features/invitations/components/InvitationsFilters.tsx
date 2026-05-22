import { Group, NativeSelect } from '@mantine/core';
import type { InvitationStatus } from '../types';

interface InvitationsFiltersProps {
  statusFilter: InvitationStatus | 'all';
  onFilterChange: (status: InvitationStatus | 'all') => void;
}

const STATUS_OPTIONS = [
  { value: 'all', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'accepted', label: 'Accepted' },
  { value: 'expired', label: 'Expired' },
  { value: 'revoked', label: 'Revoked' },
];

export function InvitationsFilters({ statusFilter, onFilterChange }: InvitationsFiltersProps) {
  return (
    <Group gap="lg" mb="xl">
      <NativeSelect
        label="Filter by status"
        data={STATUS_OPTIONS}
        value={statusFilter}
        onChange={(event) => onFilterChange(event.currentTarget.value as InvitationStatus | 'all')}
        data-testid="status-filter"
      />
    </Group>
  );
}
