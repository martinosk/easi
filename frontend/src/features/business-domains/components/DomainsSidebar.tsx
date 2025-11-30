import type { BusinessDomain, BusinessDomainId } from '../../../api/types';
import { DomainList } from './DomainList';

interface DomainsSidebarProps {
  isCollapsed: boolean;
  domains: BusinessDomain[];
  selectedDomainId: BusinessDomainId | undefined;
  onToggle: () => void;
  onCreateClick: () => void;
  onVisualize: (domain: BusinessDomain) => void;
  onContextMenu: (e: React.MouseEvent, domain: BusinessDomain) => void;
}

export function DomainsSidebar({
  isCollapsed,
  domains,
  selectedDomainId,
  onToggle,
  onCreateClick,
  onVisualize,
  onContextMenu,
}: DomainsSidebarProps) {
  if (isCollapsed) {
    return (
      <button
        type="button"
        className="sidebar-toggle-btn-collapsed left"
        onClick={onToggle}
        aria-label="Expand sidebar"
      >
        ›
      </button>
    );
  }

  return (
    <aside className="collapsible-sidebar open">
      <div className="sidebar-content">
        <div className="sidebar-header">
          <h3>Business Domains</h3>
          <button
            type="button"
            className="sidebar-toggle-btn"
            onClick={onToggle}
            aria-label="Collapse sidebar"
          >
            ‹
          </button>
        </div>
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
    </aside>
  );
}
