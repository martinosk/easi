import { useMemo } from 'react';
import type { Capability, CapabilityId, CapabilityRealization, ComponentId } from '../../../api/types';
import type { DepthLevel } from './DepthSelector';
import { useResponsive, RESPONSIVE_GRID_COLUMNS, RESPONSIVE_SPACING, getResponsiveValue } from '../../../hooks/useResponsive';
import { buildTree, compareNodesByPosition, type PositionMap } from './grid/gridUtils';
import { CapabilityItem } from './grid/CapabilityItem';
import './visualization.css';

export type { PositionMap };

export interface NestedCapabilityGridProps {
  capabilities: Capability[];
  depth: DepthLevel;
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu?: (capability: Capability, event: React.MouseEvent) => void;
  positions?: PositionMap;
  showApplications?: boolean;
  getRealizationsForCapability?: (capabilityId: CapabilityId) => CapabilityRealization[];
  onApplicationClick?: (componentId: ComponentId) => void;
  isDragOver?: boolean;
  onDragOver?: (e: React.DragEvent) => void;
  onDragLeave?: () => void;
  onDrop?: (e: React.DragEvent) => void;
  selectedCapabilities?: Set<CapabilityId>;
}

export function NestedCapabilityGrid({
  capabilities,
  depth,
  onCapabilityClick,
  onContextMenu,
  positions,
  showApplications = false,
  getRealizationsForCapability,
  onApplicationClick,
  isDragOver = false,
  onDragOver,
  onDragLeave,
  onDrop,
  selectedCapabilities,
}: NestedCapabilityGridProps) {
  const { currentBreakpoint } = useResponsive();
  const tree = useMemo(() => buildTree(capabilities), [capabilities]);

  const sortedTree = useMemo(() => {
    if (positions && Object.keys(positions).length > 0) {
      return [...tree].sort((a, b) => compareNodesByPosition(a, b, positions));
    }
    return tree;
  }, [tree, positions]);

  const containerPadding = getResponsiveValue(RESPONSIVE_SPACING.containerPadding, currentBreakpoint) || '1rem';
  const gridGap = getResponsiveValue(RESPONSIVE_SPACING.gridGap, currentBreakpoint) || '1rem';
  const rootGridColumns = getResponsiveValue(RESPONSIVE_GRID_COLUMNS.L1, currentBreakpoint) || 'repeat(auto-fill, minmax(250px, 1fr))';

  return (
    <div
      className="nested-capability-grid"
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
      style={{
        minHeight: '200px',
        border: isDragOver ? '2px dashed #3b82f6' : '2px dashed transparent',
        borderRadius: '0.5rem',
        transition: 'border-color 0.2s, background-color 0.2s',
        backgroundColor: isDragOver ? 'rgba(59, 130, 246, 0.05)' : 'transparent',
      }}
    >
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: rootGridColumns,
          gap: gridGap,
          padding: containerPadding,
        }}
      >
        {sortedTree.map((node) => (
          <CapabilityItem
            key={node.capability.id}
            node={node}
            depth={depth}
            onClick={onCapabilityClick}
            onContextMenu={onContextMenu}
            showApplications={showApplications}
            getRealizationsForCapability={getRealizationsForCapability}
            onApplicationClick={onApplicationClick}
            selectedCapabilities={selectedCapabilities}
            breakpoint={currentBreakpoint}
          />
        ))}
      </div>
    </div>
  );
}
