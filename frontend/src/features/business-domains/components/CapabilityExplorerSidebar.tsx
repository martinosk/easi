import type { Capability, CapabilityId, BusinessDomain } from '../../../api/types';
import { CapabilityExplorer } from './CapabilityExplorer';

interface CapabilityExplorerSidebarProps {
  visualizedDomain: BusinessDomain | null;
  capabilities: Capability[];
  assignedCapabilityIds: Set<CapabilityId>;
  isLoading: boolean;
  onDragStart?: (capability: Capability) => void;
  onDragEnd?: () => void;
}

export function CapabilityExplorerSidebar({
  visualizedDomain,
  capabilities,
  assignedCapabilityIds,
  isLoading,
  onDragStart,
  onDragEnd,
}: CapabilityExplorerSidebarProps) {
  return (
    <div className="sidebar-content">
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
          onDragStart={onDragStart}
          onDragEnd={onDragEnd}
        />
      </div>
    </div>
  );
}
