import { useState, useCallback } from 'react';
import { useCapabilityTree } from '../../business-domains/hooks/useCapabilityTree';
import type { CapabilityTreeNode } from '../../business-domains/hooks/useCapabilityTree';
import type { Capability } from '../../../api/types';

interface CapabilitySidebarProps {
  mappedCapabilityIds: Set<string>;
  onDragCapability: (capability: Capability) => void;
}

export function CapabilitySidebar({ mappedCapabilityIds, onDragCapability }: CapabilitySidebarProps) {
  const { tree, isLoading } = useCapabilityTree();
  const [filter, setFilter] = useState('');

  const matchesFilter = useCallback((node: CapabilityTreeNode): boolean => {
    if (!filter) return true;
    const lower = filter.toLowerCase();
    if (node.capability.name.toLowerCase().includes(lower)) return true;
    return node.children.some(child => matchesFilter(child));
  }, [filter]);

  const filteredTree = filter ? tree.filter(node => matchesFilter(node)) : tree;

  if (isLoading) {
    return (
      <div className="cap-sidebar">
        <div className="cap-sidebar-header">
          <h3>Capabilities</h3>
        </div>
        <div className="cap-sidebar-loading">Loading capabilities...</div>
      </div>
    );
  }

  return (
    <div className="cap-sidebar" data-testid="capability-sidebar">
      <div className="cap-sidebar-header">
        <h3>Capabilities</h3>
      </div>
      <input
        type="text"
        className="cap-sidebar-filter"
        placeholder="Filter capabilities..."
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
        data-testid="capability-filter"
      />
      <div className="cap-sidebar-tree">
        {filteredTree.length === 0 ? (
          <div className="cap-sidebar-empty">
            {filter ? 'No capabilities match your filter' : 'No capabilities found'}
          </div>
        ) : (
          filteredTree.map((node) => (
            <TreeNode
              key={node.capability.id}
              node={node}
              mappedCapabilityIds={mappedCapabilityIds}
              onDragCapability={onDragCapability}
              filter={filter}
              matchesFilter={matchesFilter}
            />
          ))
        )}
      </div>
    </div>
  );
}

interface TreeNodeProps {
  node: CapabilityTreeNode;
  mappedCapabilityIds: Set<string>;
  onDragCapability: (capability: Capability) => void;
  filter: string;
  matchesFilter: (node: CapabilityTreeNode) => boolean;
  depth?: number;
}

const LEVEL_COLORS: Record<string, string> = {
  L1: 'var(--color-gray-600)',
  L2: '#8b5cf6',
  L3: '#ec4899',
  L4: '#f97316',
};

function TreeNode({ node, mappedCapabilityIds, onDragCapability, filter, matchesFilter, depth = 0 }: TreeNodeProps) {
  const [expanded, setExpanded] = useState(depth === 0);
  const isMapped = mappedCapabilityIds.has(node.capability.id);
  const hasChildren = node.children.length > 0;
  const visibleChildren = filter ? node.children.filter(child => matchesFilter(child)) : node.children;
  const level = node.capability.level || 'L1';

  const handleDragStart = useCallback((e: React.DragEvent) => {
    e.dataTransfer.setData('application/json', JSON.stringify(node.capability));
    e.dataTransfer.effectAllowed = 'copy';
    onDragCapability(node.capability);
  }, [node.capability, onDragCapability]);

  return (
    <div className="cap-tree-node" style={{ paddingLeft: depth > 0 ? `${depth * 12}px` : undefined }}>
      <div
        className={`cap-tree-item ${isMapped ? 'cap-tree-mapped' : ''}`}
        draggable={!isMapped}
        onDragStart={handleDragStart}
        data-testid={`cap-tree-${node.capability.id}`}
      >
        {hasChildren && (
          <button
            type="button"
            className="cap-tree-toggle"
            onClick={() => setExpanded(!expanded)}
          >
            <svg viewBox="0 0 24 24" fill="none" width="12" height="12" style={{ transform: expanded ? 'rotate(90deg)' : undefined, transition: 'transform 0.15s ease' }}>
              <path d="M9 18l6-6-6-6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
        )}
        <span className="cap-tree-level" style={{ color: LEVEL_COLORS[level] }}>{level}</span>
        <span className="cap-tree-name">{node.capability.name}</span>
        {isMapped && <span className="cap-tree-badge">Mapped</span>}
      </div>
      {expanded && hasChildren && visibleChildren.map((child) => (
        <TreeNode
          key={child.capability.id}
          node={child}
          mappedCapabilityIds={mappedCapabilityIds}
          onDragCapability={onDragCapability}
          filter={filter}
          matchesFilter={matchesFilter}
          depth={depth + 1}
        />
      ))}
    </div>
  );
}
