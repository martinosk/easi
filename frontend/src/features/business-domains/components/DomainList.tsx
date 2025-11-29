import { DomainCard } from './DomainCard';
import type { BusinessDomain } from '../../../api/types';

interface DomainListProps {
  domains: BusinessDomain[];
  onEdit: (domain: BusinessDomain) => void;
  onDelete: (domain: BusinessDomain) => void;
  onView: (domain: BusinessDomain) => void;
}

export function DomainList({ domains, onEdit, onDelete, onView }: DomainListProps) {
  if (domains.length === 0) {
    return (
      <div className="empty-state" data-testid="domains-empty-state">
        <p>No business domains yet.</p>
        <p>Create your first domain to get started.</p>
      </div>
    );
  }

  return (
    <div className="domain-list" data-testid="domain-list">
      {domains.map((domain) => (
        <DomainCard
          key={domain.id}
          domain={domain}
          onEdit={onEdit}
          onDelete={onDelete}
          onView={onView}
        />
      ))}
    </div>
  );
}
