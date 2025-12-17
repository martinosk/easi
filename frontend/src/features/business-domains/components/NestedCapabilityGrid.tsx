import { useMemo } from 'react';
import type { Capability, CapabilityId, CapabilityRealization, ComponentId, Position } from '../../../api/types';
import type { DepthLevel } from './DepthSelector';
import { ApplicationChipList } from './ApplicationChipList';
import './visualization.css';

const LEVEL_COLORS = {
  L1: '#3b82f6',
  L2: '#8b5cf6',
  L3: '#ec4899',
  L4: '#f97316',
};

const LEVEL_SIZES = {
  L1: { minHeight: '200px', padding: '1rem' },
  L2: { minHeight: '120px', padding: '0.75rem' },
  L3: { minHeight: '80px', padding: '0.5rem' },
  L4: { minHeight: '50px', padding: '0.375rem' },
};

const GRID_COLUMNS = {
  L1: 'repeat(auto-fill, minmax(150px, 1fr))',
  L2: 'repeat(auto-fill, minmax(120px, 1fr))',
  L3: 'repeat(auto-fill, minmax(100px, 1fr))',
  L4: 'repeat(auto-fill, minmax(100px, 1fr))',
};

function getGridColumns(level: Capability['level']): string {
  return GRID_COLUMNS[level] || GRID_COLUMNS.L3;
}

export interface PositionMap {
  [capabilityId: string]: Position;
}

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

interface CapabilityNode {
  capability: Capability;
  children: CapabilityNode[];
}

function buildTree(capabilities: Capability[]): CapabilityNode[] {
  const byId = new Map<CapabilityId, Capability>();
  const childrenMap = new Map<CapabilityId | undefined, Capability[]>();

  for (const cap of capabilities) {
    byId.set(cap.id, cap);
    const parentId = cap.parentId;
    if (!childrenMap.has(parentId)) {
      childrenMap.set(parentId, []);
    }
    childrenMap.get(parentId)!.push(cap);
  }

  function buildNode(cap: Capability): CapabilityNode {
    const children = (childrenMap.get(cap.id) || [])
      .sort((a, b) => a.name.localeCompare(b.name))
      .map(buildNode);
    return { capability: cap, children };
  }

  const l1Caps = capabilities.filter((c) => c.level === 'L1');
  return l1Caps.sort((a, b) => a.name.localeCompare(b.name)).map(buildNode);
}

function levelToNumber(level: Capability['level']): number {
  return parseInt(level.substring(1), 10);
}

function compareNodesByPosition(
  a: CapabilityNode,
  b: CapabilityNode,
  positions: PositionMap
): number {
  const posA = positions[a.capability.id];
  const posB = positions[b.capability.id];

  if (posA && posB) {
    if (posA.y !== posB.y) return posA.y - posB.y;
    return posA.x - posB.x;
  }
  if (posA) return -1;
  if (posB) return 1;
  return a.capability.name.localeCompare(b.capability.name);
}

interface NestedCapabilityItemProps {
  node: CapabilityNode;
  depth: DepthLevel;
  onClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu?: (capability: Capability, event: React.MouseEvent) => void;
  showApplications?: boolean;
  getRealizationsForCapability?: (capabilityId: CapabilityId) => CapabilityRealization[];
  onApplicationClick?: (componentId: ComponentId) => void;
  selectedCapabilities?: Set<CapabilityId>;
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
}: ChildrenGridProps) {
  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: getGridColumns(level),
        gap: '0.5rem',
        flex: 1,
        overflow: 'auto',
      }}
    >
      {children.map((child) => (
        <NestedCapabilityItem
          key={child.capability.id}
          node={child}
          depth={depth}
          onClick={onClick}
          onContextMenu={onContextMenu}
          showApplications={showApplications}
          getRealizationsForCapability={getRealizationsForCapability}
          onApplicationClick={onApplicationClick}
          selectedCapabilities={selectedCapabilities}
        />
      ))}
    </div>
  );
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

function isCapabilitySelected(
  selectedCapabilities: Set<CapabilityId> | undefined,
  capabilityId: CapabilityId
): boolean {
  return selectedCapabilities?.has(capabilityId) ?? false;
}

function getSelectionStyles(isSelected: boolean): { border?: string; boxShadow?: string } {
  if (!isSelected) return {};
  return {
    border: '3px solid #2563eb',
    boxShadow: '0 0 0 3px rgba(37, 99, 235, 0.2)',
  };
}

function getClassName(isSelected: boolean): string {
  return isSelected ? 'capability-item selected' : 'capability-item';
}

function NestedCapabilityItem({
  node,
  depth,
  onClick,
  onContextMenu,
  showApplications = false,
  getRealizationsForCapability,
  onApplicationClick,
  selectedCapabilities,
}: NestedCapabilityItemProps) {
  const { capability, children } = node;
  const realizations = getCapabilityRealizations(showApplications, getRealizationsForCapability, capability.id);
  const visibleChildren = getVisibleChildren(children, depth);
  const hasContent = visibleChildren.length > 0 || realizations.length > 0;
  const sizes = LEVEL_SIZES[capability.level];
  const isSelected = isCapabilitySelected(selectedCapabilities, capability.id);
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

  const renderRealizations = () => {
    if (!canShowRealizations) return null;
    const marginBottom = visibleChildren.length > 0 ? '0.5rem' : 0;
    return (
      <div style={{ marginBottom }}>
        <ApplicationChipList realizations={realizations} onApplicationClick={onApplicationClick!} />
      </div>
    );
  };

  const renderChildren = () => {
    if (visibleChildren.length === 0) return null;
    return (
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
      />
    );
  };

  return (
    <div
      className={getClassName(isSelected)}
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
      {renderRealizations()}
      {renderChildren()}
    </div>
  );
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
  const tree = useMemo(() => buildTree(capabilities), [capabilities]);

  const sortedTree = useMemo(() => {
    if (positions && Object.keys(positions).length > 0) {
      return [...tree].sort((a, b) => compareNodesByPosition(a, b, positions));
    }
    return tree;
  }, [tree, positions]);

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
          gridTemplateColumns: 'repeat(auto-fill, minmax(250px, 1fr))',
          gap: '1rem',
          padding: '1rem',
        }}
      >
        {sortedTree.map((node) => (
          <NestedCapabilityItem
            key={node.capability.id}
            node={node}
            depth={depth}
            onClick={onCapabilityClick}
            onContextMenu={onContextMenu}
            showApplications={showApplications}
            getRealizationsForCapability={getRealizationsForCapability}
            onApplicationClick={onApplicationClick}
            selectedCapabilities={selectedCapabilities}
          />
        ))}
      </div>
    </div>
  );
}
