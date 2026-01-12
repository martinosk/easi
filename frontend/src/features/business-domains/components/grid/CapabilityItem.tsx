import type { Capability, CapabilityId, CapabilityRealization, ComponentId } from '../../../../api/types';
import type { DepthLevel } from '../DepthSelector';
import type { Breakpoint } from '../../../../hooks/useResponsive';
import { RESPONSIVE_GRID_COLUMNS, RESPONSIVE_SPACING, getResponsiveValue } from '../../../../hooks/useResponsive';
import { ApplicationChipList } from '../ApplicationChipList';
import { LEVEL_COLORS, LEVEL_SIZES, levelToNumber, type CapabilityNode } from './gridUtils';

function getResponsiveGridColumns(level: Capability['level'], breakpoint: Breakpoint): string {
  const columns = RESPONSIVE_GRID_COLUMNS[level] || RESPONSIVE_GRID_COLUMNS.L3;
  return getResponsiveValue(columns, breakpoint) || columns.base;
}

function getVisibleChildren(children: CapabilityNode[], depth: DepthLevel): CapabilityNode[] {
  return children.filter((child) => levelToNumber(child.capability.level) <= depth);
}

function getCapabilityRealizations(
  showApplications: boolean,
  getRealizationsForCapability: ((id: CapabilityId) => CapabilityRealization[]) | undefined,
  capabilityId: CapabilityId
): CapabilityRealization[] {
  if (!showApplications || !getRealizationsForCapability) return [];
  return getRealizationsForCapability(capabilityId);
}

function getSelectionStyles(isSelected: boolean): { border?: string; boxShadow?: string } {
  if (!isSelected) return {};
  return {
    border: '3px solid #2563eb',
    boxShadow: '0 0 0 3px rgba(37, 99, 235, 0.2)',
  };
}

export interface CapabilityItemProps {
  node: CapabilityNode;
  depth: DepthLevel;
  onClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu?: (capability: Capability, event: React.MouseEvent) => void;
  showApplications?: boolean;
  getRealizationsForCapability?: (capabilityId: CapabilityId) => CapabilityRealization[];
  onApplicationClick?: (componentId: ComponentId) => void;
  selectedCapabilities?: Set<CapabilityId>;
  breakpoint: Breakpoint;
}

interface ChildrenGridProps {
  children: CapabilityNode[];
  level: Capability['level'];
  depth: DepthLevel;
  onClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu?: (capability: Capability, event: React.MouseEvent) => void;
  showApplications: boolean;
  getRealizationsForCapability?: (capabilityId: CapabilityId) => CapabilityRealization[];
  onApplicationClick?: (componentId: ComponentId) => void;
  selectedCapabilities?: Set<CapabilityId>;
  breakpoint: Breakpoint;
}

function ChildrenGrid({
  children,
  level,
  depth,
  onClick,
  onContextMenu,
  showApplications,
  getRealizationsForCapability,
  onApplicationClick,
  selectedCapabilities,
  breakpoint,
}: ChildrenGridProps) {
  const gridGap = getResponsiveValue(RESPONSIVE_SPACING.gridGap, breakpoint) || '0.5rem';

  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: getResponsiveGridColumns(level, breakpoint),
        gap: gridGap,
        flex: 1,
        minWidth: 0,
        overflow: 'hidden',
      }}
    >
      {children.map((child) => (
        <CapabilityItem
          key={child.capability.id}
          node={child}
          depth={depth}
          onClick={onClick}
          onContextMenu={onContextMenu}
          showApplications={showApplications}
          getRealizationsForCapability={getRealizationsForCapability}
          onApplicationClick={onApplicationClick}
          selectedCapabilities={selectedCapabilities}
          breakpoint={breakpoint}
        />
      ))}
    </div>
  );
}

export function CapabilityItem({
  node,
  depth,
  onClick,
  onContextMenu,
  showApplications = false,
  getRealizationsForCapability,
  onApplicationClick,
  selectedCapabilities,
  breakpoint,
}: CapabilityItemProps) {
  const { capability, children } = node;
  const realizations = getCapabilityRealizations(showApplications, getRealizationsForCapability, capability.id);
  const visibleChildren = getVisibleChildren(children, depth);
  const hasContent = visibleChildren.length > 0 || realizations.length > 0;
  const sizes = LEVEL_SIZES[capability.level];
  const isSelected = selectedCapabilities?.has(capability.id) ?? false;
  const selectionStyles = getSelectionStyles(isSelected);

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onClick(capability, e);
  };

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onContextMenu?.(capability, e);
  };

  const canShowRealizations = realizations.length > 0 && onApplicationClick;

  return (
    <div
      className={isSelected ? 'capability-item selected' : 'capability-item'}
      data-testid={`capability-${capability.id}`}
      onClick={handleClick}
      onContextMenu={handleContextMenu}
      style={{
        backgroundColor: LEVEL_COLORS[capability.level],
        color: 'white',
        padding: sizes.padding,
        borderRadius: '0.5rem',
        minHeight: sizes.minHeight,
        cursor: 'pointer',
        display: 'flex',
        flexDirection: 'column',
        ...selectionStyles,
      }}
    >
      <div style={{ fontWeight: 500, marginBottom: hasContent ? '0.5rem' : 0 }}>
        {capability.name}
      </div>
      {canShowRealizations && (
        <div style={{ marginBottom: visibleChildren.length > 0 ? '0.5rem' : 0 }}>
          <ApplicationChipList realizations={realizations} onApplicationClick={onApplicationClick!} />
        </div>
      )}
      {visibleChildren.length > 0 && (
        <ChildrenGrid
          children={visibleChildren}
          level={capability.level}
          depth={depth}
          onClick={onClick}
          onContextMenu={onContextMenu}
          showApplications={showApplications}
          getRealizationsForCapability={getRealizationsForCapability}
          onApplicationClick={onApplicationClick}
          selectedCapabilities={selectedCapabilities}
          breakpoint={breakpoint}
        />
      )}
    </div>
  );
}
