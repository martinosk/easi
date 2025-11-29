import { useMemo } from 'react';
import { useDraggable } from '@dnd-kit/core';
import type { Capability, CapabilityId } from '../../../api/types';

interface TreeNode {
  capability: Capability;
  children: TreeNode[];
}

export interface CapabilityExplorerProps {
  capabilities: Capability[];
  assignedCapabilityIds: Set<CapabilityId>;
  isLoading: boolean;
}

function buildTree(capabilities: Capability[]): TreeNode[] {
  const map = new Map<CapabilityId, TreeNode>();

  capabilities.forEach((cap) => {
    map.set(cap.id, { capability: cap, children: [] });
  });

  const roots: TreeNode[] = [];

  capabilities.forEach((cap) => {
    const node = map.get(cap.id)!;
    if (cap.parentId && map.has(cap.parentId)) {
      map.get(cap.parentId)!.children.push(node);
    } else if (cap.level === 'L1') {
      roots.push(node);
    }
  });

  roots.sort((a, b) => a.capability.name.localeCompare(b.capability.name));
  return roots;
}

interface DraggableL1ItemProps {
  capability: Capability;
  isAssigned: boolean;
  children: TreeNode[];
}

function DraggableL1Item({ capability, isAssigned, children }: DraggableL1ItemProps) {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: capability.id,
    data: { capability },
  });

  return (
    <div
      ref={setNodeRef}
      data-testid={`draggable-${capability.id}`}
      data-draggable="true"
      {...listeners}
      {...attributes}
      style={{
        padding: '0.5rem',
        marginBottom: '0.25rem',
        backgroundColor: isDragging ? '#e0e7ff' : '#f3f4f6',
        borderRadius: '0.25rem',
        cursor: 'grab',
        opacity: isDragging ? 0.5 : 1,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <span style={{ fontWeight: 500 }}>{capability.name}</span>
        {isAssigned && (
          <span
            data-testid={`assigned-indicator-${capability.id}`}
            style={{
              fontSize: '0.75rem',
              backgroundColor: '#dbeafe',
              color: '#1e40af',
              padding: '0.125rem 0.5rem',
              borderRadius: '9999px',
            }}
          >
            Assigned
          </span>
        )}
      </div>
      {children.length > 0 && (
        <div style={{ marginLeft: '1rem', marginTop: '0.5rem' }}>
          {children.map((child) => (
            <CapabilityTreeItem key={child.capability.id} node={child} />
          ))}
        </div>
      )}
    </div>
  );
}

interface CapabilityTreeItemProps {
  node: TreeNode;
}

function CapabilityTreeItem({ node }: CapabilityTreeItemProps) {
  const levelColors: Record<string, string> = {
    L2: '#8b5cf6',
    L3: '#ec4899',
    L4: '#f97316',
  };

  return (
    <div
      style={{
        padding: '0.25rem 0.5rem',
        marginBottom: '0.25rem',
        borderLeft: `3px solid ${levelColors[node.capability.level] || '#6b7280'}`,
        backgroundColor: '#ffffff',
      }}
    >
      <span style={{ fontSize: '0.875rem', color: '#374151' }}>{node.capability.name}</span>
      {node.children.length > 0 && (
        <div style={{ marginLeft: '0.75rem', marginTop: '0.25rem' }}>
          {node.children.map((child) => (
            <CapabilityTreeItem key={child.capability.id} node={child} />
          ))}
        </div>
      )}
    </div>
  );
}

export function CapabilityExplorer({
  capabilities,
  assignedCapabilityIds,
  isLoading,
}: CapabilityExplorerProps) {
  const tree = useMemo(() => buildTree(capabilities), [capabilities]);

  if (isLoading) {
    return (
      <div style={{ padding: '1rem', color: '#6b7280' }}>Loading capabilities...</div>
    );
  }

  if (tree.length === 0) {
    return (
      <div style={{ padding: '1rem', color: '#6b7280' }}>No capabilities available</div>
    );
  }

  return (
    <div className="capability-explorer" style={{ padding: '0.5rem' }}>
      {tree.map((node) => (
        <DraggableL1Item
          key={node.capability.id}
          capability={node.capability}
          isAssigned={assignedCapabilityIds.has(node.capability.id)}
          children={node.children}
        />
      ))}
    </div>
  );
}
