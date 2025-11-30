import { useMemo } from 'react';
import { useDroppable } from '@dnd-kit/core';
import { SortableContext, useSortable, rectSortingStrategy } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import type { Capability, CapabilityId, Position } from '../../../api/types';
import type { DepthLevel } from './DepthSelector';
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

export interface PositionMap {
  [capabilityId: string]: Position;
}

export interface NestedCapabilityGridProps {
  capabilities: Capability[];
  depth: DepthLevel;
  onCapabilityClick: (capability: Capability) => void;
  positions?: PositionMap;
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

interface NestedCapabilityItemProps {
  node: CapabilityNode;
  depth: DepthLevel;
  onClick: (capability: Capability) => void;
  sortable?: boolean;
}

function NestedCapabilityItem({ node, depth, onClick, sortable = false }: NestedCapabilityItemProps) {
  const { capability, children } = node;
  const level = capability.level;
  const color = LEVEL_COLORS[level];
  const sizes = LEVEL_SIZES[level];

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: capability.id, disabled: !sortable });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    backgroundColor: color,
    color: 'white',
    padding: sizes.padding,
    borderRadius: '0.5rem',
    minHeight: sizes.minHeight,
    opacity: isDragging ? 0.5 : 1,
    cursor: sortable ? 'grab' : 'pointer',
    display: 'flex',
    flexDirection: 'column' as const,
  };

  const visibleChildren = children.filter((child) => {
    const childLevel = levelToNumber(child.capability.level);
    return childLevel <= depth;
  });

  return (
    <div
      ref={setNodeRef}
      className="capability-item"
      data-testid={`capability-${capability.id}`}
      onClick={(e) => {
        e.stopPropagation();
        onClick(capability);
      }}
      style={style}
      {...(sortable ? { ...attributes, ...listeners } : {})}
    >
      <div style={{ fontWeight: 500, marginBottom: visibleChildren.length > 0 ? '0.5rem' : 0 }}>
        {capability.name}
      </div>

      {visibleChildren.length > 0 && (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: level === 'L1' ? 'repeat(auto-fill, minmax(150px, 1fr))' :
                                level === 'L2' ? 'repeat(auto-fill, minmax(120px, 1fr))' :
                                'repeat(auto-fill, minmax(100px, 1fr))',
            gap: '0.5rem',
            flex: 1,
            overflow: 'auto',
          }}
        >
          {visibleChildren.map((child) => (
            <NestedCapabilityItem
              key={child.capability.id}
              node={child}
              depth={depth}
              onClick={onClick}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function NestedCapabilityGrid({
  capabilities,
  depth,
  onCapabilityClick,
  positions,
}: NestedCapabilityGridProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: 'nested-grid-droppable',
  });

  const tree = useMemo(() => buildTree(capabilities), [capabilities]);

  const sortedTree = useMemo(() => {
    if (positions && Object.keys(positions).length > 0) {
      return [...tree].sort((a, b) => {
        const posA = positions[a.capability.id];
        const posB = positions[b.capability.id];

        if (posA && posB) {
          if (posA.y !== posB.y) return posA.y - posB.y;
          return posA.x - posB.x;
        }
        if (posA) return -1;
        if (posB) return 1;
        return a.capability.name.localeCompare(b.capability.name);
      });
    }
    return tree;
  }, [tree, positions]);

  const sortableIds = useMemo(
    () => sortedTree.map((n) => n.capability.id),
    [sortedTree]
  );

  const hasSortablePositions = positions && Object.keys(positions).length > 0;

  return (
    <div
      ref={setNodeRef}
      className="nested-capability-grid"
      data-dnd-context="true"
      style={{
        minHeight: '200px',
        border: isOver ? '2px dashed #3b82f6' : '2px dashed transparent',
        borderRadius: '0.5rem',
        transition: 'border-color 0.2s',
      }}
    >
      {hasSortablePositions ? (
        <SortableContext items={sortableIds} strategy={rectSortingStrategy}>
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
                sortable
              />
            ))}
          </div>
        </SortableContext>
      ) : (
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
            />
          ))}
        </div>
      )}
    </div>
  );
}
