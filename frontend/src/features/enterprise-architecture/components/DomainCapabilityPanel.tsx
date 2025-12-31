import { useMemo } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';
import type { CapabilityLinkStatusResponse, CapabilityLinkStatus } from '../types';

interface TreeNode {
  capability: Capability;
  children: TreeNode[];
}

const LEVEL_COLORS: Record<string, string> = {
  L2: '#8b5cf6',
  L3: '#ec4899',
  L4: '#f97316',
};

const STATUS_STYLES = {
  available: { bg: '#ffffff', text: '#374151', opacity: 1 },
  linked: { bg: '#fef3c7', text: '#374151', opacity: 1 },
  blocked: { bg: '#f3f4f6', text: '#9ca3af', opacity: 0.8 },
} as const;

function isBlockedStatus(status: CapabilityLinkStatus): boolean {
  return status === 'blocked_by_parent' || status === 'blocked_by_child';
}

function getStylesForStatus(status: CapabilityLinkStatus) {
  if (isBlockedStatus(status)) return STATUS_STYLES.blocked;
  if (status === 'linked') return STATUS_STYLES.linked;
  return STATUS_STYLES.available;
}

const STATUS_DISPLAY_TEMPLATES = {
  linked: { withName: (name: string) => `──► ${name}`, fallback: '──► Linked' },
  blocked_by_parent: { withName: (name: string) => `Parent linked to ${name}`, fallback: 'Blocked by parent' },
  blocked_by_child: { withName: (name: string) => `Child linked to ${name}`, fallback: 'Blocked by child' },
} as const;

function getStatusDisplayText(status: CapabilityLinkStatus, linkStatus?: CapabilityLinkStatusResponse): string | null {
  if (status === 'available') return null;
  if (status === 'linked') {
    const template = STATUS_DISPLAY_TEMPLATES.linked;
    return linkStatus?.linkedTo ? template.withName(linkStatus.linkedTo.name) : template.fallback;
  }
  const template = STATUS_DISPLAY_TEMPLATES[status];
  return linkStatus?.blockingCapability ? template.withName(linkStatus.blockingCapability.name) : template.fallback;
}

function getBlockedText(status: CapabilityLinkStatus, linkStatus?: CapabilityLinkStatusResponse): string | null {
  if (!isBlockedStatus(status)) return null;
  const name = linkStatus?.blockingCapability?.name;
  if (!name) return '(blocked)';
  const prefix = status === 'blocked_by_parent' ? 'parent' : 'child';
  return `(blocked: ${prefix} "${name}" is linked)`;
}

function getStatusDisplayStyle(status: CapabilityLinkStatus) {
  const isLinked = status === 'linked';
  return {
    fontSize: '0.75rem',
    color: isLinked ? '#92400e' : '#6b7280',
    fontStyle: isLinked ? 'normal' : 'italic',
  } as const;
}

function ChildrenTreeList({
  children,
  linkStatuses
}: {
  children: TreeNode[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>
}) {
  if (children.length === 0) return null;
  return (
    <div style={{ marginLeft: '1rem', marginTop: '0.25rem' }}>
      {children.map((child) => (
        <CapabilityTreeItem
          key={child.capability.id}
          node={child}
          linkStatus={linkStatuses.get(child.capability.id)}
          linkStatuses={linkStatuses}
        />
      ))}
    </div>
  );
}

export interface DomainCapabilityPanelProps {
  capabilities: Capability[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  isLoading: boolean;
  onDragStart?: (capability: Capability) => void;
  onDragEnd?: () => void;
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

interface DraggableCapabilityItemProps {
  capability: Capability;
  linkStatus: CapabilityLinkStatusResponse | undefined;
  children: TreeNode[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  onDragStart?: (capability: Capability) => void;
  onDragEnd?: () => void;
}

function DraggableCapabilityItem({
  capability,
  linkStatus,
  children,
  linkStatuses,
  onDragStart,
  onDragEnd,
}: DraggableCapabilityItemProps) {
  const status = linkStatus?.status || 'available';
  const isDraggable = status === 'available';
  const styles = getStylesForStatus(status);
  const statusDisplay = getStatusDisplayText(status, linkStatus);

  const handleDragStart = (e: React.DragEvent<HTMLDivElement>) => {
    if (!isDraggable) {
      e.preventDefault();
      return;
    }
    e.dataTransfer.setData('application/json', JSON.stringify(capability));
    e.dataTransfer.effectAllowed = 'move';
    onDragStart?.(capability);
  };

  return (
    <div style={{ marginBottom: '0.25rem' }}>
      <div
        draggable={isDraggable}
        onDragStart={handleDragStart}
        onDragEnd={onDragEnd}
        style={{
          padding: '0.5rem',
          backgroundColor: styles.bg,
          borderRadius: '0.25rem',
          cursor: isDraggable ? 'grab' : 'not-allowed',
          userSelect: 'none',
          WebkitUserSelect: 'none',
          border: '1px solid #e5e7eb',
          position: 'relative',
        }}
        title={isBlockedStatus(status) ? statusDisplay || undefined : undefined}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          {isDraggable && (
            <span style={{ color: '#9ca3af', fontSize: '0.875rem' }}>⋮⋮</span>
          )}
          <span style={{ fontWeight: 500, color: styles.text, flex: 1 }}>
            {capability.name}
          </span>
          {statusDisplay && (
            <span style={getStatusDisplayStyle(status)}>{statusDisplay}</span>
          )}
        </div>
      </div>

      <ChildrenTreeList children={children} linkStatuses={linkStatuses} />
    </div>
  );
}

interface CapabilityTreeItemProps {
  node: TreeNode;
  linkStatus: CapabilityLinkStatusResponse | undefined;
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
}

function CapabilityTreeItem({ node, linkStatus, linkStatuses }: CapabilityTreeItemProps) {
  const status = linkStatus?.status || 'available';
  const isBlocked = isBlockedStatus(status);
  const blockedText = getBlockedText(status, linkStatus);
  const borderColor = LEVEL_COLORS[node.capability.level] || '#6b7280';

  return (
    <div
      style={{
        padding: '0.25rem 0.5rem',
        marginBottom: '0.25rem',
        borderLeft: `3px solid ${borderColor}`,
        backgroundColor: isBlocked ? '#fef2f2' : '#ffffff',
        opacity: isBlocked ? 0.8 : 1,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <span
          style={{
            fontSize: '0.875rem',
            color: isBlocked ? '#9ca3af' : '#374151',
          }}
        >
          {node.capability.name}
        </span>
        {blockedText && (
          <span style={{ fontSize: '0.7rem', color: '#ef4444', fontStyle: 'italic' }}>
            {blockedText}
          </span>
        )}
      </div>
      {node.children.length > 0 && (
        <div style={{ marginLeft: '0.75rem', marginTop: '0.25rem' }}>
          {node.children.map((child) => (
            <CapabilityTreeItem
              key={child.capability.id}
              node={child}
              linkStatus={linkStatuses.get(child.capability.id)}
              linkStatuses={linkStatuses}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function DomainCapabilityPanel({
  capabilities,
  linkStatuses,
  isLoading,
  onDragStart,
  onDragEnd,
}: DomainCapabilityPanelProps) {
  const tree = useMemo(() => buildTree(capabilities), [capabilities]);

  if (isLoading) {
    return (
      <div style={{ padding: '1rem' }}>
        <div style={{ color: '#6b7280' }}>Loading domain capabilities...</div>
      </div>
    );
  }

  if (tree.length === 0) {
    return (
      <div style={{ padding: '1rem' }}>
        <div style={{ color: '#6b7280' }}>No domain capabilities available</div>
      </div>
    );
  }

  return (
    <div style={{ padding: '1rem' }}>
      <h2 style={{ fontSize: '1.25rem', fontWeight: 600, marginBottom: '1rem' }}>
        Domain Capabilities
      </h2>
      <p style={{ fontSize: '0.875rem', color: '#6b7280', marginBottom: '1.5rem' }}>
        Drag L1 capabilities to enterprise capabilities
      </p>
      <div>
        {tree.map((node) => (
          <DraggableCapabilityItem
            key={node.capability.id}
            capability={node.capability}
            linkStatus={linkStatuses.get(node.capability.id)}
            children={node.children}
            linkStatuses={linkStatuses}
            onDragStart={onDragStart}
            onDragEnd={onDragEnd}
          />
        ))}
      </div>
    </div>
  );
}
