import { useCallback, useState } from 'react';
import { arrayMove } from '@dnd-kit/sortable';
import type { DragEndEvent, DragStartEvent } from '@dnd-kit/core';
import type { Capability, CapabilityId, BusinessDomainId } from '../../../api/types';

export interface PendingReassignment {
  capability: Capability;
  newParent: Capability;
}

const GRID_DROPPABLE_IDS = ['domain-grid-droppable', 'nested-grid-droppable'];

function isGridDropTarget(id: string): boolean {
  return GRID_DROPPABLE_IDS.includes(id);
}

function findReassignmentPair(
  allCapabilities: Capability[],
  activeId: string,
  targetId: string
): PendingReassignment | null {
  if (activeId === targetId) return null;
  const draggedCap = allCapabilities.find((c) => c.id === activeId);
  const targetCap = allCapabilities.find((c) => c.id === targetId);
  if (!draggedCap || !targetCap) return null;
  return { capability: draggedCap, newParent: targetCap };
}

interface UseDragHandlersProps {
  domainId: BusinessDomainId | null;
  capabilities: Capability[];
  assignedCapabilityIds: Set<CapabilityId>;
  positions: Record<CapabilityId, { x: number; y: number }>;
  updatePosition: (capabilityId: CapabilityId, x: number, y: number) => Promise<void>;
  associateCapability: (capabilityId: CapabilityId, capability: Capability) => Promise<void>;
  refetchCapabilities: () => Promise<void>;
  allCapabilities?: Capability[];
  onReassignment?: (reassignment: PendingReassignment) => void;
}

export function useDragHandlers({
  domainId,
  capabilities,
  assignedCapabilityIds,
  positions,
  updatePosition,
  associateCapability,
  refetchCapabilities,
  allCapabilities,
  onReassignment,
}: UseDragHandlersProps) {
  const [activeCapability, setActiveCapability] = useState<Capability | null>(null);

  const handleDragStart = useCallback((event: DragStartEvent) => {
    const capability = event.active.data.current?.capability as Capability | undefined;
    if (capability) {
      setActiveCapability(capability);
    }
  }, []);

  const handleSortDrag = useCallback(
    (activeId: string, overId: string): boolean => {
      const l1Caps = capabilities.filter((c) => c.level === 'L1');
      const sortedL1Caps = [...l1Caps].sort((a, b) => {
        const posA = positions[a.id]?.x ?? Infinity;
        const posB = positions[b.id]?.x ?? Infinity;
        return posA - posB;
      });

      const oldIndex = sortedL1Caps.findIndex((c) => c.id === activeId);
      const newIndex = sortedL1Caps.findIndex((c) => c.id === overId);

      if (oldIndex === -1 || newIndex === -1) return false;

      const newOrder = arrayMove(sortedL1Caps, oldIndex, newIndex);
      newOrder.forEach((cap, index) => {
        updatePosition(cap.id, index, 0);
      });
      return true;
    },
    [capabilities, positions, updatePosition]
  );

  const handleReassignDrag = useCallback(
    (activeId: string, overId: string): boolean => {
      if (!allCapabilities || !onReassignment) return false;

      const pair = findReassignmentPair(allCapabilities, activeId, overId);
      if (!pair) return false;

      onReassignment(pair);
      return true;
    },
    [allCapabilities, onReassignment]
  );

  const handleAssociateDrag = useCallback(
    async (capability: Capability): Promise<boolean> => {
      if (!domainId || capability.level !== 'L1') return false;
      if (assignedCapabilityIds.has(capability.id)) return false;

      try {
        await associateCapability(capability.id, capability);
        await refetchCapabilities();
        const currentCount = capabilities.filter((c) => c.level === 'L1').length;
        await updatePosition(capability.id, currentCount, 0);
        return true;
      } catch (err) {
        console.error('Failed to assign capability:', err);
        return false;
      }
    },
    [domainId, assignedCapabilityIds, associateCapability, refetchCapabilities, capabilities, updatePosition]
  );

  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      setActiveCapability(null);

      const { active, over } = event;
      if (!over || !domainId) return;

      const activeId = active.id as string;
      const overId = over.id as string;
      const isDifferentTarget = activeId !== overId && !isGridDropTarget(overId);

      if (isDifferentTarget) {
        if (handleSortDrag(activeId, overId)) return;
        handleReassignDrag(activeId, overId);
        return;
      }

      const capability = active.data.current?.capability as Capability | undefined;
      if (!capability) return;

      await handleAssociateDrag(capability);
    },
    [domainId, handleSortDrag, handleReassignDrag, handleAssociateDrag]
  );

  return {
    activeCapability,
    handleDragStart,
    handleDragEnd,
  };
}
