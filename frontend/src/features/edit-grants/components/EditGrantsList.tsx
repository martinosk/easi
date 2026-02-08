import { useState } from 'react';
import { useEditGrantsForArtifact, useRevokeEditGrant } from '../hooks/useEditGrants';
import { EditGrantsEmptyState } from './EditGrantsEmptyState';
import type { ArtifactType, EditGrantStatus } from '../types';
import { hasLink } from '../../../utils/hateoas';

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

export function EditGrantsList({ artifactType, artifactId }: EditGrantsListProps) {
  const [statusFilter, setStatusFilter] = useState<EditGrantStatus | 'all'>('all');
  const { data: grants, isLoading } = useEditGrantsForArtifact(artifactType, artifactId);
  const revokeGrant = useRevokeEditGrant();

  const filteredGrants = grants?.filter(
    g => statusFilter === 'all' || g.status === statusFilter
  ) ?? [];

  if (isLoading) {
    return <div className="loading-spinner" data-testid="edit-grants-loading" />;
  }

  return (
    <div data-testid="edit-grants-list">
      <div className="filter-bar">
        {STATUS_OPTIONS.map(option => (
          <button
            key={option.value}
            className={`btn btn-sm ${statusFilter === option.value ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setStatusFilter(option.value)}
            data-testid={`filter-${option.value}`}
          >
            {option.label}
          </button>
        ))}
      </div>

      {filteredGrants.length === 0 ? (
        <EditGrantsEmptyState statusFilter={statusFilter} />
      ) : (
        <table className="table" data-testid="edit-grants-table">
          <thead>
            <tr>
              <th>Grantee</th>
              <th>Granted By</th>
              <th>Status</th>
              <th>Expires</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredGrants.map(grant => (
              <tr key={grant.id} data-testid={`edit-grant-row-${grant.id}`}>
                <td>{grant.granteeEmail}</td>
                <td>{grant.grantorEmail}</td>
                <td>
                  <span className={`badge badge-${grant.status === 'active' ? 'success' : grant.status === 'revoked' ? 'danger' : 'warning'}`}>
                    {grant.status}
                  </span>
                </td>
                <td>{new Date(grant.expiresAt).toLocaleDateString()}</td>
                <td>
                  {hasLink(grant, 'delete') && (
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={() => revokeGrant.mutate(grant.id)}
                      disabled={revokeGrant.isPending}
                      data-testid={`revoke-grant-${grant.id}`}
                    >
                      Revoke
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
