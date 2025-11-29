import type { BusinessDomain } from '../../../api/types';

interface DomainCardProps {
  domain: BusinessDomain;
  onEdit: (domain: BusinessDomain) => void;
  onDelete: (domain: BusinessDomain) => void;
  onView: (domain: BusinessDomain) => void;
}

export function DomainCard({ domain, onEdit, onDelete, onView }: DomainCardProps) {
  const canDelete = domain.capabilityCount === 0 && domain._links.delete;

  return (
    <div className="domain-card" data-testid={`domain-card-${domain.id}`}>
      <div className="domain-card-header">
        <h3 className="domain-card-title">{domain.name}</h3>
        <div className="domain-card-badge">
          {domain.capabilityCount} {domain.capabilityCount === 1 ? 'capability' : 'capabilities'}
        </div>
      </div>

      <p className="domain-card-description">{domain.description || 'No description'}</p>

      <div className="domain-card-meta">
        <span className="domain-card-date">Created: {new Date(domain.createdAt).toLocaleDateString()}</span>
        {domain.updatedAt && (
          <span className="domain-card-date">Updated: {new Date(domain.updatedAt).toLocaleDateString()}</span>
        )}
      </div>

      <div className="domain-card-actions">
        <button
          type="button"
          className="btn btn-secondary btn-sm"
          onClick={() => onView(domain)}
          data-testid={`domain-view-${domain.id}`}
        >
          Manage
        </button>
        {domain._links.update && (
          <button
            type="button"
            className="btn btn-secondary btn-sm"
            onClick={() => onEdit(domain)}
            data-testid={`domain-edit-${domain.id}`}
          >
            Edit
          </button>
        )}
        {canDelete && (
          <button
            type="button"
            className="btn btn-danger btn-sm"
            onClick={() => onDelete(domain)}
            data-testid={`domain-delete-${domain.id}`}
          >
            Delete
          </button>
        )}
      </div>
    </div>
  );
}
