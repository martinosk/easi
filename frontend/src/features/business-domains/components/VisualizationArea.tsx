import type { BusinessDomain, Capability, CapabilityId, CapabilityRealization, ComponentId } from '../../../api/types';
import { NestedCapabilityGrid } from './NestedCapabilityGrid';
import { DepthSelector, type DepthLevel } from './DepthSelector';
import { ShowApplicationsToggle } from './ShowApplicationsToggle';

interface VisualizationAreaProps {
  visualizedDomain: BusinessDomain | null;
  capabilities: Capability[];
  capabilitiesLoading: boolean;
  depth: DepthLevel;
  positions: Record<CapabilityId, { x: number; y: number }>;
  onDepthChange: (depth: DepthLevel) => void;
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  selectedCapabilities: Set<CapabilityId>;
  showApplications: boolean;
  onShowApplicationsChange: (value: boolean) => void;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
  onApplicationClick: (componentId: ComponentId) => void;
  isDragOver?: boolean;
  onDragOver?: (e: React.DragEvent) => void;
  onDragLeave?: () => void;
  onDrop?: (e: React.DragEvent) => void;
}

export function VisualizationArea({
  visualizedDomain,
  capabilities,
  capabilitiesLoading,
  depth,
  positions,
  onDepthChange,
  onCapabilityClick,
  onContextMenu,
  selectedCapabilities,
  showApplications,
  onShowApplicationsChange,
  getRealizationsForCapability,
  onApplicationClick,
  isDragOver = false,
  onDragOver,
  onDragLeave,
  onDrop,
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
          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
            <ShowApplicationsToggle
              showApplications={showApplications}
              onShowApplicationsChange={onShowApplicationsChange}
            />
            <DepthSelector value={depth} onChange={onDepthChange} />
          </div>
        </div>
        <NestedCapabilityGrid
          capabilities={capabilities}
          depth={depth}
          onCapabilityClick={onCapabilityClick}
          onContextMenu={onContextMenu}
          selectedCapabilities={selectedCapabilities}
          positions={positions}
          showApplications={showApplications}
          getRealizationsForCapability={getRealizationsForCapability}
          onApplicationClick={onApplicationClick}
          isDragOver={isDragOver}
          onDragOver={onDragOver}
          onDragLeave={onDragLeave}
          onDrop={onDrop}
        />
      </div>
    </main>
  );
}
