import { Box, Center, Group, Stack, Text, Title } from '@mantine/core';
import { useMemo } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';
import type { CapabilityLinkStatus, CapabilityLinkStatusResponse } from '../types';
import classes from './DomainCapabilityPanel.module.css';

interface TreeNode {
  capability: Capability;
  children: TreeNode[];
}

const BLOCKED_STATUSES: ReadonlySet<CapabilityLinkStatus> = new Set(['blocked_by_parent', 'blocked_by_child']);

function isBlockedStatus(status: CapabilityLinkStatus): boolean {
  return BLOCKED_STATUSES.has(status);
}

const STATUS_DISPLAY_TEMPLATES = {
  linked: { withName: (name: string) => `──► ${name}`, fallback: '──► Linked' },
  blocked_by_parent: { withName: (name: string) => `Parent linked to ${name}`, fallback: 'Blocked by parent' },
  blocked_by_child: { withName: (name: string) => `Child linked to ${name}`, fallback: 'Blocked by child' },
} as const;

type LabelledStatus = keyof typeof STATUS_DISPLAY_TEMPLATES;

const STATUS_NAME_LOOKUP: Record<LabelledStatus, (s: CapabilityLinkStatusResponse) => string | undefined> = {
  linked: (s) => s.linkedTo?.name,
  blocked_by_parent: (s) => s.blockingCapability?.name,
  blocked_by_child: (s) => s.blockingCapability?.name,
};

function getStatusDisplayText(status: CapabilityLinkStatus, linkStatus?: CapabilityLinkStatusResponse): string | null {
  if (status === 'available') return null;
  const template = STATUS_DISPLAY_TEMPLATES[status];
  const name = linkStatus ? STATUS_NAME_LOOKUP[status](linkStatus) : undefined;
  return name ? template.withName(name) : template.fallback;
}

const BLOCKED_PREFIX: Record<'blocked_by_parent' | 'blocked_by_child', string> = {
  blocked_by_parent: 'parent',
  blocked_by_child: 'child',
};

function getBlockedText(status: CapabilityLinkStatus, linkStatus?: CapabilityLinkStatusResponse): string | null {
  if (status !== 'blocked_by_parent' && status !== 'blocked_by_child') return null;
  const name = linkStatus?.blockingCapability?.name;
  return name ? `(blocked: ${BLOCKED_PREFIX[status]} "${name}" is linked)` : '(blocked)';
}

function attachToTree(
  cap: Capability,
  node: TreeNode,
  nodesById: Map<CapabilityId, TreeNode>,
  roots: TreeNode[],
): void {
  const parent = cap.parentId ? nodesById.get(cap.parentId) : undefined;
  if (parent) parent.children.push(node);
  else if (cap.level === 'L1') roots.push(node);
}

function buildTree(capabilities: Capability[]): TreeNode[] {
  const nodesById = new Map<CapabilityId, TreeNode>();
  for (const cap of capabilities) {
    nodesById.set(cap.id, { capability: cap, children: [] });
  }

  const roots: TreeNode[] = [];
  for (const cap of capabilities) {
    attachToTree(cap, nodesById.get(cap.id)!, nodesById, roots);
  }

  roots.sort((a, b) => a.capability.name.localeCompare(b.capability.name));
  return roots;
}

export interface DomainCapabilityPanelProps {
  capabilities: Capability[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  isLoading: boolean;
  onDragStart?: (capability: Capability) => void;
  onDragEnd?: () => void;
}

function StatusLabel({ status, text }: { status: CapabilityLinkStatus; text: string }) {
  const isLinked = status === 'linked';
  return (
    <Text size="xs" c={isLinked ? 'yellow.8' : 'dimmed'} fs={isLinked ? 'normal' : 'italic'}>
      {text}
    </Text>
  );
}

function buildDragStartHandler(capability: Capability, onDragStart?: (capability: Capability) => void) {
  return (e: React.DragEvent<HTMLDivElement>) => {
    e.dataTransfer.setData('application/json', JSON.stringify(capability));
    e.dataTransfer.effectAllowed = 'move';
    onDragStart?.(capability);
  };
}

interface DraggableCapabilityItemProps {
  capability: Capability;
  linkStatus: CapabilityLinkStatusResponse | undefined;
  childNodes: TreeNode[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  onDragStart?: (capability: Capability) => void;
  onDragEnd?: () => void;
}

function DraggableCapabilityItem({
  capability,
  linkStatus,
  childNodes,
  linkStatuses,
  onDragStart,
  onDragEnd,
}: DraggableCapabilityItemProps) {
  const status = linkStatus?.status || 'available';
  const isDraggable = status === 'available';
  const statusDisplay = getStatusDisplayText(status, linkStatus);
  const handleDragStart = isDraggable ? buildDragStartHandler(capability, onDragStart) : undefined;

  return (
    <Box mb="xs">
      <Box
        draggable={isDraggable}
        onDragStart={handleDragStart}
        onDragEnd={onDragEnd}
        data-status={status}
        className={classes.draggable}
        title={isBlockedStatus(status) ? statusDisplay || undefined : undefined}
      >
        <Group gap="xs" wrap="nowrap">
          {isDraggable && (
            <Text size="sm" c="gray.5">
              ⋮⋮
            </Text>
          )}
          <Text size="sm" fw={500} flex={1} className={classes.label}>
            {capability.name}
          </Text>
          {statusDisplay && <StatusLabel status={status} text={statusDisplay} />}
        </Group>
      </Box>
      <ChildrenTreeList nodes={childNodes} linkStatuses={linkStatuses} />
    </Box>
  );
}

function ChildrenTreeList({
  nodes,
  linkStatuses,
}: {
  nodes: TreeNode[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
}) {
  if (nodes.length === 0) return null;
  return (
    <Box pl="md" mt={4}>
      {nodes.map((child) => (
        <CapabilityTreeItem
          key={child.capability.id}
          node={child}
          linkStatus={linkStatuses.get(child.capability.id)}
          linkStatuses={linkStatuses}
        />
      ))}
    </Box>
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

  return (
    <Box data-level={node.capability.level} data-blocked={isBlocked || undefined} className={classes.treeItem}>
      <Group gap="xs" wrap="nowrap">
        <Text size="sm" c={isBlocked ? 'gray.5' : 'gray.8'}>
          {node.capability.name}
        </Text>
        {blockedText && (
          <Text size="xs" c="red.6" fs="italic">
            {blockedText}
          </Text>
        )}
      </Group>
      <ChildrenTreeList nodes={node.children} linkStatuses={linkStatuses} />
    </Box>
  );
}

function EmptyMessage({ children }: { children: React.ReactNode }) {
  return (
    <Center p="md">
      <Text c="dimmed">{children}</Text>
    </Center>
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

  if (isLoading) return <EmptyMessage>Loading domain capabilities...</EmptyMessage>;
  if (tree.length === 0) return <EmptyMessage>No domain capabilities available</EmptyMessage>;

  return (
    <Stack p="md" gap="sm">
      <Title order={3}>Domain Capabilities</Title>
      <Text size="sm" c="dimmed">
        Drag L1 capabilities to enterprise capabilities
      </Text>
      <Box>
        {tree.map((node) => (
          <DraggableCapabilityItem
            key={node.capability.id}
            capability={node.capability}
            linkStatus={linkStatuses.get(node.capability.id)}
            childNodes={node.children}
            linkStatuses={linkStatuses}
            onDragStart={onDragStart}
            onDragEnd={onDragEnd}
          />
        ))}
      </Box>
    </Stack>
  );
}
