import { DomainCard } from './DomainCard';
import type { BusinessDomain, BusinessDomainId } from '../../../api/types';

interface DomainListProps {
  domains: BusinessDomain[];
  onVisualize: (domain: BusinessDomain) => void;
  onContextMenu: (e: React.MouseEvent, domain: BusinessDomain) => void;
  selectedDomainId?: BusinessDomainId | null;
}

export function DomainList({ domains, onVisualize, onContextMenu, selectedDomainId }: DomainListProps) {
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
          onVisualize={onVisualize}
          onContextMenu={onContextMenu}
          isSelected={domain.id === selectedDomainId}
        />
      ))}
    </div>
  );
}
