import { useCallback, useState } from 'react';
import toast from 'react-hot-toast';
import type { Capability, CapabilityId, BusinessDomainId } from '../../../api/types';

export interface PendingReassignment {
  capability: Capability;
  newParent: Capability;
}

interface UseDragHandlersProps {
  domainId: BusinessDomainId | null;
  capabilities: Capability[];
  assignedCapabilityIds: Set<CapabilityId>;
  positions: Record<CapabilityId, { x: number; y: number }>;
  updatePosition: (capabilityId: CapabilityId, x: number, y: number) => Promise<void>;
  associateCapability: (capabilityId: CapabilityId) => Promise<void>;
  refetchCapabilities: () => Promise<void>;
  refetchRealizations?: () => Promise<void>;
}

export function useDragHandlers(props: UseDragHandlersProps) {
  const [activeCapability, setActiveCapability] = useState<Capability | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);

  const handleDragStart = useCallback((capability: Capability) => {
    setActiveCapability(capability);
  }, []);

  const handleDragEnd = useCallback(() => {
    setActiveCapability(null);
    setIsDragOver(false);
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback(() => {
    setIsDragOver(false);
  }, []);

  const handleDrop = useCallback(
    async (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragOver(false);

      const capabilityJson = e.dataTransfer.getData('application/json');
      if (!capabilityJson || !props.domainId) {
        setActiveCapability(null);
        return;
      }

      try {
        const capability = JSON.parse(capabilityJson) as Capability;

        if (capability.level !== 'L1') {
          setActiveCapability(null);
          return;
        }

        if (props.assignedCapabilityIds.has(capability.id)) {
          setActiveCapability(null);
          return;
        }

        await props.associateCapability(capability.id);
        await props.refetchCapabilities();
        await props.refetchRealizations?.();
        const currentCount = props.capabilities.filter((c) => c.level === 'L1').length;
        await props.updatePosition(capability.id, currentCount, 0);
      } catch {
        toast.error('Failed to assign capability');
      } finally {
        setActiveCapability(null);
      }
    },
    [props]
  );

  return {
    activeCapability,
    isDragOver,
    handleDragStart,
    handleDragEnd,
    handleDragOver,
    handleDragLeave,
    handleDrop,
  };
}
