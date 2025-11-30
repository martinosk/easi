import type { Capability, CapabilityId, BusinessDomain } from '../../../api/types';
import { CapabilityExplorer } from './CapabilityExplorer';

interface CapabilityExplorerSidebarProps {
  isCollapsed: boolean;
  visualizedDomain: BusinessDomain | null;
  capabilities: Capability[];
  assignedCapabilityIds: Set<CapabilityId>;
  isLoading: boolean;
  onToggle: () => void;
}

export function CapabilityExplorerSidebar({
  isCollapsed,
  visualizedDomain,
  capabilities,
  assignedCapabilityIds,
  isLoading,
  onToggle,
}: CapabilityExplorerSidebarProps) {
  if (isCollapsed) {
    return (
      <button
        type="button"
        className="sidebar-toggle-btn-collapsed right"
        onClick={onToggle}
        aria-label="Expand sidebar"
      >
        ‹
      </button>
    );
  }

  return (
    <aside className="collapsible-sidebar right open narrow">
      <div className="sidebar-content">
        <div className="sidebar-header">
          <h3>Capability Explorer</h3>
          <button
            type="button"
            className="sidebar-toggle-btn"
            onClick={onToggle}
            aria-label="Collapse sidebar"
          >
            ›
          </button>
        </div>
        <p style={{ fontSize: '0.875rem', color: '#6b7280', marginBottom: '1rem' }}>
          {visualizedDomain
            ? 'Drag L1 capabilities to the grid to assign them'
            : 'Select a domain to visualize, then drag capabilities to assign them'}
        </p>
        <div className="sidebar-scrollable">
          <CapabilityExplorer
            capabilities={capabilities}
            assignedCapabilityIds={assignedCapabilityIds}
            isLoading={isLoading}
          />
        </div>
      </div>
    </aside>
  );
}
