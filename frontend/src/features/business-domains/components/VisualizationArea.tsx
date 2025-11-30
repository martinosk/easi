import type { BusinessDomain, Capability, CapabilityId } from '../../../api/types';
import { DomainGrid } from './DomainGrid';
import { NestedCapabilityGrid } from './NestedCapabilityGrid';
import { DepthSelector, type DepthLevel } from './DepthSelector';

interface VisualizationAreaProps {
  visualizedDomain: BusinessDomain | null;
  capabilities: Capability[];
  capabilitiesWithDescendants: Capability[];
  capabilitiesLoading: boolean;
  depth: DepthLevel;
  positions: Record<CapabilityId, { x: number; y: number }>;
  onDepthChange: (depth: DepthLevel) => void;
  onCapabilityClick: (capability: Capability) => void;
}

export function VisualizationArea({
  visualizedDomain,
  capabilities,
  capabilitiesWithDescendants,
  capabilitiesLoading,
  depth,
  positions,
  onDepthChange,
  onCapabilityClick,
}: VisualizationAreaProps) {
  if (!visualizedDomain) {
    return (
      <main className="business-domains-main" style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
        <div style={{ textAlign: 'center', marginTop: '4rem' }}>
          <h2>Grid Visualization</h2>
          <p style={{ color: '#6b7280', marginTop: '1rem' }}>
            Click a domain to see its capabilities
          </p>
        </div>
      </main>
    );
  }

  if (capabilitiesLoading) {
    return (
      <main className="business-domains-main" style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
        <div className="loading-message">Loading capabilities...</div>
      </main>
    );
  }

  return (
    <main className="business-domains-main" style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
      <div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
          <h2>{visualizedDomain.name}</h2>
          <DepthSelector value={depth} onChange={onDepthChange} />
        </div>
        {depth === 1 ? (
          <DomainGrid capabilities={capabilities} onCapabilityClick={onCapabilityClick} positions={positions} />
        ) : (
          <NestedCapabilityGrid
            capabilities={capabilitiesWithDescendants}
            depth={depth}
            onCapabilityClick={onCapabilityClick}
            positions={positions}
          />
        )}
      </div>
    </main>
  );
}
