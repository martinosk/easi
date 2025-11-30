import { useCallback, useState } from 'react';
import { arrayMove } from '@dnd-kit/sortable';
import type { DragEndEvent, DragStartEvent } from '@dnd-kit/core';
import type { Capability, CapabilityId, BusinessDomain } from '../../../api/types';

interface UseDragHandlersProps {
  visualizedDomain: BusinessDomain | null;
  capabilities: Capability[];
  assignedCapabilityIds: Set<CapabilityId>;
  positions: Record<CapabilityId, { x: number; y: number }>;
  updatePosition: (capabilityId: CapabilityId, x: number, y: number) => Promise<void>;
  associateCapability: (capabilityId: CapabilityId, capability: Capability) => Promise<void>;
  refetchCapabilities: () => Promise<void>;
}

export function useDragHandlers({
  visualizedDomain,
  capabilities,
  assignedCapabilityIds,
  positions,
  updatePosition,
  associateCapability,
  refetchCapabilities,
}: UseDragHandlersProps) {
  const [activeCapability, setActiveCapability] = useState<Capability | null>(null);

  const handleDragStart = useCallback((event: DragStartEvent) => {
    const capability = event.active.data.current?.capability as Capability | undefined;
    if (capability) {
      setActiveCapability(capability);
    }
  }, []);

  const handleSortDrag = useCallback(
    async (activeId: string, overId: string) => {
      const l1Caps = capabilities.filter((c) => c.level === 'L1');
      const sortedL1Caps = [...l1Caps].sort((a, b) => {
        const posA = positions[a.id]?.x ?? Infinity;
        const posB = positions[b.id]?.x ?? Infinity;
        return posA - posB;
      });

      const oldIndex = sortedL1Caps.findIndex((c) => c.id === activeId);
      const newIndex = sortedL1Caps.findIndex((c) => c.id === overId);

      if (oldIndex !== -1 && newIndex !== -1) {
        const newOrder = arrayMove(sortedL1Caps, oldIndex, newIndex);
        newOrder.forEach((cap, index) => {
          updatePosition(cap.id, index, 0);
        });
        return true;
      }
      return false;
    },
    [capabilities, positions, updatePosition]
  );

  const handleAssociateDrag = useCallback(
    async (capability: Capability) => {
      if (!visualizedDomain || capability.level !== 'L1') return false;
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
    [visualizedDomain, assignedCapabilityIds, associateCapability, refetchCapabilities, capabilities, updatePosition]
  );

  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      setActiveCapability(null);

      const { active, over } = event;
      if (!over || !visualizedDomain) return;

      const droppedOnId = over.id as string;
      const isDroppedOnGrid = droppedOnId === 'domain-grid-droppable' || droppedOnId === 'nested-grid-droppable';

      if (active.id !== over.id && !isDroppedOnGrid) {
        await handleSortDrag(active.id as string, over.id as string);
        return;
      }

      const capability = active.data.current?.capability as Capability | undefined;
      if (!capability) return;

      await handleAssociateDrag(capability);
    },
    [visualizedDomain, handleSortDrag, handleAssociateDrag]
  );

  return {
    activeCapability,
    handleDragStart,
    handleDragEnd,
  };
}
