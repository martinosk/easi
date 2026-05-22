import { Badge, Box, Group, Paper, Stack, Text } from '@mantine/core';
import { useMemo } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';
import classes from './CapabilityExplorer.module.css';

interface TreeNode {
  capability: Capability;
  children: TreeNode[];
}

export interface CapabilityExplorerProps {
  capabilities: Capability[];
  assignedCapabilityIds: Set<CapabilityId>;
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

interface DraggableL1ItemProps {
  capability: Capability;
  isAssigned: boolean;
  subItems: TreeNode[];
  onDragStart?: (capability: Capability) => void;
  onDragEnd?: () => void;
}

function DraggableL1Item({ capability, isAssigned, subItems, onDragStart, onDragEnd }: DraggableL1ItemProps) {
  const handleDragStart = (e: React.DragEvent<HTMLDivElement>) => {
    e.dataTransfer.setData('application/json', JSON.stringify(capability));
    e.dataTransfer.effectAllowed = 'move';
    onDragStart?.(capability);
  };

  const handleDragEnd = () => {
    onDragEnd?.();
  };

  return (
    <Paper
      withBorder={false}
      bg="gray.1"
      radius="sm"
      p="xs"
      className={classes.draggable}
      draggable
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      data-testid={`draggable-${capability.id}`}
      data-draggable="true"
    >
      <Group gap="xs" wrap="nowrap">
        <Text size="sm" fw={500}>
          {capability.name}
        </Text>
        {isAssigned && (
          <Badge color="blue" variant="light" size="xs" data-testid={`assigned-indicator-${capability.id}`}>
            Assigned
          </Badge>
        )}
      </Group>
      {subItems.length > 0 && (
        <Stack gap={4} mt="xs" ml="md">
          {subItems.map((child) => (
            <CapabilityTreeItem key={child.capability.id} node={child} />
          ))}
        </Stack>
      )}
    </Paper>
  );
}

interface CapabilityTreeItemProps {
  node: TreeNode;
}

function levelClass(level: Capability['level']): string {
  if (level === 'L2') return classes.levelL2;
  if (level === 'L3') return classes.levelL3;
  if (level === 'L4') return classes.levelL4;
  return classes.levelDefault;
}

function CapabilityTreeItem({ node }: CapabilityTreeItemProps) {
  return (
    <Box bg="white" px="xs" py={4} className={`${classes.levelBar} ${levelClass(node.capability.level)}`}>
      <Text size="sm">{node.capability.name}</Text>
      {node.children.length > 0 && (
        <Stack gap={4} mt={4} ml="sm">
          {node.children.map((child) => (
            <CapabilityTreeItem key={child.capability.id} node={child} />
          ))}
        </Stack>
      )}
    </Box>
  );
}

export function CapabilityExplorer({
  capabilities,
  assignedCapabilityIds,
  isLoading,
  onDragStart,
  onDragEnd,
}: CapabilityExplorerProps) {
  const tree = useMemo(() => buildTree(capabilities), [capabilities]);

  if (isLoading) {
    return (
      <Box p="md">
        <Text c="dimmed">Loading capabilities...</Text>
      </Box>
    );
  }

  if (tree.length === 0) {
    return (
      <Box p="md">
        <Text c="dimmed">No capabilities available</Text>
      </Box>
    );
  }

  return (
    <Stack gap="xs" p="xs">
      {tree.map((node) => (
        <DraggableL1Item
          key={node.capability.id}
          capability={node.capability}
          isAssigned={assignedCapabilityIds.has(node.capability.id)}
          subItems={node.children}
          onDragStart={onDragStart}
          onDragEnd={onDragEnd}
        />
      ))}
    </Stack>
  );
}
