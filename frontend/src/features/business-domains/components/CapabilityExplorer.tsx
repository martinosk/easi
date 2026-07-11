import { Badge } from '@mantine/core';
import { useCallback, useMemo } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';
import { CapabilityTree } from '../../capabilities/components/CapabilityTree';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';

export interface CapabilityExplorerProps {
  capabilities: Capability[];
  assignedCapabilityIds: Set<CapabilityId>;
  isLoading: boolean;
  onDragStart?: (capability: Capability) => void;
  onDragEnd?: () => void;
}

export function CapabilityExplorer({
  capabilities,
  assignedCapabilityIds,
  isLoading,
  onDragStart,
  onDragEnd,
}: CapabilityExplorerProps) {
  const tree = useMemo(() => buildCapabilityTree(capabilities), [capabilities]);

  const getRowProps = useCallback(
    (node: CapabilityTreeNode) => {
      if (node.capability.level !== 'L1') return {};
      return {
        draggable: true,
        testId: `draggable-${node.capability.id}`,
        onDragStart: (e: React.DragEvent) => {
          e.dataTransfer.setData('application/json', JSON.stringify(node.capability));
          e.dataTransfer.effectAllowed = 'move';
          onDragStart?.(node.capability);
        },
        onDragEnd: () => onDragEnd?.(),
      };
    },
    [onDragStart, onDragEnd],
  );

  const renderRight = useCallback(
    (node: CapabilityTreeNode) =>
      assignedCapabilityIds.has(node.capability.id) ? (
        <Badge color="blue" variant="light" size="xs" data-testid={`assigned-indicator-${node.capability.id}`}>
          Assigned
        </Badge>
      ) : null,
    [assignedCapabilityIds],
  );

  return (
    <CapabilityTree
      tree={tree}
      isLoading={isLoading}
      emptyText="No capabilities available"
      getRowProps={getRowProps}
      renderRight={renderRight}
    />
  );
}
