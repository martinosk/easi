import { Badge, Box, Paper, Text } from '@mantine/core';
import { useCallback } from 'react';
import type { Capability } from '../../../api/types';
import { CapabilityTree } from '../../capabilities/components/CapabilityTree';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { useCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import classes from './CapabilitySidebar.module.css';

interface CapabilitySidebarProps {
  mappedCapabilityIds: Set<string>;
  onDragCapability?: (capability: Capability) => void;
}

export function CapabilitySidebar({ mappedCapabilityIds, onDragCapability }: CapabilitySidebarProps) {
  const { tree, isLoading } = useCapabilityTree();

  const getRowProps = useCallback(
    (node: CapabilityTreeNode) => {
      const isMapped = mappedCapabilityIds.has(node.capability.id);
      return {
        draggable: !isMapped,
        dimmed: isMapped,
        testId: `cap-tree-${node.capability.id}`,
        onDragStart: (e: React.DragEvent) => {
          e.dataTransfer.setData('application/json', JSON.stringify(node.capability));
          e.dataTransfer.effectAllowed = 'copy';
          onDragCapability?.(node.capability);
        },
      };
    },
    [mappedCapabilityIds, onDragCapability],
  );

  const renderRight = useCallback(
    (node: CapabilityTreeNode) =>
      mappedCapabilityIds.has(node.capability.id) ? (
        <Badge size="xs" variant="light" color="blue">
          Mapped
        </Badge>
      ) : null,
    [mappedCapabilityIds],
  );

  return (
    <Paper className={classes.sidebar} radius="lg" shadow="sm" data-testid="capability-sidebar">
      <Box className={classes.header}>
        <Text size="sm" fw={700} c="gray.7" tt="uppercase">
          Capabilities
        </Text>
      </Box>
      <CapabilityTree
        tree={tree}
        isLoading={isLoading}
        className={classes.tree}
        searchPlaceholder="Filter capabilities..."
        searchTestId="capability-filter"
        emptyText="No capabilities found"
        noMatchText="No capabilities match your filter"
        getRowProps={getRowProps}
        renderRight={renderRight}
      />
    </Paper>
  );
}
