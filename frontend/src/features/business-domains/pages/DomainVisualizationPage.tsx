import { useState, useMemo } from 'react';
import { useBusinessDomains } from '../hooks/useBusinessDomains';
import { useDomainCapabilities } from '../hooks/useDomainCapabilities';
import { useCapabilityTree } from '../hooks/useCapabilityTree';
import { useGridPositions } from '../hooks/useGridPositions';
import { useDragHandlers } from '../hooks/useDragHandlers';
import { DomainFilter } from '../components/DomainFilter';
import { NestedCapabilityGrid } from '../components/NestedCapabilityGrid';
import { DepthSelector, type DepthLevel } from '../components/DepthSelector';
import { CapabilityExplorer } from '../components/CapabilityExplorer';
import { usePersistedDepth } from '../hooks/usePersistedDepth';
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
  onCapabilityClick: (capability: Capability) => void;
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
}: MainContentProps) {
  if (!selectedDomainId) {
    return (
      <div style={{ textAlign: 'center', marginTop: '4rem' }}>
        <h1>Grid Visualization</h1>
        <p style={{ color: '#6b7280', marginTop: '1rem' }}>
          Select a domain from the left sidebar
        </p>
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
        positions={positions}
        isDragOver={dragHandlers.isDragOver}
        onDragOver={dragHandlers.handleDragOver}
        onDragLeave={dragHandlers.handleDragLeave}
        onDrop={dragHandlers.handleDrop}
      />
    </div>
  );
}

export function DomainVisualizationPage({ initialDomainId }: DomainVisualizationPageProps) {
  const [selectedDomainId, setSelectedDomainId] = useState<BusinessDomainId | null>(initialDomainId ?? null);
  const [selectedCapability, setSelectedCapability] = useState<Capability | null>(null);
  const [depth, setDepth] = usePersistedDepth();
  const { domains, isLoading: domainsLoading } = useBusinessDomains();
  const { tree, isLoading: treeLoading } = useCapabilityTree();
  const { positions, updatePosition } = useGridPositions(selectedDomainId);

  const allCapabilities = useMemo(() => {
    const flatten = (nodes: typeof tree): Capability[] => {
      return nodes.flatMap((node) => [node.capability, ...flatten(node.children)]);
    };
    return flatten(tree);
  }, [tree]);

  const selectedDomain = useMemo(
    () => domains.find((d) => d.id === selectedDomainId),
    [domains, selectedDomainId]
  );

  const {
    capabilities,
    isLoading: capabilitiesLoading,
    associateCapability,
    refetch: refetchCapabilities,
  } = useDomainCapabilities(selectedDomain?._links.capabilities);

  const assignedCapabilityIds = useMemo(
    () => new Set<CapabilityId>(capabilities.map((c) => c.id)),
    [capabilities]
  );

  const dragHandlers = useDragHandlers({
    domainId: selectedDomainId,
    capabilities,
    assignedCapabilityIds,
    positions,
    updatePosition,
    associateCapability,
    refetchCapabilities,
  });

  if (domainsLoading) {
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
        <DomainFilter
          domains={domains}
          selected={selectedDomainId}
          onSelect={(id) => { setSelectedDomainId(id); setSelectedCapability(null); }}
        />
      </aside>

      <main style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
        <MainContent
          selectedDomainId={selectedDomainId}
          selectedDomain={selectedDomain}
          capabilitiesLoading={capabilitiesLoading}
          capabilities={capabilities}
          depth={depth}
          setDepth={setDepth}
          positions={positions}
          dragHandlers={dragHandlers}
          onCapabilityClick={setSelectedCapability}
        />
      </main>

      <aside style={{ width: '300px', borderLeft: '1px solid #e5e7eb', padding: '1rem', overflow: 'auto' }}>
        <h3 style={{ marginBottom: '1rem' }}>Capability Explorer</h3>
        <p style={{ fontSize: '0.875rem', color: '#6b7280', marginBottom: '1rem' }}>
          Drag L1 capabilities to the grid to assign them to the selected domain
        </p>
        <CapabilityExplorer
          capabilities={allCapabilities}
          assignedCapabilityIds={assignedCapabilityIds}
          isLoading={treeLoading}
          onDragStart={dragHandlers.handleDragStart}
          onDragEnd={dragHandlers.handleDragEnd}
        />
      </aside>

      {selectedCapability && (
        <CapabilityDetailPanel
          capability={selectedCapability}
          onClose={() => setSelectedCapability(null)}
        />
      )}
    </div>
  );
}
