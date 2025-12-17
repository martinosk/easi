import { useDomainVisualization } from '../hooks/useDomainVisualization';
import { useDragHandlers } from '../hooks/useDragHandlers';
import { DomainFilter } from '../components/DomainFilter';
import { NestedCapabilityGrid } from '../components/NestedCapabilityGrid';
import { DepthSelector, type DepthLevel } from '../components/DepthSelector';
import { CapabilityExplorer } from '../components/CapabilityExplorer';
import { ContextMenu } from '../../../components/shared/ContextMenu';
import { DeleteCapabilityDialog } from '../../capabilities/components/DeleteCapabilityDialog';
import type { BusinessDomain, BusinessDomainId, Capability, CapabilityId } from '../../../api/types';

interface DomainVisualizationPageProps {
  initialDomainId?: BusinessDomainId;
}

interface CapabilityDetailPanelProps {
  capability: Capability;
  onClose: () => void;
}

function CapabilityDetailPanel({ capability, onClose }: CapabilityDetailPanelProps) {
  return (
    <aside style={{ width: '300px', borderLeft: '1px solid #e5e7eb', padding: '1rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h3>Capability Details</h3>
        <button
          type="button"
          onClick={onClose}
          style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.5rem' }}
        >
          &times;
        </button>
      </div>
      <div style={{ marginTop: '1rem' }}>
        <p><strong>Name:</strong> {capability.name}</p>
        <p><strong>Level:</strong> {capability.level}</p>
        {capability.description && (
          <p><strong>Description:</strong> {capability.description}</p>
        )}
      </div>
    </aside>
  );
}

interface MainContentProps {
  selectedDomainId: BusinessDomainId | null;
  selectedDomain: BusinessDomain | undefined;
  capabilitiesLoading: boolean;
  capabilities: Capability[];
  depth: DepthLevel;
  setDepth: (depth: DepthLevel) => void;
  positions: Record<CapabilityId, { x: number; y: number }>;
  dragHandlers: ReturnType<typeof useDragHandlers>;
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onCapabilityContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  selectedCapabilities: Set<CapabilityId>;
}

function MainContent({
  selectedDomainId,
  selectedDomain,
  capabilitiesLoading,
  capabilities,
  depth,
  setDepth,
  positions,
  dragHandlers,
  onCapabilityClick,
  onCapabilityContextMenu,
  selectedCapabilities,
}: MainContentProps) {
  if (!selectedDomainId) {
    return (
      <div style={{ textAlign: 'center', marginTop: '4rem' }}>
        <h1>Grid Visualization</h1>
        <p style={{ color: '#6b7280', marginTop: '1rem' }}>Select a domain from the left sidebar</p>
      </div>
    );
  }

  if (capabilitiesLoading) {
    return <div className="loading-message">Loading capabilities...</div>;
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h1>{selectedDomain?.name}</h1>
        <DepthSelector value={depth} onChange={setDepth} />
      </div>
      <NestedCapabilityGrid
        capabilities={capabilities}
        depth={depth}
        onCapabilityClick={onCapabilityClick}
        onContextMenu={onCapabilityContextMenu}
        positions={positions}
        isDragOver={dragHandlers.isDragOver}
        onDragOver={dragHandlers.handleDragOver}
        onDragLeave={dragHandlers.handleDragLeave}
        onDrop={dragHandlers.handleDrop}
        selectedCapabilities={selectedCapabilities}
      />
    </div>
  );
}

export function DomainVisualizationPage({ initialDomainId }: DomainVisualizationPageProps) {
  const state = useDomainVisualization(initialDomainId);

  if (state.domainsLoading) {
    return (
      <div className="page-container">
        <div className="loading-message">Loading domains...</div>
      </div>
    );
  }

  return (
    <div className="visualization-container" style={{ display: 'flex', height: '100vh' }}>
      <aside style={{ width: '250px', borderRight: '1px solid #e5e7eb', padding: '1rem' }}>
        <h2 style={{ marginBottom: '1rem' }}>Business Domains</h2>
        <DomainFilter domains={state.domains} selected={state.selectedDomainId} onSelect={state.handleDomainSelect} />
      </aside>

      <main style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
        <MainContent
          selectedDomainId={state.selectedDomainId}
          selectedDomain={state.selectedDomain}
          capabilitiesLoading={state.capabilitiesLoading}
          capabilities={state.capabilities}
          depth={state.depth}
          setDepth={state.setDepth}
          positions={state.positions}
          dragHandlers={state.dragHandlers}
          onCapabilityClick={state.handleCapabilityClick}
          onCapabilityContextMenu={state.handleCapabilityContextMenu}
          selectedCapabilities={state.selectedCapabilities}
        />
      </main>

      <aside style={{ width: '300px', borderLeft: '1px solid #e5e7eb', padding: '1rem', overflow: 'auto' }}>
        <h3 style={{ marginBottom: '1rem' }}>Capability Explorer</h3>
        <p style={{ fontSize: '0.875rem', color: '#6b7280', marginBottom: '1rem' }}>
          Drag L1 capabilities to the grid to assign them to the selected domain
        </p>
        <CapabilityExplorer
          capabilities={state.allCapabilities}
          assignedCapabilityIds={state.assignedCapabilityIds}
          isLoading={state.treeLoading}
          onDragStart={state.dragHandlers.handleDragStart}
          onDragEnd={state.dragHandlers.handleDragEnd}
        />
      </aside>

      {state.selectedCapability && (
        <CapabilityDetailPanel capability={state.selectedCapability} onClose={state.closeCapabilityDetails} />
      )}

      {state.contextMenu && (
        <ContextMenu
          x={state.contextMenu.x}
          y={state.contextMenu.y}
          items={state.contextMenuItems}
          onClose={state.closeContextMenu}
        />
      )}

      <DeleteCapabilityDialog
        isOpen={state.capabilityToDelete !== null}
        onClose={state.closeDeleteDialog}
        capability={state.capabilityToDelete}
        onConfirm={state.handleDeleteConfirm}
        capabilitiesToDelete={state.capabilitiesToDelete}
      />
    </div>
  );
}
