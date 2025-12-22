import type { InvitationStatus } from '../types';

interface InvitationsFiltersProps {
  statusFilter: InvitationStatus | 'all';
  onFilterChange: (status: InvitationStatus | 'all') => void;
}

export function InvitationsFilters({ statusFilter, onFilterChange }: InvitationsFiltersProps) {
  return (
    <div className="invitations-filters">
      <div className="filter-group">
        <label htmlFor="status-filter" className="filter-label">Filter by status:</label>
        <select
          id="status-filter"
          className="filter-select"
          value={statusFilter}
          onChange={(e) => onFilterChange(e.target.value as InvitationStatus | 'all')}
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
  );
}
