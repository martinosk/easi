import type { BusinessDomain, BusinessDomainId } from '../../../api/types';
import { DomainList } from './DomainList';

interface DomainsSidebarProps {
  domains: BusinessDomain[];
  selectedDomainId: BusinessDomainId | undefined;
  onCreateClick: () => void;
  onVisualize: (domain: BusinessDomain) => void;
  onContextMenu: (e: React.MouseEvent, domain: BusinessDomain) => void;
}

export function DomainsSidebar({
  domains,
  selectedDomainId,
  onCreateClick,
  onVisualize,
  onContextMenu,
}: DomainsSidebarProps) {
  return (
    <div className="sidebar-content">
      <div style={{ marginBottom: '1rem' }}>
        <button
          type="button"
          className="btn btn-primary"
          onClick={onCreateClick}
          data-testid="create-domain-button"
        >
          Create Domain
        </button>
      </div>
      <div className="sidebar-scrollable">
        <DomainList
          domains={domains}
          onVisualize={onVisualize}
          onContextMenu={onContextMenu}
          selectedDomainId={selectedDomainId}
        />
      </div>
    </div>
  );
}
